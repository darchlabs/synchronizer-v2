package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTablesMoveNameColumnToSmartcontractUsers, downAlterTablesMoveNameColumnToSmartcontractUsers)
}

func upAlterTablesMoveNameColumnToSmartcontractUsers(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec("ALTER TABLE smartcontracts DROP COLUMN name;")
	if err != nil {
		return err
	}

	_, err = tx.Exec("ALTER TABLE smartcontract_users ADD name TEXT;")
	if err != nil {
		return err
	}

	return nil
}

func downAlterTablesMoveNameColumnToSmartcontractUsers(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	_, err := tx.Exec("ALTER TABLE smartcontract_users DROP COLUMN name;")
	if err != nil {
		return err
	}

	_, err = tx.Exec("ALTER TABLE smartcontracts ADD name TEXT;")
	if err != nil {
		return err
	}

	return nil
}
