package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTableSmartcontractsDeleteUpdatedAtNullConstraint, downAlterTableSmartcontractsDeleteUpdatedAtNullConstraint)
}

func upAlterTableSmartcontractsDeleteUpdatedAtNullConstraint(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec("ALTER TABLE smartcontracts DROP COLUMN updated_at;")
	if err != nil {
		return err
	}

	_, err = tx.Exec("ALTER TABLE smartcontracts ADD COLUMN updated_at TIMESTAMPTZ;")
	if err != nil {
		return err
	}
	return nil
}

func downAlterTableSmartcontractsDeleteUpdatedAtNullConstraint(tx *sql.Tx) error {
	// THIS MIGRATION IS LEFT EMPTY BECAUSE THE COLUMN WAS DEFINED WRONG FROM THE ORIGIN
	// AND WE DONT NEET A REVERT.
	return nil
}
