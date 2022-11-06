package cronjob

import (
	"github.com/darchlabs/synchronizer-v2/pkg/api"
	"github.com/gofiber/fiber/v2"
)

func statusCronjobHandler(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// get cronjob status
		status := ctx.Cronjob.GetStatus()

		// prepare response
		response := struct{
			status string
		}{
			status: status,
		}

		return c.Status(fiber.StatusOK).JSON(api.Response{
			Data: response,
		})
	}
}
