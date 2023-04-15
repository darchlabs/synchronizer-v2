package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/darchlabs/synchronizer-v2"
	"github.com/darchlabs/synchronizer-v2/internal/cronjob"
	"github.com/darchlabs/synchronizer-v2/internal/env"
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	eventstorage "github.com/darchlabs/synchronizer-v2/internal/storage/event"
	smartcontractstorage "github.com/darchlabs/synchronizer-v2/internal/storage/smartcontract"
	transactionstorage "github.com/darchlabs/synchronizer-v2/internal/storage/transaction"
	txsengine "github.com/darchlabs/synchronizer-v2/internal/txs-engine"
	CronjobAPI "github.com/darchlabs/synchronizer-v2/pkg/api/cronjob"
	EventAPI "github.com/darchlabs/synchronizer-v2/pkg/api/event"
	smartcontractsAPI "github.com/darchlabs/synchronizer-v2/pkg/api/smartcontracts"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	"github.com/pressly/goose/v3"

	_ "github.com/darchlabs/synchronizer-v2/migrations"
)

var (
	eventStorage        synchronizer.EventStorage
	smartContactStorage synchronizer.SmartContractStorage
	cronjobSvc          synchronizer.Cronjob
	txsEngine           txsengine.TxsEngine
)

func main() {
	// load env values
	var env env.Env
	err := envconfig.Process("", &env)
	if err != nil {
		log.Fatal("invalid env values, error: ", err)
	}

	// initialize storage
	s, err := storage.New(env.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}

	// run migrations
	err = goose.Up(s.DB.DB, env.MigrationDir)
	if err != nil {
		log.Fatal(err)
	}

	// initialize storages
	eventStorage = eventstorage.New(s)
	smartContactStorage = smartcontractstorage.New(s)
	transactionStorage := transactionstorage.New(s)

	// parse seconds from string to int64
	seconds, err := strconv.ParseInt(env.IntervalSeconds, 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	// initialize fiber
	api := fiber.New()
	api.Use(logger.New())
	api.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))

	// create clients map
	clients := make(map[string]*ethclient.Client)

	// initialize the cronjob
	cronjobSvc = cronjob.New(seconds, eventStorage, &clients, env.Debug, uuid.NewString, time.Now)

	// Initialize the transactions engine
	txsEngine := txsengine.New(smartContactStorage, transactionStorage, uuid.NewString, time.Now)

	// configure routers
	smartcontractsAPI.Route(api, smartcontractsAPI.Context{
		Storage:      smartContactStorage,
		EventStorage: eventStorage,
		IDGen:        uuid.NewString,
		DateGen:      time.Now,
		Env:          env,
	})
	EventAPI.Route(api, EventAPI.Context{
		Storage: eventStorage,
		Cronjob: cronjobSvc,
		Clients: &clients,
		IDGen:   uuid.NewString,
		DateGen: time.Now,
	})
	CronjobAPI.Route(api, CronjobAPI.Context{
		Cronjob: cronjobSvc,
	})

	// run process
	err = cronjobSvc.Start()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		api.Listen(fmt.Sprintf(":%s", env.Port))
	}()

	// TODO(nb): This should be inside a txs engine function
	go func() {
		for {
			txsEngine.Run()
			fmt.Println("---- sleeping ---")
			time.Sleep(time.Duration(seconds) * time.Second)
			fmt.Println("---- sleept ---")
		}
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
		log.Println("Signal received", s.String())
		quit <- struct{}{}
	}()
}

// gracefullShutdown method used to close all synchronizer processes
func gracefullShutdown() {
	log.Println("Gracefully shutdown")

	// stop cronjob ticker
	cronjobSvc.Halt()

	// stop txs engine
	txsEngine.Halt()

	// close databanse connection
	err := eventStorage.Stop()
	if err != nil {
		log.Println(err)
	}
}
