package event

type Abi struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Anonymous bool `json:"anonymous"`
	Inputs []*Input `json:"inputs"`
}

type Input struct {
	Indexed bool `json:"indexed"`
	InternalType string `json:"internalType"`
	Name string `json:"name"`
	Type string `json:"type"`
}