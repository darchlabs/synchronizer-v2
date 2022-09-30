package cronjob

import (
	"github.com/darchlabs/synchronizer-v2/internal/api"
	"github.com/gofiber/fiber/v2"
)

func stopCronjobHandler(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// stop cronjob
		err := ctx.Cronjob.Stop()
		if err != nil {
			c.Status(fiber.StatusUnprocessableEntity).JSON(api.Response{
				Error: err.Error(),
			})
		}

		// prepare response
		return c.Status(fiber.StatusOK).JSON(api.Response{})
	}
}
