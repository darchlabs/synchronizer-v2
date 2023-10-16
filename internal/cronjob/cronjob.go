package cronjob

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/darchlabs/synchronizer-v2/internal/blockchain"
	customlogger "github.com/darchlabs/synchronizer-v2/internal/custom-logger"
	"github.com/darchlabs/synchronizer-v2/internal/webhooksender"
	"github.com/darchlabs/synchronizer-v2/pkg/event"
	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/darchlabs/synchronizer-v2/pkg/webhook"
	"github.com/ethereum/go-ethereum/ethclient"
)

type idGenerator func() string
type dateGenerator func() time.Time

type EventStorage interface {
	ListEventsByAddress(address string, sort string, limit int64, offset int64) ([]*event.Event, error)
	InsertEventData(data []*event.EventData) error
	UpdateEvent(e *event.Event) error
	GetEvent(address string, eventName string) (*event.Event, error)
}

type SmartContractStorage interface {
	GetSmartContractByAddress(address string) (*smartcontract.SmartContract, error)
	ListAllSmartContracts() ([]*smartcontract.SmartContract, error)
	UpdateStatusAndError(id string, status smartcontract.SmartContractStatus, err error) error
	UpdateSmartContract(sc *smartcontract.SmartContract) (*smartcontract.SmartContract, error)
}

type WebhookSender interface {
	CreateAndSendWebhook(wh *webhook.Webhook) error
}

type CronjobStatus string

const (
	StatusIdle     CronjobStatus = "idle"
	StatusRunning  CronjobStatus = "running"
	StatusStopping CronjobStatus = "stopping"
	StatusStopped  CronjobStatus = "stopped"
	StatusError    CronjobStatus = "error"
)

type cronjob struct {
	ticker *time.Ticker
	quit   chan struct{}
	error  error

	seconds       int64
	clients       *map[string]*ethclient.Client
	storage       EventStorage
	scStorage     SmartContractStorage
	debug         bool
	status        CronjobStatus
	webhookSender WebhookSender

	idGen   idGenerator
	dateGen dateGenerator

	log *customlogger.CustomLogger
}

func New(seconds int64, storage EventStorage, scStorage SmartContractStorage, clients *map[string]*ethclient.Client, debug bool, idGen idGenerator, dateGen dateGenerator, webhookSender *webhooksender.WebhookSender) *cronjob {
	// initialize custom logger
	customLogger, err := customlogger.NewCustomLogger("cyan", os.Stdout)
	if err != nil {
		panic(err)
	}

	return &cronjob{
		seconds: seconds,
		status:  StatusIdle,
		clients: clients,

		storage:       storage,
		scStorage:     scStorage,
		debug:         debug,
		idGen:         idGen,
		dateGen:       dateGen,
		webhookSender: webhookSender,
		log:           customLogger,
	}
}

func (c *cronjob) Start() error {
	if c.status == StatusRunning {
		return errors.New("cronjob - its already running")
	}

	if c.status == StatusStopping {
		return errors.New("cronjob - is stopping now, wait few seconds")
	}

	c.log.Printf("cronjob - running ticker each %d seconds", c.seconds)

	// initialize ticker
	c.ticker = time.NewTicker(time.Duration(time.Duration(c.seconds) * time.Second))
	c.quit = make(chan struct{})
	c.status = StatusRunning
	c.error = nil

	// run gourutine associated to the ticker
	go func() {
		for {
			select {
			case <-c.ticker.C:
				// call job method to run de ticker process
				err := c.job()
				if err != nil {
					c.status = StatusError
					c.error = err

					c.log.Printf("cronjob - error executing job: %v", err.Error())
					return
				}

			case <-c.quit:
				c.ticker.Stop()
				c.status = StatusStopped
				c.ticker = nil
				return
			}
		}
	}()

	return nil
}

func (c *cronjob) Restart() error {
	c.log.Println("cronjob - restarting ticker")

	if c.status == StatusIdle {
		return errors.New("cronjob - isn't ready yet, wait few seconds")
	}

	if c.status == StatusStopping {
		return errors.New("cronjob - is stopping now, wait few seconds")
	}

	if c.status == StatusRunning {
		err := c.Stop()
		if err != nil {
			return err
		}
	}

	// wait c.Seconds for restart, always the neccesary time is < c.Seconds when the cronjob are stopping
	if c.status == StatusStopping {
		time.Sleep(time.Duration(c.seconds) * time.Second)
	}

	err := c.Start()
	if err != nil {
		return err
	}

	return nil
}

func (c *cronjob) Stop() error {
	if c.status == StatusIdle {
		return errors.New("cronjob - isn't ready yet, wait few seconds")
	}

	if c.status == StatusStopping {
		return errors.New("cronjob - is stopping now, wait few seconds")
	}

	if c.status == StatusStopped {
		return errors.New("cronjob - is already stopped")
	}

	c.status = StatusStopping
	c.quit <- struct{}{}

	return nil
}

