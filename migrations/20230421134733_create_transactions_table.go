package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upCreateTransactionsTable, downCreateTransactionsTable)
}

func upCreateTransactionsTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS transactions (
		id TEXT PRIMARY KEY,
		contract_id TEXT NOT NULL,
		hash TEXT NOT NULL UNIQUE,
		block_number TEXT NOT NULL,
		"from" TEXT NOT NULL,
		from_balance TEXT,
		from_is_whale TEXT,
		value TEXT NOT NULL,
		contract_balance TEXT,
		gas TEXT NOT NULL,
		gas_price TEXT NOT NULL,
		gas_used TEXT NOT NULL,
		cumulative_gas_used TEXT NOT NULL,
		confirmations TEXT NOT NULL,
		is_error TEXT NOT NULL,
		tx_receipt_status TEXT NOT NULL,
		function_name TEXT NOT NULL,
		timestamp TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL
	);
	`)

	return err
}

func downCreateTransactionsTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
	DROP TABLE IF EXISTS transactions;
	`)

	return err
}
