package smartcontracts

import (
	"github.com/darchlabs/synchronizer-v2/pkg/event"
	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/gofiber/fiber/v2"
)

type restartSmartContractResponse struct {
	Error string `json:"error,omitemty"`
}

func restartSmartContractHandler(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		address := c.Params("address")
		if address == "" {
			return c.Status(fiber.StatusOK).JSON(restartSmartContractResponse{
				Error: "address cannot be nil",
			})
		}

		// get smart contract
		sc, err := ctx.Storage.GetSmartContractByAddress(address)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(
				createSmartContractResponse{
					Error: err.Error(),
				},
			)
		}

		if sc.Status == smartcontract.StatusError {
			err = ctx.Storage.UpdateStatusAndError(sc.ID, smartcontract.StatusRunning, nil)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(
					createSmartContractResponse{
						Error: err.Error(),
					},
				)
			}
		}

		events, err := ctx.EventStorage.ListEventsByAddress(sc.Address, "DESC", 999, 0)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(
				createSmartContractResponse{
					Error: err.Error(),
				},
			)
		}

		// restart all events
		for _, ev := range events {
			if ev.Status == event.StatusError {
				// if status is error, change to running and make update
				ev.Status = event.StatusRunning
				ev.Error = ""
				err := ctx.EventStorage.UpdateEvent(ev)
				if err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(
						createSmartContractResponse{
							Error: err.Error(),
						},
					)
				}
			}
		}

		return c.Status(fiber.StatusOK).JSON(struct{}{})
	}
}
