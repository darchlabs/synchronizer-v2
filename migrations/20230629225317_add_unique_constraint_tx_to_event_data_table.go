package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAddUniqueConstraintTx, downRemoveUniqueConstraintTx)
}

func upAddUniqueConstraintTx(tx *sql.Tx) error {
	// Add unique constraint to 'tx' column
	_, err := tx.Exec(`
		ALTER TABLE event_data ADD CONSTRAINT unique_tx UNIQUE(tx);
	`)
	return err
}

func downRemoveUniqueConstraintTx(tx *sql.Tx) error {
	// Remove the unique constraint from 'tx' column
	_, err := tx.Exec(`
		ALTER TABLE event_data DROP CONSTRAINT unique_tx;
	`)
	return err
}
