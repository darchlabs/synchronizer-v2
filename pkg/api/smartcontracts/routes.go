package smartcontracts

import (
	"net/http"
	"time"

	"github.com/darchlabs/backoffice/pkg/client"
	"github.com/darchlabs/backoffice/pkg/middleware"
	"github.com/darchlabs/synchronizer-v2"
	"github.com/darchlabs/synchronizer-v2/internal/env"
	txsengine "github.com/darchlabs/synchronizer-v2/internal/txsengine"
	"github.com/darchlabs/synchronizer-v2/pkg/api"
	"github.com/gofiber/fiber/v2"
)

type idGenerator func() string
type dateGenerator func() time.Time

type Context struct {
	Storage      synchronizer.SmartContractStorage
	EventStorage synchronizer.EventStorage
	Env          *env.Env
	TxsEngine    txsengine.TxsEngine

	IDGen   idGenerator
	DateGen dateGenerator
}

func Route(app *fiber.App, ctx Context) {
	cl := client.New(&client.Config{
		Client:  http.DefaultClient,
		BaseURL: ctx.Env.BackofficeApiURL,
	})

	auth := middleware.NewAuth(cl)

	apiContext := &api.Context{
		ScStorage:    ctx.Storage,
		EventStorage: ctx.EventStorage,
		Env:          ctx.Env,
		TxsEngine:    ctx.TxsEngine,
		IDGen:        api.IDGenerator(ctx.IDGen),
		DateGen:      api.DateGenerator(ctx.DateGen),
	}

	app.Post("/api/v1/smartcontracts", auth.Middleware, api.HandleFunc(apiContext, insertSmartContractHandler))
	app.Post("/api/v1/smartcontracts/:address/restart", restartSmartContractHandler(ctx))
	app.Get("/api/v1/smartcontracts", listSmartContracts(ctx))
	app.Delete("/api/v1/smartcontracts/:address", deleteSmartContractHandler(ctx))
	app.Patch("/api/v1/smartcontracts/:address", updateSmartContractHandler(ctx))
}
