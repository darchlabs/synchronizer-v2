package query

import (
	"errors"

	"github.com/darchlabs/synchronizer-v2/internal/logger"
	"github.com/darchlabs/synchronizer-v2/internal/wrapper"
)

var (
	ErrInvalidDate = errors.New("query: date cannot be zero error")
)

type SmartContractStatus string

// SMART CONTRACT QUERIER
type SmartContractQuerier struct {
	idGen   wrapper.IDGenerator
	dateGen wrapper.DateGenerator
	logger  logger.Client
}

func NewSmartContractQuerier(logger logger.Client, idGen wrapper.IDGenerator, dateGen wrapper.DateGenerator) *SmartContractQuerier {
	return &SmartContractQuerier{
		idGen:   idGen,
		dateGen: dateGen,
		logger:  logger,
	}
}

// ABI QUERIER
type ABIQuerier struct {
	idGen   wrapper.IDGenerator
	dateGen wrapper.DateGenerator
	logger  logger.Client
}

func NewABIQuerier(logger logger.Client, idGen wrapper.IDGenerator, dateGen wrapper.DateGenerator) *ABIQuerier {
	return &ABIQuerier{
		idGen:   idGen,
		dateGen: dateGen,
		logger:  logger,
	}
}

// INPUT QUERIER
type InputQuerier struct {
	idGen   wrapper.IDGenerator
	dateGen wrapper.DateGenerator
	logger  logger.Client
}

func NewInputQuerier(logger logger.Client, idGen wrapper.IDGenerator, dateGen wrapper.DateGenerator) *InputQuerier {
	return &InputQuerier{
		idGen:   idGen,
		dateGen: dateGen,
		logger:  logger,
	}
}

// SMART CONTRACT USER QUERIER
type SmartContractUserQuerier struct {
	idGen   wrapper.IDGenerator
	dateGen wrapper.DateGenerator

	logger logger.Client
}

func NewSmartContractUserQuerier(logger logger.Client, idGen wrapper.IDGenerator, dateGen wrapper.DateGenerator) *SmartContractUserQuerier {
	return &SmartContractUserQuerier{
		idGen:   idGen,
		dateGen: dateGen,
		logger:  logger,
	}
}

// EVENTS QUERIER
type EventQuerier struct {
	idGen   wrapper.IDGenerator
	dateGen wrapper.DateGenerator
	logger  logger.Client
}

func NewEventsQuerier(logger logger.Client, idGen wrapper.IDGenerator, dateGen wrapper.DateGenerator) *EventQuerier {
	return &EventQuerier{
		idGen:   idGen,
		dateGen: dateGen,
		logger:  logger,
	}
}

// EVENTS QUERIER
type SmartcontractUserQuerier struct {
	idGen   wrapper.IDGenerator
	dateGen wrapper.DateGenerator
	logger  logger.Client
}

func NewSmartcontractUserQuerier(logger logger.Client, idGen wrapper.IDGenerator, dateGen wrapper.DateGenerator) *SmartcontractUserQuerier {
	return &SmartcontractUserQuerier{
		idGen:   idGen,
		dateGen: dateGen,
		logger:  logger,
	}
}

// EVENT DATA
type EventDataQuerier struct {
	idGen   wrapper.IDGenerator
	dateGen wrapper.DateGenerator
	logger  logger.Client
}

func NewEventDataQuerier(logger logger.Client, idGen wrapper.IDGenerator, dateGen wrapper.DateGenerator) *EventDataQuerier {
	return &EventDataQuerier{
		idGen:   idGen,
		dateGen: dateGen,
		logger:  logger,
	}
}