func (c *cronjob) job() error {
	// get all smartcontracts
	smartContracts, err := c.scStorage.ListAllSmartContracts()
	if err != nil {
		return err
	}

	// define waitgroup for proccessing the events logs
	var wg sync.WaitGroup
	wg.Add(len(smartContracts))

	// iterate over events
	for _, contract := range smartContracts {
		go func(contract *smartcontract.SmartContract) {
			defer wg.Done()

			// get all events from storage
			events, err := c.storage.ListEventsByAddress(contract.Address, "ASC", 100, 0)
			if err != nil {
				updateSmartContractError(contract, err, c.scStorage)
				return
			}

			// generate map of smartcontract events
			eventsNameMap := make(map[string]*event.Event)
			eventsIdMap := make(map[string]*event.Event)
			for _, ev := range events {
				eventsNameMap[ev.Abi.Name] = ev
				eventsIdMap[ev.ID] = ev
			}

			// get client from map or create and save
			client, ok := (*c.clients)[contract.NodeURL]
			if !ok {
				// validate client works
				client, err = ethclient.Dial(contract.NodeURL)
				if err != nil {
					updateSmartContractError(contract, err, c.scStorage)
					return
				}

				// validate client is working correctly
				_, err = client.ChainID(context.Background())
				if err != nil {
					updateSmartContractError(contract, err, c.scStorage)
					return
				}

				// save client in map
				(*c.clients)[contract.NodeURL] = client
			}

			// get current block number
			blockNumber, err := client.BlockNumber(context.Background())
			if err != nil {
				updateSmartContractError(contract, err, c.scStorage)
				return
			}

			c.log.Printf("cronjob - address: %s last_synced_block_number: %d current_block_number: %d", contract.Address[:6]+"..."+contract.Address[len(contract.Address)-5:], contract.LastTxBlockSynced, blockNumber)

			// get abi from events
			abi := make([]*event.Abi, 0)
			for _, ev := range events {
				abi = append(abi, ev.Abi)
			}

			// marshal abi to json
			abiB, err := json.Marshal(abi)
			if err != nil {
				updateSmartContractError(contract, err, c.scStorage)
				return
			}

			// define final block number
			finalBlockNumber := int64(blockNumber)

			// define and read channel with log data in go routine
			logsChannel := make(chan []blockchain.LogData)
			go func() {
				defer func() {
					// update block number in smartcontract
					contract.LastTxBlockSynced = int64(finalBlockNumber)
					_, err := c.scStorage.UpdateSmartContract(contract)
					if err != nil {
						updateSmartContractError(contract, err, c.scStorage)
						return
					}
				}()

				for logs := range logsChannel {
					// parse each log to EventData
					eventDatas := make([]*event.EventData, 0)
					for _, l := range logs {
						// get event from map
						ev, ok := eventsNameMap[l.EventName]
						if !ok {
							c.log.Printf("cronjob - warning event_name=%s not found in eventsNameMap", l.EventName)
							continue
						}

						ed := &event.EventData{}
						err := ed.FromLogData(l, c.idGen(), ev.ID, c.dateGen())
						if err != nil {
							updateEventError(ev, err, c.dateGen(), c.storage)
							return
						}

						eventDatas = append(eventDatas, ed)
					}

					// insert logs data to event
					err := c.storage.InsertEventData(eventDatas)
					if err != nil {
						updateSmartContractError(contract, err, c.scStorage)
						return
					}

					// send webhook for any event data if contract is synced and has webhook
					if contract.Webhook != "" && contract.IsSynced() {
						c.log.Println("cronjob - sending webhooks")
						for _, evData := range eventDatas {
							// get event from map
							ev, ok := eventsIdMap[evData.EventID]
							if !ok {
								c.log.Printf("cronjob - warning event_id=%s not found in eventsIdMap", evData.EventID)
								continue
							}

							// check if sc is synced and has available webhook
							wh, err := evData.ToWebhookEvent(c.idGen(), ev, contract.Webhook, c.dateGen())
							if err != nil {
								updateEventError(ev, err, c.dateGen(), c.storage)
								return
							}

							err = c.webhookSender.CreateAndSendWebhook(wh)
							if err != nil {
								updateEventError(ev, err, c.dateGen(), c.storage)
								return
							}
						}
					}
				}
			}()

			// define context with timeout for getting log proccess
			// TODO(ca): should to use env value
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // 10 minutos de límite
			defer cancel()

			// get event logs from contract
			count, latestBlockNumber, err := blockchain.GetLogs(ctx, blockchain.Config{
				Client:          client,
				ABI:             string(abiB),
				Address:         contract.Address,
				FromBlockNumber: &contract.LastTxBlockSynced,
				ToBlockNumber:   &blockNumber,
				LogsChannel:     logsChannel,
				Logger:          c.debug,
			})
			if err == context.Canceled || err == context.DeadlineExceeded {
				c.log.Println("cronjob - warning context was cancelled/deadline_exceeded")
			} else if err != nil {
				updateSmartContractError(contract, err, c.scStorage)
				return
			}

			// show count log
			if count > 0 {
				c.log.Printf("cronjob - %d new events have been inserted into the database with %d latest block number", count, latestBlockNumber)
			}

			// replace final block number if latestBlockNumber is less than finalBlockNumber
			if latestBlockNumber < finalBlockNumber {
				finalBlockNumber = latestBlockNumber + 1 // add 1 to avoid duplicate logs
			}
		}(contract)
	}

	wg.Wait()

	return nil
}

func (c *cronjob) Halt() {
	c.ticker.Stop()
	c.status = StatusStopped
	c.ticker = nil
}

func (c *cronjob) GetStatus() string {
	return string(c.status)
}

func (c *cronjob) GetSeconds() int64 {
	return c.seconds
}

func (c *cronjob) GetError() string {
	if c.error != nil {
		return c.error.Error()
	}

	return ""
}

func updateEventError(ev *event.Event, err error, date time.Time, storage EventStorage) {
	// update event error
	ev.Status = event.StatusError
	ev.Error = err.Error()
	ev.UpdatedAt = date
	_ = storage.UpdateEvent(ev)
}

func updateSmartContractError(sc *smartcontract.SmartContract, err error, storage SmartContractStorage) {
	// update sc error
	err = storage.UpdateStatusAndError(sc.ID, smartcontract.StatusError, err)
	if err != nil {
		fmt.Println("🔥 cronjob - error updating smartcontract", err.Error())
	}
}
