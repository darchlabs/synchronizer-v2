package query

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

type SelectSmartContractQueryOutput struct {
	ID                 string                          `db:"id"`
	Network            storage.EventNetwork            `db:"network"`
	Address            string                          `db:"address"`
	LastTxBlockSynced  int64                           `db:"last_tx_block_synced"`
	InitialBlockNumber int64                           `db:"initial_block_number"`
	Status             storage.SmartContractUserStatus `db:"status"`
}

func (sq *SmartContractQuerier) SelectSmartContractsQuery(
	tx storage.Transaction,
) ([]*SelectSmartContractQueryOutput, error) {
	output := make([]*SelectSmartContractQueryOutput, 0)

	err := tx.Select(&output, `
		SELECT 
			sc.id AS id,
			sc.network AS network,
			sc.address AS address,
			sc.last_tx_block_synced AS last_tx_block_synced,
			sc.initial_block_number AS initial_block_number,
			scu.status AS status
		FROM 
			smartcontracts sc,
			(SELECT DISTINCT(sc_address) AS address, status AS status FROM smartcontract_users) scu
		WHERE scu.address = sc.address
	);`)
	if err != nil {
		return nil, errors.Wrap(err, "query: SmartContractQuerier.SelectSmartContractsQuery tx.Select error")
	}

	return output, nil
}
