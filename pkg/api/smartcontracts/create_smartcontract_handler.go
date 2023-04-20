package smartcontracts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/darchlabs/synchronizer-v2/internal/blockchain"
	"github.com/darchlabs/synchronizer-v2/pkg/event"
	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
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

		// validate client works
		client, err := ethclient.Dial(body.SmartContract.NodeURL)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				createSmartContractResponse{
					Error: fmt.Sprintf("can't valid ethclient error=%s", err),
				},
			)
		}

		// validate client is working correctly
		_, err = client.ChainID(context.Background())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				createSmartContractResponse{
					Error: fmt.Sprintf("can't valid ethclient error=%s", err),
				},
			)
		}

		// Validate contract exists on the given address
		code, err := client.CodeAt(context.Background(), common.HexToAddress(body.SmartContract.Address), nil)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				createSmartContractResponse{
					Error: fmt.Sprintf("can't get contract code at given address error=%s", err),
				},
			)
		}

		if len(code) == 0 {
			return c.Status(fiber.StatusInternalServerError).JSON(
				createSmartContractResponse{
					Error: fmt.Sprint("the contract does not exist at the given address"),
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

				// CLIENTE NODEJS -> GOLANG

				// send post to synchronizers
				url := fmt.Sprintf("http://localhost:%s/api/v1/events/%s", ctx.Env.Port, body.SmartContract.Address)
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
				response := &createEventResponse{}
				err = json.NewDecoder(res.Body).Decode(response)
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

		fmt.Println("body sc: ", body.SmartContract)
		fmt.Println("body sc: upat", body.SmartContract.UpdatedAt)

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

		go func() {
			// Get the block number from the first emitted logs of the contract (probably 1st event)
			var maxRetry int64 = 10
			firstEventBlock, err := blockchain.GetFirstLogBlockNum(client, body.SmartContract.Address, maxRetry)
			// This error only breaks the go routine and won't be returned since it will be finished after sending the response to the client
			if err != nil {
				fmt.Println("err in create sc handler: ", err)
				return
			}

			err = ctx.Storage.UpdateLastBlockNumber(body.SmartContract.ID, int64(firstEventBlock))
			// This error only breaks the go routine and won't be returned since it will be finished after sending the response to the client
			if err != nil {
				fmt.Println("err in create sc handler: ", err)
				return
			}
		}()

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
