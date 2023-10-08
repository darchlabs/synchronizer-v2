package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTableAbiUpdateScAddressColumn, downAlterTableAbiUpdateScAddressColumn)
}

func upAlterTableAbiUpdateScAddressColumn(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec(`ALTER TABLE abi DROP COLUMN smartcontract_id;`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`ALTER TABLE abi ADD COLUMN sc_address TEXT REFERENCES smartcontracts(address);`)
	if err != nil {
		return err
	}
	return nil
}

func downAlterTableAbiUpdateScAddressColumn(tx *sql.Tx) error {
	// THIS MIGRATION IS LEFT EMPTY BECAUSE THE COLUMN WAS DEFINED WRONG FROM THE ORIGIN
	// AND WE DONT NEET A REVERT.
	return nil
}
