package metrics

import (
	txsengine "github.com/darchlabs/synchronizer-v2/internal/txs-engine"
	"github.com/gofiber/fiber/v2"
)

type getEngineStatusRes struct {
	Data txsengine.StatusEngine `json:"data"`
}

func getEngineStatus(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		engineStatus := ctx.Engine.GetStatus()

		return c.Status(fiber.StatusOK).JSON(getEngineStatusRes{
			Data: engineStatus,
		})
	}
}
