package synchronizer

import (
	"github.com/darchlabs/synchronizer-v2/pkg/event"
	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/darchlabs/synchronizer-v2/pkg/transaction"
)

type EventStorage interface {
	ListAllEvents() ([]*event.Event, error)
	ListEvents(sort string, limit int64, offset int64) ([]*event.Event, error)
	ListEventsByAddress(address string, sort string, limit int64, offset int64) ([]*event.Event, error)
	GetEvent(address string, eventName string) (*event.Event, error)
	GetEventByID(id string) (*event.Event, error)
	InsertEvent(e *event.Event) (*event.Event, error)
	UpdateEvent(e *event.Event) error
	DeleteEvent(address string, eventName string) error
	ListEventData(address string, eventName string, sort string, limit int64, offset int64) ([]*event.EventData, error)
	InsertEventData(e *event.Event, data []*event.EventData) error
	GetEventsCount() (int64, error)
	GetEventCountByAddress(address string) (int64, error)
	GetEventDataCount(address string, eventName string) (int64, error)
	Stop() error
}

type Cronjob interface {
	Stop() error
	Restart() error
	Start() error
	Halt()
	GetStatus() string
	GetSeconds() int64
	GetError() string
}

type SmartContractStorage interface {
	ListSmartContracts(sort string, limit int64, offset int64) ([]*smartcontract.SmartContract, error)
	InsertSmartContract(s *smartcontract.SmartContract) (*smartcontract.SmartContract, error)
	UpdateLastBlockNumber(id string, blockNumber int64) error
	DeleteSmartContractByAddress(address string) error
	GetSmartContractByID(id string) (*smartcontract.SmartContract, error)
	GetSmartContractByAddress(address string) (*smartcontract.SmartContract, error)
	GetSmartContractsCount() (int64, error)
	UpdateStatusAndError(id string, status smartcontract.SmartContractStatus, err error) error
	Stop() error
}

type TransactionStorage interface {
	InsertTxsByContract([]*transaction.Transaction) error

	ListTxs(sort string, limit int64, offset int64) ([]*transaction.Transaction, error)
	ListContractTxs(id string) ([]*transaction.Transaction, error)
	GetContractTotalTxs(id string) (int64, error)
	GetTxById(id string) (*transaction.Transaction, error)

	ListContractFailedTxs(id string) ([]*transaction.Transaction, error)
	GetContractTotalFailedTxs(id string) (int64, error)

	GetContractTotalAddresses(id string) (int64, error)
	ListContractUniqueAddresses(id string) ([]string, error)

	GetContractCurrentTVL(id string) (int64, error)
	ListContractTVLs(id string) ([]int64, error)

	GetContractTotalGasSpent(id string) (int64, error)
	ListContractGasSpent(id string) ([]string, error)

	GetContractTotalValueTransferred(id string) (int64, error)
}
