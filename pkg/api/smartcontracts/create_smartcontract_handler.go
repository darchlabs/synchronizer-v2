package smartcontracts

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/darchlabs/synchronizer-v2/pkg/event"
	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/darchlabs/synchronizer-v2/pkg/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
)

type createSmartContractResponse struct {
	Data  *smartcontract.SmartContract `json:"data,omitempty"`
	Error string                       `json:"error,omitemty"`
}

type createEventResponse struct {
	Data *event.Event `json:"data"`
}

func insertSmartContractHandler(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		// prepate body request struct
		body := struct {
			SmartContract *smartcontract.SmartContract `json:"smartcontract"`
		}{}

		// parse body to smartcontract struct
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(
				createSmartContractResponse{
					Error: err.Error(),
				},
			)
		}

		// validate body
		validate := validator.New()
		err = validate.Struct(body.SmartContract)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(
				createSmartContractResponse{
					Error: err.Error(),
				},
			)
		}

		// get and validate node url
		nodeUrl := body.SmartContract.NodeURL
		network := string(body.SmartContract.Network)
		err = util.NodeURLIsValid(nodeUrl, network)
		if err != nil {
			networksEtherscanURL, err := util.ParseStringifiedMap(ctx.Env.NetworksNodeURL)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(
					createSmartContractResponse{
						Error: fmt.Sprintf("can't valid ethclient error=%s", err),
					},
				)
			}

			nodeUrl = networksEtherscanURL[network]
		}

		// instance client
		client, err := ethclient.Dial(nodeUrl)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				createSmartContractResponse{
					Error: fmt.Sprintf("can't valid ethclient error=%s", err),
				},
			)
		}

		// validate contract exists at the given address
		code, err := client.CodeAt(context.Background(), common.HexToAddress(body.SmartContract.Address), nil)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				createSmartContractResponse{
					Error: fmt.Sprintf("can't validate contract exists error=%s", err),
				},
			)
		}

		if len(code) == 0 {
			return c.Status(fiber.StatusInternalServerError).JSON(
				createSmartContractResponse{
					Error: "contract does not exist at the given address",
				},
			)
		}

		// Declare variables that could be updated for the contract if it already exists
		lastTxBlockSynced := int64(0)
		var events []*event.Event
		status := smartcontract.StatusIdle

		// Get contract
		dbContract, _ := ctx.Storage.GetSmartContractByAddress(body.SmartContract.Address)
		// if contract exists, update the variables with the ones we already have
		if dbContract != nil {
			lastTxBlockSynced = dbContract.LastTxBlockSynced
			events = dbContract.Events
			status = dbContract.Status
		}

		// Update smartcontract
		body.SmartContract.ID = ctx.IDGen()
		body.SmartContract.CreatedAt = ctx.DateGen()
		body.SmartContract.UpdatedAt = ctx.DateGen()
		body.SmartContract.Events = events
		body.SmartContract.Status = status
		body.SmartContract.LastTxBlockSynced = lastTxBlockSynced

		for _, input := range body.SmartContract.Abi {
			input.ID = ctx.IDGen()
		}

		// save smartcontract struct on database
		createdSmartContract, err := ctx.Storage.InsertSmartContract(body.SmartContract)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				createSmartContractResponse{
					Error: err.Error(),
				},
			)
		}

		// update response
		createdSmartContract.Events = events

		// prepare response
		return c.Status(fiber.StatusOK).JSON(struct {
			Data *smartcontract.SmartContract `json:"data"`
		}{
			Data: createdSmartContract,
		})
	}
}
