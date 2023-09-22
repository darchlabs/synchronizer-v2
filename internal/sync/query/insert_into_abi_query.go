package query

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

func (aq *ABIQuerier) InsertABIQuery(qCtx storage.QueryContext, input *storage.ABIRecord) error {
	_, err := qCtx.Exec(`
		INSERT INTO abi (id, sc_address, name, type, anonymous)
		VALUES ($1, $2, $3, $4, $5);`,
		input.ID,
		input.SmartContractAddress,
		input.Name,
		input.Type,
		input.Anonymous,
	)
	if err != nil {
		return errors.Wrap(err, "query ABIQuerier.InsertABIQuery abi tx.Exec error")
	}

	return nil
}
