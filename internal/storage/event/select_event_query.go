package eventstorage

import (
	"github.com/darchlabs/synchronizer-v2/pkg/event"
	"github.com/pkg/errors"
)

func (s *Storage) GetEvent(address string, eventName string) (*event.Event, error) {
	// get event from db
	e := &event.Event{}
	err := s.storage.DB.Get(e, `
		SELECT event.*
		FROM event
		INNER JOIN abi
		ON event.abi_id = abi.id
		WHERE event.address = $1
		AND abi.name = $2;`,
		address,
		eventName,
	)
	if err != nil {
		return nil, errors.Wrap(err, "eventstorage: Storage.GetEvent event s.storage.DB.Get error")
	}

	// get event abi from db
	abi := &event.Abi{}
	err = s.storage.DB.Get(abi, "SELECT * FROM abi WHERE ID = $1", e.AbiID)
	if err != nil {
		return nil, errors.Wrap(err, "eventstorage: Storage.GetEvent abi s.storage.DB.Get error")
	}
	e.Abi = abi

	// get event abi inputs from db
	inputs := []*event.Input{}
	err = s.storage.DB.Select(&inputs, "SELECT * FROM input WHERE abi_id = $1", e.AbiID)
	if err != nil {
		return nil, errors.Wrap(err, "eventstorage: Storage.GetEvent input s.storage.DB.Select error")
	}
	e.Abi.Inputs = inputs

	return e, nil
}
