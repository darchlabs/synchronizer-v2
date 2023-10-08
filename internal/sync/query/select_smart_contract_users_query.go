package query

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

func (sq *SmartContractUserQuerier) SelectSmartContractUserQuery(tx storage.Transaction, address string) ([]*storage.SmartContractUserRecord, error) {
	records := make([]*storage.SmartContractUserRecord, 0)
	err := tx.Select(&records, `
		SELECT *
		FROM smartcontract_users
		WHERE sc_address = $1;`,
		address,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query: SmartContractUserQuerier.SelectSmartContractUserQuery")
	}

	return records, nil
}
