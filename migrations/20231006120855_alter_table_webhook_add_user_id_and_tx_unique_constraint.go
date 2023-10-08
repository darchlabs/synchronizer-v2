package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upAlterTableWebhookAddUserIdAndTxUniqueConstraint, downAlterTableWebhookAddUserIdAndTxUniqueConstraint)
}

func upAlterTableWebhookAddUserIdAndTxUniqueConstraint(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	_, err := tx.Exec("ALTER TABLE webhooks ADD COLUMN tx TEXT NOT NULL;")
	if err != nil {
		return err
	}
	_, err = tx.Exec("ALTER TABLE webhooks ADD COLUMN user_id TEXT NOT NULL;")
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		ALTER TABLE webhooks
		ADD CONSTRAINT unique_userid_user_id_payload
		UNIQUE(user_id, tx);`,
	)
	if err != nil {
		return err
	}

	return nil
}

func downAlterTableWebhookAddUserIdAndTxUniqueConstraint(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
