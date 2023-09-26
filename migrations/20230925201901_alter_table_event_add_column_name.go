package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTableEventAddColumnName, downAlterTableEventAddColumnName)
}

func upAlterTableEventAddColumnName(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec("ALTER TABLE event ADD COLUMN name TEXT NOT NULL;")
	if err != nil {
		return err
	}
	return nil
}

func downAlterTableEventAddColumnName(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
