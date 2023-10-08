package sync

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/darchlabs/synchronizer-v2/internal/sync/query"
	"github.com/pkg/errors"
)

func (ng *Engine) SelectEventsByStatus(status storage.EventStatus) ([]*storage.EventRecord, error) {
	records, err := ng.EventQuerier.SelectEventsQuery(ng.database, &query.SelectEventsQueryFilters{
		Status: string(status),
	})
	if err != nil {
		return nil, errors.Wrap(err, "sync: Engine.SelectEventsByFilters ng.EventQuerier.SelectEventsQuery error")
	}

	return records, nil
}
