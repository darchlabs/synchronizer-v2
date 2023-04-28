package transactionstorage

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
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

	// Return an empty array and not null in case there are no rows
	if len(txs) == 0 {
		return []*transaction.Transaction{}, nil
	}

	return txs, nil
}

func (s *Storage) GetTotalTxsCount() (int64, error) {
	// define events response
	var totalTxs []int64

	// get txs from db
	eventQuery := "SELECT COUNT(*) FROM transactions"
	err := s.storage.DB.Select(&totalTxs, eventQuery)
	if err != nil {
		return 0, err
	}

	return totalTxs[0], nil
}

func (s *Storage) ListContractTxs(id string, sort string, limit int64, offset int64) ([]*transaction.Transaction, error) {
	// define events response
	var txs []*transaction.Transaction

	// get txs from db
	eventQuery := fmt.Sprintf("SELECT * FROM transactions WHERE contract_id = $1 ORDER BY block_number %s LIMIT $2 OFFSET $3", sort)
	err := s.storage.DB.Select(&txs, eventQuery, id, limit, offset)
	if err != nil {
		return nil, err
	}

	// Return an empty array and not null in case there are no rows
	if len(txs) == 0 {
		return []*transaction.Transaction{}, nil
	}

	return txs, nil
}

func (s *Storage) GetContractTotalTxsCount(id string) (int64, error) {
	// define events response
	var totalTxsNum []int64

	// get txs from db
	eventQuery := "SELECT COUNT(*) FROM transactions WHERE contract_id = $1"
	err := s.storage.DB.Select(&totalTxsNum, eventQuery, id)
	if err != nil {
		return 0, err
	}

	return totalTxsNum[0], nil
}

func (s *Storage) GetContractCurrentTVL(id string) (int64, error) {
	// define events response
	var lastTVL []string

	// get txs from db
	eventQuery := "SELECT contract_balance FROM transactions WHERE contract_id = $1 ORDER BY block_number DESC LIMIT 1"
	err := s.storage.DB.Select(&lastTVL, eventQuery, id)
	if err != nil {
		return 0, err
	}

	// Return 0 if there is no registers
	if lastTVL[0] == "" {
		return 0, nil
	}

	currentTVL, err := strconv.ParseInt(lastTVL[0], 10, 64)
	if err != nil {
		return 0, err
	}

	return currentTVL, nil
}

func (s *Storage) ListContractTVLs(id string, sort string, limit int64, offset int64) ([]string, error) {
	// define events response
	var tvlArr []string

	// get txs from db
	eventQuery := fmt.Sprintf("SELECT contract_balance FROM transactions WHERE contract_id = $1 ORDER BY block_number %s LIMIT $2 OFFSET $3", sort)
	err := s.storage.DB.Select(&tvlArr, eventQuery, id, limit, offset)
	if err != nil {
		return nil, err
	}

	// Return an empty array and not null in case there are no rows
	if len(tvlArr) == 0 {
		return []string{}, nil
	}

	return tvlArr, nil
}

func (s *Storage) ListContractUniqueAddresses(id string, sort string, limit int64, offset int64) ([]string, error) {
	var uniqueAddresses []string

	query := fmt.Sprintf("SELECT DISTINCT t.from FROM (SELECT t.from, t.block_number FROM transactions AS t WHERE contract_id = $1 ORDER BY t.block_number %s) t LIMIT $2 OFFSET $3", sort)

	// execute query and retrieve result
	err := s.storage.DB.Select(&uniqueAddresses, query, id, limit, offset)
	if err != nil {
		return nil, err
	}

	// Return an empty array and not null in case there are no rows
	if len(uniqueAddresses) == 0 {
		return []string{}, nil
	}

	return uniqueAddresses, nil
}

func (s *Storage) GetContractTotalAddressesCount(id string) (int64, error) {
	// define events response
	var totalAddr []int64

	query := "SELECT COUNT(DISTINCT t.from) FROM transactions as T WHERE contract_id = $1"

	// execute query and retrieve result
	err := s.storage.DB.Select(&totalAddr, query, id)
	if err != nil {
		return 0, err
	}

	return totalAddr[0], nil
}

