package query

import (
	"fmt"
	"time"

	"github.com/darchlabs/synchronizer-v2/internal/pagination"
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

type SelectSmartContractUserQueryOutput struct {
	ID                 string                `db:"id"`
	Name               string                `db:"name"`
	Status             string                `db:"status"`
	Error              *string               `db:"error"`
	WebhookURL         string                `db:"webhook"`
	Network            string                `db:"network"`
	Address            string                `db:"address"`
	LastTxBlockSynced  int64                 `db:"last_tx_block_synced"`
	InitialBlockNumber int64                 `db:"initial_block_number"`
	CreatedAt          time.Time             `db:"created_at"`
	UpdatedAt          time.Time             `db:"updated_at"`
	Events             []storage.EventRecord `db:"-"`
}

func (sq *SmartContractQuerier) SelectSmartContractUserQuery(tx storage.Transaction, userID string, p *pagination.Pagination) ([]*SelectSmartContractUserQueryOutput, error) {
	records := make([]*SelectSmartContractUserQueryOutput, 0)

	err := tx.Select(
		&records,
		fmt.Sprintf(`
			SELECT
				sc.id as id,
				scu.name as name,
				scu.status as status,
				scu.error as error,
				scu.webhook as webhook,
				sc.network as network,
				sc.address as address,
				sc.last_tx_block_synced as last_tx_block_synced,
				sc.initial_block_number as initial_block_number,
				sc.created_at as created_at
			FROM smartcontracts sc
			JOIN smartcontract_users scu
			ON sc.address = scu.sc_address
			WHERE scu.user_id = $3
			ORDER BY sc.created_at %s
			LIMIT $1
			OFFSET $2`, p.Sort),
		p.Limit,
		p.Offset,
		userID,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query: SmartcontractQuerier.SelectSmartContractUserQuery tx.Get error")
	}

	return records, nil
}
