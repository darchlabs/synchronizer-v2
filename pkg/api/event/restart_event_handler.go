package event

import (
	"github.com/darchlabs/synchronizer-v2/pkg/api"
	"github.com/gofiber/fiber/v2"
)

func restartEventHandler(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		// get and valid params
		address := c.Params("address")
		eventName := c.Params("event_name")
		if address == "" || eventName == "" {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(api.Response{
				Error: "invalid params",
			})
		}

		// get event from storage
		event, err := ctx.Storage.GetEvent(address, eventName)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(api.Response{
				Error: err.Error(),
			})
		}

		// remove error from event
		err = event.UpdateError(nil, ctx.Storage)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(api.Response{
				Error: err.Error(),
			})
		}

		// prepare reponse
		return c.Status(fiber.StatusOK).JSON(api.Response{})
	}
}
