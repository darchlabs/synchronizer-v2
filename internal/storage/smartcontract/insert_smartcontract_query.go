package smartcontractstorage

import (
	"fmt"

	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/pkg/errors"
)

func (s *Storage) InsertSmartContract(sc *smartcontract.SmartContract) (*smartcontract.SmartContract, error) {
	// get current sc
	current, _ := s.GetSmartContractByAddress(sc.Address)
	if current != nil {
		return nil, fmt.Errorf("smartcontract already exists with address=%s", sc.Address)
	}

	// insert new smartcontract in database
	var smartcontractId string
	err := s.storage.DB.Get(&smartcontractId, `
		INSERT INTO smartcontracts (
			id,
			name,
			network,
			node_url,
			address,last_tx_block_synced,
			status,
			error,
			created_at,
			updated_at,
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id;`,
		sc.ID,
		sc.Name,
		sc.Network,
		sc.NodeURL,
		sc.Address,
		sc.LastTxBlockSynced,
		sc.Status,
		sc.Error,
		sc.CreatedAt,
		sc.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "SmartContractstorage: Storage.InsertSmartContract s.storage.DB.Get error")
	}

	// get created smartcontract
	createdSmartcontract, err := s.GetSmartContractById(smartcontractId)
	if err != nil {
		return nil, errors.Wrap(err, "SmartContractstorage: Storage.InsertSmartContract s.GetSmartContractById error")

	}

	return createdSmartcontract, nil
}
