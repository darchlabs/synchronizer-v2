package cronjob

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/darchlabs/synchronizer-v2/internal/blockchain"
	"github.com/darchlabs/synchronizer-v2/internal/event"
	"github.com/ethereum/go-ethereum/ethclient"
)

type cronjob struct {
	ticker *time.Ticker
	quit chan struct{}
	isRunning bool
	seconds int64
	
	storage Storage
	client *ethclient.Client
}

type Storage interface {
	ListEvents() ([]*event.Event, error)
	InsertEventData(e *event.Event, data []blockchain.LogData) (int64, error)
	UpdateEvent(e *event.Event) error
}

func NewCronJob(seconds int64, storage Storage, client *ethclient.Client) *cronjob {
	return &cronjob{
		isRunning: false,
		seconds: seconds,
		
		storage: storage,
		client: client,
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

	// run gourutine associated to the ticker
	go func() {
    for {
       select {
        case <- c.ticker.C:
					// call job method to run de ticker process
					err := c.job()
					if err != nil {
						log.Fatal(err)
					}
        case <- c.quit:
            c.ticker.Stop()
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

		// get event logs from contract
		data, latestBlockNumber, err := blockchain.GetLogs(blockchain.Config{
			Client: c.client,
			ABI: fmt.Sprintf("[%s]", string(b)),
			EventName: event.Abi.Name,
			Address: event.Address,
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
		
		// show logger when counter is greather than 0
		if count > 0 {
			log.Printf("%d new events have been inserted into the database with %d latest block number \n", count, latestBlockNumber)
		}

		// update latest block number in event
		err = event.UpdateLatestBlock(latestBlockNumber, c.storage)
		if err != nil {
			return err
		}
	}

	return nil
}