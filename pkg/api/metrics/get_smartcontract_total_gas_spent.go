package metrics

import (
	"github.com/gofiber/fiber/v2"
)

func getSmartContractTotalGasSpent(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		res := struct {
			Data  int64  `json:"data"`
			Error string `json:"error,omitempty"`
		}{}

		// Get address
		address := c.Params("address")
		if address == "" {
			res.Error = "address cannot be nil"
			return c.Status(fiber.StatusOK).JSON(res)
		}

		// get contract by address
		contract, err := ctx.SmartContractStorage.GetSmartContractByAddress(address)
		if err != nil {
			res.Error = err.Error()
			return c.Status(fiber.StatusInternalServerError).JSON(res)
		}

		// check if contras is defined
		if contract == nil {
			res.Error = "smart contract not found in the given address"
			return c.Status(fiber.StatusInternalServerError).JSON(res)
		}

		// Get the total gas spent
		gasSpent, err := ctx.TransactionStorage.GetTotalGasSpentById(contract.ID)
		if err != nil {
			res.Error = err.Error()
			return c.Status(fiber.StatusInternalServerError).JSON(res)
		}

		// prepare response
		res.Data = gasSpent
		return c.Status(fiber.StatusOK).JSON(res)
	}
}
