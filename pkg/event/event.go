package event

import (
	"time"

	"github.com/darchlabs/synchronizer-v2/internal/blockchain"
)

type EventNetwork string

const (
	Ethereum EventNetwork = "ethereum"
	Polygon  EventNetwork = "polygon"
)

type EventDataStorage interface {
	InsertEventData(e *Event, data []blockchain.LogData) (int64, error)
	UpdateEvent(e *Event) error
}

type Event struct {
	ID                string       `json:"id"`
	Network           EventNetwork `json:"network"`
	NodeURL           string       `json:"nodeURL"`
	Address           string       `json:"address"`
	LatestBlockNumber int64        `json:"latestBlockNumber"`
	Abi               *Abi         `json:"abi"`
	Error             string       `json:"error"`

	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

func (e *Event) UpdateError(eventErr error, storage EventDataStorage) error {
	// update error value, can be a string or nil
	if eventErr != nil {
		e.Error = eventErr.Error()
	} else {
		e.Error = ""
	}

	// update event in database
	err := storage.UpdateEvent(e)
	if err != nil {
		return err
	}

	return nil
}

func (e *Event) UpdateLatestBlock(lbn int64, storage EventDataStorage) error {
	// change latest block number value
	e.LatestBlockNumber = lbn

	// chanche updated at value
	e.UpdatedAt = time.Now()

	// update event in database
	err := storage.UpdateEvent(e)
	if err != nil {
		return err
	}

	return nil
}

func (e *Event) InsertData(data []blockchain.LogData, storage EventDataStorage) (int64, error) {
	// insert event data to event
	count, err := storage.InsertEventData(e, data)
	if err != nil {
		return 0, err
	}

	return count, nil
}
