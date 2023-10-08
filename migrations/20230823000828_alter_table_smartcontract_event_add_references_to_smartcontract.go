package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTableSmartcontractEventAddReferencesToSmartcontract, downAlterTableSmartcontractEventAddReferencesToSmartcontract)
}

func upAlterTableSmartcontractEventAddReferencesToSmartcontract(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec(`
		ALTER TABLE smartcontract_event
		ADD CONSTRAINT fk_sc_event_smartcontract_id
		FOREIGN KEY (smartcontract_id) REFERENCES smartcontracts(id);`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_smartcontract_id ON smartcontract_event(smartcontract_id);`)
	if err != nil {
		return err
	}
	return nil
}

func downAlterTableSmartcontractEventAddReferencesToSmartcontract(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	_, err := tx.Exec("DROP INDEX IF EXISTS idx_smartcontract_id;")
	if err != nil {
		return err
	}

	_, err = tx.Exec("ALTER TABLE smartcontract_event DROP CONSTRAINT fk_smartcontract_id;")
	if err != nil {
		return err
	}
	return nil
}
