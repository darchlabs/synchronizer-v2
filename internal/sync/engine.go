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
	SelectEventsAndABI(input *SelectEventsAndABIInput) (*SelectEventsAndABIOutput, error)
	SelectUserSmartContractsWithEvents(input *SelectUserSmartContractsWithEventsInput) (*SelectUserSmartContractsWithEventsOutput, error)
	SelectEventData(input *SelectEventDataInput) (*SelectEventDataOutput, error)
}

type Engine struct {
	database storage.Database

	ABIQuerier               ABIQuerier
	SmartContractQuerier     SmartContractQuerier
	SmartContractUserQuerier SmartContractUserQuerier
	InputQuerier             InputQuerier
	EventQuerier             EventQuerier
	EventDataQuerier         EventDataQuerier

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

		ABIQuerier:               query.NewABIQuerier(nil, uuid.NewString, time.Now),
		SmartContractQuerier:     query.NewSmartContractQuerier(nil, uuid.NewString, time.Now),
		SmartContractUserQuerier: query.NewSmartContractUserQuerier(nil, uuid.NewString, time.Now),
		InputQuerier:             query.NewInputQuerier(nil, uuid.NewString, time.Now),
		EventQuerier:             query.NewEventsQuerier(nil, uuid.NewString, time.Now),
		EventDataQuerier:         query.NewEventDataQuerier(nil, uuid.NewString, time.Now),
	}
}

func (ng *Engine) GetDatabase() storage.Database {
	return ng.database
}
