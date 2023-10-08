package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTableEventSetErrorColumnDefaultValue, downAlterTableEventSetErrorColumnDefaultValue)
}

func upAlterTableEventSetErrorColumnDefaultValue(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec("ALTER TABLE event ALTER COLUMN error SET DEFAULT '';")
	if err != nil {
		return err
	}
	return nil
}

func downAlterTableEventSetErrorColumnDefaultValue(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
