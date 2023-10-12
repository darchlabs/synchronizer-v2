package events

import (
	"encoding/json"
	"time"
)

type EventRes struct {
	ID                   string     `json:"id"`
	AbiID                string     `json:"abi_id"`
	Network              string     `json:"network"`
	Name                 string     `json:"name"`
	NodeURL              string     `json:"nodeURL"`
	Address              string     `json:"address"`
	LatestBlockNumber    int64      `json:"latestBlockNumber"`
	SmartContractAddress string     `json:"scAddress"`
	Status               string     `json:"status"`
	Error                string     `json:"error"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            *time.Time `json:"updated_at"`
	ABI                  *AbiRes    `json:"abi"`
}

type AbiRes struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Anonymous bool   `json:"anonymous"`

	Inputs []InputRes `json:"inputs"`
}

type InputRes struct {
	Indexed      bool   `json:"indexed"`
	InternalType string `json:"internalType"`
	Name         string `json:"name"`
	Type         string `json:"type"`
}

type EventDataRes struct {
	ID          string          `json:"id"`
	EventID     string          `json:"eventId"`
	Tx          string          `json:"tx"`
	Data        json.RawMessage `json:"data"`
	BlockNumber int64           `json:"block_number"`
	CreatedAt   time.Time       `json:"created_at"`
}
