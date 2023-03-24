package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/darchlabs/synchronizer-v2/internal/blockchain"
	"github.com/darchlabs/synchronizer-v2/pkg/api"
	"github.com/darchlabs/synchronizer-v2/pkg/event"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
)

func insertEventHandler(ctx Context) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		// prepate body request struct
		body := struct {
			Event *event.Event `json:"event"`
		}{}

		// parse body to event struct
		err := json.Unmarshal(c.Body(), &body)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(api.Response{
				Error: err.Error(),
			})
		}

		// get, valid and set address to event struct
		address := c.Params("address")
		if address == "" {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(api.Response{
				Error: "invalid param",
			})
		}

		// Validate body
		validate := validator.New()
		err = validate.Struct(body.Event)
		if err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(api.Response{
				Error: err.Error(),
			})
		}

		// Update address
		body.Event.Address = address
		body.Event.Status = event.StatusSynching

		// Validate abi is not nil
		if body.Event.Abi == nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(api.Response{
				Error: "abi cannot be nil",
			})
		}

		// Validate network is one of the supported
		network := body.Event.Network
		if network != event.Ethereum && network != event.Polygon {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(api.Response{
				Error: "invalid network",
			})
		}

		// Validate node url is correct
		nodeURL := body.Event.NodeURL
		if nodeURL == "" {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(api.Response{
				Error: "invalid nodeURL",
			})
		}

		// get or create eth client in client
		client, ok := (*ctx.Clients)[nodeURL]
		if !ok {
			// validate client works
			client, err = ethclient.Dial(nodeURL)
			if err != nil {
				return fmt.Errorf("can't getting ethclient error=%s", err)
			}

			// validate client is working correctly
			_, err = client.ChainID(context.Background())
			if err != nil {
				return fmt.Errorf("can't valid ethclient error=%s", err)
			}

			// save client in map
			(*ctx.Clients)[nodeURL] = client
		}

		// save event struct on database
		err = ctx.Storage.InsertEvent(body.Event)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				api.Response{
					Error: err.Error(),
				},
			)
		}

		// get initial logs in background
		go func() {
			ev := body.Event

			// parse abi to string
			b, err := json.Marshal(ev.Abi)
			if err != nil {
				// update event error
				_ = ev.UpdateStatus(event.StatusError, err, ctx.Storage)
				return
			}

			// define and read channel with log data in go routine
			logsChannel := make(chan []blockchain.LogData)
			go func() {
				for logs := range logsChannel {
					log.Printf("received logs data=%+v \n", logs)

					// insert logs data to event
					_, err := ev.InsertData(logs, ctx.Storage)
					if err != nil {
						// update event error
						_ = ev.UpdateStatus(event.StatusError, err, ctx.Storage)
						return
					}
				}
			}()

			// get event logs from contract
			count, latestBlockNumber, err := blockchain.GetLogs(blockchain.Config{
				Client:          client,
				ABI:             fmt.Sprintf("[%s]", string(b)),
				EventName:       ev.Abi.Name,
				Address:         ev.Address,
				FromBlockNumber: &ev.LatestBlockNumber,
				LogsChannel:     logsChannel,
				Logger:          true,
			})
			if err != nil {
				// update event error
				_ = ev.UpdateStatus(event.StatusError, err, ctx.Storage)
				return
			}

			// finish when the contract dont have new events
			if count > 0 {
				log.Printf("%d new events have been inserted into the database with %d latest block number \n", count, latestBlockNumber)
			}

			// update latest block number in event
			err = ev.UpdateLatestBlock(latestBlockNumber, ctx.Storage)
			if err != nil {
				// update event error
				_ = ev.UpdateStatus(event.StatusError, err, ctx.Storage)
				return
			}

			// update event status to running
			err = ev.UpdateStatus(event.StatusRunning, err, ctx.Storage)
			if err != nil {
				// update event error
				_ = ev.UpdateStatus(event.StatusError, err, ctx.Storage)
				return
			}
		}()

		// prepare response
		return c.Status(fiber.StatusOK).JSON(api.Response{
			Data: body.Event,
		})
	}
}
