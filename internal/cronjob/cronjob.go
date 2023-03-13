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
	ticker    *time.Ticker
	quit      chan struct{}
	isRunning bool
	seconds   int64
	Status    CronjobStatus `json:"status"`
	clients   *map[string]*ethclient.Client

	storage EventDataStorage
}

func New(seconds int64, storage EventDataStorage, clients *map[string]*ethclient.Client) *cronjob {
	return &cronjob{
		isRunning: false,
		seconds:   seconds,
		Status:    StatusIdle,

		storage: storage,
		clients: clients,
	}
}

func (c *cronjob) Start() error {
	if c.isRunning {
		return errors.New("cronjob its already running")
	}

	log.Printf("Running ticker each %d seconds \n", c.seconds)

	// initialize ticker
	c.ticker = time.NewTicker(time.Duration(time.Duration(c.seconds) * time.Second))
	c.quit = make(chan struct{})
	c.Status = StatusRunning

	// TODO(ca): should manage status by EACH synchronizer

	// run gourutine associated to the ticker
	go func() {
		for {
			select {
			case <-c.ticker.C:
				// call job method to run de ticker process
				c.Status = StatusRunning
				err := c.job()
				if err != nil {
					c.Status = StatusError
					log.Printf("WARNING: %s", err.Error())
					return
				}

			case <-c.quit:
				c.ticker.Stop()
				c.Status = StatusStopped
				return
			}
		}
	}()

	return nil
}

func (c *cronjob) Restart() error {
	log.Println("Restarting ticker")

	err := c.Stop()
	if err != nil {
		return err
	}

	err = c.Start()
	if err != nil {
		return err
	}

	return nil
}

func (c *cronjob) Stop() error {
	if !c.isRunning {
		return errors.New("cronjob its already stopped")
	}

	log.Println("Stoping ticker")

	c.Status = StatusStopping
	c.quit <- struct{}{}
	c.ticker = nil

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
		// parse abi to string
		b, err := json.Marshal(event.Abi)
		if err != nil {
			return err
		}

		// get client from map or create and save
		client, ok := (*c.clients)[event.NodeURL]
		if !ok {
			// Validate client works
			log.Println("geetting client ...")
			client, err = ethclient.Dial(event.NodeURL)
			if err != nil {
				return fmt.Errorf("can't getting ethclient error=%s", err)
			}

			// Validate client is working correctly
			_, err = client.ChainID(context.Background())
			if err != nil {
				return fmt.Errorf("can't valid ethclient error=%s", err)
			}

			// TODO: Validate it matches the given body network

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
			return err
		}

		// insert data to event
		count, err := event.InsertData(data, c.storage)
		if err != nil {
			return err
		}

		// finish when the contract dont have new events
		if count == 0 {
			continue
		}

		log.Printf("%d new events have been inserted into the database with %d latest block number \n", count, latestBlockNumber)

		// update latest block number in event
		err = event.UpdateLatestBlock(latestBlockNumber, c.storage)
		if err != nil {
			return err
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
