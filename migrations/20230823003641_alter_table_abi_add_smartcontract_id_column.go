package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTableAbiAddSmartcontractIdColumn, downAlterTableAbiAddSmartcontractIdColumn)
}

func upAlterTableAbiAddSmartcontractIdColumn(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec(`
		ALTER TABLE abi
		ADD COLUMN smartcontract_id TEXT NOT NULL;`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		ALTER TABLE abi
		ADD CONSTRAINT fk_abi_smartcontract_id
		FOREIGN KEY (smartcontract_id) REFERENCES smartcontracts(id);`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_sc_event_smartcontract_id ON abi(smartcontract_id);`)
	if err != err {
		return err
	}
	return nil
}

func downAlterTableAbiAddSmartcontractIdColumn(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	_, err := tx.Exec("DROP INDEX IF EXISTS idx_sc_event_smartcontract_id;")
	if err != nil {
		return err
	}

	_, err = tx.Exec("ALTER TABLE abi DROP CONSTRAINT fk_abi_smartcontract_id;")
	if err != nil {
		return err
	}
	return nil
}
