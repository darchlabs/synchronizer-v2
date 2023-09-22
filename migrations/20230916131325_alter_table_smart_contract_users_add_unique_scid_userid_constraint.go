package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTableSmartContractUsersAddUniqueScidUseridConstraint, downAlterTableSmartContractUsersAddUniqueScidUseridConstraint)
}

func upAlterTableSmartContractUsersAddUniqueScidUseridConstraint(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec(`
		ALTER TABLE smartcontract_users RENAME COLUMN smartcontract_id TO sc_address;`,
	)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		ALTER TABLE smartcontract_users
		ADD CONSTRAINT unique_userid_sc_address
		UNIQUE(user_id, sc_address);`,
	)
	if err != nil {
		return err
	}

	return nil
}

func downAlterTableSmartContractUsersAddUniqueScidUseridConstraint(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	_, err := tx.Exec(`
		ALTER TABLE smartcontract_users DROP unique_userid_sc_address;`,
	)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		ALTER TABLE smartcontract_users RENAME COLUMN sc_address TO smartcontract_id;`,
	)
	if err != nil {
		return err
	}

	return nil
}
