package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAddColumnWebhookToSmartcontractTable, downAddColumnWebhookToSmartcontractTable)
}

func upAddColumnWebhookToSmartcontractTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE smartcontracts
		ADD COLUMN webhook TEXT;
	`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		UPDATE smartcontracts
		SET webhook = '';
	`)
	if err != nil {
		return err
	}

	return nil
}

func downAddColumnWebhookToSmartcontractTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE smartcontracts
		DROP COLUMN webhook;
	`)
	if err != nil {
		return err
	}

	return nil
}
