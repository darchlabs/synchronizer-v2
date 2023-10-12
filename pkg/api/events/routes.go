package events

import (
	"net/http"
	"time"

	"github.com/darchlabs/backoffice/pkg/client"
	"github.com/darchlabs/backoffice/pkg/middleware"
	"github.com/darchlabs/synchronizer-v2/internal/env"
	"github.com/darchlabs/synchronizer-v2/internal/sync"
	"github.com/darchlabs/synchronizer-v2/pkg/api"
	"github.com/gofiber/fiber/v2"
)

type idGenerator func() string
type dateGenerator func() time.Time

type Context struct {
	Env    *env.Env
	Engine *sync.Engine

	IDGen   idGenerator
	DateGen dateGenerator
}

func Route(app *fiber.App, apiContext *api.Context) {
	cl := client.New(&client.Config{
		Client:  http.DefaultClient,
		BaseURL: apiContext.Env.BackofficeApiURL,
	})
	auth := middleware.NewAuth(cl)

	// V2 ROUTES
	// handlers
	getEventsByAddressV2Handler := &getEventsByAddressV2Handler{}
	getEventDataV2Handler := &getEventDataV2Handler{}

	// routing
	app.Get("/api/v2/events/:address", auth.Middleware, api.HandleFunc(apiContext, getEventsByAddressV2Handler.Invoke))
	app.Get("/api/v2/events/:address/data/:event_name", auth.Middleware, api.HandleFunc(apiContext, getEventDataV2Handler.Invoke))
}
