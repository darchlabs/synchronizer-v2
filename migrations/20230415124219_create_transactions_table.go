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
		from_addr TEXT NOT NULL,
		from_balance TEXT NOT NULL,
		from_is_whale TEXT,
		contract_balance TEXT NOT NULL,
		gas_paid TEXT NOT NULL,
		gas_price TEXT NOT NULL,
		gas_cost TEXT NOT NULL,
		succeded BOOL NOT NULL,
		block_number BIGINT NOT NULL,
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
