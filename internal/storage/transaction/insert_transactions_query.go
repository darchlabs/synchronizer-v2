package transactionstorage

import (
	"time"

	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/darchlabs/synchronizer-v2/pkg/transaction"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

func (s *Storage) InsertTxs(transactions []*transaction.Transaction) (err error) {
	// check it has enough len
	if len(transactions) == 0 {
		return errors.Wrap(ErrTransactionsEmpty, "transactionstorage: Storage.InsertTxs error")
	}

	// prepare transaction to create event on db
	tx, err := s.storage.DB.Beginx()
	if err != nil {
		return errors.Wrap(err, "transactionstorage: Storage.InsertTxs s.storage.DB.Beginx error")
	}

	defer func() {
		if err != nil && tx != nil {
			txErr := tx.Rollback()
			if txErr != nil {
				err = errors.WithMessagef(txErr, "transactionstorage: Storage.InsertTxs rollback transaction error: %s", err.Error())
			}
		}
	}()

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
		return errors.Wrap(err, "transactionstorage: Storage.InsertTxs tx.Exec batch insert error")
	}

	// Get the smart contract id and the last block number from its txs array
	contractID := transactions[len(transactions)-1].ContractID
	latestBlockNumber := transactions[len(transactions)-1].BlockNumber

	// Update the smart contract with the latest block number, status and error
	smartContractQuery := `UPDATE smartcontracts SET last_tx_block_synced = $1, status = $2, error = $3 WHERE id = $4`
	_, err = tx.Exec(smartContractQuery, latestBlockNumber, smartcontract.StatusRunning, "", contractID)
	if err != nil {
		return errors.Wrap(err, "transactionstorage: Storage.InsertTxs tx.Exec update error")
	}

	// Commit the whole tx when it finishes
	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "transactionstorage: Storage.InsertTxs tx.Commit error")
	}

	return nil
}
