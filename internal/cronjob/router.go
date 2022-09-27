package cronjob

import "github.com/gofiber/fiber/v2"

type Cronjob interface {
	Start() error
	Restart() error
	Stop() error
}

type Context struct {
	Cronjob Cronjob
}

func Router(app *fiber.App, ctx Context) {
	app.Post("/api/v1/cronjob/start", startCronjobHandler(ctx))
	app.Post("/api/v1/cronjob/stop", stopCronjobHandler(ctx))
	app.Post("/api/v1/cronjob/restart", restartCronjobHandler(ctx))
}