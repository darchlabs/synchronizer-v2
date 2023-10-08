package query

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

func (eq *EventQuerier) SelectEventsByAddressQuery(tx storage.Transaction, address string) ([]*storage.EventRecord, error) {
	records := make([]*storage.EventRecord, 0)
	err := tx.Select(&records, `
		SELECT * FROM event WHERE address = $1;`,
		address,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query: EventQuerier.SelectEventsByAddressQuery tx.Select error")
	}

	return records, nil
}
