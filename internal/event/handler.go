package event

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Data interface{} `json:"data,omitempty"`
	Meta interface{} `json:"meta,omitempty"`
	Error interface{} `json:"error,omitempty"`
}

func listEventsByAddressHandler(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// get and vald params
		address := c.Params("address")
		if address == "" {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(Response{
				Error: "invalid param",
			})
		}	

		// get elements from database
		// events, err := ctx.Storage.ListEventsByAddress(address)
		events, err := ctx.Storage.ListEvents()
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(Response{
				Error: err.Error(),
			})
		}

		// prepare response
		return c.Status(fiber.StatusOK).JSON(Response{
			Data: events,
		})
	}
}

func getEventHandler(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		// get and vald params
		address := c.Params("address")
		eventName := c.Params("event_name")
		if address == "" || eventName == "" {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(Response{
				Error: "invalid params",
			})
		}	
		
		// get event from storage
		event, err  := ctx.Storage.GetEvent(address, eventName)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(Response{
				Error: err.Error(),
			})
		}

		// prepare response
		return c.Status(fiber.StatusOK).JSON(Response{
			Data: event,
		})
	}
}

func createEventHandler(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		// prepate body request struct
		body := struct {
			Event *Event `json:"event"`
		}{}

		// parse body to event struct
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(Response{
				Error: err.Error(),
			})
		}

		// get, valid and set address to event struct
		address := c.Params("address")
		if address == "" {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(Response{
				Error: "invalid param",
			})
		}	
		body.Event.Address = address

		// TODO(ca): validate event stuct

		// save event struct on database
		err = ctx.Storage.CreateEvent(body.Event)
		if err != nil {
			return c.Status(fiber.StatusForbidden).JSON(
				Response{
					Error: err.Error(),
				},
			)
		}

		// prepare response
		return c.Status(fiber.StatusOK).JSON(Response{
			Data: body.Event,
		})
	}
}

func deleteEventHandler(ctx Context) func (c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		// get and vald params
		address := c.Params("address")
		eventName := c.Params("event_name")
		if address == "" || eventName == "" {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(Response{
				Error: "invalid params",
			})
		}	

		// delete event data from storage
		err := ctx.Storage.DeleteEventData(address, eventName)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(Response{
				Error: err.Error(),
			})
		}

		// delete event from storage
		err = ctx.Storage.DeleteEvent(address, eventName)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(Response{
				Error: err.Error(),
			})
		}

		// prepare response
		return c.Status(fiber.StatusOK).JSON(Response{})
	}
}

func ListEventDataHandler(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		// get and valid params
		address := c.Params("address")
		eventName := c.Params("event_name")
		if address == "" || eventName == "" {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(Response{
				Error: "invalid params",
			})
		}	

		// get event from storage
		event, err  := ctx.Storage.GetEvent(address, eventName)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(Response{
				Error: err.Error(),
			})
		}
		
		// get event data from storage
		data, err := ctx.Storage.ListEventData(address, eventName)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(Response{
				Error: err.Error(),
			})
		}

		// prepare response
		return c.Status(fiber.StatusOK).JSON(Response{
			Data: data,
			Meta: event,
		})
	}
}