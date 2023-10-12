package smartcontracts

// import (
// 	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
// 	"github.com/gofiber/fiber/v2"
// )

// type deleteSmartContractResponse struct {
// 	Data  *smartcontract.SmartContract `json:"data,omitempty"`
// 	Error string                       `json:"error,omitempty"`
// }

// func deleteSmartContractHandler(ctx Context) func(c *fiber.Ctx) error {
// 	return func(c *fiber.Ctx) error {
// 		c.Accepts("application/json")

// 		address := c.Params("address")
// 		if address == "" {
// 			return c.Status(fiber.StatusOK).JSON(deleteSmartContractResponse{
// 				Error: "address cannot be nil",
// 			})
// 		}

// 		// get smart contract
// 		sc, err := ctx.Storage.GetSmartContractByAddress(address)
// 		if err != nil {
// 			return c.Status(fiber.StatusBadRequest).JSON(
// 				createSmartContractResponse{
// 					Error: err.Error(),
// 				},
// 			)
// 		}

// 		// get all event count by address
// 		count, err := ctx.EventStorage.GetEventCountByAddress(sc.Address)
// 		if err != nil {
// 			return c.Status(fiber.StatusInternalServerError).JSON(
// 				listSmartContractsResponse{
// 					Error: err.Error(),
// 				},
// 			)
// 		}

// 		// get events by address
// 		events, err := ctx.EventStorage.ListEventsByAddress(sc.Address, "desc", count, 0)
// 		if err != nil {
// 			return c.Status(fiber.StatusInternalServerError).JSON(
// 				listSmartContractsResponse{
// 					Error: err.Error(),
// 				},
// 			)
// 		}

// 		// iterate over events and delete from sychornizer service
// 		for _, e := range events {
// 			err := ctx.EventStorage.DeleteEvent(e.Address, e.Abi.Name)
// 			if err != nil {
// 				return c.Status(fiber.StatusBadRequest).JSON(
// 					createSmartContractResponse{
// 						Error: err.Error(),
// 					},
// 				)
// 			}
// 		}

// 		// delete smartcontract from database
// 		err = ctx.Storage.DeleteSmartContractByAddress(address)
// 		if err != nil {
// 			return c.Status(fiber.StatusBadRequest).JSON(
// 				createSmartContractResponse{
// 					Error: err.Error(),
// 				},
// 			)
// 		}

// 		return c.Status(fiber.StatusOK).JSON(struct{}{})
// 	}
// }
