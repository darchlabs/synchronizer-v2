package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAddIndexToBlockNumberColumnInEventDataTable, downAddIndexToBlockNumberColumnInEventDataTable)
}

func upAddIndexToBlockNumberColumnInEventDataTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE INDEX idx_block_number ON event_data (block_number);
	`)

	return err
}

func downAddIndexToBlockNumberColumnInEventDataTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		DROP INDEX idx_block_number;
	`)

	return err
}