func (s *Storage) ListContractFailedTxs(id string, sort string, limit int64, offset int64) ([]*transaction.Transaction, error) {
	var failedTxs []*transaction.Transaction

	query := fmt.Sprintf("SELECT * FROM transactions WHERE contract_id = $1 AND (is_error = '1' OR tx_receipt_status = '0') ORDER BY block_number %s LIMIT $2 OFFSET $3", sort)

	// execute query and retrieve result
	err := s.storage.DB.Select(&failedTxs, query, id, limit, offset)
	if err != nil {
		return nil, err
	}

	// Return an empty array and not null in case there are no rows
	if len(failedTxs) == 0 {
		return []*transaction.Transaction{}, nil
	}

	return failedTxs, nil
}

func (s *Storage) GetContractTotalFailedTxsCount(id string) (int64, error) {
	var totalFailedTxs []int64

	query := "SELECT COUNT(*) FROM transactions WHERE contract_id = $1 AND (is_error = '1' OR tx_receipt_status = '0')"

	// execute query and retrieve result
	err := s.storage.DB.Select(&totalFailedTxs, query, id)
	if err != nil {
		return 0, err
	}

	return totalFailedTxs[0], nil
}

func (s *Storage) ListContractGasSpent(id string, sort string, limit int64, offset int64) ([]string, error) {
	var gasSpentArr []string

	query := fmt.Sprintf("SELECT gas_used FROM transactions WHERE contract_id = $1 ORDER BY block_number %s LIMIT $2 OFFSET $3", sort)

	// execute query and retrieve result
	err := s.storage.DB.Select(&gasSpentArr, query, id, limit, offset)
	if err != nil {
		return nil, err
	}

	// Return an empty array and not null in case there are no rows
	if len(gasSpentArr) == 0 {
		return []string{}, nil
	}

	return gasSpentArr, nil
}

func (s *Storage) GetContractTotalValueTransferred(id string) (int64, error) {
	var totalValueTransferred []int64

	query := "SELECT SUM(value::bigint) FROM transactions WHERE contract_id = $1"

	// execute query and retrieve result
	err := s.storage.DB.Select(&totalValueTransferred, query, id)
	if err != nil {
		return 0, err
	}

	return totalValueTransferred[0], nil
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
		ids, contractIds, blockNumbers, hashes, chainIds, fromAddresses,
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
		chainIds = append(chainIds, txData.ChainID)
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
		id, contract_id, hash, chain_id, block_number, "from", from_balance, from_is_whale, value,  contract_balance, gas, gas_price, gas_used, cumulative_gas_used, confirmations, is_error, tx_receipt_status, function_name, timestamp, created_at, updated_at
		)
		SELECT * FROM unnest(
			$1::text[], $2::text[], $3::text[], $4::text[], $5::text[], $6::text[], $7::text[], $8::text[], $9::text[], $10::text[], $11::text[], $12::text[], $13::text[], $14::text[],
			$15::text[], $16::text[], $17::text[], $18::text[], $19::text[], $20::timestamp with time zone[], $21::timestamp with time zone[]
		)
		ON CONFLICT (hash, chain_id) DO NOTHING`

	// Insert all of the values in the table, and then obtain each smart contract id with its last block number as response
	_, err = tx.Exec(
		transactionsQuery,
		pq.Array(ids), pq.Array(contractIds), pq.Array(hashes), pq.Array(chainIds), pq.Array(blockNumbers),
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

	// Update the smart contract with the latest block number, status and error
	smartContractQuery := `UPDATE smartcontracts SET last_tx_block_synced = $1, status = $2, error = $3 WHERE id = $4`
	_, err = tx.Exec(smartContractQuery, latestBlockNumber, smartcontract.StatusRunning, "", contractID)
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
