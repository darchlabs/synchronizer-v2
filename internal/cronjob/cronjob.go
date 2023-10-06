package cronjob

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/darchlabs/synchronizer-v2/internal/blockchain"
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	syncng "github.com/darchlabs/synchronizer-v2/internal/sync"
	"github.com/darchlabs/synchronizer-v2/internal/sync/query"
	"github.com/darchlabs/synchronizer-v2/internal/webhooksender"
	"github.com/darchlabs/synchronizer-v2/internal/wrapper"
	"github.com/darchlabs/synchronizer-v2/pkg/event"
	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/darchlabs/synchronizer-v2/pkg/webhook"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type idGenerator func() string
type dateGenerator func() time.Time

type EventDataStorage interface {
	ListAllEvents() ([]*event.Event, error)
	InsertEventData(e *event.Event, data []*event.EventData) error
	UpdateEvent(e *event.Event) error
}

type SmartContractStorage interface {
	GetSmartContractByAddress(address string) (*smartcontract.SmartContract, error)
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
	clients       *sync.Map
	storage       EventDataStorage
	scStorage     SmartContractStorage
	debug         bool
	status        CronjobStatus
	webhookSender WebhookSender

	// sync engine
	syncEngine *syncng.Engine

	idGen   wrapper.IDGenerator
	dateGen wrapper.DateGenerator
}

type Config struct {
	Seconds          int64
	EventDataStorage EventDataStorage
	SCStorage        SmartContractStorage
	Clients          *sync.Map
	Debug            bool
	IDGen            wrapper.IDGenerator
	DateGen          wrapper.DateGenerator
	WebhookSender    *webhooksender.WebhookSender
	Engine           *syncng.Engine
}

func New(config *Config) *cronjob {
	return &cronjob{
		seconds: config.Seconds,
		status:  StatusIdle,
		clients: config.Clients,

		storage:       config.EventDataStorage,
		scStorage:     config.SCStorage,
		debug:         config.Debug,
		idGen:         config.IDGen,
		dateGen:       config.DateGen,
		webhookSender: config.WebhookSender,
		syncEngine:    config.Engine,
	}
}

