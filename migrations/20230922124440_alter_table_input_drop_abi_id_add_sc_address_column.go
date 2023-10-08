package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTableInputDropAbiIdAddScAddressColumn, downAlterTableInputDropAbiIdAddScAddressColumn)
}

func upAlterTableInputDropAbiIdAddScAddressColumn(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec("ALTER TABLE input DROP COLUMN abi_id;")
	if err != nil {
		return err
	}

	_, err = tx.Exec("ALTER TABLE input ADD COLUMN sc_address TEXT REFERENCES smartcontracts(address);")
	if err != nil {
		return err
	}
	return nil
}

func downAlterTableInputDropAbiIdAddScAddressColumn(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
