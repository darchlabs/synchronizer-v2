package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upCreateInputTable, downCreateInputTable)
}

func upCreateInputTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS input (
			id TEXT PRIMARY KEY,
			indexed BOOLEAN NOT NULL,
			internal_type TEXT NOT NULL,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			abi_id TEXT NOT NULL,
			FOREIGN KEY (abi_id) REFERENCES abi (id) ON DELETE CASCADE
		)
	`)

	return err
}

func downCreateInputTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP TABLE IF EXISTS input
	`)

	return err
}
