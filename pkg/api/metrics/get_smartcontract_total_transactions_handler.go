package metrics

import (
	"github.com/gofiber/fiber/v2"
)

type getSmartContractTotalTransactionsRes struct {
	Data  int64  `json:"data"`
	Error string `json:"error,omitempty"`
}

func getSmartContractTotalTransactions(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		// Get address
		address := c.Params("address")
		if address == "" {
			return c.Status(fiber.StatusOK).JSON(getSmartContractTotalTransactionsRes{
				Error: "address cannot be nil",
			})
		}

		contract, err := ctx.SmartContractStorage.GetSmartContractByAddress(address)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				getSmartContractTotalTransactionsRes{
					Error: err.Error(),
				},
			)
		}

		if contract == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				getSmartContractTotalTransactionsRes{
					Error: "smart contract not found in the given address",
				},
			)

		}

		// Get the transactions
		totalTransactions, err := ctx.TransactionStorage.GetContractTotalTxs(contract.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				getSmartContractTotalTransactionsRes{
					Error: err.Error(),
				},
			)
		}

		// prepare response
		return c.Status(fiber.StatusOK).JSON(getSmartContractTotalTransactionsRes{
			Data: totalTransactions,
		})
	}
}