package smartcontracts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/darchlabs/synchronizer-v2/pkg/api"
	"github.com/darchlabs/synchronizer-v2/pkg/event"
	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/darchlabs/synchronizer-v2/pkg/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
)

type createSmartContractResponse struct {
	Data  *smartcontract.SmartContract `json:"data"`
	Error string                       `json:"error,omitempty"`
}

func insertSmartContractHandler(ctx *api.Context, c *fiber.Ctx) (interface{}, interface{}, int, error) {
	c.Accepts("application/json")

	// prepate body request struct
	body := struct {
		SmartContract *smartcontract.SmartContract `json:"smartcontract"`
	}{}

	// parse body to smartcontract struct
	err := c.BodyParser(&body)
	if err != nil {
		return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"smartcontracts: insertSmartContractHandler json.Unmarshal error",
		)
	}

	// validate body
	validate := validator.New()
	err = validate.Struct(body.SmartContract)
	if err != nil {
		return nil, nil, fiber.StatusBadRequest, errors.Wrap(
			err,
			"smartcontracts: insertSmartContractHandler validate.Struct error",
		)
	}

	smartContract := body.SmartContract
	smartContract.ID = ctx.IDGen()

	userID, err := api.GetUserIDFromRequestCtx(c)
	if err != nil {
		return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"smartcontracts: insertSmartContractHandler api.GetUserIDFromRequestCtx error",
		)
	}

	// Get contract and check if already exist
	dbContract, _ := ctx.ScStorage.GetSmartContractByAddress(body.SmartContract.Address)
	if dbContract != nil {
		return nil, nil, fiber.StatusBadRequest, errors.Errorf(
			"smartcontracts: insertSmartContractHandler ctx.Storage.GetSmartContractByAddress smartcontract already exists with address=%s error",
			body.SmartContract.Address,
		)
	}

	// get and validate node url
	nodeURL := body.SmartContract.NodeURL
	network := string(body.SmartContract.Network)
	err = util.NodeURLIsValid(nodeURL, network)
	if err != nil {
		networksEtherscanURL, err := util.ParseStringifiedMap(ctx.Env.NetworksNodeURL)
		if err != nil {
			// CHECKPOINT
			return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
				err,
				"smartcontracts: insertSmartContractHandler util.ParseStringifiedMap can't valid ethclient error",
			)
		}

		nodeURL = networksEtherscanURL[network]
	}

	// instance client
	client, err := ethclient.Dial(nodeURL)
	if err != nil {
		return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"smartcontracts: insertSmartContractHandler ethclient.Dial can't valid ethclient error",
		)
	}

	// validate contract exists at the given address
	code, err := client.CodeAt(context.Background(), common.HexToAddress(body.SmartContract.Address), nil)
	if err != nil {
		return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"smartcontracts: insertSmartContractHandler clien.CodeAt can't valid ethclient error",
		)
	}

	if len(code) == 0 {
		return nil, nil, fiber.StatusInternalServerError, errors.New(
			"smartcontracts: insertSmartContractHandler contract does not exist at the given address",
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
					Abi: &event.Abi{
						ID:              a.ID,
						SmartContractID: smartContract.ID,
						Name:            a.Name,
						Type:            a.Type,
						Anonymous:       a.Anonymous,
						Inputs:          a.Inputs,
					},
				},
			}

			b, err := json.Marshal(ev)
			if err != nil {
				return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
					err,
					"smartcontracts: insertSmartContractHandler json.Marshal event error",
				)
			}

			// send post to synchronizers
			url := fmt.Sprintf("%s/api/v1/events/%s", "http://localhost:5555", body.SmartContract.Address)
			res, err := http.Post(url, "application/json", bytes.NewBuffer(b))
			if err != nil {
				return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
					err,
					"smartcontracts: insertSmartContractHandler http.Post error",
				)
			}
			defer res.Body.Close()

			// check status code
			if res.StatusCode != http.StatusOK {
				io.Copy(os.Stdout, res.Body)

				return nil, nil, res.StatusCode, errors.Errorf(
					"smartcontracts: insertSmartContractHandler error creating the event=%s with smartcontract=%s",
					a.Name,
					body.SmartContract.Name,
				)
			}

			// parse response
			response := struct {
				Data *event.Event `json:"data"`
			}{}
			err = json.NewDecoder(res.Body).Decode(&response)
			if err != nil {
				return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
					err,
					"smartcontracts: insertSmartContractHandler json.NewDecoder.Decode error",
				)
			}

			fmt.Println("response.Data.Event", response.Data)

			// add to event and append to slice
			ev.Event = response.Data
			events = append(events, ev.Event)
		}
	}

	// Update smartcontract
	// TODO: define crearly with @cagodoy which fields are mandatory to make it work and prefer
	// definitions like `smartContract := &smartcontract.SmartContract{}` instead multiple assignment
	smartContract.UserID = userID
	smartContract.CreatedAt = ctx.DateGen()
	smartContract.UpdatedAt = ctx.DateGen()
	smartContract.Events = events
	smartContract.Status = smartcontract.StatusIdle
	smartContract.LastTxBlockSynced = int64(0)
	for _, input := range smartContract.Abi {
		input.ID = ctx.IDGen()
	}

	// get and set latest block number from node client
	blockNumber, err := client.BlockNumber(context.Background())
	if err != nil {
		return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"smartcontracts: insertSmartContractHandler client.BlockNumber error",
		)
	}

	smartContract.InitialBlockNumber = int64(blockNumber)
	// save smartcontract struct on database
	err = ctx.ScStorage.InsertSmartContractQuery(smartContract)
	if err != nil {
		return nil, nil, fiber.StatusInternalServerError, errors.Wrap(
			err,
			"smartcontracts: insertSmartContractHandler ctx.Storage.InsertSmartContractQuery error",
		)
	}

	// update response
	smartContract.Events = events

	// prepare response
	// TODO: Check impact of changing StatusOK to StatusCreated since we are creating the smart contract
	return smartContract, nil, fiber.StatusOK, nil
}
