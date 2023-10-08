package eventstorage

import "github.com/pkg/errors"

func (s *Storage) DeleteEvent(address string, eventName string) error {
	// get event using address and eventName
	e, err := s.GetEvent(address, eventName)
	if err != nil {
		return errors.Wrapf(ErrEventNotFound, "eventstorage: Storage.DeleteEvent address=%s event_name=%s", address, eventName)
	}

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
				err = errors.WithMessagef(txErr, "eventstorage: Storage.DeleteEvent rollback transaction error: %s", err.Error())
			}
		}
	}()

	// delete event data
	eventDataQuery := "DELETE FROM event_data WHERE event_id = $1"
	_, err = tx.Exec(eventDataQuery, e.ID)
	if err != nil {
		return errors.Wrap(err, "eventstorage: Storage.DeleteEvent event_data tx.Exec error")
	}

	// delete event from db
	eventQuery := "DELETE FROM event WHERE id = $1"
	_, err = tx.Exec(eventQuery, e.ID)
	if err != nil {
		return errors.Wrap(err, "eventstorage: Storage.DeleteEvent event tx.Exec error")
	}

	// delete inputs from db
	inputQuery := "DELETE FROM input WHERE abi_id = $1"
	_, err = tx.Exec(inputQuery, e.AbiID)
	if err != nil {
		return errors.Wrap(err, "eventstorage: Storage.DeleteEvent input tx.Exec error")
	}

	// delete abi from db
	abiQuery := "DELETE FROM abi WHERE ID = $1"
	_, err = tx.Exec(abiQuery, e.AbiID)
	if err != nil {
		return errors.Wrap(err, "eventstorage: Storage.DeleteEvent abi tx.Exec error")
	}

	// commit transaction
	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "eventstorage: Storage.DeleteEvent tx.Commit error")
	}

	return nil
}
