package eventstorage

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/darchlabs/synchronizer-v2/internal/blockchain"
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/darchlabs/synchronizer-v2/pkg/event"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type Storage struct {
	storage *storage.S
}

func New(s *storage.S) *Storage {
	return &Storage{
		storage: s,
	}
}

func (s *Storage) InsertEvent(e *event.Event) error {
	// format the composed key used in db
	key := fmt.Sprintf("event:%s:%s", e.Address, e.Abi.Name)

	// check if key already exists in database
	current, _ := s.GetEvent(e.Address, e.Abi.Name)
	if current != nil {
		return fmt.Errorf("key=%s already exists in db", key)
	}

	// set defalut values to event
	e.ID = key
	e.LatestBlockNumber = 0
	e.CreatedAt = time.Now()

	// parse struct to bytes
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}

	// save in database
	err = s.storage.DB.Put([]byte(key), b, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) UpdateEvent(e *event.Event) error {
	fmt.Println("update event -----> :")
	// format the composed key used in db
	key := fmt.Sprintf("event:%s:%s", e.Address, e.Abi.Name)
	fmt.Println("1")

	// parse struct to bytes
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}

	fmt.Println("2")
	// save in database
	err = s.storage.DB.Put([]byte(key), b, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) ListEvents() ([]*event.Event, error) {
	// format the composed prefix key used in db
	prefix := "event:"

	// prepare slice of events
	events := make([]*event.Event, 0)

	// iterate over db elements and push in slice
	iter := s.storage.DB.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		// parse bytes to event struct
		var event *event.Event
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

func (s *Storage) ListEventsByAddress(address string) ([]*event.Event, error) {
	// format the composed prefix key used in db
	prefix := fmt.Sprintf("event:%s:", address)

	// prepare slice of events
	events := make([]*event.Event, 0)

	// iterate over db elements and push in slice
	iter := s.storage.DB.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		// parse bytes to Event struct
		var event *event.Event
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

func (s *Storage) GetEvent(address string, eventName string) (*event.Event, error) {
	// format the composed key used in db
	key := fmt.Sprintf("event:%s:%s", address, eventName)

	// get bytes by composed key to db
	b, err := s.storage.DB.Get([]byte(key), nil)
	if err != nil {
		return nil, err
	}

	// parse bytes to Event struct
	var event *event.Event
	err = json.Unmarshal(b, &event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (s *Storage) DeleteEvent(address string, eventName string) error {
	// format the composed key used in db
	key := fmt.Sprintf("event:%s:%s", address, eventName)

	// delete data on db using composed key
	err := s.storage.DB.Delete([]byte(key), nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) Stop() error {
	return s.storage.DB.Close()
}

func (s *Storage) ListEventData(address string, eventName string) ([]interface{}, error) {
	// format the composed prefix used in db
	prefix := fmt.Sprintf("data:%s:%s:", address, eventName)

	// prepare slice of event data
	data := make([]interface{}, 0)

	// iterate over db elements and push in event data slice
	iter := s.storage.DB.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		// parse bytes to blockchain.LogData struct
		var logData blockchain.LogData
		err := json.Unmarshal(iter.Value(), &logData)
		if err != nil {
			return nil, err
		}

		// append new elemento to data slice
		data = append(data, logData)
	}
	iter.Release()

	// check if iteration has error
	err := iter.Error()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Storage) InsertEventData(e *event.Event, data []blockchain.LogData) (int64, error) {
	// define counter for new events
	count := int64(0)

	for _, d := range data {
		// format the composed prefix used in db
		key := fmt.Sprintf("data:%s:%s:%s", e.Address, e.Abi.Name, d.Tx)

		// get exist value from database
		exist, err := s.storage.DB.Has([]byte(key), nil)
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
		err = s.storage.DB.Put([]byte(key), b, nil)
		if err != nil {
			return 0, err
		}

		// increase in one the counter
		count++
	}

	return count, nil
}

func (s *Storage) DeleteEventData(address string, eventName string) error {
	// format the composed prefix key used in db
	prefix := fmt.Sprintf("data:%s:%s:", address, eventName)

	// iterate over db elements and delete each one
	iter := s.storage.DB.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		// delete data on db using composed key
		err := s.storage.DB.Delete([]byte(iter.Key()), nil)
		if err != nil {
			return err
		}
	}

	return nil
}
