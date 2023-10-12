package query

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

func (q *SmartContractQuerier) SelectCountUserSmartContractsQuery(db storage.Database, userID string) (int64, error) {
	var count int64

	err := db.Get(
		&count, `
		SELECT COUNT(sc.id)
		FROM smartcontracts sc
		JOIN smartcontract_users scu
		ON sc.address = scu.sc_address
		WHERE scu.user_id = $1`,
		userID,
	)
	if err != nil {
		return 0, errors.Wrap(err, "query: SmartContractQuerier.SelectCountUserSmartContractsQuery db.Get error")
	}

	return count, nil
}
