package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upCreateEventTable, downCreateEventTable)
}

func upCreateEventTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS event (
			id TEXT PRIMARY KEY,
			network TEXT NOT NULL,
			node_url TEXT NOT NULL,
			address TEXT NOT NULL,
			latest_block_number BIGINT NOT NULL,
			abi_id TEXT NOT NULL,
			status TEXT NOT NULL,
			error TEXT,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
			FOREIGN KEY (abi_id) REFERENCES abi (id) ON DELETE CASCADE
		)
	`)

	return err
}

func downCreateEventTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP TABLE IF EXISTS event
	`)

	return err
}
