package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTableAbiAddInputColumn, downAlterTableAbiAddInputColumn)
}

func upAlterTableAbiAddInputColumn(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec(`ALTER TABLE abi ADD COLUMN inputs jsonb DEFAULT '[]'::jsonb;`)
	if err != nil {
		return err
	}
	return nil
}

func downAlterTableAbiAddInputColumn(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
