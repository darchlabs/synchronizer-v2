package sync

import (
	"time"

	"github.com/darchlabs/synchronizer-v2/internal/logger"
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/darchlabs/synchronizer-v2/internal/sync/query"
	"github.com/darchlabs/synchronizer-v2/internal/wrapper"
	"github.com/google/uuid"
)

type SyncEngine interface {
	InsertAtomicSmartContract(input *InsertAtomicSmartContractInput) (*InsertAtomicSmartContractOutput, error)
}

type Engine struct {
	database storage.Database

	abiQuerier               ABIQuerier
	smartContractQuerier     SmartContractQuerier
	smartContractUserQuerier SmartContractUserQuerier
	inputQuerier             InputQuerier
	eventQuerier             EventQuerier

	dateGen wrapper.DateGenerator
	idGen   wrapper.IDGenerator

	logger logger.Client
}

type EngineConfig struct {
	Database storage.Database
	Logger   logger.Client
}

func NewEngine(conf *EngineConfig) *Engine {
	return &Engine{
		database: conf.Database,
		logger:   conf.Logger,
		dateGen:  time.Now,
		idGen:    uuid.NewString,

		abiQuerier:               query.NewABIQuerier(nil, uuid.NewString, time.Now),
		smartContractQuerier:     query.NewSmartContractQuerier(nil, uuid.NewString, time.Now),
		smartContractUserQuerier: query.NewSmartContractUserQuerier(nil, uuid.NewString, time.Now),
		inputQuerier:             query.NewInputQuerier(nil, uuid.NewString, time.Now),
		eventQuerier:             query.NewEventsQuerier(nil, uuid.NewString, time.Now),
	}
}
