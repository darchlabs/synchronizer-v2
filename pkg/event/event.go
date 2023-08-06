package event

import (
	"encoding/json"
	"time"

	"github.com/darchlabs/synchronizer-v2/internal/blockchain"
)

type EventNetwork string

const (
	Ethereum EventNetwork = "ethereum"
	Polygon  EventNetwork = "polygon"
)

type EventStatus string

const (
	StatusSynching EventStatus = "synching"
	StatusRunning  EventStatus = "running"
	StatusStopped  EventStatus = "stopped"
	StatusError    EventStatus = "error"
)

type Event struct {
	ID                string       `json:"id" db:"id"`
	UserID            string       `json:"user_id" db:"user_id"`
	Network           EventNetwork `json:"network" db:"network"`
	NodeURL           string       `json:"nodeURL" db:"node_url"`
	Address           string       `json:"address" db:"address"`
	LatestBlockNumber int64        `json:"latestBlockNumber" db:"latest_block_number"`
	AbiID             string       `json:"abiId" db:"abi_id"`
	Status            EventStatus  `json:"status" db:"status"`
	Error             string       `json:"error" db:"error"`
	CreatedAt         time.Time    `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time    `json:"updatedAt" db:"updated_at"`

	Abi *Abi `json:"abi"`
}

type Abi struct {
	ID        string `id:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	Type      string `json:"type" db:"type"`
	Anonymous bool   `json:"anonymous" db:"anonymous"`

	Inputs []*Input `json:"inputs"`
}

type Input struct {
	ID           string `json:"id" db:"id"`
	Indexed      bool   `json:"indexed" db:"indexed"`
	InternalType string `json:"internalType" db:"internal_type"`
	Name         string `json:"name" db:"name"`
	Type         string `json:"type" db:"type"`
	AbiId        string `json:"abiId" db:"abi_id"`
}

type EventData struct {
	ID          string          `json:"id" db:"id"`
	EventID     string          `json:"eventId" db:"event_id"`
	Tx          string          `json:"tx" db:"tx"`
	BlockNumber int64           `json:"blockNumber" db:"block_number"`
	Data        json.RawMessage `json:"data" db:"data"`
	CreatedAt   time.Time       `json:"createdAt" db:"created_at"`
}

func (ed *EventData) FromLogData(logData blockchain.LogData, id string, eventID string, createdAt time.Time) error {
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
