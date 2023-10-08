package synchronizer

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/darchlabs/synchronizer-v2/pkg/event"
	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/darchlabs/synchronizer-v2/pkg/transaction"
	"github.com/darchlabs/synchronizer-v2/pkg/webhook"
)

type EventStorage interface {
	ListAllEvents() ([]*event.Event, error)
	ListEvents(sort string, limit int64, offset int64) ([]*event.Event, error)
	ListEventsByAddress(address string, sort string, limit int64, offset int64) ([]*event.Event, error)
	GetEvent(address string, eventName string) (*event.Event, error)
	GetEventById(id string) (*event.Event, error)
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
	ListAllSmartContracts() ([]*smartcontract.SmartContract, error)
	ListSmartContracts(sort string, limit int64, offset int64) ([]*smartcontract.SmartContract, error)
	ListUniqueSmartContractsByNetwork() ([]*smartcontract.SmartContract, error)
	InsertSmartContractQuery(s *smartcontract.SmartContract) error
	UpdateLastBlockNumber(id string, blockNumber int64) error
	DeleteSmartContractByAddress(address string) error
	GetSmartContractById(id string) (*smartcontract.SmartContract, error)
	GetSmartContractByAddress(address string) (*smartcontract.SmartContract, error)
	GetSmartContractsCount() (int64, error)
	UpdateStatusAndError(id string, status smartcontract.SmartContractStatus, err error) error
	UpdateSmartContract(sc *smartcontract.SmartContract) (*smartcontract.SmartContract, error)
	Stop() error
}

type TransactionStorage interface {
	InsertTxs([]*transaction.Transaction) error
	DeleteTransactionsByContractId(Id string) error
	ListTxs(sort string, limit int64, offset int64) ([]*transaction.Transaction, error)
	GetTxsCount() (int64, error)
	ListTxsById(id string, ctx *ListItemsInRangeCtx) ([]*transaction.Transaction, error)
	GetTxsCountById(id string) (int64, error)
	ListFailedTxsById(id string, ctx *ListItemsInRangeCtx) ([]*transaction.Transaction, error)
	GetFailedTxsCountById(id string) (int64, error)
	ListUniqueAddresses(id string, ctx *ListItemsInRangeCtx) ([]string, error)
	GetAddressesCountById(id string) (int64, error)
	ListTvlsById(id string, ctx *ListItemsInRangeCtx) ([][]string, error)
	GetTvlById(id string) (int64, error)
	ListGasSpentById(id string, startTs int64, endTs int64, interval int64) ([][]string, error)
	GetTotalGasSpentById(id string) (int64, error)
	GetValueTransferredById(id string) (int64, error)
}

type WebhookStorage interface {
	CreateWebhook(wh *webhook.Webhook) (*webhook.Webhook, error)
	UpdateWebhook(wh *webhook.Webhook) (*webhook.Webhook, error)
	GetWebhookByID(id string) (*webhook.Webhook, error)
	ListAllWebhooks() ([]*webhook.Webhook, error)
	ListWebhooks(smartcontractID string) ([]*webhook.Webhook, error)
	GetWebhooksForRetry() ([]*webhook.Webhook, error)
}

type SmartcontractUserStorage interface {
	InsertSmartContractUserQuery(tx storage.Transaction, input *storage.SmartContractUserRecord) error
}

type ListItemsInRangeCtx struct {
	StartTime string
	EndTime   string
	Sort      string
	Limit     int64
	Offset    int64
}

type GasTimestamp struct {
	GasUsed   int64 `db:"gas_used"`
	Timestamp int64 `db:"timestamp"`
}

type ContractBalanceTimestamp struct {
	ContractBalance string `db:"contract_balance"`
	Timestamp       string `db:"timestamp"`
}
