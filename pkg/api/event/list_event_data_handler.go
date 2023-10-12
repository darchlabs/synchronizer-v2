package event

import (
	"github.com/darchlabs/synchronizer-v2/pkg/api"
	"github.com/darchlabs/synchronizer-v2/pkg/util"
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

		// get pagination
		pagination := &util.Pagination{}
		err := pagination.GetPaginationFromFiber(c)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(api.Response{
				Error: err.Error(),
			})
		}

		// get event from storage
		event, err := ctx.EventStorage.GetEvent(address, eventName)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(api.Response{
				Error: err.Error(),
			})
		}

		// get event data from storage
		data, err := ctx.EventStorage.ListEventData(address, eventName, pagination.Sort, pagination.Limit, pagination.Offset)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(api.Response{
				Error: err.Error(),
			})
		}

		// get all events by address count from database
		count, err := ctx.EventStorage.GetEventDataCount(address, eventName)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(api.Response{
				Error: err.Error(),
			})
		}

		// define meta response
		meta := make(map[string]interface{})
		meta["event"] = event
		meta["cronjob"] = CronjobMeta{
			Status:  ctx.Cronjob.GetStatus(),
			Seconds: ctx.Cronjob.GetSeconds(),
			Error:   ctx.Cronjob.GetError(),
		}
		meta["pagination"] = pagination.GetPaginationMeta(count)

		// prepare response
		return c.Status(fiber.StatusOK).JSON(api.Response{
			Data: data,
			Meta: meta,
		})
	}
}