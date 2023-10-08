package query

import (
	"time"

	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

type UpdateEventQueryInput struct {
	ID                   *string
	AbiID                *string
	Network              *storage.EventNetwork
	Name                 *string
	NodeURL              *string
	Address              *string
	LatestBlockNumber    *int64
	SmartContractAddress *string
	Status               *storage.EventStatus
	Error                *string
	UpdatedAt            *time.Time
}

func (eq *EventQuerier) UpdateEventQuery(tx storage.Transaction, input *UpdateEventQueryInput) (*storage.EventRecord, error) {
	var record storage.EventRecord
	err := tx.Get(&record, `
		UPDATE event
		SET
			abi_id = COALESCE($2, abi_id),
			network = COALESCE($3, network),
			name = COALESCE($4, name),
			node_url = COALESCE($5, node_url),
			address = COALESCE($6, address),
			latest_block_number = COALESCE($7, latest_block_number),
			sc_address = COALESCE($8, sc_address),
			status = COALESCE($9, status),
			error = COALESCE($10, error),
			updated_at = COALESCE($11, updated_at)
		WHERE id = $1
		RETURNING *;`,
		input.ID,
		input.AbiID,
		input.Network,
		input.Name,
		input.NodeURL,
		input.Address,
		input.LatestBlockNumber,
		input.SmartContractAddress,
		input.Status,
		input.Error,
		input.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "sync: Engine.UpdateEventQuery tx.Get error")
	}

	return &record, nil
}
