package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterSmarcontractsName, downAlterSmarcontractsName)
}

func upAlterSmarcontractsName(tx *sql.Tx) error {
	_, err := tx.Exec(`
	ALTER TABLE smartcontract RENAME TO smartcontracts;
	`)

	return err
}

func downAlterSmarcontractsName(tx *sql.Tx) error {
	_, err := tx.Exec(`
	ALTER TABLE smartcontracts RENAME TO smartcontract;
`)
	return err
}
