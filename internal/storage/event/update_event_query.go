package eventstorage

import (
	"github.com/darchlabs/synchronizer-v2/pkg/event"
	"github.com/pkg/errors"
)

// TODO:
// 1. Add user_id to the query
// 2. Consider use a specific structure with pointers and leverage COALESCE sql method
func (s *Storage) UpdateEvent(e *event.Event) error {
	// update event on db
	_, err := s.storage.DB.Exec(`
		UPDATE event
		SET network = $1,
			node_url = $2,
			address = $3,
			latest_block_number = $4,
			abi_id = $5,
			status = $6,
			error = $7,
			updated_at = $8
		WHERE id = $9;`,
		e.Network,
		e.NodeURL,
		e.Address,
		e.LatestBlockNumber,
		e.AbiID,
		e.Status,
		e.Error,
		e.UpdatedAt,
		e.ID,
	)
	if err != nil {
		return errors.Wrap(err, "eventstorage: Storage.UpdateEvent s.storage.DB.Exec error")
	}

	return nil
}
