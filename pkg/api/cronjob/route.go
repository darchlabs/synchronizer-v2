package cronjob

import (
	"fmt"

	"github.com/darchlabs/synchronizer-v2"
	"github.com/gofiber/fiber/v2"
)

type Context struct {
	Cronjob synchronizer.Cronjob
	BaseURL string
}

func Route(app *fiber.App, ctx Context) {
	app.Post(fmt.Sprintf("%s/api/v1/cronjob/start", ctx.BaseURL), startCronjobHandler(ctx))
	app.Post(fmt.Sprintf("%s/api/v1/cronjob/stop", ctx.BaseURL), stopCronjobHandler(ctx))
	app.Post(fmt.Sprintf("%s/api/v1/cronjob/restart", ctx.BaseURL), restartCronjobHandler(ctx))
	app.Post(fmt.Sprintf("%s/api/v1/cronjob/status", ctx.BaseURL), statusCronjobHandler(ctx))
}
