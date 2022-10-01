package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	CronjobAPI "github.com/darchlabs/synchronizer-v2/internal/api/cronjob"
	EventAPI "github.com/darchlabs/synchronizer-v2/internal/api/event"
	"github.com/darchlabs/synchronizer-v2/internal/blockchain"
	"github.com/darchlabs/synchronizer-v2/internal/cronjob"
	"github.com/darchlabs/synchronizer-v2/internal/event"
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	EventStorage "github.com/darchlabs/synchronizer-v2/internal/storage/event"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type Stopper interface {
	ListEventsByAddress(address string) ([]*event.Event, error)
	ListEvents() ([]*event.Event, error)
	GetEvent(address string, eventName string) (*event.Event, error)
	InsertEvent(e *event.Event) error
	UpdateEvent(e *event.Event) error
	DeleteEvent(address string, eventName string) error
	DeleteEventData(address string, eventName string) error
	ListEventData(address string, eventName string) ([]interface{}, error)
	InsertEventData(e *event.Event, data []blockchain.LogData) (int64, error)
	Stop() error
}

type Cronjob interface {
	Stop() error
	Restart() error
	Start() error
}


var (
	eventStorage Stopper
	cronjobSvc Cronjob
)

func main() {
	var err error
	
	// get NODE_URL environment value
	nodeUrl := os.Getenv("NODE_URL")
	if nodeUrl == "" {
		log.Fatal("invalid NODE_URL environment value")
	}

	// get INTERVAL_SECONDS environment value
	intervalSeconds := os.Getenv("INTERVAL_SECONDS")
	if intervalSeconds == "" {
		log.Fatal("invalid INTERVAL_SECONDS environment value")
	}

	// get DATABASE_FILEPATH environment value
	databaseFilepath := os.Getenv("DATABASE_FILEPATH")
	if databaseFilepath == "" {
		log.Fatal("invalid DATABASE_FILEPATH environment value")
	}

	// initialize storage
	if err != nil {
		log.Fatal(err)
	}

	s, err := storage.New(databaseFilepath)
	if err != nil {
		log.Fatal(err)
	}

	// initialize event storage
	eventStorage = EventStorage.New(s)

	// parse seconds from string to int64
	seconds, err := strconv.ParseInt(intervalSeconds, 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	// initialize eth client
	client, err := ethclient.Dial(nodeUrl)
	if err != nil {
		log.Fatal(err)
	}

	// initialize the cronjob
	cronjobSvc = cronjob.New(seconds, eventStorage, client)

	// initialize fiber
	api := fiber.New()
	api.Use(logger.New())
	api.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))

	// configure routers
	EventAPI.Route(api, EventAPI.Context{Storage: eventStorage})
	CronjobAPI.Route(api, CronjobAPI.Context{
		Cronjob: cronjobSvc,
	})

	// run process
	err = cronjobSvc.Start()
	if err != nil {
		log.Fatal(err)
	}
	go func () {
		api.Listen(":3000")
	}()

	// listen interrupt
	quit := make(chan struct{})
	listenInterrupt(quit)
	<-quit
	gracefullShutdown()
}

// listenInterrupt method used to listen SIGTERM OS Signal
func listenInterrupt(quit chan struct{}) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-c
		fmt.Println("Signal received", s.String())
		quit <- struct{}{}
	}()
}

// gracefullShutdown method used to close all synchronizer processes
func gracefullShutdown() {
	log.Println("Gracefully shutdown")

	// stop cronjob ticker
	err := cronjobSvc.Stop()
	if err != nil {
		log.Println(err)
	}

	// close databanse connection
	err = eventStorage.Stop()
	if err != nil {
		log.Println(err)
	}
}