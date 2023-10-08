package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upCreateTableSmartcontractUser, downCreateTableSmartcontractUser)
}

func upCreateTableSmartcontractUser(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec(`
		CREATE TABLE smartcontract_users (
			id                  TEXT PRIMARY KEY NOT NULL,
			user_id             TEXT NOT NULL,
			smartcontract_id    TEXT REFERENCES smartcontracts(id) NOT NULL,
			created_at           TIMESTAMPTZ,
			updated_at           TIMESTAMPTZ,
			deleted_at           TIMESTAMPTZ DEFAULT NULL
		);`)
	if err != nil {
		return err
	}
	return nil
}

func downCreateTableSmartcontractUser(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	_, err := tx.Exec(`DROP TABLE smartcontract_users`)
	if err != nil {
		return err
	}
	return nil
}
