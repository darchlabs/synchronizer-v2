package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upCreateEventDataTable, downCreateEventDataTable)
}

func upCreateEventDataTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS event_data (
			id SERIAL PRIMARY KEY,
			event_id TEXT NOT NULL,
			tx TEXT NOT NULL,
			block_number BIGINT NOT NULL,
			data JSONB NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			FOREIGN KEY (event_id) REFERENCES event (id) ON DELETE CASCADE
		)
	`)

	return err
}

func downCreateEventDataTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP TABLE IF EXISTS event_data
	`)

	return err
}
