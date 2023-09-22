package query

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

func (eq *EventQuerier) InsertEventBatchQuery(qCtx storage.QueryContext, inputs []*storage.EventRecord, abiID, userID string) error {
	for _, input := range inputs {
		input.CreatedAt = eq.dateGen()
		input.ID = eq.idGen()
		input.AbiID = abiID
		input.UserID = userID
		err := eq.InsertEventQuery(qCtx, input)
		if err != nil {
			return errors.Wrap(err, "query: EventQuerier.InsertEventQuery eq.InsertEventQuery error")
		}
	}

	return nil
}
