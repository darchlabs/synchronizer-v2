package query

import (
	"github.com/Masterminds/squirrel"
	"github.com/darchlabs/synchronizer-v2/internal/pagination"
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

type SelectEventsQueryFilters struct {
	SmartContractAddress string
	EventName            string
	Status               string
	Pagination           *pagination.Pagination
}

func (eq *EventQuerier) SelectEventsQuery(
	tx storage.Transaction,
	filters *SelectEventsQueryFilters,
) ([]*storage.EventRecord, error) {
	records := make([]*storage.EventRecord, 0)
	q := squirrel.Select("*").From("event")

	if filters.Status != "" {
		q = q.Where("status = ?", filters.Status)
	}

	if filters.SmartContractAddress != "" {
		q = q.Where("sc_address = ?", filters.SmartContractAddress)
	}

	if filters.EventName != "" {
		q = q.Where("name = ?", filters.EventName)
	}

	if filters.Pagination != nil {
		q = q.OrderBy("created_at " + filters.Pagination.Sort)
		q = q.Limit(uint64(filters.Pagination.Limit))
		q = q.Offset(uint64(filters.Pagination.Offset))
	}

	query, args, err := q.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "query: EventQuerier.SelectEventsQuery q.PlaceholderFormat().ToSql error")
	}

	err = tx.Select(&records, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "query: EventQuerier.SelectEventsQuery tx.Select error")
	}

	return records, nil
}
