package event

import (
	"github.com/darchlabs/synchronizer-v2/pkg/api"
	"github.com/gofiber/fiber/v2"
)

func deleteEventHandler(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		// get and vald params
		address := c.Params("address")
		eventName := c.Params("event_name")
		if address == "" || eventName == "" {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(api.Response{
				Error: "invalid params",
			})
		}

		// delete event from storage
		err := ctx.EventStorage.DeleteEvent(address, eventName)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(api.Response{
				Error: err.Error(),
			})
		}

		// prepare response
		return c.Status(fiber.StatusOK).JSON(api.Response{})
	}
}