package query

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

func (aq *ABIQuerier) InsertABIBatchQuery(qCtx storage.QueryContext, inputs []*storage.ABIRecord, scAddress string) error {
	for _, input := range inputs {
		if input.ID == "" {
			input.ID = aq.idGen()
		}
		input.SmartContractAddress = scAddress
		err := aq.InsertABIQuery(qCtx, input)
		if err != nil {
			return errors.Wrap(err, "query ABIQuerier.InsertABIQuery abi tx.Exec error")
		}
	}

	return nil
}
