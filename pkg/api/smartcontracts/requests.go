package smartcontracts

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
)

type SmartContractReq struct {
	UserID     string    `json:"-"`
	Network    string    `json:"network" validate:"required"`
	Name       string    `json:"name" validate:"required"`
	Address    string    `json:"address" validate:"required"`
	NodeURL    string    `json:"nodeUrl"`
	WebhookURL string    `json:"webhook"`
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
