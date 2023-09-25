package query

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

func (eq *EventQuerier) InsertEventBatchQuery(qCtx storage.QueryContext, inputs []*storage.EventRecord, scAddress string) error {
	now := eq.dateGen()
	for _, input := range inputs {
		input.CreatedAt = now
		input.ID = eq.idGen()
		input.SmartContractAddress = scAddress
		err := eq.InsertEventQuery(qCtx, input)
		if err != nil {
			return errors.Wrap(err, "query: EventQuerier.InsertEventQuery eq.InsertEventQuery error")
		}
	}

	return nil
}
