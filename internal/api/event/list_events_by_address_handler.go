package event

import (
	"github.com/darchlabs/synchronizer-v2/internal/api"
	"github.com/gofiber/fiber/v2"
)

func listEventsByAddressHandler(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// get and vald params
		address := c.Params("address")
		if address == "" {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(api.Response{
				Error: "invalid param",
			})
		}	

		// get elements from database
		// events, err := ctx.Storage.ListEventsByAddress(address)
		events, err := ctx.Storage.ListEvents()
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(api.Response{
				Error: err.Error(),
			})
		}

		// prepare response
		return c.Status(fiber.StatusOK).JSON(api.Response{
			Data: events,
		})
	}
}
