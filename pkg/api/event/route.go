package event

import (
	"fmt"

	"github.com/darchlabs/synchronizer-v2"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gofiber/fiber/v2"
)

type Context struct {
	Clients *map[string]*ethclient.Client
	Storage synchronizer.EventStorage
	Cronjob synchronizer.Cronjob
	BaseURL string
}

func Route(app *fiber.App, ctx Context) {
	app.Post(fmt.Sprintf("%s/api/v1/events/:address", ctx.BaseURL), insertEventHandler(ctx))
	app.Get(fmt.Sprintf("%s/api/v1/events", ctx.BaseURL), listEvents(ctx))
	app.Get(fmt.Sprintf("%s/api/v1/events/:address", ctx.BaseURL), listEventsByAddressHandler(ctx))
	app.Get(fmt.Sprintf("%s/api/v1/events/:address/:event_name", ctx.BaseURL), getEventHandler(ctx))
	app.Get(fmt.Sprintf("%s/api/v1/events/:address/:event_name/data", ctx.BaseURL), listEventDataHandler(ctx))
	app.Delete(fmt.Sprintf("%s/api/v1/events/:address/:event_name", ctx.BaseURL), deleteEventHandler(ctx))
}
