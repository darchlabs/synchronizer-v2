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

		// Update event
		body.Event.Address = address
		body.Event.ID = ctx.IDGen()
		body.Event.Abi.ID = ctx.IDGen()
		body.Event.LatestBlockNumber = 0
		body.Event.Status = event.StatusSynching
		body.Event.Error = ""
		body.Event.CreatedAt = ctx.DateGen()
		body.Event.UpdatedAt = ctx.DateGen()
		for _, input := range body.Event.Abi.Inputs {
			input.ID = ctx.IDGen()
		}

		// Validate abi is not nil
		if body.Event.Abi == nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(api.Response{
				Error: "abi cannot be nil",
			})
		}

		// Validate network is one of the supported
		network := body.Event.Network
		if !event.IsValidEventNetwork(network) {
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
		createdEvent, err := ctx.Storage.InsertEvent(body.Event)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				api.Response{
					Error: err.Error(),
				},
			)
		}

		// get initial logs in background
		go func() {
			// get created event from database
			ev, err := ctx.Storage.GetEvent(body.Event.Address, body.Event.Abi.Name)
			if err != nil {
				// update event error
				ev.Status = event.StatusError
				ev.Error = err.Error()
				ev.UpdatedAt = ctx.DateGen()
				_ = ctx.Storage.UpdateEvent(ev)
				return
			}

			// parse abi to string
			b, err := json.Marshal(ev.Abi)
			if err != nil {
				// update event error
				ev.Status = event.StatusError
				ev.Error = err.Error()
				ev.UpdatedAt = ctx.DateGen()
				_ = ctx.Storage.UpdateEvent(ev)
				return
			}

			// define and read channel with log data in go routine
			logsChannel := make(chan []blockchain.LogData)
			go func() {
				for logs := range logsChannel {
					log.Printf("received logs len_data=%+v \n", len(logs))

					// parse each log to EventData
					eventDatas := make([]*event.EventData, 0)
					for _, log := range logs {
						ed := &event.EventData{}
						err := ed.FromLogData(log, ctx.IDGen(), ev.ID, ctx.DateGen())
						if err != nil {
							// update event error
							ev.Status = event.StatusError
							ev.Error = err.Error()
							ev.UpdatedAt = ctx.DateGen()
							_ = ctx.Storage.UpdateEvent(ev)
							return
						}

						eventDatas = append(eventDatas, ed)
					}

					// insert logs data to event
					err := ctx.Storage.InsertEventData(ev, eventDatas)
					if err != nil {
						// update event error
						ev.Status = event.StatusError
						ev.Error = err.Error()
						ev.UpdatedAt = ctx.DateGen()
						_ = ctx.Storage.UpdateEvent(ev)
						return
					}

					// update latest block number using last dataLog
					if len(logs) > 0 {
						ev.LatestBlockNumber = int64(logs[len(logs)-1].BlockNumber)
						ev.UpdatedAt = ctx.DateGen()
						err = ctx.Storage.UpdateEvent(ev)
						if err != nil {
							// update event error
							ev.Status = event.StatusError
							ev.Error = err.Error()
							ev.UpdatedAt = ctx.DateGen()
							_ = ctx.Storage.UpdateEvent(ev)
							return
						}
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
				ev.Status = event.StatusError
				ev.Error = err.Error()
				ev.UpdatedAt = ctx.DateGen()
				_ = ctx.Storage.UpdateEvent(ev)
				return
			}

			// finish when the contract dont have new events
			if count > 0 {
				log.Printf("%d new events have been inserted into the database with %d latest block number \n", count, latestBlockNumber)
			}

			// update latest block number in event
			ev.LatestBlockNumber = latestBlockNumber
			ev.UpdatedAt = ctx.DateGen()
			err = ctx.Storage.UpdateEvent(ev)
			if err != nil {
				// update event error
				ev.Status = event.StatusError
				ev.Error = err.Error()
				ev.UpdatedAt = ctx.DateGen()
				_ = ctx.Storage.UpdateEvent(ev)
				return
			}

			// update event status to running
			ev.Status = event.StatusRunning
			ev.Error = ""
			ev.UpdatedAt = ctx.DateGen()
			_ = ctx.Storage.UpdateEvent(ev)
			if err != nil {
				// update event error
				ev.Status = event.StatusError
				ev.Error = err.Error()
				ev.UpdatedAt = ctx.DateGen()
				_ = ctx.Storage.UpdateEvent(ev)
				return
			}
		}()

		// prepare response
		return c.Status(fiber.StatusOK).JSON(api.Response{
			Data: createdEvent,
		})
	}
}
