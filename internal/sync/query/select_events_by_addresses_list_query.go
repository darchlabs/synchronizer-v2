package query

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

func (eq *EventQuerier) SelectEventsByAddressesListQuery(tx storage.Transaction, addresses []string) ([]*storage.EventRecord, error) {
	records := make([]*storage.EventRecord, 0)
	query := `
		SELECT * FROM event WHERE address = ANY($1::text[]);`
	err := tx.Select(&records, query, pq.Array(addresses))
	if err != nil {
		return nil, errors.Wrap(err, "query: EventQuerier.SelectEventsByAddressesListQuery tx.Select error")
	}

	return records, nil
}
