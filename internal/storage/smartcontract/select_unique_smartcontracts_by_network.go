package smartcontractstorage

import (
	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/pkg/errors"
)

func (s *Storage) ListUniqueSmartContractsByNetwork() ([]*smartcontract.SmartContract, error) {
	// define smartcontracts response
	smartcontracts := []*smartcontract.SmartContract{}

	// get unique smartcontracts by network from db
	/* @dev: It creates a sub table with a partition with only address and network fields.
	 * This partition makes a counter for each row of smart contracts that has the same address
	 * and network. Then from that partition we only get the first row, ensuring that we are not
	 * getting any smart contract with this repeated info using the row number counter.
	 */
	err := s.storage.DB.Select(
		&smartcontracts, `
			SELECT id, name, network, node_url, address, last_tx_block_synced, status, error, created_at, updated_At
			FROM (
				SELECT *, ROW_NUMBER() OVER (PARTITION BY address, network) AS rn
				FROM smartcontracts
			) AS sq
			WHERE sq.rn = 1;`,
	)
	if err != nil {
		return nil, errors.Wrap(err, "smartcontractstorage: Storage.ListUniqueSmartContractsByNetwork s.storage.DB.Select error")
	}

	return smartcontracts, nil
}
