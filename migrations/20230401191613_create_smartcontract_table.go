package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upCreateSmartcontractTable, downCreateSmartcontractTable)
}

func upCreateSmartcontractTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS smartcontract (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			network TEXT NOT NULL,
			node_url TEXT NOT NULL,
			address TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL
		)
	`)

	return err
}

func downCreateSmartcontractTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP TABLE IF EXISTS smartcontract
	`)

	return err
}
