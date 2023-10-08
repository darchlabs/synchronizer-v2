package query

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

func (eq *EventQuerier) InsertEventQuery(qCtx storage.QueryContext, input *storage.EventRecord) error {
	_, err := qCtx.Exec(`
		INSERT INTO event (
			id,
			network,
			node_url,
			address,
			latest_block_number,
			sc_address,
			status,
			created_at,
			name,
			abi_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);`,
		input.ID,
		input.Network,
		input.NodeURL,
		input.Address,
		input.LatestBlockNumber,
		input.SmartContractAddress,
		input.Status,
		input.CreatedAt,
		input.Name,
		input.AbiID,
	)
	if err != nil {
		return errors.Wrap(err, "query: EventQuerier.InsertEventQuery qCtx.Exec error")
	}
	return nil
}
