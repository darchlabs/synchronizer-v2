package query

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

func (sq *SmartContractUserQuerier) UpsertSmartContractUserQuery(
	tx storage.Transaction,
	input *storage.SmartContractUserRecord,
) error {
	err := tx.Get(input, `
		INSERT INTO smartcontract_users (id, user_id, sc_address, webhook, node_url, status, created_at, name)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT(user_id, sc_address)
		DO UPDATE SET
				created_at = excluded.created_at,
				webhook = COALESCE(smartcontract_users.webhook, excluded.webhook)
		RETURNING *;`,
		input.ID,
		input.UserID,
		input.SmartContractAddress,
		input.WebhookURL,
		input.NodeURL,
		input.Status,
		input.CreatedAt,
		input.Name,
	)
	if err != nil {
		return errors.Wrap(err, "query: SmartContractUserRecord.UpsertSmartContractUserQuery tx.Get error")
	}

	return nil
}
