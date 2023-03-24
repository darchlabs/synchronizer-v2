package cronjob

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/darchlabs/synchronizer-v2/internal/blockchain"
	"github.com/darchlabs/synchronizer-v2/pkg/event"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EventDataStorage interface {
	ListEvents() ([]*event.Event, error)
	InsertEventData(e *event.Event, data []blockchain.LogData) (int64, error)
	UpdateEvent(e *event.Event) error
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

	seconds int64
	clients *map[string]*ethclient.Client
	storage EventDataStorage
	debug   bool
	status  CronjobStatus
}

func New(seconds int64, storage EventDataStorage, clients *map[string]*ethclient.Client, debug bool) *cronjob {
	return &cronjob{
		seconds: seconds,
		status:  StatusIdle,
		clients: clients,

		storage: storage,
		debug:   debug,
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

func (c *cronjob) job() error {
	// get all events from storage
	events, err := c.storage.ListEvents()
	if err != nil {
		return err
	}

	// filter by only for running events
	runningEvents := make([]*event.Event, 0)
	for _, e := range events {
		if e.Status == event.StatusRunning {
			runningEvents = append(runningEvents, e)
		}
	}

	// define waitgroup for proccessing the events logs
	var wg sync.WaitGroup
	wg.Add(len(runningEvents))

	// iterate over events
	for _, e := range runningEvents {
		go func(ev *event.Event) {
			// parse abi to string
			b, err := json.Marshal(ev.Abi)
			if err != nil {
				// update event error
				_ = ev.UpdateStatus(event.StatusError, err, c.storage)
				return
			}

			// get client from map or create and save
			client, ok := (*c.clients)[ev.NodeURL]
			if !ok {
				// validate client works
				client, err = ethclient.Dial(ev.NodeURL)
				if err != nil {
					// update event error
					_ = ev.UpdateStatus(event.StatusError, err, c.storage)
					return
				}

				// validate client is working correctly
				_, err = client.ChainID(context.Background())
				if err != nil {
					// update event error
					_ = ev.UpdateStatus(event.StatusError, err, c.storage)
					return
				}

				// save client in map
				(*c.clients)[ev.NodeURL] = client
			}

			// define and read channel with log data in go routine
			logsChannel := make(chan []blockchain.LogData)
			go func() {
				for logs := range logsChannel {
					// insert logs data to event
					_, err := ev.InsertData(logs, c.storage)
					if err != nil {
						// update event error
						_ = ev.UpdateStatus(event.StatusError, err, c.storage)
						return
					}
				}
			}()

			// get event logs from contract
			count, latestBlockNumber, err := blockchain.GetLogs(blockchain.Config{
				Client:          client,
				ABI:             fmt.Sprintf("[%s]", string(b)),
				EventName:       ev.Abi.Name,
				Address:         ev.Address,
				FromBlockNumber: &ev.LatestBlockNumber,
				LogsChannel:     logsChannel,
				Logger:          c.debug,
			})
			if err != nil {
				// update event error
				_ = ev.UpdateStatus(event.StatusError, err, c.storage)
				return
			}

			// show count log
			if count > 0 {
				log.Printf("%d new events have been inserted into the database with %d latest block number \n", count, latestBlockNumber)
			}

			// update latest block number
			err = ev.UpdateLatestBlock(latestBlockNumber, c.storage)
			if err != nil {
				// update event error
				_ = ev.UpdateStatus(event.StatusError, err, c.storage)
				return
			}

			defer wg.Done()
			return
		}(e)
	}

	wg.Wait()

	return nil
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
