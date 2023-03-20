package cronjob

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
	ticker  *time.Ticker
	quit    chan struct{}
	seconds int64
	Status  CronjobStatus `json:"status"`
	clients *map[string]*ethclient.Client
	error   error

	storage EventDataStorage
}

func New(seconds int64, storage EventDataStorage, clients *map[string]*ethclient.Client) *cronjob {
	return &cronjob{
		seconds: seconds,
		Status:  StatusIdle,

		storage: storage,
		clients: clients,
	}
}

func (c *cronjob) Start() error {
	if c.Status == StatusRunning {
		return errors.New("cronjob its already running")
	}

	if c.Status == StatusStopping {
		return errors.New("cronjob is stopping now, wait few seconds")
	}

	log.Printf("Running ticker each %d seconds \n", c.seconds)

	// initialize ticker
	c.ticker = time.NewTicker(time.Duration(time.Duration(c.seconds) * time.Second))
	c.quit = make(chan struct{})
	c.Status = StatusRunning
	c.error = nil

	// run gourutine associated to the ticker
	go func() {
		for {
			select {
			case <-c.ticker.C:
				// call job method to run de ticker process
				err := c.job()
				if err != nil {
					c.Status = StatusError
					c.error = err

					log.Printf("Cronjob has error: %s", err.Error())
					return
				}

			case <-c.quit:
				c.ticker.Stop()
				c.Status = StatusStopped
				c.ticker = nil
				return
			}
		}
	}()

	return nil
}

func (c *cronjob) Restart() error {
	log.Println("Restarting ticker")

	if c.Status == StatusIdle {
		return errors.New("cronjob isn't ready yet, wait few seconds")
	}

	if c.Status == StatusStopping {
		return errors.New("cronjob is stopping now, wait few seconds")
	}

	if c.Status == StatusRunning {
		err := c.Stop()
		if err != nil {
			return err
		}
	}

	// wait c.Seconds for restart, always the neccesary time is < c.Seconds when the cronjob are stopping
	if c.Status == StatusStopping {
		time.Sleep(time.Duration(c.seconds) * time.Second)
	}

	err := c.Start()
	if err != nil {
		return err
	}

	return nil
}

func (c *cronjob) Stop() error {
	if c.Status == StatusIdle {
		return errors.New("cronjob isn't ready yet, wait few seconds")
	}

	if c.Status == StatusStopping {
		return errors.New("cronjob is stopping now, wait few seconds")
	}

	if c.Status == StatusStopped {
		return errors.New("cronjob is already stopped")
	}

	c.Status = StatusStopping
	c.quit <- struct{}{}

	return nil
}

func (c *cronjob) job() error {
	// get all events from storage
	events, err := c.storage.ListEvents()
	if err != nil {
		return err
	}

	// iterate over events
	for _, event := range events {
		// if event has error, continue
		if event.Error != "" {
			continue
		}

		// parse abi to string
		b, err := json.Marshal(event.Abi)
		if err != nil {
			// update event error
			_ = event.UpdateError(err, c.storage)
			continue
		}

		// get client from map or create and save
		client, ok := (*c.clients)[event.NodeURL]
		if !ok {
			// Validate client works
			client, err = ethclient.Dial(event.NodeURL)
			if err != nil {
				// update event error
				_ = event.UpdateError(err, c.storage)
				continue
			}

			// Validate client is working correctly
			_, err = client.ChainID(context.Background())
			if err != nil {
				// update event error
				_ = event.UpdateError(err, c.storage)
				continue
			}

			// TODO(nb): validate it matches the given body network

			// Save client in map
			(*c.clients)[event.NodeURL] = client
		}

		// get event logs from contract
		data, latestBlockNumber, err := blockchain.GetLogs(blockchain.Config{
			Client:          client,
			ABI:             fmt.Sprintf("[%s]", string(b)),
			EventName:       event.Abi.Name,
			Address:         event.Address,
			FromBlockNumber: &event.LatestBlockNumber,
		})
		if err != nil {
			// update event error
			_ = event.UpdateError(err, c.storage)
			continue
		}

		// insert data to event
		count, err := event.InsertData(data, c.storage)
		if err != nil {
			// update event error
			_ = event.UpdateError(err, c.storage)
			continue
		}

		// finish when the contract dont have new events
		if count == 0 {
			continue
		}

		log.Printf("%d new events have been inserted into the database with %d latest block number \n", count, latestBlockNumber)

		// update latest block number in event
		err = event.UpdateLatestBlock(latestBlockNumber, c.storage)
		if err != nil {
			// update event error
			_ = event.UpdateError(err, c.storage)
			continue
		}
	}

	return nil
}

func (c *cronjob) GetStatus() string {
	return string(c.Status)
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
