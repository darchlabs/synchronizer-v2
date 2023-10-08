package query

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

func (sc *SmartContractQuerier) InsertSmartContractQuery(
	qCtx storage.QueryContext,
	input *storage.SmartContractRecord,
) error {
	// date validation
	if input.CreatedAt.IsZero() {
		return ErrInvalidDate
	}

	// insert new smartcontract in database
	_, err := qCtx.Exec(`
		INSERT INTO smartcontracts (
			id,
			network,
			address,
			last_tx_block_synced,
			created_at
		) VALUES ($1, $2, $3, $4, $5);`,
		input.ID,
		input.Network,
		input.Address,
		input.LastTxBlockSynced,
		input.CreatedAt,
	)
	if err != nil {
		return errors.Wrap(err, "query: SmartContractQuerier.InsertSmartContractQuery tx.Exec error")
	}

	return nil

}
