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
			abi_id,
			status,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`,
		input.ID,
		input.Network,
		input.NodeURL,
		input.Address,
		input.LatestBlockNumber,
		input.AbiID,
		input.Status,
		input.CreatedAt,
	)
	if err != nil {
		return errors.Wrap(err, "query: EventQuerier.InsertEventQuery qCtx.Exec error")
	}
	return nil
}
