package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTableAbiChangeColumnInputsType, downAlterTableAbiChangeColumnInputsType)
}

func upAlterTableAbiChangeColumnInputsType(tx *sql.Tx) error {
	//This code is executed when the migration is applied.
	_, err := tx.Exec("ALTER TABLE abi ALTER COLUMN inputs TYPE TEXT;")
	if err != nil {
		return err
	}

	// MIGRATION HERE
	//_, err = tx.Exec("ALTER TABLE abi ALTER COLUMN inputs DROP DEFAULT;")
	//if err != nil {
	//return err
	//}

	//_, err = tx.Exec("ALTER TABLE abi ALTER COLUMN inputs SET DEFAULT '[]';")
	//if err != nil {
	//return err
	//}
	return nil
}

func downAlterTableAbiChangeColumnInputsType(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
