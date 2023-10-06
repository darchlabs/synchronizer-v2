package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTableEventDataAddTxUniqueConstraint, downAlterTableEventDataAddTxUniqueConstraint)
}

func upAlterTableEventDataAddTxUniqueConstraint(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec(`
		ALTER TABLE event_data ADD CONSTRAINT unique_tx_event_data UNIQUE(tx);`,
	)
	if err != nil {
		return err
	}
	return nil
}

func downAlterTableEventDataAddTxUniqueConstraint(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
