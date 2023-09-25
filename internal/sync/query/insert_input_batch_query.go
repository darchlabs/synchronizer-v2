package query

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

func (iq *InputQuerier) InsertInputBatchQuery(
	qCtx storage.QueryContext,
	inputs []*storage.InputRecord,
	abiID string,
) error {

	for _, input := range inputs {
		// TODO: this should be into its own method
		input.ID = iq.idGen()
		input.SmartContractAddress = abiID
		// input.CreatedAt = iq.dateGen()
		err := iq.InsertInputQuery(qCtx, input)

		if err != nil {
			return errors.Wrap(err, "eventstorage: Storage.InsertEvent input tx.Exec error")
		}
	}

	return nil
}
