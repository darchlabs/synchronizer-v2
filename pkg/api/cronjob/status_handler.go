package cronjob

import (
	"github.com/darchlabs/synchronizer-v2/pkg/api"
	"github.com/gofiber/fiber/v2"
)

func statusCronjobHandler(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// get cronjob status
		status := ctx.Cronjob.GetStatus()
		seconds := ctx.Cronjob.GetSeconds()
		error := ctx.Cronjob.GetError()

		// prepare response
		response := struct {
			Status  string `json:"status"`
			Seconds int64  `json:"seconds"`
			Error   string `json:"error"`
		}{
			Status:  status,
			Seconds: seconds,
			Error:   error,
		}

		return c.Status(fiber.StatusOK).JSON(api.Response{
			Data: response,
		})
	}
}
