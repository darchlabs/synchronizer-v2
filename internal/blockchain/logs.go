package blockchain

import (
	"context"
	"errors"
	"fmt"
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
}

type LogData struct {
	Tx          common.Hash            `json:"tx"`
	BlockNumber uint64                 `json:"blockNumber"`
	Data        map[string]interface{} `json:"data"`
}

func GetLogs(c Config) ([]LogData, int64, error) {
	fmt.Println(1)
	// check config params
	if c.Client == nil {
		return nil, 0, errors.New("invalid Client config param")
	}
	if c.ABI == "" {
		return nil, 0, errors.New("invalid ABI config param")
	}
	if c.EventName == "" {
		return nil, 0, errors.New("invalid EventName config param")
	}
	if c.Address == "" {
		return nil, 0, errors.New("invalid Address config param")
	}

	fmt.Println(2)
	// prepare form block number params
	var from *big.Int
	if c.FromBlockNumber != nil {
		from = big.NewInt(*c.FromBlockNumber)
	} else {
		from = nil
	}

	fmt.Println(3)
	// prepare query params
	query := ethereum.FilterQuery{
		FromBlock: from,
		ToBlock:   nil,
		Addresses: []common.Address{
			common.HexToAddress(c.Address),
		},
	}

	fmt.Println(3.5)

	// code, err := c.Client.CodeAt(context.Background(),
	// 	common.HexToAddress(c.Address),
	// 	nil)
	// if err != nil {
	// 	fmt.Println("error in code at ---> :", err)
	// }

	fmt.Println(4)
	// get logs from contract
	logs, err := c.Client.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, 0, err
	}

	fmt.Println(5)
	// prepare contract instance using ABI definition
	contractWithAbi, err := abi.JSON(strings.NewReader(c.ABI))
	if err != nil {
		return nil, 0, err
	}
	fmt.Println(6)

	// prepare data slice
	data := make([]LogData, 0)

	fmt.Println(7)
	// iterate over logs
	latestBlockNumber := int64(0)
	for _, vLog := range logs {
		// get event from contract log
		eventData := make(map[string]interface{})

		err := contractWithAbi.UnpackIntoMap(eventData, c.EventName, vLog.Data)
		// err := contractWithAbi.UnpackIntoMap(eventData, c.EventName, []byte(c.ABI))
		if err != nil {
			return nil, 0, err
		}

		fmt.Println(8)
		// filter only indexed elements from events inputs
		indexedInputs := make([]abi.Argument, 0)
		for _, e := range contractWithAbi.Events[c.EventName].Inputs {
			if e.Indexed {
				indexedInputs = append(indexedInputs, e)
			}
		}

		fmt.Println(9)
		// get indexed topics from log and parse to map
		topics := make(map[string]interface{})
		err = abi.ParseTopicsIntoMap(topics, indexedInputs, vLog.Topics[1:])
		if err != nil {
			return nil, 0, err
		}

		fmt.Printf("Logs: %+v", logs)
		fmt.Println(10)
		// iterate indexed topics and add to eventData map
		for key, t := range topics {
			eventData[key] = t
		}

		fmt.Println(11)
		// prepare event data
		d := LogData{
			Tx:          vLog.TxHash,
			BlockNumber: vLog.BlockNumber,
			Data:        eventData,
		}

		fmt.Println(12)
		// append event data to slice
		data = append(data, d)

		fmt.Println(13)
		// check if current block number is greater than global counter
		if int64(vLog.BlockNumber) > latestBlockNumber {
			latestBlockNumber = int64(vLog.BlockNumber)
		}
	}

	return data, latestBlockNumber, nil
}
