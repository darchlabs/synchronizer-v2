package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTransactionsUniqueAddNetwork, downAlterTransactionsUniqueAddNetwork)
}

func upAlterTransactionsUniqueAddNetwork(tx *sql.Tx) error {
	// Add chain_id column
	_, err := tx.Exec(`
			ALTER TABLE transactions
			ADD COLUMN chain_id TEXT NOT NULL;
		`)
	if err != nil {
		return err
	}

	// Drop UNIQUE constraint over hash column
	_, err = tx.Exec(`
			ALTER TABLE transactions
			DROP CONSTRAINT transactions_hash_key;
			`)
	if err != nil {
		return err
	}

	// Add UNIQUE constraint over the combination between hash and chain_id columns
	_, err = tx.Exec(`
			ALTER TABLE transactions
			ADD CONSTRAINT transactions_hash_network_key UNIQUE (hash, chain_id);
		`)
	if err != nil {
		return err
	}

	// This code is executed when the migration is applied.
	return nil
}

func downAlterTransactionsUniqueAddNetwork(tx *sql.Tx) error {
	_, err := tx.Exec(`
			ALTER TABLE transactions
			DROP CONSTRAINT transactions_hash_cahin_id_key;
		`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
			ALTER TABLE transactions
			ADD CONSTRAINT transactions_hash_key UNIQUE (hash);
		`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
	ALTER TABLE transactions
	DROP COLUMN chain_id;
	`)
	if err != nil {
		return err
	}

	// This code is executed when the migration is rolled back.
	return nil
}
