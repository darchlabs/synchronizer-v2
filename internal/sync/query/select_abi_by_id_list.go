package query

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

func (aq *ABIQuerier) SelectABIByIDs(tx storage.Transaction, ids []string) ([]*storage.ABIRecord, error) {
	records := make([]*storage.ABIRecord, 0)
	if len(ids) == 0 {
		return records, nil
	}

	for _, id := range ids {
		query := "SELECT * FROM abi WHERE id = $1;"
		var r storage.ABIRecord
		err := tx.Get(&r, query, id)
		if err != nil {
			return nil, errors.Wrap(err, "query: ABIQuerier.SelectABIByIDs tx.Get error")
		}

		records = append(records, &r)
	}

	return records, nil
}
