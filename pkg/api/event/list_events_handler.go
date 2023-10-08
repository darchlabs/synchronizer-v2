package event

import (
	"github.com/darchlabs/synchronizer-v2/pkg/api"
	"github.com/darchlabs/synchronizer-v2/pkg/util"
	"github.com/gofiber/fiber/v2"
)

type CronjobMeta struct {
	Status  string `json:"status"`
	Seconds int64  `json:"seconds"`
	Error   string `json:"error"`
}

func listEvents(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		// get pagination
		pagination := &util.Pagination{}
		err := pagination.GetPaginationFromFiber(c)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(api.Response{
				Error: err.Error(),
			})
		}

		// get elements from database
		events, err := ctx.EventStorage.ListEvents(pagination.Sort, pagination.Limit, pagination.Offset)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(api.Response{
				Error: err.Error(),
			})
		}

		// get all events count from database
		count, err := ctx.EventStorage.GetEventsCount()
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
