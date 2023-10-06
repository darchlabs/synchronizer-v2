package query

import (
	"github.com/Masterminds/squirrel"
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

type SelectEventsQueryFilters struct {
	SmartContractAddress string
	Status               string
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
