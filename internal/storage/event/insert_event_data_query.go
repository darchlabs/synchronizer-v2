package eventstorage

import (
	"github.com/darchlabs/synchronizer-v2/pkg/event"
	"github.com/pkg/errors"
)

func (s *Storage) InsertEventData(e *event.Event, data []*event.EventData) error {
	// prepare transaction
	tx, err := s.storage.DB.Beginx()
	if err != nil {
		return err
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

	// insert event data in db
	batch, err := tx.Preparex(`
		INSERT INTO event_data (id, event_id, tx, block_number, data, created_at)
		VALUES ($1, $2, $3, $4, $5, $6);`)
	if err != nil {
		return errors.Wrap(err, "eventstorage: Storage.InsertEventData batch.Exec error")
	}
	defer batch.Close()

	// iterate over logsData array for inserting on db
	for _, ed := range data {
		// execute que batch into the db
		_, err = batch.Exec(ed.ID, e.ID, ed.Tx, ed.BlockNumber, ed.Data, ed.CreatedAt)
		if err != nil {
			return errors.Wrap(err, "eventstorage: Storage.InsertEventData batch.Exec error")
		}
	}

	// send transaction to db
	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "eventstorage: Storage.InsertEventData tx.Commit error")
	}

	return nil
}
