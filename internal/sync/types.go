package sync

import (
	"github.com/darchlabs/synchronizer-v2/internal/pagination"
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/darchlabs/synchronizer-v2/internal/sync/query"
)

// Smart contracts
type SmartContractQuerier interface {
	InsertSmartContractQuery(storage.QueryContext, *storage.SmartContractRecord) error
	SelectSmartContractByAddressQuery(storage.Transaction, string) (*storage.SmartContractRecord, error)
	SelectSmartContractsByAddressesList(tx storage.Transaction, addresses []string) ([]*storage.SmartContractRecord, error)
	SelectUserSmartContractsQuery(tx storage.Transaction, userID string, p *pagination.Pagination) ([]*query.UserSmartContractOutput, error)
	SelectCountUserSmartContractsQuery(db storage.Database, userID string) (int64, error)
}

// ABI
type ABIQuerier interface {
	InsertABIBatchQuery(storage.QueryContext, []*storage.ABIRecord, string) error
	SelectABIByAddressQuery(storage.Transaction, string) ([]*storage.ABIRecord, error)
	SelectABIByIDs(tx storage.Transaction, ids []string) ([]*storage.ABIRecord, error)
}

// Input
type InputQuerier interface {
	InsertInputBatchQuery(storage.QueryContext, []*storage.InputRecord, string) error
	SelectInputByABIIDQuery(storage.Transaction, string) ([]*storage.InputRecord, error)
}

type SmartContractUserQuerier interface {
	UpsertSmartContractUserQuery(storage.Transaction, *storage.SmartContractUserRecord) error
	SelectSmartContractUserQuery(storage.Transaction, string) ([]*storage.SmartContractUserRecord, error)
	SmartContractUsersByIDListQuery(storage.Transaction, []string) ([]*storage.SmartContractUserRecord, error)
}

type EventQuerier interface {
	InsertEventBatchQuery(storage.QueryContext, []*storage.EventRecord, string) error
	SelectEventsByAddressQuery(storage.Transaction, string) ([]*storage.EventRecord, error)
	SelectEventsByAddressesListQuery(storage.Transaction, []string) ([]*storage.EventRecord, error)
	SelectEventsQuery(storage.Transaction, *query.SelectEventsQueryFilters) ([]*storage.EventRecord, error)
	UpdateEventQuery(storage.Transaction, *query.UpdateEventQueryInput) (*storage.EventRecord, error)
	SelectCountEventsQuery(storage.Transaction, *query.SelectEventsQueryFilters) (int64, error)
}

type EventDataQuerier interface {
	InsertEventDataQuery(storage.QueryContext, *storage.EventDataRecord) error
	InsertEventDataBatchQuery(storage.Transaction, []*storage.EventDataRecord) error
	SelectCountEventDataQuery(tx storage.Transaction, input *query.SelectCountEventDataQueryFilters) (int64, error)
	SelectEventDataQuery(tx storage.Transaction, input *query.SelectEventDataQueryFilters) ([]*storage.EventDataRecord, error)
}
