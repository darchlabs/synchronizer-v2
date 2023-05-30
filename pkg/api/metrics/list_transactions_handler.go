package metrics

import (
	"github.com/darchlabs/synchronizer-v2/pkg/transaction"
	"github.com/darchlabs/synchronizer-v2/pkg/util"
	"github.com/gofiber/fiber/v2"
)

type listTransactionsRes struct {
	Data  []*transaction.Transaction `json:"data"`
	Meta  interface{}                `json:"meta,omitempty"`
	Error string                     `json:"error,omitempty"`
}

func listTransactions(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		// Get pagination
		p := &util.Pagination{}
		err := p.GetPaginationFromFiber(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractTransactionsRes{
					Error: err.Error(),
				},
			)
		}

		// Get transactions
		txs, err := ctx.TransactionStorage.ListTxs(p.Sort, p.Limit, p.Offset)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listTransactionsRes{
					Error: err.Error(),
				},
			)
		}

		// Get the number of transactions of the contract
		totalTxs, err := ctx.TransactionStorage.GetTxsCount()
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
		return c.Status(fiber.StatusOK).JSON(listTransactionsRes{
			Data: txs,
			Meta: meta,
		})
	}
}
