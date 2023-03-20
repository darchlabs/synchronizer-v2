package event

import (
	"github.com/darchlabs/synchronizer-v2/pkg/api"
	"github.com/darchlabs/synchronizer-v2/pkg/event"
	"github.com/gofiber/fiber/v2"
)

type MetaCronjob struct {
	Status  string `json:"status"`
	Seconds int64  `json:"seconds"`
	Error   string `json:"error"`
}

type Meta struct {
	Cronjob *MetaCronjob `json:"cronjob"`
}

type ListEventResponse struct {
	Data []*event.Event `json:"data"`
	Meta *Meta          `json:"meta"`
}

func listEvents(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// get elements from database
		events, err := ctx.Storage.ListEvents()
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(api.Response{
				Error: err.Error(),
			})
		}

		// define meta response
		meta := &Meta{
			Cronjob: &MetaCronjob{
				Status:  ctx.Cronjob.GetStatus(),
				Seconds: ctx.Cronjob.GetSeconds(),
				Error:   ctx.Cronjob.GetError(),
			},
		}

		// prepare response
		return c.Status(fiber.StatusOK).JSON(ListEventResponse{
			Data: events,
			Meta: meta,
		})
	}
}
