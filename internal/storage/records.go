package storage

import "time"

type Network string
type SmartContractUserStatus string
type EventNetwork string

type EventStatus string

const (
	// SmartContractStatus
	SmartContractStatusIdle          SmartContractUserStatus = "idle"
	SmartContractStatusRunning       SmartContractUserStatus = "running"
	SmartContractStatusStopping      SmartContractUserStatus = "stopping"
	SmartContractStatusSynching      SmartContractUserStatus = "synching"
	SmartContractStatusStopped       SmartContractUserStatus = "stopped"
	SmartContractStatusError         SmartContractUserStatus = "error"
	SmartContractStatusQuotaExceeded SmartContractUserStatus = "quota_exceeded"

	// EventNetwork
	Ethereum EventNetwork = "ethereum"
	Polygon  EventNetwork = "polygon"
	Mumbai   EventNetwork = "mumbai"

	// Event status
	EventStatusSynching EventStatus = "synching"
	EventStatusRunning  EventStatus = "running"
	EventStatusStopped  EventStatus = "stopped"
	EventStatusError    EventStatus = "error"
)

type SmartContractRecord struct {
	ID                 string  `db:"id"`
	Network            Network `db:"network"`
	Address            string  `db:"address"`
	LastTxBlockSynced  int64   `db:"last_tx_block_synced"`
	InitialBlockNumber int64   `db:"initial_block_number"`

	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

type SmartContractUserRecord struct {
	ID                   string                  `db:"id"`
	UserID               string                  `db:"user_id"`
	SmartContractAddress string                  `db:"sc_address"`
	Name                 string                  `db:"name"`
	ErrorMessage         *string                 `db:"error"`
	WebhookURL           string                  `db:"webhook"`
	NodeURL              string                  `db:"node_url"`
	Status               SmartContractUserStatus `db:"status"`
	CreatedAt            time.Time               `db:"created_at"`
	DeletedAt            *time.Time              `db:"deleted_at"`
	UpdatedAt            *time.Time              `db:"updated_at"`
}

type ABIRecord struct {
	ID                   string `db:"id"`
	SmartContractAddress string `db:"sc_address"`
	Name                 string `db:"name"`
	Type                 string `db:"type"`
	Anonymous            bool   `db:"anonymous"`
	Inputs               string `json:"inputs"`
}

type InputRecord struct {
	ID                   string `db:"id"`
	Indexed              bool   `db:"indexed"`
	InternalType         string `db:"internal_type"`
	Name                 string `db:"name"`
	Type                 string `db:"type"`
	SmartContractAddress string `db:"sc_address"`
}

type EventRecord struct {
	ID                   string       `db:"id"`
	UserID               string       `db:"user_id"`
	Network              EventNetwork `db:"network"`
	Name                 string       `db:"name"`
	NodeURL              string       `db:"node_url"`
	Address              string       `db:"address"`
	LatestBlockNumber    int64        `db:"latest_block_number"`
	SmartContractAddress string       `db:"sc_address"`
	Status               EventStatus  `db:"status"`
	Error                string       `db:"error"`
	CreatedAt            time.Time    `db:"created_at"`
	UpdatedAt            *time.Time   `db:"updated_at"`
}
