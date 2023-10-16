package blockchain

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	customlogger "github.com/darchlabs/synchronizer-v2/internal/custom-logger"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Config struct {
	Client          *ethclient.Client
	ABI             string
	Address         string
	FromBlockNumber *int64
	ToBlockNumber   *uint64
	MaxRetry        int64
	LogsChannel     chan []LogData
	Logger          bool
}

type LogData struct {
	Tx          common.Hash            `json:"tx"`
	BlockNumber uint64                 `json:"blockNumber"`
	Data        map[string]interface{} `json:"data"`

	EventName string
}

func GetLogs(ctx context.Context, c Config) (int64, int64, error) {
	// check config params
	if c.Client == nil {
		return 0, 0, errors.New("invalid Client config param")
	}
	if c.ABI == "" {
		return 0, 0, errors.New("invalid ABI config param")
	}
	if c.Address == "" {
		return 0, 0, errors.New("invalid Address config param")
	}
	if c.FromBlockNumber == nil {
		return 0, 0, errors.New("invalid FromBlock config param")
	}
	if c.ToBlockNumber != nil && *c.FromBlockNumber > int64(*c.ToBlockNumber) {
		return 0, 0, errors.New("invalid ToBlockNumber number because is lower than FromBlockNumber")
	}
	if c.LogsChannel == nil {
		return 0, 0, errors.New("invalid LogsChannel config param")
	}

	log, err := customlogger.NewCustomLogger("green", os.Stdout)
	if err != nil {
		panic(err)
	}

	// prepare contract instance using ABI definition
	contractWithAbi, err := abi.JSON(strings.NewReader(c.ABI))
	if err != nil {
		return 0, 0, err
	}

	// iterate over contractWithAbi events
	eventsIDs := make([]common.Hash, 0)
	eventsIdToName := make(map[string]string)
	for _, event := range contractWithAbi.Events {
		eventsIDs = append(eventsIDs, event.ID)
		eventsIdToName[event.ID.String()] = event.Name
	}

	// define from block and interval numbers
	logsCount := int64(0)
	interval := int64(0)
	fromBlock := int64(*c.FromBlockNumber)
	retry := int64(0)

	// set toBlock and temporalToBlock using config or lastest value from node
	var toBlock int64
	if c.ToBlockNumber == nil {
		// get current latest block from
		blockNumber, err := c.Client.BlockNumber(context.Background())
		if err != nil {
			return 0, 0, err
		}
		toBlock = int64(blockNumber)
	} else {
		toBlock = int64(*c.ToBlockNumber)
	}
	temporalToBlock := toBlock

	// define values to manage the ticker
	// TODO(ca): should to implement rate limit approach
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

loop:
	for count := 0; ; count++ {
		select {
		case <-ctx.Done():
			// close log channel and finish
			log.Println("indexer - context Done() signal received")
			close(c.LogsChannel)
			return logsCount, int64(temporalToBlock - interval), ctx.Err()
		case <-ticker.C:
			// prepare query params
			query := ethereum.FilterQuery{
				FromBlock: big.NewInt(fromBlock),
				ToBlock:   big.NewInt(temporalToBlock),
				Addresses: []common.Address{
					common.HexToAddress(c.Address),
				},
				Topics: [][]common.Hash{eventsIDs},
			}

			// get logs from contract
			logs, err := c.Client.FilterLogs(context.Background(), query)
			if err != nil {
				// TODO(ca): should implement a better recongnition for node log limit error
				//
				// Infura(Polygon): query returned more than 10000 results
				//
				// Alchemy(Ethereum): Log response size exceeded. You can make eth_getLogs requests
				// with up to a 2K block range and no limit on the response size, or you can request
				// any block range with a cap of 10K logs in the response. Based on your parameters
				// and the response size limit, this block range should work: [0x0, 0x88c025]
				//
				// Alchemy(Polygon): Query git out exceeded. Consider reducing your block range. Based
				// on your parameters and the response size limit, this block range should work: [0x0, 0x1360b8a]
				//
				// QuickNode(Polygon): 413 Request Entity Too Large: {"jsonrpc":"2.0","id":2,"result":null,"error":
				// {"code":-32602,"message":"eth_getLogs and eth_newFilter are limited to a 10,000 blocks range"}}
				//
				if strings.Contains(err.Error(), "returned") || strings.Contains(err.Error(), "exceeded") || strings.Contains(err.Error(), "reducing") || strings.Contains(err.Error(), "limited") {
					interval = (temporalToBlock - fromBlock) / 2
					temporalToBlock = temporalToBlock - interval

					continue
				}

				// retry process
				retry++
				if retry > c.MaxRetry {
					return 0, 0, fmt.Errorf("indexer - error retrying to get logs from contract, retry: %d, error: %s", retry, err.Error())
				} else {
					log.Printf("indexer - c.Client.FilterLogs(context.Background(), query), err%s", err.Error())
				}

				continue
			}

			// define reset and data log slice
			retry = 0
			data := make([]LogData, 0)

			if c.Logger {
				log.Printf("indexer - address: %s iteration: %d from: %d to: %d interval: %d logs: %d", c.Address[:6]+"..."+c.Address[len(c.Address)-5:], count, fromBlock, temporalToBlock, interval, len(logs))
			}

			// iterate over logs
			for _, vLog := range logs {
				// continue if event data are empty
				if len(vLog.Data) == 0 {
					continue
				}

				// get event name
				eventName, ok := eventsIdToName[vLog.Topics[0].String()]
				if !ok {
					// show warning message and continue
					log.Printf("indexer - warning event_name: %s not found in eventsIdToName", vLog.Topics[0].String())
					continue
				}

				// get event from contract log
				eventData := make(map[string]interface{})
				err := contractWithAbi.UnpackIntoMap(eventData, eventName, vLog.Data)
				if err != nil {
					return 0, 0, err
				}

				// filter only indexed elements from events inputs
				indexedInputs := make([]abi.Argument, 0)
				for _, e := range contractWithAbi.Events[eventName].Inputs {
					if e.Indexed {
						indexedInputs = append(indexedInputs, e)
					}
				}

				// get indexed topics from log and parse to map
				topics := make(map[string]interface{})
				err = abi.ParseTopicsIntoMap(topics, indexedInputs, vLog.Topics[1:])
				if err != nil {
					return 0, 0, err
				}

				// iterate indexed topics and add to eventData map
				for key, t := range topics {
					eventData[key] = t
				}

				// prepare event data
				d := LogData{
					Tx:          vLog.TxHash,
					BlockNumber: vLog.BlockNumber,
					Data:        eventData,
					EventName:   eventName,
				}

				// append log in data log slice and increase the counter
				data = append(data, d)
				logsCount++
			}

			// continue if data log slice is empty
			if len(data) > 0 {
				c.LogsChannel <- data
			}

			// condition for finish the bucle
			if temporalToBlock == int64(toBlock) {
				break loop
			}

			// add interval value to fromBlock and toBlock numbers
			// TODO(ca): maybe is interval + 1
			fromBlock = fromBlock + interval
			if temporalToBlock+interval > int64(toBlock) {
				temporalToBlock = int64(toBlock)
			} else {
				temporalToBlock = temporalToBlock + interval
			}
		}
	}

	// close log channel
	close(c.LogsChannel)
	return logsCount, int64(temporalToBlock), nil
}
