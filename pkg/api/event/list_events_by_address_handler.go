package event

import (
	"github.com/darchlabs/synchronizer-v2/pkg/api"
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
		events, err := ctx.Storage.ListEventsByAddress(address)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(api.Response{
				Error: err.Error(),
			})
		}

		// define meta response
		meta := make(map[string]interface{})
		meta["cronjob"] = struct{
			Status string `json:"status"`
			Seconds int64 `json:"seconds"`
		} {
			Status: ctx.Cronjob.GetStatus(),
			Seconds: ctx.Cronjob.GetSeconds(),
		}

		// prepare response
		return c.Status(fiber.StatusOK).JSON(api.Response{
			Data: events,
			Meta: meta,
		})
	}
}
