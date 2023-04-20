package smartcontract

import (
	"time"

	"github.com/darchlabs/synchronizer-v2/pkg/event"
)

type SmartContract struct {
	ID                string              `json:"id" db:"id"`
	Name              string              `json:"name" db:"name" validate:"required"`
	Network           event.EventNetwork  `json:"network" db:"network" validate:"required"`
	NodeURL           string              `json:"nodeURL" db:"node_url" validate:"required,url"`
	Address           string              `json:"address" db:"address" validate:"required"`
	LastTxBlockSynced int64               `json:"last_tx_block_synced" db:"last_tx_block_synced"`
	Status            SmartContractStatus `json:"status" db:"status"`
	Error             *string             `json:"error" db:"error"`

	Abi    []*event.Abi   `json:"abi,omitempty" validate:"required,gt=0,dive"`
	Events []*event.Event `json:"events"`

	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

type SmartContractStatus string

const (
	StatusIdle     SmartContractStatus = "idle"
	StatusRunning  SmartContractStatus = "running"
	StatusStopping SmartContractStatus = "stopping"
	StatusStopped  SmartContractStatus = "stopped"
	StatusSynching SmartContractStatus = "synching"
	StatusError    SmartContractStatus = "error"
)
