package sync

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
)

type SmartContractQuerier interface {
	InsertSmartContractQuery(storage.QueryContext, *storage.SmartContractRecord) error
	SelectSmartContractByAddressQuery(storage.Transaction, string) (*storage.SmartContractRecord, error)
}

type ABIQuerier interface {
	InsertABIBatchQuery(storage.QueryContext, []*storage.ABIRecord, string) error
	SelectABIByAddressQuery(storage.Transaction, string) ([]*storage.ABIRecord, error)
}

type InputQuerier interface {
	InsertInputBatchQuery(storage.QueryContext, []*storage.InputRecord, string) error
	SelectInputByABIIDQuery(storage.Transaction, string) ([]*storage.InputRecord, error)
}

type SmartContractUserQuerier interface {
	UpsertSmartContractUserQuery(storage.Transaction, *storage.SmartContractUserRecord) error
}

type EventQuerier interface {
	InsertEventBatchQuery(storage.QueryContext, []*storage.EventRecord, string) error
	SelectEventsByAddressQuery(storage.Transaction, string) ([]*storage.EventRecord, error)
}
