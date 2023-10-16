package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAddEngineError, downRemoveEngineError)
}

func upAddEngineError(tx *sql.Tx) error {
	// Add column 'engine_error'
	_, err := tx.Exec(`
		ALTER TABLE smartcontracts ADD COLUMN engine_error TEXT NOT NULL DEFAULT '';
	`)
	if err != nil {
		return err
	}

	// Set the value 'running' for all existing records
	_, err = tx.Exec(`
		UPDATE smartcontracts SET engine_error = '';
	`)
	return err
}

func downRemoveEngineError(tx *sql.Tx) error {
	// Remove the 'engine_error' column
	_, err := tx.Exec(`
		ALTER TABLE smartcontracts DROP COLUMN engine_error;
	`)
	return err
}
