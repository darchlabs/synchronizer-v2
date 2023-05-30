package transaction

import (
	"time"
)

type Transaction struct {
	ID                string    `json:"id" db:"id"`
	ContractID        string    `json:"contractId" db:"contract_id"`
	Hash              string    `json:"hash" db:"hash"`
	ChainID           string    `json:"chainId" db:"chain_id"`
	BlockNumber       string    `json:"blockNumber" db:"block_number"`
	From              string    `json:"from" db:"from"`
	FromBalance       string    `json:"fromBalance" db:"from_balance"`
	FromIsWhale       string    `json:"fromIsWhale" db:"from_is_whale"`
	Value             string    `json:"value" db:"value"`
	ContractBalance   string    `json:"contract_balance" db:"contract_balance"`
	Gas               string    `json:"gas" db:"gas"`
	GasPrice          string    `json:"gasPrice" db:"gas_price"`
	GasUsed           string    `json:"gasUsed" db:"gas_used"`
	CumulativeGasUsed string    `json:"cumulativeGasUsed" db:"cumulative_gas_used"`
	Confirmations     string    `json:"confirmations" db:"confirmations"`
	IsError           string    `json:"isError" db:"is_error"`
	TxReceiptStatus   string    `json:"txReceiptStatus" db:"tx_receipt_status"`
	FunctionName      string    `json:"functionName" db:"function_name"`
	Timestamp         string    `json:"timestamp" db:"timestamp"`
	CreatedAt         time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time `json:"updatedAt" db:"updated_at"`
}
