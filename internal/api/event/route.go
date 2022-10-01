package event

import (
	"github.com/darchlabs/synchronizer-v2"
	"github.com/gofiber/fiber/v2"
)

type Context struct {
	Storage synchronizer.EventStorage
}

func Route(app *fiber.App, ctx Context) {
	app.Post("/api/v1/events/:address", insertEventHandler(ctx))
	app.Get("/api/v1/events/:address", listEventsByAddressHandler(ctx))
	app.Get("/api/v1/events/:address/:event_name", getEventHandler(ctx))
	app.Get("/api/v1/events/:address/:event_name/data", listEventDataHandler(ctx))
	app.Delete("/api/v1/events/:address/:event_name", deleteEventHandler(ctx))
}