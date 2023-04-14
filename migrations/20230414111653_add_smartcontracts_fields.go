package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAddSmartContractsFields, downAddSmartContractsFields)
}

func upAddSmartContractsFields(tx *sql.Tx) error {
	_, err := tx.Exec(`
ALTER TABLE smartcontracts
ADD COLUMN last_tx_block_synced BIGINT NOT NULL DEFAULT 0,
ADD COLUMN status TEXT NOT NULL DEFAULT 'idle',
ADD COLUMN error TEXT;
`)

	return err
}

func downAddSmartContractsFields(tx *sql.Tx) error {
	_, err := tx.Exec(`
	ALTER TABLE smartcontracts
	DROP COLUMN last_tx_block_synced,
	DROP COLUMN status,
	DROP COLUMN error;
	`)

	return err
}
