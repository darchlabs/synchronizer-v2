package smartcontracts

import (
	"time"

	"github.com/darchlabs/synchronizer-v2"
	"github.com/darchlabs/synchronizer-v2/internal/env"
	"github.com/gofiber/fiber/v2"
)

type idGenerator func() string
type dateGenerator func() time.Time

type Context struct {
	Storage      synchronizer.SmartContractStorage
	EventStorage synchronizer.EventStorage
	Env          env.Env

	IDGen   idGenerator
	DateGen dateGenerator
}

func Route(app *fiber.App, ctx Context) {
	app.Post("/api/v1/smartcontracts", insertSmartContractHandler(ctx))
	app.Get("/api/v1/smartcontracts", listSmartContracts(ctx))
	app.Delete("/api/v1/smartcontracts/:address", deleteSmartContractHandler(ctx))
}
