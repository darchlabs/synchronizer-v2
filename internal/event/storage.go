package event

import (
	"encoding/json"
	"fmt"

	"github.com/darchlabs/synchronizer-v2/internal/blockchain"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type storage struct {
	db *leveldb.DB
}

func NewStorage(filepath string) (*storage, error) {
	// read db from file
	db, err := leveldb.OpenFile(fmt.Sprintf("./%s", filepath), nil)
	if err != nil {
		return nil, err
	}

	return &storage{
		db: db,
	}, nil
}

func (s *storage) CreateEvent(e *Event) error {
	// format the composed key used in db
	key := fmt.Sprintf("event:%s:%s", e.Address, e.Abi.Name)

	// check if key already exists in database
	current, _ := s.GetEvent(e.Address, e.Abi.Name)
	if current != nil {
		return fmt.Errorf("key=%s already exists in db", key)
	}

	// set defalut values to event
	e.LatestBlockNumber = 0

	// parse struct to bytes
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}

	// save in database
	err = s.db.Put([]byte(key), b, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *storage) UpdateEvent(e *Event) error {
	// format the composed key used in db
	key := fmt.Sprintf("event:%s:%s", e.Address, e.Abi.Name)

	// parse struct to bytes
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}

	// save in database
	err = s.db.Put([]byte(key), b, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *storage) ListEvents() ([]*Event, error) {
	// format the composed prefix key used in db
	prefix := "event:"

	// prepare slice of events
	events := make([]*Event, 0)

	// iterate over dbn elementas and push in slice
	iter := s.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		// parse bytes to event struct
		var event *Event
		err := json.Unmarshal(iter.Value(), &event)
		if err!= nil {
			return nil, err
		}

		// append new element to events slice
		events = append(events, event)
	}
	iter.Release()

	// check if iteration has error
	err := iter.Error()
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (s *storage) ListEventsByAddress(address string) ([]*Event, error) {
	// format the composed prefix key used in db
	prefix := fmt.Sprintf("event:%s:", address)

	// prepare slice of events
	events := make([]*Event, 0)

	// iterate over db elements and push in slice
	iter := s.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		// parse bytes to Event struct
		var event *Event
		err := json.Unmarshal(iter.Value(), &event)
		if err != nil {
			return nil, err
		}

		// append new element to events slice
		events = append(events, event)
	}
	iter.Release()

	// check if iteration has error
	err := iter.Error()
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (s *storage) GetEvent(address string, eventName string) (*Event, error) {
	// format the composed key used in db
	key := fmt.Sprintf("event:%s:%s", address, eventName)
	
	// get bytes by composed key to db
	b, err := s.db.Get([]byte(key), nil)
	if err != nil {
		return nil, err
	}
	
	// parse bytes to Event struct
	var event *Event
	err = json.Unmarshal(b, &event)
	if err != nil {
		return nil, err
	}
	
	return event, nil
}

func (s *storage) DeleteEvent(address string, eventName string) error {
	// format the composed key used in db
	key := fmt.Sprintf("event:%s:%s", address, eventName)

	// delete data on db using composed key
	err := s.db.Delete([]byte(key), nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *storage) Stop() error {
	return s.db.Close()
}

func (s *storage) ListEventData(address string, eventName string) ([]interface{}, error) {	
	// format the composed prefix used in db
	prefix := fmt.Sprintf("data:%s:%s:", address, eventName)

	// prepare slice of event data
	data := make([]interface{}, 0)

	// iterate over db elements and push in event data slice
	iter := s.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		// parse bytes to blockchain.LogData struct
		var logData blockchain.LogData
		err := json.Unmarshal(iter.Value(), &logData)
		if err != nil {
			return nil, err
		}

		// append new elemento to data slice
		data = append(data,	logData)
	}
	iter.Release()

	// check if iteration has error
	err := iter.Error()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *storage) InsertEventData(e *Event, data []blockchain.LogData) (int64, error) {
	// define counter for new events
	count := int64(0)

	for _, d := range data {
		// format the composed prefix used in db
		key := fmt.Sprintf("data:%s:%s:%s", e.Address, e.Abi.Name, d.Tx)

		// get exist value from database
		exist, err := s.db.Has([]byte(key), nil)
		if err != nil {
			return 0, err
		}

		// check if key is present in database
		if exist {
			continue
		}

		// parse struct to bytes
		b, err := json.Marshal(d)
		if err != nil {
			return 0, err
		}

		// save in database
		err = s.db.Put([]byte(key), b, nil)
		if err != nil {
			return 0, err
		}

		// increase in one the counter
		count++;
	}
	
	return count, nil
}

func (s *storage) DeleteEventData(address string, eventName string) error {
	// format the composed prefix key used in db
	prefix := fmt.Sprintf("data:%s:%s:", address, eventName)

	// iterate over db elements and delete each one
	iter := s.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		// delete data on db using composed key
		err := s.db.Delete([]byte(iter.Key()), nil) 
		if err != nil {
			return err
		} 
	}

	return nil
}

