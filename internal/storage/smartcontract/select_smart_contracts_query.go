package smartcontractstorage

import (
	"fmt"

	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/pkg/errors"
)

func (s *Storage) ListSmartContracts(sort string, limit int64, offset int64) ([]*smartcontract.SmartContract, error) {
	// define smartcontracts response
	smartcontracts := []*smartcontract.SmartContract{}

	// get smartcontracts from db
	err := s.storage.DB.Select(
		&smartcontracts,
		fmt.Sprintf("SELECT * FROM smartcontracts ORDER BY created_at %s LIMIT $1 OFFSET $2", sort),
		limit,
		offset,
	)
	if err != nil {
		return nil, errors.Wrap(err, "smartcontractstorage: Storage.ListSmartContracts s.storage.DB.Select error")
	}

	return smartcontracts, nil
}
