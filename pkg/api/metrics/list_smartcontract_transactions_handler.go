package metrics

import (
	"fmt"

	"github.com/darchlabs/synchronizer-v2"
	"github.com/darchlabs/synchronizer-v2/internal/pagination"
	"github.com/darchlabs/synchronizer-v2/pkg/transaction"
	"github.com/gofiber/fiber/v2"
)

type listSmartContractTransactionsRes struct {
	Data  []*transaction.Transaction `json:"data"`
	Meta  interface{}                `json:"meta,omitempty"`
	Error string                     `json:"error,omitempty"`
}

func listSmartContractTransactions(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		// Get address
		address := c.Params("address")
		if address == "" {
			return c.Status(fiber.StatusOK).JSON(listSmartContractTransactionsRes{
				Error: "address cannot be nil",
			})
		}

		contract, err := ctx.SmartContractStorage.GetSmartContractByAddress(address)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractTransactionsRes{
					Error: err.Error(),
				},
			)
		}

		if contract == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractTransactionsRes{
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

		// Get the transactions
		transactions, err := ctx.TransactionStorage.ListTxsById(contract.ID, queryCtx)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractTransactionsRes{
					Error: err.Error(),
				},
			)
		}

		// Get the number of transactions of the contract
		totalTxs, err := ctx.TransactionStorage.GetTxsCountById(contract.ID)
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
		return c.Status(fiber.StatusOK).JSON(listSmartContractTransactionsRes{
			Data: transactions,
			Meta: meta,
		})
	}
}
