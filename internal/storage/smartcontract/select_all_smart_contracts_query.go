package smartcontractstorage

import (
	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/pkg/errors"
)

func (s *Storage) ListAllSmartContracts() ([]*smartcontract.SmartContract, error) {
	// define smartcontracts response
	smartcontracts := []*smartcontract.SmartContract{}

	// get smartcontracts from db
	err := s.storage.DB.Select(&smartcontracts, "SELECT * FROM smartcontracts")
	if err != nil {
		return nil, errors.Wrap(err, "smartcontractstorage: Storage.ListAllSmartContracts s.storage.DB.Select error")
	}

	return smartcontracts, nil
}
