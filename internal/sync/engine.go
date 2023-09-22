package sync

import (
	"time"

	"github.com/darchlabs/synchronizer-v2/internal/logger"
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/darchlabs/synchronizer-v2/internal/wrapper"
)

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
	}
}
