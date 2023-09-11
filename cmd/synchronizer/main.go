package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/darchlabs/synchronizer-v2"
	"github.com/darchlabs/synchronizer-v2/internal/cronjob"
	"github.com/darchlabs/synchronizer-v2/internal/env"
	"github.com/darchlabs/synchronizer-v2/internal/httpclient"
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	eventstorage "github.com/darchlabs/synchronizer-v2/internal/storage/event"
	smartcontractstorage "github.com/darchlabs/synchronizer-v2/internal/storage/smartcontract"
	transactionstorage "github.com/darchlabs/synchronizer-v2/internal/storage/transaction"
	webhookstorage "github.com/darchlabs/synchronizer-v2/internal/storage/webhook"
	txsengine "github.com/darchlabs/synchronizer-v2/internal/txsengine"
	"github.com/darchlabs/synchronizer-v2/internal/webhooksender"
	CronjobAPI "github.com/darchlabs/synchronizer-v2/pkg/api/cronjob"
	EventAPI "github.com/darchlabs/synchronizer-v2/pkg/api/event"
	"github.com/darchlabs/synchronizer-v2/pkg/api/metrics"
	smartcontractsAPI "github.com/darchlabs/synchronizer-v2/pkg/api/smartcontracts"
	"github.com/darchlabs/synchronizer-v2/pkg/util"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	uuid "github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	"github.com/pressly/goose/v3"

	_ "github.com/darchlabs/synchronizer-v2/migrations"
)

var (
	eventStorage        synchronizer.EventStorage
	smartContactStorage synchronizer.SmartContractStorage
	cronjobSvc          synchronizer.Cronjob
	transactionStorage  synchronizer.TransactionStorage
	txsEngine           txsengine.TxsEngine
)

func main() {
	// load env values
	var env env.Env
	err := envconfig.Process("", &env)
	if err != nil {
		log.Fatal("invalid env values, error: ", err)
	}

	networksEtherscanURL, err := util.ParseStringifiedMap(env.NetworksEtherscanURL)
	if err != nil {
		log.Fatal(err)
	}

	networksEtherscanAPIKey, err := util.ParseStringifiedMap(env.NetworksEtherscanAPIKey)
	if err != nil {
		log.Fatal(err)
	}

	networksNodeURL, err := util.ParseStringifiedMap(env.NetworksNodeURL)
	if err != nil {
		log.Fatal(err)
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
	transactionStorage = transactionstorage.New(s)
	smartContactStorage = smartcontractstorage.New(s, eventStorage, transactionStorage)
	webhookStorage := webhookstorage.New(s)

	// initialize webhook sender, start processing events and retrying failed webhooks
	webhookSender := webhooksender.NewWebhookSender(webhookStorage, &http.Client{}, time.Duration(env.WebhooksIntervalSeconds+2))

	// Inicializar los webhooks desde el almacenamiento persistente
	if err := webhookSender.InitializeFromStorage(); err != nil {
		log.Fatalf("Error initializing webhooks from storage: %v", err)
	}
	go webhookSender.ProcessWebhooks()
	go webhookSender.StartRetries()

	// initialize fiber
	api := fiber.New()
	api.Use(logger.New())
	api.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))

	// create clients map
	clients := make(map[string]*ethclient.Client)

	// initialize the cronjob
	cronjobSvc = cronjob.New(env.CronjobIntervalSeconds, eventStorage, smartContactStorage, &clients, env.Debug, uuid.NewString, time.Now, webhookSender)

	// initialize http client with rate limiter
	client := httpclient.NewClient(&httpclient.Options{
		MaxRetry:        2,
		MaxRequest:      5,
		WindowInSeconds: 1,
	}, http.DefaultClient)

	// Initialize the transactions engine
	txsEngine = txsengine.New(txsengine.Config{
		ContractStorage:    smartContactStorage,
		TransactionStorage: transactionStorage,
		IdGen:              uuid.NewString,
		EtherscanUrlMap:    networksEtherscanURL,
		ApiKeyMap:          networksEtherscanAPIKey,
		NodesUrlMap:        networksNodeURL,
		Client:             client,
		MaxTransactions:    env.MaxTransactions,
	})

	// configure routers
	smartcontractsAPI.Route(api, smartcontractsAPI.Context{
		Storage:      smartContactStorage,
		EventStorage: eventStorage,
		TxsEngine:    txsEngine,
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
	metrics.Route(api, metrics.Context{
		SmartContractStorage: smartContactStorage,
		TransactionStorage:   transactionStorage,
		EventStorage:         eventStorage,
		Engine:               txsEngine,
	})

	// run process
	err = cronjobSvc.Start()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		api.Listen(fmt.Sprintf(":%s", env.Port))
	}()

	// Run txs engine process
	txsEngine.Start(env.CronjobIntervalSeconds + 1)

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
