package transactionstorage

import (
	"errors"
	"fmt"
	"time"

	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/darchlabs/synchronizer-v2/pkg/transaction"
	"github.com/lib/pq"
)

type Storage struct {
	storage *storage.S
}

func New(s *storage.S) *Storage {
	return &Storage{
		storage: s,
	}
}

func (s *Storage) ListTxs(sort string, limit int64, offset int64) ([]*transaction.Transaction, error) {
	// define events response
	txs := []*transaction.Transaction{}

	// get txs from db
	eventQuery := fmt.Sprintf("SELECT * FROM transactions ORDER BY block_number %s LIMIT $1 OFFSET $2", sort)
	err := s.storage.DB.Select(&txs, eventQuery, limit, offset)
	if err != nil {
		return nil, err
	}

	return txs, nil
}

func (s *Storage) GetTxById(id string) (*transaction.Transaction, error) {
	// define events response
	tx := transaction.Transaction{}

	// get txs from db
	eventQuery := "SELECT * FROM transactions WHERE id = $1"
	err := s.storage.DB.Select(&tx, eventQuery, id)
	if err != nil {
		return nil, err
	}

	return &tx, nil

}

func (s *Storage) ListCurrentHashes() (*[]string, error) {
	// define events response
	var hashesArr []string

	// get txs from db
	eventQuery := "SELECT tx FROM transactions"
	err := s.storage.DB.Select(&hashesArr, eventQuery)
	if err != nil {
		return nil, err
	}

	return &hashesArr, nil
}

func (s *Storage) GetTVL(contractAddr string) (*int64, error) {
	// define events response
	var tvl int64

	// get txs from db
	eventQuery := "SELECT SUM(contract_balance) FROM transaction WHERE address = $1"
	err := s.storage.DB.Select(&tvl, eventQuery, contractAddr)
	if err != nil {
		return nil, err
	}

	return &tvl, nil

}

func (s *Storage) ListTotalAddresses(contractAddr string) (*int64, error) {
	// define events response
	var totalAddr int64

	query := "SELECT COUNT(DISTINCT address) FROM transactions WHERE contract_addr = $1"

	// execute query and retrieve result
	err := s.storage.DB.Select(&totalAddr, query, contractAddr)
	if err != nil {
		return nil, err
	}

	return &totalAddr, nil
}

// get the last synced tx and its block before executing it
func (s *Storage) InsertTxsByContract(transactions []*transaction.Transaction) error {
	// check it has enough len
	if len(transactions) == 0 {
		return errors.New("the transactions array to insert is empty")
	}

	// prepare transaction to create event on db
	tx, err := s.storage.DB.Beginx()
	if err != nil {
		return err
	}

	/* Prepare the query values */
	// Make an array of each field from the transactions array
	var (
		ids, contractIds, blockNumbers, hashes, fromAddresses,
		fromBalances, contractBalances, txsGases,
		gasPrices, gasUsed, isErrorTxs, fromWhales,
		txsValues, cumulativeGasesUsed, confirmations, txsReceipts,
		functionNames, timestamps, createdAtTxs, updatedAtTxs []string
	)

	// Create the array for each transaction field
	for _, txData := range transactions {
		ids = append(ids, txData.ID)
		contractIds = append(contractIds, txData.ContractID)
		hashes = append(hashes, txData.Hash)
		blockNumbers = append(blockNumbers, txData.BlockNumber)
		fromAddresses = append(fromAddresses, txData.From)
		fromBalances = append(fromBalances, txData.FromBalance)
		fromWhales = append(fromWhales, txData.FromIsWhale)
		txsValues = append(txsValues, txData.Value)
		contractBalances = append(contractBalances, txData.ContractBalance)
		txsGases = append(txsGases, txData.Gas)
		gasPrices = append(gasPrices, txData.GasPrice)
		gasUsed = append(gasUsed, txData.GasUsed)
		cumulativeGasesUsed = append(cumulativeGasesUsed, txData.CumulativeGasUsed)
		confirmations = append(confirmations, txData.Confirmations)
		isErrorTxs = append(isErrorTxs, txData.IsError)
		txsReceipts = append(txsReceipts, txData.TxReceiptStatus)
		functionNames = append(functionNames, txData.FunctionName)
		timestamps = append(timestamps, txData.Timestamp)
		updatedAtTxs = append(updatedAtTxs, txData.UpdatedAt.Format(time.RFC3339))
		createdAtTxs = append(createdAtTxs, txData.CreatedAt.Format(time.RFC3339))
	}

	// Insert the txs on the query
	/// @notice: `unnest` improves query performance
	transactionsQuery := `INSERT INTO transactions (
		id, contract_id, hash, block_number, "from", from_balance, from_is_whale, value,  contract_balance, gas, gas_price, gas_used, cumulative_gas_used, confirmations, is_error, tx_receipt_status, function_name, timestamp, created_at, updated_at
		)
		SELECT *
		FROM unnest($1::text[], $2::text[], $3::text[], $4::text[], $5::text[], $6::text[], $7::text[], $8::text[], $9::text[], $10::text[], $11::text[], $12::text[], $13::text[], $14::text[],
			$15::text[], $16::text[], $17::text[], $18::text[], $19::timestamp with time zone[], $20::timestamp with time zone[]
		)
		ON CONFLICT (hash) DO NOTHING`

	// Insert all of the values in the table, and then obtain each smart contract id with its last block number as response
	_, err = tx.Exec(
		transactionsQuery,
		pq.Array(ids), pq.Array(contractIds), pq.Array(hashes), pq.Array(blockNumbers),
		pq.Array(fromAddresses), pq.Array(fromBalances), pq.Array(fromWhales), pq.Array(txsValues),
		pq.Array(contractBalances), pq.Array(txsGases), pq.Array(gasPrices), pq.Array(gasUsed),
		pq.Array(cumulativeGasesUsed), pq.Array(confirmations), pq.Array(isErrorTxs), pq.Array(txsReceipts),
		pq.Array(functionNames), pq.Array(timestamps), pq.Array(createdAtTxs), pq.Array(updatedAtTxs))
	if err != nil {
		tx.Rollback()
		return err
	}

	// Get the smart contract id and the last block number from its txs array
	contractID := transactions[len(transactions)-1].ContractID
	latestBlockNumber := transactions[len(transactions)-1].BlockNumber

	// Update the smart contract with the latest block number
	smartContractQuery := `UPDATE smartcontracts SET last_tx_block_synced = $1 WHERE id = $2`
	_, err = tx.Exec(smartContractQuery, latestBlockNumber, contractID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit the whole tx when it finishes
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
