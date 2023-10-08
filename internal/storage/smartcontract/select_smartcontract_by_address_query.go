package smartcontractstorage

import (
	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/pkg/errors"
)

func (s *Storage) GetSmartContractByAddress(address string) (*smartcontract.SmartContract, error) {
	// get smartcontract from db
	sc := &smartcontract.SmartContract{}
	err := s.storage.DB.Get(sc, "SELECT * FROM smartcontracts WHERE address = $1", address)
	if err != nil {
		return nil, errors.Wrap(err, "smartcontracts: Storage.GetSmartContractByAddress s.storage.DB.Get error")
	}

	return sc, nil
}
