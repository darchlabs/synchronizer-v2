package synchronizer

import (
	"github.com/darchlabs/synchronizer-v2/internal/blockchain"
	"github.com/darchlabs/synchronizer-v2/internal/event"
)

type EventStorage interface {
	ListEventsByAddress(address string) ([]*event.Event, error)
	ListEvents() ([]*event.Event, error)
	GetEvent(address string, eventName string) (*event.Event, error)
	InsertEvent(e *event.Event) error
	UpdateEvent(e *event.Event) error
	DeleteEvent(address string, eventName string) error
	DeleteEventData(address string, eventName string) error
	ListEventData(address string, eventName string) ([]interface{}, error)
	InsertEventData(e *event.Event, data []blockchain.LogData) (int64, error)
	Stop() error
}

type Cronjob interface {
	Stop() error
	Restart() error
	Start() error
	GetStatus() string
	GetSeconds() int64 
}