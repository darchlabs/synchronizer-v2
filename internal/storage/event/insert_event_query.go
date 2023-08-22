package eventstorage

import (
	"github.com/darchlabs/synchronizer-v2/pkg/event"
	"github.com/pkg/errors"
)

func (s *Storage) InsertEvent(e *event.Event) (_ *event.Event, err error) {
	// check if event already exist and is using the same address and name
	ev, _ := s.GetEvent(e.Address, e.Abi.Name)
	if ev != nil {
		return nil, ErrEventAlreadyExist
	}

	// prepare transaction to create event on db
	tx, err := s.storage.DB.Beginx()
	if err != nil {
		return nil, errors.Wrap(err, "eventstorage: Storage.InsertEvent s.storage.DB.Beginx")
	}

	// rollback into the defer func avoids duplicate code if any error
	defer func() {
		if err != nil && tx != nil {
			txErr := tx.Rollback()
			if txErr != nil {
				err = errors.WithMessagef(txErr, "eventstorage: Storage.InsertEvent rollback transaction error: %s", err.Error())
			}
		}
	}()

	// inser abi to use in db
	// TODO: this should be into its own method
	var abiID string
	err = tx.Get(&abiID, `
		INSERT INTO abi (id, name, type, anonymous)
		VALUES ($1, $2, $3, $4)
		RETURNING id;`,
		e.Abi.ID,
		e.Abi.Name,
		e.Abi.Type,
		e.Abi.Anonymous,
	)
	if err != nil {
		return nil, errors.Wrap(err, "eventstorage: Storage.InsertEvent abi tx.Get error")
	}

	// iterate over inputs for inserting on db
	for _, input := range e.Abi.Inputs {
		// TODO: this should be into its own method
		_, err = tx.Exec(`
			Insert INTO input (id, indexed, internal_type, name, type, abi_id)
			VALUES ($1, $2, $3, $4, $5, $6);`,
			input.ID,
			input.Indexed,
			input.InternalType,
			input.Name,
			input.Type,
			abiID,
		)
		if err != nil {
			return nil, errors.Wrap(err, "eventstorage: Storage.InsertEvent input tx.Exec error")
		}
	}

	// insert new event in database
	var eventID string
	// TODO: this should be into its own method
	err = tx.Get(&eventID, `
		INSERT INTO event (
			id,
			network,
			node_url,
			address,
			latest_block_number,
			abi_id,
			status,
			error,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id;`,
		e.ID,
		e.Network,
		e.NodeURL,
		e.Address,
		e.LatestBlockNumber,
		abiID,
		e.Status,
		e.Error,
		e.CreatedAt,
		e.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "eventstorage: Storage.InsertEvent event tx.Get error")
	}

	// commit transaction
	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "eventstorage: Storage.InsertEvent tx.Commit error")
	}

	// get created event
	// TODO: drop this. since the event inserte is already alocated into a variable
	// there is no need to get the event back and load the database with one more query
	createdEvent, err := s.GetEventById(eventID)
	if err != nil {
		return nil, errors.Wrap(err, "eventstorage: Storage.InsertEvent tx.Commit error")
	}

	return createdEvent, nil
}
