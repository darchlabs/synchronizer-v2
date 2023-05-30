package metrics

import (
	txsengine "github.com/darchlabs/synchronizer-v2/internal/txsengine"
	"github.com/gofiber/fiber/v2"
)

type startEngineRes struct {
	Data  txsengine.StatusEngine `json:"data"`
	Error string                 `json:"error,omitempty"`
}

func startEngine(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		// Get status
		engineStatus := ctx.Engine.GetStatus()
		if engineStatus == txsengine.StatusRunning {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(startEngineRes{
				Error: "the engine is already running",
			})

		}

		// Set status to running
		ctx.Engine.SetStatus(txsengine.StatusRunning)

		engineStatus = ctx.Engine.GetStatus()
		return c.Status(fiber.StatusOK).JSON(getEngineStatusRes{
			Data: engineStatus,
		})
	}
}
