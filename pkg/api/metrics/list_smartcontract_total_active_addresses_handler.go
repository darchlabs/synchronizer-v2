package metrics

import (
	"github.com/gofiber/fiber/v2"
)

type listSmartContractTotalActiveAddressesRes struct {
	Data  int64  `json:"data"`
	Error string `json:"error,omitempty"`
}

func listSmartContractTotalActiveAddresses(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		// Get address
		address := c.Params("address")
		if address == "" {
			return c.Status(fiber.StatusOK).JSON(listSmartContractTotalActiveAddressesRes{
				Error: "address cannot be nil",
			})
		}

		contract, err := ctx.SmartContractStorage.GetSmartContractByAddress(address)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractTotalActiveAddressesRes{
					Error: err.Error(),
				},
			)
		}

		if contract == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractTotalActiveAddressesRes{
					Error: "smart contract not found in the given address",
				},
			)

		}

		// Get the transactions
		totalActiveAddresses, err := ctx.TransactionStorage.GetContractTotalAddresses(contract.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractTotalActiveAddressesRes{
					Error: err.Error(),
				},
			)
		}

		// prepare response
		return c.Status(fiber.StatusOK).JSON(listSmartContractTotalActiveAddressesRes{
			Data: totalActiveAddresses,
		})
	}
}
