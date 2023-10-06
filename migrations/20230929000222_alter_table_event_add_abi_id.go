package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTableEventAddAbiId, downAlterTableEventAddAbiId)
}

func upAlterTableEventAddAbiId(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec("ALTER TABLE event ADD COLUMN abi_id TEXT NOT NULL REFERENCES abi(id);")
	if err != nil {
		return nil
	}

	return nil
}

func downAlterTableEventAddAbiId(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
