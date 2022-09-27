package event

import (
	"github.com/gofiber/fiber/v2"
)

type ContextStorage interface {
	CreateEvent(e *Event) error
	ListEvents() ([]*Event, error)
	ListEventsByAddress(address string) ([]*Event, error)
	GetEvent(address string, eventName string) (*Event, error)
	DeleteEvent(address string, eventName string) error
	ListEventData(address string, eventName string) ([]interface{}, error) 
	DeleteEventData(address string, eventName string) error
}

type Context struct {
	Storage ContextStorage
}

func Router(app *fiber.App, ctx Context) {
	app.Post("/api/v1/events/:address", createEventHandler(ctx))
	app.Get("/api/v1/events/:address", listEventsByAddressHandler(ctx))
	app.Get("/api/v1/events/:address/:event_name", getEventHandler(ctx))
	app.Get("/api/v1/events/:address/:event_name/data", ListEventDataHandler(ctx))
	app.Delete("/api/v1/events/:address/:event_name", deleteEventHandler(ctx))
}