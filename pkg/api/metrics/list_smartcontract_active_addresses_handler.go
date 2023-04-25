package metrics

import (
	"github.com/darchlabs/synchronizer-v2/internal/pagination"
	"github.com/gofiber/fiber/v2"
)

type listSmartContractActiveAddressesRes struct {
	Data  []string    `json:"data,omitempty"`
	Meta  interface{} `json:"meta,omitempty"`
	Error string      `json:"error,omitempty"`
}

func listSmartContractActiveAddresses(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		// Get address
		address := c.Params("address")
		if address == "" {
			return c.Status(fiber.StatusOK).JSON(listSmartContractActiveAddressesRes{
				Error: "address cannot be nil",
			})
		}

		contract, err := ctx.SmartContractStorage.GetSmartContractByAddress(address)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractActiveAddressesRes{
					Error: err.Error(),
				},
			)
		}

		if contract == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractActiveAddressesRes{
					Error: "smart contract not found in the given address",
				},
			)

		}

		// Get pagination
		p := &pagination.Pagination{}
		err = p.GetPaginationFromFiber(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractTransactionsRes{
					Error: err.Error(),
				},
			)
		}

		// Get the transactions
		activeAddresses, err := ctx.TransactionStorage.ListContractUniqueAddresses(contract.ID, p.Sort, p.Limit, p.Offset)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractActiveAddressesRes{
					Error: err.Error(),
				},
			)
		}

		// define meta response with pagination
		meta := make(map[string]interface{})
		meta["pagination"] = p.GetPaginationMeta(int64(len(activeAddresses)))

		// prepare response
		return c.Status(fiber.StatusOK).JSON(listSmartContractActiveAddressesRes{
			Data: activeAddresses,
			Meta: meta,
		})
	}
}
