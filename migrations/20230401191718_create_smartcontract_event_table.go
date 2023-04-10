package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upCreateSmartcontractEventTable, downCreateSmartcontractEventTable)
}

func upCreateSmartcontractEventTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS smartcontract_event (
			id TEXT PRIMARY KEY,
			smartcontract_id TEXT NOT NULL,
			event_id TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL
		)
	`)

	return err
}

func downCreateSmartcontractEventTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP TABLE IF EXISTS smartcontract_event
	`)

	return err
}
