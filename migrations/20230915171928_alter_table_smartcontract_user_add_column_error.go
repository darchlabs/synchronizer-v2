package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTableSmartcontractUserAddColumnError, downAlterTableSmartcontractUserAddColumnError)
}

func upAlterTableSmartcontractUserAddColumnError(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec("ALTER TABLE smartcontract_users ADD error TEXT;")
	if err != nil {
		return err
	}

	return nil
}

func downAlterTableSmartcontractUserAddColumnError(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	_, err := tx.Exec("ALTER TABLE smartcontract_users DROP COLUMN error;")
	if err != nil {
		return err
	}

	return nil
}
