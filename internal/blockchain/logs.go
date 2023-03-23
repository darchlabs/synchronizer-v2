package blockchain

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Config struct {
	Client          *ethclient.Client
	ABI             string
	EventName       string
	Address         string
	FromBlockNumber *int64
	ToBlockNumber   *int64
	MaxRetry        int64
	LogsChannel     chan []LogData
	Logger          bool
}

type LogData struct {
	Tx          common.Hash            `json:"tx"`
	BlockNumber uint64                 `json:"blockNumber"`
	Data        map[string]interface{} `json:"data"`
}

func GetLogs(c Config) (int64, int64, error) {
	// check config params
	if c.Client == nil {
		return 0, 0, errors.New("invalid Client config param")
	}
	if c.ABI == "" {
		return 0, 0, errors.New("invalid ABI config param")
	}
	if c.EventName == "" {
		return 0, 0, errors.New("invalid EventName config param")
	}
	if c.Address == "" {
		return 0, 0, errors.New("invalid Address config param")
	}
	if c.FromBlockNumber == nil {
		return 0, 0, errors.New("invalid FromBlock config param")
	}
	if c.ToBlockNumber != nil && *c.FromBlockNumber > *c.ToBlockNumber {
		return 0, 0, errors.New("invalid ToBlockNumber number because is lower than FromBlockNumber")
	}
	if c.LogsChannel == nil {
		return 0, 0, errors.New("invalid LogsChannel config param")
	}

	// prepare contract instance using ABI definition
	contractWithAbi, err := abi.JSON(strings.NewReader(c.ABI))
	if err != nil {
		return 0, 0, err
	}

	// get event definition for getting event id
	event, ok := contractWithAbi.Events[c.EventName]
	if !ok {
		return 0, 0, fmt.Errorf("event_name=%s is not defined in abi", c.EventName)
	}

	// define from block and interval numbers
	logsCount := int64(0)
	// latestBlockNumber := int64(0)
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
		toBlock = *c.ToBlockNumber
	}
	temporalToBlock := toBlock

	// we need to request log by batches using interval block number
	if c.Logger {
		log.Printf("\nmaking batches requests for event_name%s", c.EventName)
	}
	for count := 0; ; count++ {
		if c.Logger {
			log.Printf("\naddress=%s event_name=%s iteration=%d from=%d to=%d interval=%d ", c.Address, c.EventName, count, fromBlock, temporalToBlock, interval)
		}

		// prepare query params
		query := ethereum.FilterQuery{
			FromBlock: big.NewInt(fromBlock),
			ToBlock:   big.NewInt(temporalToBlock),
			Addresses: []common.Address{
				common.HexToAddress(c.Address),
			},
			Topics: [][]common.Hash{{event.ID}},
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
				return 0, 0, fmt.Errorf("max_retry, error=%s", err.Error())
			}

			continue
		}

		// define reset and data log slice
		retry = 0
		data := make([]LogData, 0)

		if c.Logger {
			log.Printf("logs=%d", len(logs))
		}

		// iterate over logs
		for _, vLog := range logs {
			// continue if event data are empty
			if len(vLog.Data) == 0 {
				continue
			}

			// get event from contract log
			eventData := make(map[string]interface{})
			err := contractWithAbi.UnpackIntoMap(eventData, c.EventName, vLog.Data)
			if err != nil {
				return 0, 0, err
			}

			// filter only indexed elements from events inputs
			indexedInputs := make([]abi.Argument, 0)
			for _, e := range contractWithAbi.Events[c.EventName].Inputs {
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
			}

			// append log in data log slice and increase the counter
			data = append(data, d)
			logsCount++
		}

		// send log data to channel
		c.LogsChannel <- data

		// condition for finish the bucle
		if temporalToBlock == int64(toBlock) {
			break
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

	if c.Logger {
		log.Printf("\n")
	}

	// close log channel
	close(c.LogsChannel)

	return logsCount, int64(toBlock), nil
}
