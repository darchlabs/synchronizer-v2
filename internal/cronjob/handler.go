package cronjob

import (
	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Data interface{} `json:"data,omitempty"`
	Meta interface{} `json:"meta,omitempty"`
	Error interface{} `json:"error,omitempty"`
}

func startCronjobHandler(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// start cronjob
		err := ctx.Cronjob.Start()
		if err != nil {
			c.Status(fiber.StatusUnprocessableEntity).JSON(Response{
				Error: err.Error(),
			})
		}

		// prepare response
		return c.Status(fiber.StatusOK).JSON(Response{})
	}
}

func stopCronjobHandler(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// stop cronjob
		err := ctx.Cronjob.Stop()
		if err != nil {
			c.Status(fiber.StatusUnprocessableEntity).JSON(Response{
				Error: err.Error(),
			})
		}

		// prepare response
		return c.Status(fiber.StatusOK).JSON(Response{})
	}
}

func restartCronjobHandler(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// restart cronjob
		err := ctx.Cronjob.Restart()
		if err != nil {
			c.Status(fiber.StatusUnprocessableEntity).JSON(Response{
				Error: err.Error(),
			})
		}

		// prepare response
		return c.Status(fiber.StatusOK).JSON(Response{})
	}
}