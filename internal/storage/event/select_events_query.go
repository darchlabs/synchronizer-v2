package eventstorage

import (
	"fmt"

	"github.com/darchlabs/synchronizer-v2/pkg/event"
	"github.com/pkg/errors"
)

func (s *Storage) ListAllEvents() ([]*event.Event, error) {
	// define events response
	events := []*event.Event{}

	// get events from db
	err := s.storage.DB.Select(&events, "SELECT * FROM event")
	if err != nil {
		return nil, errors.Wrap(err, "eventstorage: Storage.ListAllEvents event s.storage.DB.Select error")
	}

	// iterate over events for getting abi and input values
	for _, e := range events {
		// query for getting event abi
		var abi event.Abi
		err = s.storage.DB.Get(&abi, "SELECT * FROM abi WHERE ID = $1", e.AbiID)
		if err != nil {
			return nil, errors.Wrap(err, "eventstorage: Storage.ListAllEvents abi s.storage.DB.Get error")
		}
		e.Abi = &abi

		// query for getting event abi inputs
		inputs := []*event.Input{}
		err = s.storage.DB.Select(&inputs, "SELECT * FROM input WHERE abi_id = $1", abi.ID)
		if err != nil {
			return nil, errors.Wrap(err, "eventstorage: Storage.ListAllEvents input s.storage.DB.Select error")
		}
		e.Abi.Inputs = inputs
	}

	return events, nil
}

func (s *Storage) ListEvents(sort string, limit int64, offset int64) ([]*event.Event, error) {
	// define events response
	events := []*event.Event{}

	// get events from db
	err := s.storage.DB.Select(
		&events,
		fmt.Sprintf("SELECT * FROM event ORDER BY created_at %s LIMIT $1 OFFSET $2;", sort),
		limit,
		offset,
	)
	if err != nil {
		return nil, errors.Wrap(err, "eventstorage: Storage.ListEvents event s.storage.DB.Select error")
	}

	// iterate over events for getting abi and input values
	for _, e := range events {
		// query for getting event abi
		abi := &event.Abi{}
		err = s.storage.DB.Get(abi, "SELECT * FROM abi WHERE ID = $1", e.AbiID)
		if err != nil {
			return nil, errors.Wrap(err, "eventstorage: Storage.ListEvents abi s.storage.DB.Get error")
		}
		e.Abi = abi

		// query for getting event abi inputs
		inputs := []*event.Input{}
		inputsQuery := "SELECT * FROM input WHERE abi_id = $1"
		err = s.storage.DB.Select(&inputs, inputsQuery, abi.ID)
		if err != nil {
			return nil, errors.Wrap(err, "eventstorage: Storage.ListEvents input s.storage.DB.Select error")
		}
		e.Abi.Inputs = inputs
	}

	return events, nil
}
