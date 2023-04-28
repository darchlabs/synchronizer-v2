package metrics

import (
	"fmt"

	"github.com/darchlabs/synchronizer-v2"
	"github.com/darchlabs/synchronizer-v2/internal/pagination"
	"github.com/gofiber/fiber/v2"
)

type listSmartContractGasSpentRes struct {
	Data  []string    `json:"data"`
	Meta  interface{} `json:"meta,omitempty"`
	Error string      `json:"error,omitempty"`
}

func listSmartContractGasSpent(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		// Get address
		address := c.Params("address")
		if address == "" {
			return c.Status(fiber.StatusOK).JSON(listSmartContractGasSpentRes{
				Error: "address cannot be nil",
			})
		}

		contract, err := ctx.SmartContractStorage.GetSmartContractByAddress(address)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractGasSpentRes{
					Error: err.Error(),
				},
			)
		}

		if contract == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractGasSpentRes{
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
		queryCTX := &synchronizer.ListItemsInRangeCTX{
			Id:        contract.ID,
			StartTime: fmt.Sprint(p.StartTime),
			EndTime:   fmt.Sprint(p.EndTime),
			Sort:      p.Sort,
			Limit:     p.Limit,
			Offset:    p.Offset,
		}
		// List the array of the gas spent on the given range
		gasSpentArr, err := ctx.TransactionStorage.ListContractGasSpent(queryCTX)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractGasSpentRes{
					Error: err.Error(),
				},
			)
		}

		// Get the number of transactions of the contract
		totalTxs, err := ctx.TransactionStorage.GetContractTotalTxsCount(contract.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractGasSpentRes{
					Error: err.Error(),
				},
			)
		}

		// define meta response with pagination
		meta := make(map[string]interface{})
		meta["pagination"] = p.GetPaginationMeta(totalTxs)

		// prepare response
		return c.Status(fiber.StatusOK).JSON(listSmartContractGasSpentRes{
			Data: gasSpentArr,
			Meta: meta,
		})
	}
}
