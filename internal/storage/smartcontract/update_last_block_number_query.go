package smartcontractstorage

import (
	"time"

	"github.com/pkg/errors"
)

func (s *Storage) UpdateLastBlockNumber(id string, blockNumber int64) error {
	// get current sc
	current, _ := s.GetSmartContractById(id)
	if current == nil {
		return ErrSmartcontractNotFound
	}

	// insert new smartcontract in database
	_, err := s.storage.DB.Exec(`
		UPDATE smartcontracts
		SET last_tx_block_synced = $1,
				updated_at = $2
		WHERE id = $3
		RETURNING *;`,
		blockNumber,
		time.Now(),
		id,
	)
	if err != nil {
		return errors.Wrap(err, "smartcontractstorage: Storage.UpdateLastBlockNumber s.storage.DB.Exec")
	}

	return nil
}
