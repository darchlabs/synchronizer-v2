package sync

import (
	"time"

	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/darchlabs/synchronizer-v2/internal/sync/query"
	"github.com/pkg/errors"
)

type UpdateEventInput struct {
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
	UpdatedAt            time.Time
}

func (ng *Engine) UpdateEvent(input *UpdateEventInput) (*storage.EventRecord, error) {
	records, err := ng.EventQuerier.UpdateEventQuery(ng.database, &query.UpdateEventQueryInput{
		ID:                   input.ID,
		AbiID:                input.AbiID,
		Network:              input.Network,
		Name:                 input.Name,
		NodeURL:              input.NodeURL,
		Address:              input.Address,
		LatestBlockNumber:    input.LatestBlockNumber,
		SmartContractAddress: input.SmartContractAddress,
		Status:               input.Status,
		Error:                input.Error,
		UpdatedAt:            &input.UpdatedAt,
	})
	if err != nil {
		return nil, errors.Wrap(err, "sync: Engine.UpdateEvent ng.eventQuerier.UpdateEventQuery error")
	}

	return records, nil
}
