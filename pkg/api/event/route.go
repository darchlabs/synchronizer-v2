package event

import (
	"time"

	"github.com/darchlabs/synchronizer-v2"
	"github.com/darchlabs/synchronizer-v2/internal/env"
	"github.com/darchlabs/synchronizer-v2/internal/txsengine"
	"github.com/darchlabs/synchronizer-v2/pkg/api"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gofiber/fiber/v2"
)

type idGenerator func() string
type dateGenerator func() time.Time

type Context struct {
	Clients      *map[string]*ethclient.Client
	EventStorage synchronizer.EventStorage
	ScStorage    synchronizer.SmartContractStorage
	Cronjob      synchronizer.Cronjob
	Env          *env.Env
	TxsEngine    txsengine.TxsEngine

	IDGen   idGenerator
	DateGen dateGenerator
}

func Route(app *fiber.App, ctx Context) {
	apiContext := &api.Context{
		ScStorage:    ctx.ScStorage,
		EventStorage: ctx.EventStorage,
		Env:          ctx.Env,
		TxsEngine:    ctx.TxsEngine,
		IDGen:        api.IDGenerator(ctx.IDGen),
		DateGen:      api.DateGenerator(ctx.DateGen),
	}

	app.Post("/api/v1/events/:address", api.HandleFunc(apiContext, insertEventHandler))
	app.Get("/api/v1/events", listEvents(ctx))
	app.Get("/api/v1/events/:address", listEventsByAddressHandler(ctx)) // KEEP
	app.Get("/api/v1/events/:address/:event_name", getEventHandler(ctx))
	app.Get("/api/v1/events/:address/:event_name/data", listEventDataHandler(ctx))
	app.Delete("/api/v1/events/:address/:event_name", deleteEventHandler(ctx))
	app.Post("/api/v1/events/:address/:event_name/start", startEventHandler(ctx))
	app.Post("/api/v1/events/:address/:event_name/stop", stopEventHandler(ctx))
}
