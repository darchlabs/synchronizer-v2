package eventstorage

import (
	"fmt"

	"github.com/darchlabs/synchronizer-v2/pkg/event"
	"github.com/pkg/errors"
)

func (s *Storage) ListEventsByAddress(address string, sort string, limit int64, offset int64) ([]*event.Event, error) {
	// define events response
	events := []*event.Event{}

	// get events from db
	eventQuery := fmt.Sprintf("SELECT * FROM event WHERE address = $1 ORDER BY created_at %s LIMIT $2 OFFSET $3", sort)
	err := s.storage.DB.Select(
		&events,
		eventQuery,
		address,
		limit,
		offset,
	)
	if err != nil {
		return nil, errors.Wrap(err, "eventstorage: Stora.ListEventsByAddress events s.storage.DB.Select error")
	}

	// iterate over events for getting abi and input values
	for _, e := range events {
		// query for getting event abi
		abi := &event.Abi{}
		abiQuery := "SELECT * FROM abi WHERE ID = $1"
		err = s.storage.DB.Get(abi, abiQuery, e.AbiID)
		if err != nil {
			return nil, errors.Wrap(err, "eventstorage: Stora.ListEventsByAddress abi s.storage.DB.Get error")
		}
		e.Abi = abi

		// query for getting event abi inputs
		inputs := []*event.Input{}
		err = s.storage.DB.Select(&inputs, "SELECT * FROM input WHERE abi_id = $1", abi.ID)
		if err != nil {
			return nil, errors.Wrap(err, "eventstorage: Stora.ListEventsByAddress input s.storage.DB.Get error")
		}
		e.Abi.Inputs = inputs
	}

	return events, nil
}