func (c *cronjob) Start() error {
	if c.status == StatusRunning {
		return errors.New("cronjob its already running")
	}

	if c.status == StatusStopping {
		return errors.New("cronjob is stopping now, wait few seconds")
	}

	log.Printf("Running ticker each %d seconds \n", c.seconds)

	// initialize ticker
	c.ticker = time.NewTicker(time.Duration(time.Duration(c.seconds) * time.Second))
	c.quit = make(chan struct{})
	c.status = StatusRunning
	c.error = nil

	// run gourutine associated to the ticker
	go func() {
		for {
			log.Printf("===== \n")
			log.Printf("Here inside for ticker \n")

			select {
			case <-c.ticker.C:
				// call job method to run de ticker process
				err := c.job()
				if err != nil {
					c.status = StatusError
					c.error = err

					log.Printf("Cronjob has error: %s", err.Error())
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
	log.Println("Restarting ticker")

	if c.status == StatusIdle {
		return errors.New("cronjob isn't ready yet, wait few seconds")
	}

	if c.status == StatusStopping {
		return errors.New("cronjob is stopping now, wait few seconds")
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
		return errors.New("cronjob isn't ready yet, wait few seconds")
	}

	if c.status == StatusStopping {
		return errors.New("cronjob is stopping now, wait few seconds")
	}

	if c.status == StatusStopped {
		return errors.New("cronjob is already stopped")
	}

	c.status = StatusStopping
	c.quit <- struct{}{}

	return nil
}

func (c *cronjob) job() (err error) {
	output, err := c.syncEngine.SelectEventsAndABI(&syncng.SelectEventsAndABIInput{
		EventStatus: string(storage.EventStatusRunning),
	})
	if err != nil {
		return errors.Wrap(err, "cronjob: cronjob.job c.syncEngine.SelectEventsAndABI error")
	}

	// define waitgroup for proccessing the events logs
	var wg sync.WaitGroup
	wg.Add(len(output.Events))

	// iterate over events
	for _, e := range output.Events {
		go func(ev *storage.EventRecord) {
			now := c.dateGen()
			var err error
			defer wg.Done()
			defer func() {
				if err != nil {
					c.updateEventError(ev.ID, err, now)
				}
			}()

			log.Printf("conjob.job getting client")
			// get client from map or create and save
			cl, ok := c.clients.Load(ev.NodeURL)
			var client *ethclient.Client
			//client, ok := (*c.clients)[ev.NodeURL]
			if !ok {
				log.Printf("conjob.job client not found")
				// validate client works
				client, err = ethclient.Dial(ev.NodeURL)
				if err != nil {
					return
				}

				// save client in map
				c.clients.Store(ev.NodeURL, client)
				log.Printf("conjob.job client created")
			} else {
				client = cl.(*ethclient.Client)
			}

			// parse abi to string
			b, err := ev.ABI.MarshalJson()
			if err != nil {
				return
			}

			// define and read channel with log data in go routine
			logsChannel := make(chan []blockchain.LogData)
			go func(e *storage.EventRecord) {
				now := c.dateGen()
				defer func() {
					if err != nil {
						c.updateEventError(e.ID, err, now)
					}
				}()

				for logs := range logsChannel {
					fmt.Println("this are the logs -> ", logs)
					c.syncEngine.InTransaction(func(txx *sqlx.Tx) error {

						// parse each log to EventData
						eventDatas := make([]*storage.EventDataRecord, 0)
						for _, log := range logs {
							ed := &storage.EventDataRecord{}
							err := ed.FromLogData(&log, c.idGen(), e.ID, now)
							if err != nil {
								return err
							}

							eventDatas = append(eventDatas, ed)
						}

						// insert logs data to event
						err := c.syncEngine.EventDataQuerier.InsertEventDataBatchQuery(txx, eventDatas)
						if err != nil {
							return err
						}

						// update latest block number using last dataLog
						var logBlockNumber int64
						if len(logs) > 0 {
							logBlockNumber = int64(logs[len(logs)-1].BlockNumber)

							// update only when log_block_number is greater than event block number
							if logBlockNumber > e.LatestBlockNumber {

								_, err = c.syncEngine.EventQuerier.UpdateEventQuery(txx, &query.UpdateEventQueryInput{
									ID:                &e.ID,
									LatestBlockNumber: &logBlockNumber,
									UpdatedAt:         &now,
								})
								if err != nil {
									return err
								}
							}
						}

						// webhook related
						fmt.Println("ARRAY SCU ", e.SmartContractUsers)
						for idx, scu := range e.SmartContractUsers {
							if scu.WebhookURL != "" && logBlockNumber >= e.SmartContract.InitialBlockNumber {
								fmt.Printf("LEN %d ITERATION %d\n", len(e.SmartContractUsers), idx)
								fmt.Printf("============================= \n")
								fmt.Printf("Smart Contract Users--%+v\n", scu.UserID)
								fmt.Printf("============================= \n")

								for _, evData := range eventDatas {
									fmt.Printf("............................. \n")
									fmt.Printf("Event Data--%+v\n", evData)
									fmt.Printf("............................. \n")

									wh, err := evData.ToWebhookEvent(c.idGen(), e, scu.WebhookURL, now)
									if err != nil {
										return err
									}

									fmt.Println("Initializing webhook activity... for ", scu.UserID, scu.SmartContractAddress)
									err = c.webhookSender.CreateAndSendWebhook(&webhook.Webhook{
										ID:          wh.ID,
										Tx:          wh.Tx,
										UserID:      scu.UserID,
										EntityType:  webhook.WebhookEntityType(wh.EntityType),
										EntityID:    wh.EntityID,
										Endpoint:    wh.Endpoint,
										Payload:     wh.Payload,
										MaxAttempts: wh.MaxAttempts,
										CreatedAt:   wh.CreatedAt,
										UpdatedAt:   wh.UpdatedAt,
										SentAt:      wh.SentAt,
										Attempts:    wh.Attempts,
										NextRetryAt: wh.NextRetryAt,
										Status:      webhook.WebhookStatus(wh.Status),
									})
									if err != nil {
										return err
									}
								}
							}
						}

						return nil
					})

				}
			}(ev)

			// define context with timeout for getting log proccess
			// TODO(ca): should to use env value
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // 10 minutos de límite
			defer cancel()

			fmt.Println("==========> from block number ", ev.LatestBlockNumber)

			// get event logs from contract
			cf := blockchain.Config{
				Client:          client,
				ABI:             fmt.Sprintf("[%s]", string(b)),
				EventName:       ev.Name,
				Address:         ev.Address,
				FromBlockNumber: &ev.LatestBlockNumber,
				LogsChannel:     logsChannel,
				Logger:          c.debug,
			}
			count, latestBlockNumber, err := blockchain.GetLogs(ctx, cf)
			if err == context.Canceled || err == context.DeadlineExceeded {
			} else if err != nil {
				c.updateEventError(ev.ID, err, now)
				return
			}

			// show count log
			if count > 0 {
				log.Printf("%d new events have been inserted into the database with %d latest block number \n", count, latestBlockNumber)
			}

			// update latest block number
			_, err = c.syncEngine.EventQuerier.UpdateEventQuery(c.syncEngine.GetDatabase(), &query.UpdateEventQueryInput{
				ID:                &ev.ID,
				LatestBlockNumber: &latestBlockNumber,
				UpdatedAt:         &now,
			}) // TODO: update this
			if err != nil {
				c.updateEventError(ev.ID, err, now)
				return
			}

			return
		}(e)
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

func (c *cronjob) updateEventError(id string, err error, date time.Time) {

	ev := &syncng.UpdateEventInput{
		ID:        &id,
		UpdatedAt: date,
	}
	if err != nil {
		errString := err.Error()
		ev.Error = &errString
	}

	// update event error
	_, err = c.syncEngine.UpdateEvent(ev)
	if err != nil {
		fmt.Printf("ERROR UPDATING EVENT FAILED\nevent.ID = %s | error = %s \n", id, err.Error())
	}
}
