package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTableEventAddScAddressColumn, downAlterTableEventAddScAddressColumn)
}

func upAlterTableEventAddScAddressColumn(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec("ALTER TABLE event ADD COLUMN sc_address TEXT REFERENCES smartcontracts(address);")
	if err != nil {
		return err
	}

	_, err = tx.Exec("ALTER TABLE event DROP COLUMN abi_id;")
	if err != nil {
		return err
	}
	return nil
}

func downAlterTableEventAddScAddressColumn(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	_, err := tx.Exec("ALTER TABLE event DROP COLUMN sc_address;")
	if err != nil {
		return err
	}
	return nil
}
