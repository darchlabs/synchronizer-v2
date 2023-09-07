package smartcontract

import (
	"time"

	"github.com/darchlabs/synchronizer-v2/pkg/event"
)

type SmartContract struct {
	ID                 string              `json:"id" db:"id"`
	Name               string              `json:"name" db:"name" validate:"required"`
	Network            event.EventNetwork  `json:"network" db:"network" validate:"required"`
	NodeURL            string              `json:"nodeURL" db:"node_url"`
	Address            string              `json:"address" db:"address" validate:"required"`
	LastTxBlockSynced  int64               `json:"lastTxBlockSynced" db:"last_tx_block_synced"`
	Status             SmartContractStatus `json:"status" db:"status"`
	Error              *string             `json:"error" db:"error"`
	Webhook            string              `json:"webhook" db:"webhook" validate:"omitempty,url"`
	InitialBlockNumber int64               `json:"initialBlockNumber" db:"initial_block_number"`

	Abi    []*event.Abi   `json:"abi,omitempty" validate:"required,gt=0,dive"`
	Events []*event.Event `json:"events"`

	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

type SmartContractUpdate struct {
	Name    string `json:"name"`
	NodeURL string `json:"nodeURL" validate:"omitempty,url"`
	Webhook string `json:"webhook" validate:"omitempty,url"`
}

type SmartContractStatus string

const (
	StatusIdle          SmartContractStatus = "idle"
	StatusRunning       SmartContractStatus = "running"
	StatusStopping      SmartContractStatus = "stopping"
	StatusSynching      SmartContractStatus = "synching"
	StatusStopped       SmartContractStatus = "stopped"
	StatusError         SmartContractStatus = "error"
	StatusQuotaExceeded SmartContractStatus = "quota_exceeded"
)

func (s *SmartContract) IsSynced() bool {
	return s.LastTxBlockSynced >= s.InitialBlockNumber
}
