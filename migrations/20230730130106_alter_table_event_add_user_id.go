package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTableEventAddUserId, downAlterTableEventAddUserId)
}

func upAlterTableEventAddUserId(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec(`ALTER TABLE events ADD COLUMN user_id TEXT NOT NULL;`)
	if err != nil {
		return err
	}

	return nil
}

func downAlterTableEventAddUserId(tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE events DROP COLUMN user_id TEXT NOT NULL;`)
	if err != nil {
		return err
	}
	// This code is executed when the migration is rolled back.
	return nil
}
