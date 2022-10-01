package event

import (
	"encoding/json"

	"github.com/darchlabs/synchronizer-v2/internal/api"
	"github.com/darchlabs/synchronizer-v2/internal/event"
	"github.com/gofiber/fiber/v2"
)

func insertEventHandler(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		// prepate body request struct
		body := struct {
			Event *event.Event `json:"event"`
		}{}

		// parse body to event struct
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(api.Response{
				Error: err.Error(),
			})
		}

		// get, valid and set address to event struct
		address := c.Params("address")
		if address == "" {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(api.Response{
				Error: "invalid param",
			})
		}	
		body.Event.Address = address

		// TODO(ca): validate event stuct

		// save event struct on database
		err = ctx.Storage.InsertEvent(body.Event)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				api.Response{
					Error: err.Error(),
				},
			)
		}

		// prepare response
		return c.Status(fiber.StatusOK).JSON(api.Response{
			Data: body.Event,
		})
	}
}
