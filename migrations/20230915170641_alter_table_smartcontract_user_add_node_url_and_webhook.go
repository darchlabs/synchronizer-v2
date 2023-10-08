package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTableSmartcontractUserAddNodeUrlAndWebhook, downAlterTableSmartcontractUserAddNodeUrlAndWebhook)
}

func upAlterTableSmartcontractUserAddNodeUrlAndWebhook(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec("ALTER TABLE smartcontract_users ADD webhook TEXT;")
	if err != nil {
		return err
	}

	_, err = tx.Exec("ALTER TABLE smartcontract_users ADD node_url TEXT;")
	if err != nil {
		return err
	}

	return nil
}

func downAlterTableSmartcontractUserAddNodeUrlAndWebhook(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	// This code is executed when the migration is applied.
	_, err := tx.Exec("ALTER TABLE smartcontract_users DROP COLUMN webhook;")
	if err != nil {
		return err
	}

	_, err = tx.Exec("ALTER TABLE smartcontract_users DROP COLUMN node_url TEXT;")
	if err != nil {
		return err
	}

	return nil
}
