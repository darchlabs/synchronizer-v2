package query

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

func (eq *EventDataQuerier) InsertEventDataBatchQuery(tx storage.Transaction, records []*storage.EventDataRecord) error {
	for _, r := range records {
		err := eq.InsertEventDataQuery(tx, r)
		if err != nil {
			return errors.Wrap(err, "sync: EventDataQuerier.InsertEventDataBatchQuery error")
		}
	}

	return nil
}
