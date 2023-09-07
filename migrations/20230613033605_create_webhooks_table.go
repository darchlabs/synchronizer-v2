package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upCreateWebhooksTable, downCreateWebhooksTable)
}

func upCreateWebhooksTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS webhooks (
			id TEXT PRIMARY KEY,
			entity_type TEXT NOT NULL CHECK(entity_type IN ('event', 'transaction')),
			entity_id TEXT NOT NULL,
			endpoint TEXT NOT NULL,
			payload JSON NOT NULL,
			max_attempts INT NOT NULL DEFAULT 3,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			sent_at TIMESTAMP,
			attempts INT DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'pending',
			next_retry_at TIMESTAMP
		);
	`)
	if err != nil {
		return err
	}

	return nil
}

func downCreateWebhooksTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP TABLE IF EXISTS webhooks;
	`)

	return err
}
