package query

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

func (sq *SmartContractQuerier) SelectSmartContractByAddressQuery(tx storage.Transaction, address string) (*storage.SmartContractRecord, error) {
	var record storage.SmartContractRecord
	err := tx.Get(&record, `
		SELECT * FROM smartcontracts WHERE address = $1;`,
		address,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query: SmartcontractQuerier.SelectSmartContractByAddressQuery tx.Get error")
	}

	return &record, nil
}
