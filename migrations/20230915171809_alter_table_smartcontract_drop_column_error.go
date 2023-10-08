package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTableSmartcontractDropColumnError, downAlterTableSmartcontractDropColumnError)
}

func upAlterTableSmartcontractDropColumnError(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec("ALTER TABLE smartcontracts DROP COLUMN error;")
	if err != nil {
		return err
	}

	return nil
}

func downAlterTableSmartcontractDropColumnError(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	_, err := tx.Exec("ALTER TABLE smartcontracts ADD error TEXT;")
	if err != nil {
		return err
	}

	return nil
}
