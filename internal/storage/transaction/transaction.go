package transactionstorage

import (
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
	// prepare transaction to create event on db
	tx, err := s.storage.DB.Beginx()
	if err != nil {
		return err
	}

	/* Prepare the query values */
	// Make an array of each field from the transactions array
	var (
		ids, contractIds, hashes, fromAddresses,
		fromBalances, contractBalances, gasPaidTxs,
		gasPrices, gasCosts, createdAtTxs, updatedAtTxs []string
	)
	var blockNumbers []int64
	var succeededTxs, fromWhales []bool
	// var createdAtTxs, updatedAtTxs []time.Time

	// Create the array for each transaction field
	for _, txData := range transactions {
		ids = append(ids, txData.ID)
		contractIds = append(contractIds, txData.ContractID)
		hashes = append(hashes, txData.Hash)
		fromAddresses = append(fromAddresses, txData.FromAddr)
		fromBalances = append(fromBalances, txData.FromBalance)
		fromWhales = append(fromWhales, txData.FromIsWhale)
		contractBalances = append(contractBalances, txData.ContractBalance)
		gasPaidTxs = append(gasPaidTxs, txData.GasPaid)
		gasPrices = append(gasPrices, txData.GasPrice)
		gasCosts = append(gasCosts, txData.GasCost)
		blockNumbers = append(blockNumbers, txData.BlockNumber)
		succeededTxs = append(succeededTxs, txData.Succeded)
		createdAtTxs = append(createdAtTxs, txData.CreatedAt.Format(time.RFC3339))
		updatedAtTxs = append(updatedAtTxs, txData.UpdatedAt.Format(time.RFC3339))
	}

	// Insert the txs on the query
	/// @notice: `unnest` improves query performance
	transactionsQuery := `INSERT INTO transactions (id, contract_id, hash, from_addr, from_balance, contract_balance, gas_paid, gas_price, gas_cost, block_number, from_is_whale, succeded, created_at, updated_at) 
		SELECT id, contract_id, hash, from_addr, from_balance, contract_balance, gas_paid, gas_price, gas_cost, block_number::numeric(20, 0), from_is_whale, succeded, created_at, updated_at
		FROM unnest($1::text[], $2::text[], $3::text[], $4::text[], $5::text[], $6::text[], $7::text[], $8::text[], $9::text[], $10::numeric(20, 0)[], $11::bool[], $12::bool[], $13::timestamp with time zone[], $14::timestamp with time zone[])
		AS t(id, contract_id, hash, from_addr, from_balance, contract_balance, gas_paid, gas_price, gas_cost, block_number, from_is_whale, succeded, created_at, updated_at)
		ON CONFLICT (hash) DO NOTHING` // This is for don't insert bad or repeated hashes, but for making possible the items with correct hashes to insert

	// Insert all of the values in the table, and then obtain each smart contract id with its last block number as response
	_, err = tx.Exec(transactionsQuery, pq.Array(ids), pq.Array(contractIds), pq.Array(hashes), pq.Array(fromAddresses), pq.Array(fromBalances), pq.Array(contractBalances), pq.Array(gasPaidTxs), pq.Array(gasPrices), pq.Array(gasCosts), pq.Array(blockNumbers), pq.Array(fromWhales), pq.Array(succeededTxs), pq.Array(createdAtTxs), pq.Array(updatedAtTxs))
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

func (s *Storage) InsertTx(t *transaction.Transaction) (*transaction.Transaction, error) {
	// check if already existe an event with the same address and name
	tx, _ := s.GetTxById(t.ID)
	if tx != nil {
		return nil, fmt.Errorf("transaction already exists with hash=%s", t.Hash)
	}

	eventQuery := "INSERT INTO transactions (id, contract_id, hash, from_addr, from_balance, contract_balance, gas_paid, gas_price, gas_cost, from_is_whale, succeded, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)"
	_, err := s.storage.DB.Exec(eventQuery, t.ID, t.ContractID, t.Hash, t.FromAddr, t.FromBalance, t.ContractBalance, t.GasPaid, t.GasPrice, t.GasCost, t.FromIsWhale, t.Succeded, t.CreatedAt, t.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// get created event
	created, err := s.GetTxById(t.ID)
	if err != nil {
		return nil, err
	}

	return created, nil
}
