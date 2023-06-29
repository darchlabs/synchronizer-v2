package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/darchlabs/synchronizer-v2/pkg/api"
	"github.com/darchlabs/synchronizer-v2/pkg/event"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-playground/validator"
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

		// Validate body
		validate := validator.New()
		err = validate.Struct(body.Event)
		if err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(api.Response{
				Error: err.Error(),
			})
		}

		// Update event
		body.Event.Address = address
		body.Event.ID = ctx.IDGen()
		body.Event.Abi.ID = ctx.IDGen()
		body.Event.LatestBlockNumber = 0
		body.Event.Status = event.StatusRunning
		body.Event.Error = ""
		body.Event.CreatedAt = ctx.DateGen()
		body.Event.UpdatedAt = ctx.DateGen()
		for _, input := range body.Event.Abi.Inputs {
			input.ID = ctx.IDGen()
		}

		// Validate abi is not nil
		if body.Event.Abi == nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(api.Response{
				Error: "abi cannot be nil",
			})
		}

		// Validate network is one of the supported
		network := body.Event.Network
		if !event.IsValidEventNetwork(network) {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(api.Response{
				Error: "invalid network",
			})
		}

		// Validate node url is correct
		nodeURL := body.Event.NodeURL
		if nodeURL == "" {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(api.Response{
				Error: "invalid nodeURL",
			})
		}

		// get or create eth client in client
		client, ok := (*ctx.Clients)[nodeURL]
		if !ok {
			// validate client works
			client, err = ethclient.Dial(nodeURL)
			if err != nil {
				return fmt.Errorf("can't getting ethclient error=%s", err)
			}

			// validate client is working correctly
			_, err = client.ChainID(context.Background())
			if err != nil {
				return fmt.Errorf("can't valid ethclient error=%s", err)
			}

			// save client in map
			(*ctx.Clients)[nodeURL] = client
		}

		// save event struct on database
		createdEvent, err := ctx.Storage.InsertEvent(body.Event)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				api.Response{
					Error: err.Error(),
				},
			)
		}

		// prepare response
		return c.Status(fiber.StatusOK).JSON(api.Response{
			Data: createdEvent,
		})
	}
}
