package smartcontractstorage

import (
	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/pkg/errors"
)

func (s *Storage) GetSmartContractById(id string) (*smartcontract.SmartContract, error) {
	// get smartcontract from db
	sc := &smartcontract.SmartContract{}
	err := s.storage.DB.Get(sc, "SELECT * FROM smartcontracts WHERE id = $1", id)
	if err != nil {
		return nil, errors.Wrap(err, "smartcontractstorage: Storage.GetSmartContractById s.storage.DB.Get")
	}

	return sc, nil
}
