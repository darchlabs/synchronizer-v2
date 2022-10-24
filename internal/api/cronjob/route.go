package cronjob

import (
	"github.com/darchlabs/synchronizer-v2"
	"github.com/gofiber/fiber/v2"
)

type Context struct {
	Cronjob synchronizer.Cronjob
}

func Route(app *fiber.App, ctx Context) {
	app.Post("/api/v1/cronjob/start", startCronjobHandler(ctx))
	app.Post("/api/v1/cronjob/stop", stopCronjobHandler(ctx))
	app.Post("/api/v1/cronjob/restart", restartCronjobHandler(ctx))
	app.Post("/api/v1/cronjob/status", statusCronjobHandler(ctx))
}