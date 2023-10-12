package smartcontracts

import (
	"time"

	"github.com/darchlabs/synchronizer-v2/internal/storage"
)

type SmartContractRequest struct {
	UserID     string    `json:"-"`
	Network    string    `json:"network" validate:"required"`
	Name       string    `json:"name" validate:"required"`
	Address    string    `json:"address" validate:"required"`
	NodeURL    string    `json:"nodeUrl"`
	WebhookURL string    `json:"webhook" validate:"omitempty,url"`
	ABI        []*AbiReq `json:"abi"`
}

type AbiReq struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Anonymous bool   `json:"anonymous"`

	Inputs []InputReq `json:"inputs"`
}

type InputReq struct {
	Indexed      bool   `json:"indexed"`
	InternalType string `json:"internalType"`
	Name         string `json:"name"`
	Type         string `json:"type"`
}

func TransformInputsJsonToArray(inputs []*storage.InputABI) ([]InputReq, error) {
	inputReqs := make([]InputReq, 0)
	for _, i := range inputs {
		inputReqs = append(inputReqs, InputReq{
			Indexed:      i.Indexed,
			InternalType: i.InternalType,
			Name:         i.Name,
			Type:         i.Type,
		})
	}

	return inputReqs, nil
}

type EventResponse struct {
	ID     string              `json:"id"`
	Name   string              `json:"name"`
	Status storage.EventStatus `json:"status"`
	Error  string              `json:"error"`
}

type SmartContractResponse struct {
	ID                 string  `json:"id"`
	Network            string  `json:"network"`
	Name               string  `json:"name"`
	Address            string  `json:"address"`
	NodeURL            string  `json:"nodeUrl"`
	Status             string  `json:"status"`
	WebhookURL         string  `json:"webhook"`
	LastTxBlockSynced  int64   `json:"lastTxBlockSynced"`
	InitialBlockNumber int64   `json:"initialBlockNumber"`
	Error              *string `json:"error"`

	Events []*EventResponse `json:"events,omitempty"`

	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}
