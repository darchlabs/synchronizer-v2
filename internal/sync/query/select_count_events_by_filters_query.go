package query

import (
	"github.com/Masterminds/squirrel"
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

type SelectCountEventsQueryFilters struct {
	SmartContractAddress string
	Status               string
}

func (eq *EventQuerier) SelectCountEventsQuery(
	tx storage.Transaction,
	filters *SelectEventsQueryFilters,
) (int64, error) {
	q := squirrel.Select("COUNT(*)").From("event")

	if filters.Status != "" {
		q = q.Where("status = ?", filters.Status)
	}

	if filters.SmartContractAddress != "" {
		q = q.Where("sc_address = ?", filters.SmartContractAddress)
	}

	query, args, err := q.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "query: EventQuerier.SelectEventsQuery q.PlaceholderFormat().ToSql error")
	}

	var count int64
	err = tx.Get(&count, query, args...)
	if err != nil {
		return 0, errors.Wrap(err, "query: EventQuerier.CountEventsQuery tx.Get error")
	}

	return count, nil
}
