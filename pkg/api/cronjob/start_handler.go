package cronjob

import (
	"github.com/darchlabs/synchronizer-v2/pkg/api"
	"github.com/gofiber/fiber/v2"
)

func startCronjobHandler(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// start cronjob
		err := ctx.Cronjob.Start()
		if err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(api.Response{
				Error: err.Error(),
			})
		}

		// prepare response
		return c.Status(fiber.StatusOK).JSON(api.Response{})
	}
}
