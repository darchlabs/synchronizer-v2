package event

import (
	"github.com/darchlabs/synchronizer-v2/internal/blockchain"
)

type EventDataStorage interface {
	InsertEventData(e *Event, data []blockchain.LogData) (int64, error)
	UpdateEvent(e *Event) error
}

type Event struct {
	Address string `json:"address"`
	LatestBlockNumber int64 `json:"LatestBlockNumber"`
	Abi *Abi `json:"abi"`
}

func (e *Event) UpdateLatestBlock(lbn int64, storage EventDataStorage) error {
	// change latest block number value
	e.LatestBlockNumber = lbn;

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