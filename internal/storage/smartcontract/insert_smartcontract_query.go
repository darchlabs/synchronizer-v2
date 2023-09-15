package smartcontractstorage

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/pkg/errors"
)

/*
FOLLOW THE ORDER OF NUMBER. DONT IMPROVISE, PLEASE

BEGIN TX
	1. insert smartcontract
	2. insert smartcontract_user
	4. insert abi
	5. insert input
	3. insert events
COMMIT
*/

func (s *Storage) InsertSmartContractQuery(sc *smartcontract.SmartContract) (err error) {
	tx, err := s.storage.DB.Beginx()
	if err != nil {
		return errors.Wrap(err, "smartcontractstorage: Storage.InsertSmartContractQuery s.storage.DB.Beginx error")
	}

	// rollback into the defer func avoids duplicate code if any error
	defer func() {
		if err != nil && tx != nil {
			txErr := tx.Rollback()
			if txErr != nil {
				err = errors.WithMessagef(txErr, "eventstorage: Storage.InsertEventData rollback transaction error: %s", err.Error())
			}
		}
	}()

	// insert new smartcontract in database
	_, err = tx.Exec(`
		INSERT INTO smartcontracts (
			id,
			name,
			network,
			node_url, -- this should be moved to smartcontract_user
			address,
			last_tx_block_synced,
			status,
			error,
			created_at,
			updated_at,
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (smartcontracts.address) DO NOTHING;`, //TODO: please check this query carefuly
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
		return errors.Wrap(err, "SmartContractstorage: Storage.InsertSmartContract tx.Get error")
	}

	// TODO: Add insert for smartcontract_user record
	err = s.scuserStorage.InsertSmartContractUserQuery(s.storage.DB, &storage.SmartContractUserRecord{
		ID:              s.idGenerator(),
		UserID:          sc.UserID,
		SmartContractID: sc.ID,
	})
	if err != nil {
		return errors.Wrap(err, "smartcontractstorage: Storage.InsertSmartContractQuery s.scuserStorage.InsertSmartContractUserQuery error")
	}

	return nil
}
