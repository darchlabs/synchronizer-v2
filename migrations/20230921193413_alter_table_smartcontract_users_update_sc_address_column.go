package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTableSmartcontractUsersUpdateScAddressColumn, downAlterTableSmartcontractUsersUpdateScAddressColumn)
}

func upAlterTableSmartcontractUsersUpdateScAddressColumn(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec("ALTER TABLE smartcontracts ADD CONSTRAINT unique_smartcontract_address UNIQUE (address);")
	if err != nil {
		return err
	}

	_, err = tx.Exec("ALTER TABLE smartcontract_users DROP COLUMN sc_address;")
	if err != nil {
		return err
	}

	_, err = tx.Exec("ALTER TABLE smartcontract_users ADD COLUMN sc_address TEXT REFERENCES smartcontracts(address);")
	if err != nil {
		return err
	}

	_, err = tx.Exec("ALTER TABLE smartcontract_users ADD CONSTRAINT sc_address_user_id_unique UNIQUE (sc_address, user_id);")
	if err != nil {
		return err
	}
	return nil
}

func downAlterTableSmartcontractUsersUpdateScAddressColumn(tx *sql.Tx) error {
	// THIS MIGRATION IS LEFT EMPTY BECAUSE THE COLUMN WAS DEFINED WRONG FROM THE ORIGIN
	// AND WE DONT NEET A REVERT.
	return nil
}
