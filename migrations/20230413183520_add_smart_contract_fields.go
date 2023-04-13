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
ADD COLUMN last_tx_block_synced BIGINT NOT NULL,
ADD COLUMN status TEXT NOT NULL,
ADD COLUMN error TEXT;
`)

	return err
}

func downAddSmartContractsFields(tx *sql.Tx) error {
	_, err := tx.Exec(`
	ALTER TABLE smartcontracts
	DROP COLUMN last_tx_block_synced BIGINT NOT NULL,
	DROP COLUMN status TEXT NOT NULL,
	DROP COLUMN error TEXT;
	`)

	return err
}
