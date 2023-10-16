package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAddColumnEngineLastTxBlockSyncedToSmartcontractTable, downAddColumnEngineLastTxBlockSyncedToSmartcontractTable)
}

func upAddColumnEngineLastTxBlockSyncedToSmartcontractTable(tx *sql.Tx) error {
	// Add column 'engine_last_tx_block_synced' and set its default value to 0
	_, err := tx.Exec(`
		ALTER TABLE smartcontracts ADD COLUMN engine_last_tx_block_synced BIGINT NOT NULL DEFAULT 0;
	`)
	if err != nil {
		return err
	}

	// Set zero value to 'engine_last_tx_block_synced'
	_, err = tx.Exec(`
		UPDATE smartcontracts SET engine_last_tx_block_synced = 0;
	`)
	if err != nil {
		return err
	}

	return nil
}

func downAddColumnEngineLastTxBlockSyncedToSmartcontractTable(tx *sql.Tx) error {
	// Drop column 'engine_last_tx_block_synced'
	_, err := tx.Exec(`
		ALTER TABLE smartcontracts DROP COLUMN engine_last_tx_block_synced;
	`)

	return err
}
