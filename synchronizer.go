package synchronizer

import (
	"github.com/darchlabs/synchronizer-v2/internal/blockchain"
	"github.com/darchlabs/synchronizer-v2/pkg/event"
)

type EventStorage interface {
	ListEventsByAddress(address string) ([]*event.Event, error)
	ListEvents() ([]*event.Event, error)
	GetEvent(address string, eventName string) (*event.Event, error)
	GetEventByID(id int64) (*event.Event, error)
	InsertEvent(e *event.Event) (*event.Event, error)
	UpdateEvent(e *event.Event) error
	DeleteEvent(address string, eventName string) error
	ListEventData(address string, eventName string) ([]*event.EventData, error)
	InsertEventData(e *event.Event, data []blockchain.LogData) error
	Stop() error
}

type Cronjob interface {
	Stop() error
	Restart() error
	Start() error
	GetStatus() string
	GetSeconds() int64
	GetError() string
}
