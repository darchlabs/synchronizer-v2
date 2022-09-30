package event

import (
	"github.com/darchlabs/synchronizer-v2/internal/api"
	"github.com/gofiber/fiber/v2"
)

func listEventDataHandler(ctx Context) func(c *fiber.Ctx) error {
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
		event, err  := ctx.Storage.GetEvent(address, eventName)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(api.Response{
				Error: err.Error(),
			})
		}
		
		// get event data from storage
		data, err := ctx.Storage.ListEventData(address, eventName)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(api.Response{
				Error: err.Error(),
			})
		}

		// prepare response
		return c.Status(fiber.StatusOK).JSON(api.Response{
			Data: data,
			Meta: event,
		})
	}
}