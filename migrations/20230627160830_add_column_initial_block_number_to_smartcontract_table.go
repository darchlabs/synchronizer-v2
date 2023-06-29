package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAddColumnInitialBlockNumberToSmartcontractTable, downAddColumnInitialBlockNumberToSmartcontractTable)
}

func upAddColumnInitialBlockNumberToSmartcontractTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE smartcontracts ADD COLUMN initial_block_number BIGINT NOT NULL DEFAULT 0;
	`)

	// add the same value of 'last_tx_block_synced' to created column
	_, err = tx.Exec(`
		UPDATE smartcontracts SET initial_block_number = last_tx_block_synced;
	`)

	return err
}

func downAddColumnInitialBlockNumberToSmartcontractTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE smartcontracts DROP COLUMN initial_block_number;
	`)

	return err
}
