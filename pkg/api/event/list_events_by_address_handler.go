package event

import (
	"github.com/darchlabs/synchronizer-v2/pkg/api"
	"github.com/darchlabs/synchronizer-v2/pkg/util"
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

		// get pagination
		pagination := &util.Pagination{}
		err := pagination.GetPaginationFromFiber(c)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(api.Response{
				Error: err.Error(),
			})
		}

		// get elements from database
		events, err := ctx.EventStorage.ListEventsByAddress(address, pagination.Sort, pagination.Limit, pagination.Offset)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(api.Response{
				Error: err.Error(),
			})
		}

		// get all events by address count from database
		count, err := ctx.EventStorage.GetEventCountByAddress(address)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(api.Response{
				Error: err.Error(),
			})
		}

		// define meta response
		meta := make(map[string]interface{})
		meta["cronjob"] = CronjobMeta{
			Status:  ctx.Cronjob.GetStatus(),
			Seconds: ctx.Cronjob.GetSeconds(),
			Error:   ctx.Cronjob.GetError(),
		}
		meta["pagination"] = pagination.GetPaginationMeta(count)

		// prepare response
		return c.Status(fiber.StatusOK).JSON(api.Response{
			Data: events,
			Meta: meta,
		})
	}
}
