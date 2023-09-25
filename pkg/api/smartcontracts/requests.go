package smartcontracts

import "encoding/json"

type smartContractReq struct {
	UserID     string    `json:"-"`
	Network    string    `json:"network" validate:"required"`
	Name       string    `json:"name" validate:"required"`
	Address    string    `json:"address" validate:"required"`
	NodeURL    string    `json:"nodeUrl"`
	WebhookURL string    `json:"webhook"`
	ABI        []*abiReq `json:"abi"`
}

type abiReq struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Anonymous bool   `json:"anonymous"`

	Inputs []inputReq `json:"inputs"`
}

type inputReq struct {
	Indexed      bool   `json:"indexed"`
	InternalType string `json:"internalType"`
	Name         string `json:"name"`
	Type         string `json:"type"`
}

func transformInputsJsonToArray(jsonStr string) ([]inputReq, error) {
	var inputReqs []inputReq

	err := json.Unmarshal([]byte(jsonStr), &inputReqs)
	if err != nil {
		return nil, err
	}

	return inputReqs, nil
}
