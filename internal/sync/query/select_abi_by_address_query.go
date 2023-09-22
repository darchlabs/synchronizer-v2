package query

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

func (aq *ABIQuerier) SelectABIByAddressQuery(tx storage.Transaction, address string) (*storage.ABIRecord, error) {
	var record storage.ABIRecord
	err := tx.Get(&record, `
		SELECT * FROM abi WHERE sc_address = $1;`,
		address,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query: SmartcontractQuerier.SelectSmartContractByAddressQuery tx.Get error")
	}

	return &record, nil
}
