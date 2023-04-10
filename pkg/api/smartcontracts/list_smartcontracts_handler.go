package smartcontracts

import (
	"github.com/darchlabs/synchronizer-v2/internal/pagination"
	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/gofiber/fiber/v2"
)

type listSmartContractsResponse struct {
	Data  []*smartcontract.SmartContract `json:"data,ommitempty"`
	Meta  interface{}                    `json:"meta,ommitempty"`
	Error string                         `json:"error,omitemty"`
}

func listSmartContracts(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		// get pagination
		p := &pagination.Pagination{}
		err := p.GetPaginationFromFiber(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractsResponse{
					Error: err.Error(),
				},
			)
		}

		// get elements from database
		smartcontracts, err := ctx.Storage.ListSmartContracts(p.Sort, p.Limit, p.Offset)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractsResponse{
					Error: err.Error(),
				},
			)
		}

		for _, sc := range smartcontracts {
			// get all event count by address
			count, err := ctx.EventStorage.GetEventCountByAddress(sc.Address)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(
					listSmartContractsResponse{
						Error: err.Error(),
					},
				)
			}

			// get events by address
			events, err := ctx.EventStorage.ListEventsByAddress(sc.Address, "desc", count, 0)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(
					listSmartContractsResponse{
						Error: err.Error(),
					},
				)
			}

			// add events to smartcontract
			sc.Events = events
		}

		// get all smartcontracts count
		count, err := ctx.Storage.GetSmartContractsCount()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				listSmartContractsResponse{
					Error: err.Error(),
				},
			)
		}

		// define meta response
		meta := make(map[string]interface{})
		meta["pagination"] = p.GetPaginationMeta(count)

		// prepare response
		return c.Status(fiber.StatusOK).JSON(listSmartContractsResponse{
			Data: smartcontracts,
			Meta: meta,
		})
	}
}
