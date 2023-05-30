package smartcontracts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/darchlabs/synchronizer-v2/pkg/event"
	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/darchlabs/synchronizer-v2/pkg/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
)

type createSmartContractResponse struct {
	Data  *smartcontract.SmartContract `json:"data"`
	Error string                       `json:"error,omitempty"`
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

		// Get contract and check if already exist
		dbContract, _ := ctx.Storage.GetSmartContractByAddress(body.SmartContract.Address)
		if dbContract != nil {
			return c.Status(fiber.StatusBadRequest).JSON(
				createSmartContractResponse{
					Error: fmt.Sprintf("smartcontract already exists with address=%s", body.SmartContract.Address),
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

		// filter abi events from body
		events := make([]*event.Event, 0)
		for _, a := range body.SmartContract.Abi {
			if a.Type == "event" {
				// define new event
				ev := struct {
					Event *event.Event `json:"event"`
				}{
					Event: &event.Event{
						Network: body.SmartContract.Network,
						NodeURL: body.SmartContract.NodeURL,
						Address: body.SmartContract.Address,
						Abi:     a,
					},
				}

				b, err := json.Marshal(ev)
				if err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(
						createSmartContractResponse{
							Error: err.Error(),
						},
					)
				}

				// send post to synchronizers
				url := fmt.Sprintf("%s/api/v1/events/%s", "http://localhost:5555", body.SmartContract.Address)
				res, err := http.Post(url, "application/json", bytes.NewBuffer(b))
				if err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(
						createSmartContractResponse{
							Error: err.Error(),
						},
					)
				}
				defer res.Body.Close()

				// parse response
				response := struct {
					Data *event.Event `json:"data"`
				}{}

				err = json.NewDecoder(res.Body).Decode(&response)
				if err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(
						createSmartContractResponse{
							Error: err.Error(),
						},
					)
				}

				fmt.Println("response.Data.Event", response.Data)

				// add to event and append to slice
				ev.Event = response.Data
				events = append(events, ev.Event)
			}
		}

		// Update smartcontract
		body.SmartContract.ID = ctx.IDGen()
		body.SmartContract.CreatedAt = ctx.DateGen()
		body.SmartContract.UpdatedAt = ctx.DateGen()
		body.SmartContract.Events = events
		body.SmartContract.Status = smartcontract.StatusIdle
		body.SmartContract.LastTxBlockSynced = int64(0)
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
