package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTableSmartcontractDeleteNodeUrlAndWebhook, downAlterTableSmartcontractDeleteNodeUrlAndWebhook)
}

func upAlterTableSmartcontractDeleteNodeUrlAndWebhook(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec("ALTER TABLE smartcontracts DROP COLUMN webhook;")
	if err != nil {
		return err
	}

	_, err = tx.Exec("ALTER TABLE smartcontracts DROP COLUMN node_url;")
	if err != nil {
		return err
	}

	return nil
}

func downAlterTableSmartcontractDeleteNodeUrlAndWebhook(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	// This code is executed when the migration is applied.
	_, err := tx.Exec("ALTER TABLE smartcontracts ADD webhook TEXT;")
	if err != nil {
		return err
	}

	_, err = tx.Exec("ALTER TABLE smartcontracts ADD node_url TEXT;")
	if err != nil {
		return err
	}

	return nil
}
