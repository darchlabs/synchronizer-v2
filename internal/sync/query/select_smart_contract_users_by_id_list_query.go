package query

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

func (sq *SmartContractUserQuerier) SmartContractUsersByIDListQuery(
	tx storage.Transaction,
	addresses []string,
) ([]*storage.SmartContractUserRecord, error) {
	records := make([]*storage.SmartContractUserRecord, 0)
	if len(addresses) == 0 {
		return records, nil
	}

	// TODO: improve this
	queryMade := make(map[string]struct{})
	for _, addr := range addresses {
		if _, ok := queryMade[addr]; !ok {
			var r storage.SmartContractUserRecord
			err := tx.Get(&r, `
		SELECT *
		FROM smartcontract_users
		WHERE sc_address = $1;`,
				addr,
			)
			if err != nil {
				return nil, errors.Wrap(err, "query: SmartContractUserQuerier.SmartContractUsersByIDListQuery tx.Select error")
			}
			records = append(records, &r)
			queryMade[addr] = struct{}{}
		}
	}

	return records, nil
}
