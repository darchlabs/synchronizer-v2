package metrics

import (
	"fmt"

	"github.com/darchlabs/synchronizer-v2"
	"github.com/darchlabs/synchronizer-v2/internal/pagination"
	"github.com/darchlabs/synchronizer-v2/pkg/transaction"
	"github.com/gofiber/fiber/v2"
)

type listSmartContractFailedTransactionsRes struct {
	Data  []*transaction.Transaction `json:"data"`
	Meta  interface{}                `json:"meta,omitempty"`
	Error string                     `json:"error,omitempty"`
}

func listSmartContractFailedTransactions(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		// Get address
		address := c.Params("address")
		if address == "" {
			return c.Status(fiber.StatusOK).JSON(listSmartContractFailedTransactionsRes{
				Error: "address cannot be nil",
			})
		}

		contract, err := ctx.SmartContractStorage.GetSmartContractByAddress(address)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractFailedTransactionsRes{
					Error: err.Error(),
				},
			)
		}

		if contract == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractFailedTransactionsRes{
					Error: "smart contract not found in the given address",
				},
			)

		}

		// Get pagination
		p := &pagination.Pagination{}
		err = p.GetPaginationFromFiber(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractFailedTransactionsRes{
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
		// List the failed transactions on the given range
		failedTxs, err := ctx.TransactionStorage.ListContractFailedTxs(queryCTX)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractFailedTransactionsRes{
					Error: err.Error(),
				},
			)
		}

		// Get the total failed transactions
		totalFailedTxs, err := ctx.TransactionStorage.GetContractTotalFailedTxsCount(contract.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractFailedTransactionsRes{
					Error: err.Error(),
				},
			)
		}

		// define meta response with pagination
		meta := make(map[string]interface{})
		meta["pagination"] = p.GetPaginationMeta(totalFailedTxs)

		// prepare response
		return c.Status(fiber.StatusOK).JSON(listSmartContractFailedTransactionsRes{
			Data: failedTxs,
			Meta: meta,
		})
	}
}
