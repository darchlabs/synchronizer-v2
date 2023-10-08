package storage

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/darchlabs/synchronizer-v2/internal/blockchain"
	"github.com/darchlabs/synchronizer-v2/pkg/webhook"
	"github.com/pkg/errors"
)

type Network string
type SmartContractUserStatus string
type EventNetwork string
type EventStatus string
type WebhookEntityType string
type WebhookStatus string

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

	// Webhook status
	WebhookStatusPending   WebhookStatus = "pending"
	WebhookStatusFailed    WebhookStatus = "failed"
	WebhookStatusDelivered WebhookStatus = "delivered"

	// WebhookEntityType
	WebhookEntityTypeEvent   WebhookEntityType = "event"
	WebhookEntityTransaction WebhookEntityType = "transaction"
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

func (s *SmartContractRecord) IsSynced() bool {
	return s.LastTxBlockSynced >= s.InitialBlockNumber
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
	ID                   string      `db:"id"`
	Name                 string      `db:"name"`
	Type                 string      `db:"type"`
	Anonymous            bool        `db:"anonymous"`
	SmartContractAddress string      `db:"sc_address"`
	InputsJSON           []uint8     `db:"inputs"`
	Inputs               []*InputABI `db:"-" json:"inputs"`
}

type InputABI struct {
	Indexed      bool   `json:"indexed"`
	InternalType string `json:"internalType"`
	Name         string `json:"name"`
	Type         string `json:"type"`
}

func (i *InputABI) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, i)
}

func (i InputABI) Value() (driver.Value, error) {
	return json.Marshal(i)
}

func (r *ABIRecord) MarshalJson() ([]byte, error) {
	var ip []*InputABI
	err := json.Unmarshal([]byte(r.InputsJSON), &ip)
	if err != nil {
		return nil, errors.Wrap(err, "cannot unmarshal InputsJSON")
	}

	return json.Marshal(struct {
		Name      string      `json:"name"`
		Type      string      `json:"type"`
		Anonymous bool        `json:"anonymous"`
		Inputs    []*InputABI `json:"inputs"`
	}{
		Name:      r.Name,
		Type:      r.Type,
		Anonymous: r.Anonymous,
		Inputs:    ip,
	})
}

// TODO(mt): get rid of this record struct
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
	AbiID                string       `db:"abi_id"`
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

	// Agregation data only
	ABI                *ABIRecord                 `db:"-"`
	SmartContract      *SmartContractRecord       `db:"-"`
	SmartContractUsers []*SmartContractUserRecord `db:"-"`
}

type WebhookRecord struct {
	ID          string            `db:"id"`
	UserID      string            `db:"user_id"`
	EntityType  WebhookEntityType `db:"entity_type"`
	EntityID    string            `db:"entity_id"`
	Endpoint    string            `db:"endpoint"`
	Payload     json.RawMessage   `db:"payload"`
	MaxAttempts int               `db:"max_attempts"`
	CreatedAt   time.Time         `db:"created_at"`
	UpdatedAt   time.Time         `db:"updated_at"`
	SentAt      sql.NullTime      `db:"sent_at"`
	Attempts    int               `db:"attempts"`
	NextRetryAt sql.NullTime      `db:"next_retry_at"`
	Status      WebhookStatus     `db:"status"`
	Tx          string            `db:"tx"`
}

type EventDataRecord struct {
	ID          string          `db:"id"`
	EventID     string          `db:"event_id"`
	Tx          string          `db:"tx"`
	Data        json.RawMessage `db:"data"`
	BlockNumber int64           `db:"block_number"`
	CreatedAt   time.Time       `db:"created_at"`
}

func (ed *EventDataRecord) ToWebhookEvent(ID string, ev *EventRecord, endpoint string, date time.Time) (*WebhookRecord, error) {
	// prepare event payload
	payload := &webhook.WebhookEventPayload{
		Id:          ev.ID,
		Name:        ev.Name,
		BlockNumber: ed.BlockNumber,
		Tx:          ed.Tx,
		Data:        ed.Data,
	}

	// parse payload to raw message
	rawMessage, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &WebhookRecord{
		ID:         ID,
		Tx:         ed.Tx,
		EntityType: WebhookEntityTypeEvent,
		EntityID:   ev.ID,
		Endpoint:   endpoint,
		Payload:    rawMessage,
		CreatedAt:  date,
		UpdatedAt:  date,
	}, nil
}

func (ed *EventDataRecord) FromLogData(logData *blockchain.LogData, id string, eventID string, createdAt time.Time) error {
	// parse transaction to string
	tx := logData.Tx.Hex()

	// parse data to json
	data, err := json.Marshal(logData.Data)
	if err != nil {
		return err
	}

	ed.ID = id
	ed.EventID = eventID
	ed.Tx = tx
	ed.BlockNumber = int64(logData.BlockNumber)
	ed.Data = data
	ed.CreatedAt = createdAt

	return nil
}
