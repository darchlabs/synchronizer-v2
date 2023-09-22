package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTablesMoveStatusColumnToSmartcontractUsers, downAlterTablesMoveStatusColumnToSmartcontractUsers)
}

func upAlterTablesMoveStatusColumnToSmartcontractUsers(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec("ALTER TABLE smartcontracts DROP COLUMN status;")
	if err != nil {
		return err
	}

	_, err = tx.Exec("ALTER TABLE smartcontract_users ADD status TEXT;")
	if err != nil {
		return err
	}

	return nil
}

func downAlterTablesMoveStatusColumnToSmartcontractUsers(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	_, err := tx.Exec("ALTER TABLE smartcontract_users DROP COLUMN status;")
	if err != nil {
		return err
	}

	_, err = tx.Exec("ALTER TABLE smartcontracts ADD status TEXT;")
	if err != nil {
		return err
	}

	return nil
}
