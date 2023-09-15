package event

import (
	"context"

	"github.com/darchlabs/synchronizer-v2/pkg/api"
	"github.com/darchlabs/synchronizer-v2/pkg/event"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
)

func insertEventHandler(ctx *api.Context, c *fiber.Ctx) (interface{}, interface{}, int, error) {
	c.Accepts("application/json")

	// prepate body request struct
	body := struct {
		Event *event.Event `json:"event"`
	}{}

	// parse body to event struct
	err := c.BodyParser(&body)
	if err != nil {
		return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"event: insertEventHandler c.BodyParser error",
		)
	}

	// get, valid and set address to event struct
	address := c.Params("address")
	if address == "" {
		return nil, nil, fiber.StatusUnprocessableEntity, errors.New(
			"event: insertEventHandler invalid address parameter",
		)
	}

	// Validate body
	validate := validator.New()
	err = validate.Struct(body.Event)
	if err != nil {
		return nil, nil, fiber.StatusBadRequest, errors.Wrap(
			err,
			"event: insertEventHandler validate.Struct error",
		)
	}

	// Update event
	inputEvent := body.Event
	inputEvent.Address = address
	inputEvent.ID = ctx.IDGen()
	inputEvent.Abi.ID = ctx.IDGen()
	inputEvent.LatestBlockNumber = 0
	inputEvent.Status = event.StatusRunning
	inputEvent.Error = ""
	inputEvent.CreatedAt = ctx.DateGen()
	inputEvent.UpdatedAt = ctx.DateGen()
	for _, input := range body.Event.Abi.Inputs {
		input.ID = ctx.IDGen()
	}

	// Validate abi is not nil
	if body.Event.Abi == nil {
		return nil, nil, fiber.StatusUnprocessableEntity, errors.New(
			"event: insertEventHandler invalid abi provided error",
		)
	}

	// Validate network is one of the supported
	network := body.Event.Network
	if !event.IsValidEventNetwork(network) {
		return nil, nil, fiber.StatusUnprocessableEntity, errors.New(
			"event: insertEventHandler invalid network provided error",
		)
	}

	// Validate node url is correct
	nodeURL := body.Event.NodeURL
	if nodeURL == "" {
		return nil, nil, fiber.StatusUnprocessableEntity, errors.New(
			"event: insertEventHandler invalid node url provided error",
		)
	}

	// get or create eth client in client
	client, ok := (*ctx.Clients)[nodeURL]
	if !ok {
		// validate client works
		client, err = ethclient.Dial(nodeURL)
		if err != nil {
			return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
				err, "event: insertEventHandler ethclient.Dial error",
			)
		}

		// validate client is working correctly
		_, err = client.ChainID(context.Background())
		if err != nil {
			return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
				err, "event: insertEventHandler ethclient.ChainID error",
			)
		}

		// save client in map
		(*ctx.Clients)[nodeURL] = client
	}

	// save event struct on database
	createdEvent, err := ctx.EventStorage.InsertEvent(body.Event)
	if err != nil {
		return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
			err, "event: insertEventHandler ctx.EventStorage.InsertEvent error",
		)
	}

	// prepare response
	// TODO: Evaluate the impact of returning 201 instead 200 since the event is created
	return createdEvent, nil, fiber.StatusOK, nil
}
