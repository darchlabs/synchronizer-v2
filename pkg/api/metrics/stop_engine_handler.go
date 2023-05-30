package metrics

import (
	txsengine "github.com/darchlabs/synchronizer-v2/internal/txsengine"
	"github.com/gofiber/fiber/v2"
)

type stopEngineRes struct {
	Data  txsengine.StatusEngine `json:"data"`
	Error string                 `json:"error,omitempty"`
}

func stopEngine(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		// Get status
		engineStatus := ctx.Engine.GetStatus()
		if engineStatus == txsengine.StatusStopped || engineStatus == txsengine.StatusStopping {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(startEngineRes{
				Error: "the engine is already stopped",
			})
		}

		if engineStatus == txsengine.StatusError {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(startEngineRes{
				Error: "the engine is in error status so it is already stopped",
			})
		}

		// Set status to stopped
		ctx.Engine.SetStatus(txsengine.StatusStopping)
		ctx.Engine.SetStatus(txsengine.StatusStopped)

		engineStatus = ctx.Engine.GetStatus()
		return c.Status(fiber.StatusOK).JSON(getEngineStatusRes{
			Data: engineStatus,
		})
	}
}
