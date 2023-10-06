package query

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

func (sq *SmartContractQuerier) SelectSmartContractsByAddressesList(tx storage.Transaction, addresses []string) ([]*storage.SmartContractRecord, error) {
	records := make([]*storage.SmartContractRecord, 0)
	if len(addresses) == 0 {
		return records, nil
	}

	// TODO: improve this
	queryMade := make(map[string]struct{})
	for _, addr := range addresses {
		if _, ok := queryMade[addr]; !ok {
			var r storage.SmartContractRecord
			err := tx.Get(&r, `
		SELECT *
		FROM smartcontracts
		WHERE address IN ($1);`,
				addr,
			)
			if err != nil {
				return nil, errors.Wrap(err, "sync: SmartContractQuerier.SelectSmartContractsByAddressesList tx.Select")
			}
			records = append(records, &r)
			queryMade[addr] = struct{}{}
		}
	}

	return records, nil
}
