package query

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

type SelectCountEventDataQueryFilters struct {
	SmartContractAddress string
	EventName            string
}

func (eq *EventDataQuerier) SelectCountEventDataQuery(
	tx storage.Transaction,
	input *SelectCountEventDataQueryFilters,
) (int64, error) {
	var count int64

	err := tx.Get(
		&count, `
		SELECT COUNT(ed.id)
		FROM event_data ed
		JOIN event e
		ON ed.event_id = e.id
		WHERE e.sc_address = $1 AND e.name = $2`,
		input.SmartContractAddress, input.EventName,
	)
	if err != nil {
		return 0, errors.Wrap(err, "query: EventDataQuerier.SelectCountEventDataQuery tx.Get error")
	}

	return count, nil
}
