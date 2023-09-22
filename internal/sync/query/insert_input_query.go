package query

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

func (iq *InputQuerier) InsertInputQuery(qCtx storage.QueryContext, input *storage.InputRecord) error {
	_, err := qCtx.Exec(`
			Insert INTO input (id, indexed, internal_type, name, type, abi_id)
			VALUES ($1, $2, $3, $4, $5, $6);`,
		input.ID,
		input.Indexed,
		input.InternalType,
		input.Name,
		input.Type,
		input.AbiID,
	)
	if err != nil {
		return errors.Wrap(err, "eventstorage: Storage.InsertEvent input tx.Exec error")
	}

	return nil
}
