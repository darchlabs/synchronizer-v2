package transaction

import (
	"time"
)

type Transaction struct {
	ID              string    `json:"id" db:"id"`
	ContractID      string    `json:"contract_id" db:"contract_id"`
	Hash            string    `json:"hash" db:"hash"`
	FromAddr        string    `json:"from_addr" db:"from_addr"`
	FromBalance     string    `json:"from_balance" db:"from_balance"`
	FromIsWhale     bool      `json:"from_is_whale" db:"from_is_whale"`
	ContractBalance string    `json:"contract_balance" db:"contract_balance"`
	GasPaid         string    `json:"gas_paid" db:"gas_paid"`
	GasPrice        string    `json:"gas_price" db:"gas_price"`
	GasCost         string    `json:"gas_cost" db:"gas_cost"`
	Succeded        bool      `json:"succeded" db:"succeded"`
	BlockNumber     int64     `json:"block_number" db:"block_number"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}
