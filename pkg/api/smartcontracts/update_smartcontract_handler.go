package smartcontracts

import (
	"encoding/json"

	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
)

type updateSmartContractResponse struct {
	Data  *smartcontract.SmartContract `json:"data"`
	Error string                       `json:"error,omitempty"`
}

func updateSmartContractHandler(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		address := c.Params("address")
		if address == "" {
			return c.Status(fiber.StatusOK).JSON(updateSmartContractResponse{
				Error: "address cannot be nil",
			})
		}

		// prepate body request struct
		body := struct {
			SmartContract *smartcontract.SmartContractUpdate `json:"smartcontract"`
		}{}

		// parse body to smartcontract struct
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(
				updateSmartContractResponse{
					Error: err.Error(),
				},
			)
		}

		// validate body
		validate := validator.New()
		err = validate.Struct(body.SmartContract)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(
				updateSmartContractResponse{
					Error: err.Error(),
				},
			)
		}

		// Get contract and check if already exist
		updatedSC, err := ctx.Storage.UpdateSmartContract(&smartcontract.SmartContract{
			Address: address,
			Name:    body.SmartContract.Name,
			NodeURL: body.SmartContract.NodeURL,
			Webhook: body.SmartContract.Webhook,
		})
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(
				updateSmartContractResponse{
					Error: err.Error(),
				},
			)
		}

		// prepare response
		return c.Status(fiber.StatusOK).JSON(updateSmartContractResponse{
			Data: updatedSC,
		})
	}
}
