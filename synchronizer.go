package synchronizer

import (
	"github.com/darchlabs/synchronizer-v2/internal/blockchain"
	"github.com/darchlabs/synchronizer-v2/pkg/event"
)

type EventStorage interface {
	ListAllEvents() ([]*event.Event, error)
	ListEvents(limit int64, offset int64) ([]*event.Event, error)
	ListEventsByAddress(address string, limit int64, offset int64) ([]*event.Event, error)
	GetEvent(address string, eventName string) (*event.Event, error)
	GetEventByID(id string) (*event.Event, error)
	InsertEvent(e *event.Event) (*event.Event, error)
	UpdateEvent(e *event.Event) error
	DeleteEvent(address string, eventName string) error
	ListEventData(address string, eventName string, limit int64, offset int64) ([]*event.EventData, error)
	InsertEventData(e *event.Event, data []blockchain.LogData) error
	GetEventsCount() (int64, error)
	GetEventCountByAddress(address string) (int64, error)
	GetEventDataCount(address string, eventName string) (int64, error)
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
