package query

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

func (iq *InputQuerier) SelectInputByABIIDQuery(tx storage.Transaction, abiID string) ([]*storage.InputRecord, error) {
	records := make([]*storage.InputRecord, 0)
	err := tx.Select(&records, `
		SELECT * FROM input WHERE abi_id = $1;`,
		abiID,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query: InputQuerier.SelectInputByABIIDQuery tx.Select error")
	}

	return records, nil
}
