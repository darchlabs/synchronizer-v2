package event

import (
	"github.com/darchlabs/synchronizer-v2/internal/event"
	"github.com/gofiber/fiber/v2"
)

type ContextStorage interface {
	InsertEvent(e *event.Event) error
	ListEvents() ([]*event.Event, error)
	ListEventsByAddress(address string) ([]*event.Event, error)
	GetEvent(address string, eventName string) (*event.Event, error)
	DeleteEvent(address string, eventName string) error
	ListEventData(address string, eventName string) ([]interface{}, error) 
	DeleteEventData(address string, eventName string) error
}

type Context struct {
	Storage ContextStorage
}

func Route(app *fiber.App, ctx Context) {
	app.Post("/api/v1/events/:address", insertEventHandler(ctx))
	app.Get("/api/v1/events/:address", listEventsByAddressHandler(ctx))
	app.Get("/api/v1/events/:address/:event_name", getEventHandler(ctx))
	app.Get("/api/v1/events/:address/:event_name/data", listEventDataHandler(ctx))
	app.Delete("/api/v1/events/:address/:event_name", deleteEventHandler(ctx))
}