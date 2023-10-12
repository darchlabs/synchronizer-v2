package smartcontracts

import (
	"net/http"
	"time"

	"github.com/darchlabs/backoffice/pkg/client"
	"github.com/darchlabs/backoffice/pkg/middleware"
	"github.com/darchlabs/synchronizer-v2"
	"github.com/darchlabs/synchronizer-v2/internal/env"
	"github.com/darchlabs/synchronizer-v2/internal/sync"
	"github.com/darchlabs/synchronizer-v2/pkg/api"
	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
)

type idGenerator func() string
type dateGenerator func() time.Time

type Context struct {
	Storage      synchronizer.SmartContractStorage
	EventStorage synchronizer.EventStorage
	Env          *env.Env

	Engine *sync.Engine

	IDGen   idGenerator
	DateGen dateGenerator
}

func Route(app *fiber.App, ctx Context) {
	cl := client.New(&client.Config{
		Client:  http.DefaultClient,
		BaseURL: ctx.Env.BackofficeApiURL,
	})

	validate := validator.New()
	auth := middleware.NewAuth(cl)

	apiContext := &api.Context{
		ScStorage:    ctx.Storage,
		EventStorage: ctx.EventStorage,
		Env:          ctx.Env,
		TxsEngine:    nil, // TODO: verify if is used outside of smartcontracts.
		SyncEngine:   ctx.Engine,
		IDGen:        api.IDGenerator(ctx.IDGen),
		DateGen:      api.DateGenerator(ctx.DateGen),
	}

	// V1 ROUTES
	// routing
	app.Post("/api/v1/smartcontracts", auth.Middleware, api.HandleFunc(apiContext, insertSmartContractHandler))
	app.Post("/api/v1/smartcontracts/:address/restart", restartSmartContractHandler(ctx))
	// app.Delete("/api/v1/smartcontracts/:address", deleteSmartContractHandler(ctx))
	app.Patch("/api/v1/smartcontracts/:address", updateSmartContractHandler(ctx))

	// V2 ROUTES
	// handlers
	postSmartContractV2Handler := &postSmartContractV2Handler{validate}
	listSmartContractV2Handler := &listSmartContractV2Handler{}

	// routing
	app.Post(
		"/api/v2/smartcontracts",
		auth.Middleware,
		api.HandleFunc(apiContext, postSmartContractV2Handler.Invoke),
	)
	app.Get("/api/v2/smartcontracts", auth.Middleware, api.HandleFunc(apiContext, listSmartContractV2Handler.Invoke))
}
