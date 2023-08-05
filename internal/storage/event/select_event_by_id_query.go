package eventstorage

import (
	"github.com/darchlabs/synchronizer-v2/pkg/event"
	"github.com/pkg/errors"
)

func (s *Storage) GetEventById(id string) (*event.Event, error) {
	// get event from db
	e := &event.Event{}
	err := s.storage.DB.Get(e, "SELECT * FROM event WHERE id = $1", id)
	if err != nil {
		return nil, errors.Wrap(err, "eventstorage: Storage.GetEventById event s.storage.DB.Get error")
	}

	// get event abi from db
	abi := &event.Abi{}
	err = s.storage.DB.Get(abi, "SELECT * FROM abi WHERE ID = $1", e.AbiID)
	if err != nil {
		return nil, errors.Wrap(err, "eventstorage: Storage.GetEventById abi s.storage.DB.Get error")
	}
	e.Abi = abi

	// get event abi inputs from db
	inputs := []*event.Input{}
	err = s.storage.DB.Select(&inputs, "SELECT * FROM input WHERE abi_id = $1", e.AbiID)
	if err != nil {
		return nil, errors.Wrap(err, "eventstorage: Storage.GetEventById abi s.storage.DB.Get error")
	}
	e.Abi.Inputs = inputs

	return e, nil
}
