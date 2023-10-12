package query

import (
	"github.com/Masterminds/squirrel"
	"github.com/darchlabs/synchronizer-v2/internal/pagination"
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

type SelectEventDataQueryFilters struct {
	SmartContractAddress string
	EventName            string
	Pagination           *pagination.Pagination
}

func (eq *EventDataQuerier) SelectEventDataQuery(
	tx storage.Transaction,
	input *SelectEventDataQueryFilters,
) ([]*storage.EventDataRecord, error) {
	records := make([]*storage.EventDataRecord, 0)

	q := squirrel.
		Select("event_data.*").
		From("event").
		Join("event_data ON event.id = event_data.event_id").
		Where("event.sc_address = ?", input.SmartContractAddress).
		Where("event.name = ?", input.EventName)

	if input.Pagination != nil {
		q = q.OrderBy("created_at " + input.Pagination.Sort)
		q = q.Limit(uint64(input.Pagination.Limit))
		q = q.Offset(uint64(input.Pagination.Offset))
	}

	query, args, err := q.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "query: EventDataQuerier.SelectEventDataQuery q.PlaceholderFormat().ToSql error")
	}

	err = tx.Select(&records, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "query: EventDataQuerier.SelectEventDataQuery tx.Select error")
	}

	return records, nil
}
