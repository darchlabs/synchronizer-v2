package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTableTransactionAddReferencesToSmartcontracts, downAlterTableTransactionAddReferencesToSmartcontracts)
}

func upAlterTableTransactionAddReferencesToSmartcontracts(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec(`
		ALTER TABLE transactions
		ADD CONSTRAINT fk_contract_id
		FOREIGN KEY (contract_id) REFERENCES smartcontracts(id);`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_contract_id ON transactions(contract_id);`)
	if err != nil {
		return err
	}

	return nil
}

func downAlterTableTransactionAddReferencesToSmartcontracts(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	_, err := tx.Exec("DROP INDEX IF EXISTS idx_contract_id;")
	if err != nil {
		return err
	}

	_, err = tx.Exec("ALTER TABLE transactions DROP CONSTRAINT fk_contract_id;")
	if err != nil {
		return err
	}
	return nil
}
