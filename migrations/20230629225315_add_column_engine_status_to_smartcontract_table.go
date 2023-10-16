package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAddEngineStatus, downRemoveEngineStatus)
}

func upAddEngineStatus(tx *sql.Tx) error {
	// Add column 'engine_status'
	_, err := tx.Exec(`
		ALTER TABLE smartcontracts ADD COLUMN engine_status TEXT NOT NULL DEFAULT 'running';
	`)
	if err != nil {
		return err
	}

	// Set the value 'running' for all existing records
	_, err = tx.Exec(`
		UPDATE smartcontracts SET engine_status = 'running';
	`)
	return err
}

func downRemoveEngineStatus(tx *sql.Tx) error {
	// Remove the 'engine_status' column
	_, err := tx.Exec(`
		ALTER TABLE smartcontracts DROP COLUMN engine_status;
	`)
	return err
}
