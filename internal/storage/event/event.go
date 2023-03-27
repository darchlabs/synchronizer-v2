package eventstorage

import (
	"encoding/json"
	"fmt"

	"github.com/darchlabs/synchronizer-v2/internal/blockchain"
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/darchlabs/synchronizer-v2/pkg/event"
)

type Storage struct {
	storage *storage.S
}

func New(s *storage.S) *Storage {
	return &Storage{
		storage: s,
	}
}

func (s *Storage) InsertEvent(e *event.Event) (*event.Event, error) {
	// check if already existe an event with the same address and name
	ev, _ := s.GetEvent(e.Address, e.Abi.Name)
	if ev != nil {
		return nil, fmt.Errorf("event already exists with address=%s and eventName=%s", e.Address, e.Abi.Name)
	}

	// prepare transaction to create event on db
	tx, err := s.storage.DB.Beginx()
	if err != nil {
		return nil, err
	}

	// inser abi to use in db
	var abiID int64
	abiQuery := "INSERT INTO abi (name, type, anonymous) VALUES ($1, $2, $3) RETURNING id"
	err = tx.Get(&abiID, abiQuery, e.Abi.Name, e.Abi.Type, e.Abi.Anonymous)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// iterate over inputs for inserting on db
	for _, input := range e.Abi.Inputs {
		inputQuery := "INSERT INTO input (indexed, internal_type, name, type, abi_id) VALUES ($1, $2, $3, $4, $5)"
		_, err = tx.Exec(inputQuery, input.Indexed, input.InternalType, input.Name, input.Type, abiID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// set base/default values
	// TODO(ca): should use the creation block number of the contract
	e.LatestBlockNumber = 0
	e.Status = event.StatusSynching
	e.Error = ""

	// insert new event in database
	var eventID int64
	eventQuery := "INSERT INTO event (network, node_url, address, latest_block_number, abi_id, status, error) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id"
	err = tx.Get(&eventID, eventQuery, e.Network, e.NodeURL, e.Address, e.LatestBlockNumber, abiID, e.Status, e.Error)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// commit transaction
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// get created event
	createdEvent, err := s.GetEventByID(eventID)
	if err != nil {
		return nil, err
	}

	return createdEvent, nil
}

func (s *Storage) UpdateEvent(e *event.Event) error {
	// prepare transaction
	tx, err := s.storage.DB.Beginx()
	if err != nil {
		return nil
	}

	// update event on db
	query := "UPDATE event SET network = $1, node_url = $2, address = $3, latest_block_number = $4, abi_id = $5, status = $6, error = $7, updated_at = NOW() WHERE id = $8"
	_, err = tx.Exec(query, e.Network, e.NodeURL, e.Address, e.LatestBlockNumber, e.AbiID, e.Status, e.Error, e.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// send transaction to db
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) ListEvents() ([]*event.Event, error) {
	// define events response
	events := []*event.Event{}

	// get events from db
	eventQuery := "SELECT * FROM event"
	err := s.storage.DB.Select(&events, eventQuery)
	if err != nil {
		return nil, err
	}

	// iterate over events for getting abi and input values
	for _, e := range events {
		// query for getting event abi
		abi := &event.Abi{}
		abiQuery := "SELECT * FROM abi WHERE ID = $1"
		err = s.storage.DB.Get(abi, abiQuery, e.AbiID)
		if err != nil {
			return nil, err
		}
		e.Abi = abi

		// query for getting event abi inputs
		inputs := []*event.Input{}
		inputsQuery := "SELECT * FROM input WHERE abi_id = $1"
		err = s.storage.DB.Select(&inputs, inputsQuery, abi.ID)
		if err != nil {
			return nil, err
		}
		e.Abi.Inputs = inputs
	}

	return events, nil
}

func (s *Storage) ListEventsByAddress(address string) ([]*event.Event, error) {
	// define events response
	events := []*event.Event{}

	// get events from db
	eventQuery := "SELECT * FROM event WHERE address = $1"
	err := s.storage.DB.Select(&events, eventQuery, address)
	if err != nil {
		return nil, err
	}

	// iterate over events for getting abi and input values
	for _, e := range events {
		// query for getting event abi
		abi := &event.Abi{}
		abiQuery := "SELECT * FROM abi WHERE ID = $1"
		err = s.storage.DB.Get(abi, abiQuery, e.AbiID)
		if err != nil {
			return nil, err
		}
		e.Abi = abi

		// query for getting event abi inputs
		inputs := []*event.Input{}
		inputsQuery := "SELECT * FROM input WHERE abi_id = $1"
		err = s.storage.DB.Select(&inputs, inputsQuery, abi.ID)
		if err != nil {
			return nil, err
		}
		e.Abi.Inputs = inputs
	}

	return events, nil
}

func (s *Storage) GetEvent(address string, eventName string) (*event.Event, error) {
	// get event from db
	e := &event.Event{}
	err := s.storage.DB.Get(e, "SELECT event.* FROM event INNER JOIN abi ON event.abi_id = abi.id WHERE event.address = $1 AND abi.name = $2", address, eventName)
	if err != nil {
		return nil, err
	}

	// get event abi from db
	abi := &event.Abi{}
	err = s.storage.DB.Get(abi, "SELECT * FROM abi WHERE ID = $1", e.AbiID)
	if err != nil {
		return nil, err
	}
	e.Abi = abi

	// get event abi inputs from db
	inputs := []*event.Input{}
	err = s.storage.DB.Select(&inputs, "SELECT * FROM input WHERE abi_id = $1", e.AbiID)
	if err != nil {
		return nil, err
	}
	e.Abi.Inputs = inputs

	return e, nil
}

func (s *Storage) GetEventByID(id int64) (*event.Event, error) {
	// get event from db
	e := &event.Event{}
	err := s.storage.DB.Get(e, "SELECT * FROM event WHERE id = $1", id)
	if err != nil {
		return nil, err
	}

	// get event abi from db
	abi := &event.Abi{}
	err = s.storage.DB.Get(abi, "SELECT * FROM abi WHERE ID = $1", e.AbiID)
	if err != nil {
		return nil, err
	}
	e.Abi = abi

	// get event abi inputs from db
	inputs := []*event.Input{}
	err = s.storage.DB.Select(&inputs, "SELECT * FROM input WHERE abi_id = $1", e.AbiID)
	if err != nil {
		return nil, err
	}
	e.Abi.Inputs = inputs

	return e, nil
}

func (s *Storage) DeleteEvent(address string, eventName string) error {
	// get event using address and eventName
	e, err := s.GetEvent(address, eventName)
	if err != nil {
		return fmt.Errorf("event does not exist with address=%s event_name=%s", address, eventName)
	}

	// prepare transaction
	tx, err := s.storage.DB.Beginx()
	if err != nil {
		return err
	}

	// delete event data
	eventDataQuery := "DELETE FROM event_data WHERE event_id = $1"
	_, err = tx.Exec(eventDataQuery, e.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// delete event from db
	eventQuery := "DELETE FROM event WHERE id = $1"
	_, err = tx.Exec(eventQuery, e.ID)
	if err != nil {
		return err
	}

	// delete inputs from db
	inputQuery := "DELETE FROM input WHERE abi_id = $1"
	_, err = tx.Exec(inputQuery, e.AbiID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// delete abi from db
	abiQuery := "DELETE FROM abi WHERE ID = $1"
	_, err = tx.Exec(abiQuery, e.AbiID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// send transaction to db
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) ListEventData(address string, eventName string) ([]*event.EventData, error) {
	// define events data response
	eventsData := []*event.EventData{}

	// define and make the query on db
	eventsDataQuery := "SELECT event_data.* FROM event_data JOIN event ON event_data.event_id = event.id JOIN abi ON event.abi_id = abi.id WHERE event.address = $1 AND abi.name = $2"
	err := s.storage.DB.Select(&eventsData, eventsDataQuery, address, eventName)
	if err != nil {
		return nil, err
	}

	return eventsData, nil
}

func (s *Storage) InsertEventData(e *event.Event, data []blockchain.LogData) error {
	// prepare transaction
	tx, err := s.storage.DB.Beginx()
	if err != nil {
		return err
	}

	// insert event data in db
	eventDataQuery := "INSERT INTO event_data (event_id, tx, block_number, data, created_at) VALUES ($1, $2, $3, $4, NOW())"
	batch, err := tx.Preparex(eventDataQuery)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer batch.Close()

	// iterate over logsData array for inserting on db
	for _, logData := range data {
		// convert the map to json
		data, err := json.Marshal(logData.Data)
		if err != nil {
			tx.Rollback()
			return err
		}

		// execute que batch into the db
		_, err = batch.Exec(e.ID, logData.Tx.String(), logData.BlockNumber, data)
		if err != nil {
			tx.Rollback()

			return err
		}
	}

	// send transaction to db
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) Stop() error {
	err := s.storage.DB.Close()
	if err != nil {
		return err
	}

	return nil
}
