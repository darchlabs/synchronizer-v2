package query

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

func (eq *EventDataQuerier) InsertEventDataQuery(qCtx storage.QueryContext, record *storage.EventDataRecord) error {
	_, err := qCtx.Exec(`
		INSERT INTO event_data (id, event_id, tx, block_number, data, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT(tx) DO NOTHING;`,
		record.ID,
		record.EventID,
		record.Tx,
		record.BlockNumber,
		record.Data,
		record.CreatedAt,
	)
	if err != nil {
		return errors.Wrap(err, "query: EventDataQuerier.tnsertEventDataQuery qCtx.Exec error")
	}
	return nil
}
