package event

type Abi struct {
	ID        int64  `id:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	Type      string `json:"type" db:"type"`
	Anonymous bool   `json:"anonymous" db:"anonymous"`

	Inputs []*Input `json:"inputs"`
}

type Input struct {
	ID           int64  `json:"id" db:"id"`
	Indexed      bool   `json:"indexed" db:"indexed"`
	InternalType string `json:"internalType" db:"internal_type"`
	Name         string `json:"name" db:"name"`
	Type         string `json:"type" db:"type"`
	AbiId        int64  `json:"abiId" db:"abi_id"`
}
