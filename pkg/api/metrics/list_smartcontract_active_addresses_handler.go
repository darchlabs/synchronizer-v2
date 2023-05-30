package metrics

import (
	"fmt"

	"github.com/darchlabs/synchronizer-v2"
	"github.com/darchlabs/synchronizer-v2/internal/pagination"
	"github.com/gofiber/fiber/v2"
)

type listSmartContractActiveAddressesRes struct {
	Data  []string    `json:"data"`
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

		// Prepare the query context
		queryCtx := &synchronizer.ListItemsInRangeCtx{
			StartTime: fmt.Sprint(p.StartTime),
			EndTime:   fmt.Sprint(p.EndTime),
			Sort:      p.Sort,
			Limit:     p.Limit,
			Offset:    p.Offset,
		}
		// Get the array of unique adresses in the given range
		activeAddresses, err := ctx.TransactionStorage.ListUniqueAddresses(contract.ID, queryCtx)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractActiveAddressesRes{
					Error: err.Error(),
				},
			)
		}

		// Get the total unique addresses
		totalActiveAddresses, err := ctx.TransactionStorage.GetAddressesCountById(contract.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractActiveAddressesRes{
					Error: err.Error(),
				},
			)
		}

		// define meta response with pagination
		meta := make(map[string]interface{})
		meta["pagination"] = p.GetPaginationMeta(totalActiveAddresses)

		// prepare response
		return c.Status(fiber.StatusOK).JSON(listSmartContractActiveAddressesRes{
			Data: activeAddresses,
			Meta: meta,
		})
	}
}
