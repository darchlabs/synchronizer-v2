package util

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func GetDeployedBlockNumber(client *ethclient.Client, address common.Address, toBlock int64) (int64, error) {
	startBlock := int64(0) // Replace with a reasonable starting block number
	var wg sync.WaitGroup
	requests := make(chan int64, 100)
	result := make(chan int64, 1)

	// Create 10 goroutines to execute requests
	for i := 0; i < 100; i++ {

		wg.Add(1)
		go func() {

			for blockNumber := range requests {
				time.Sleep(1 * time.Second)
				fmt.Println("block num: ", blockNumber)
				blockNumberBigInt := big.NewInt(int64(blockNumber))

				// Fetch the contract code at the given block number
				code, err := client.CodeAt(context.Background(), address, blockNumberBigInt)
				if err != nil {
					fmt.Println("err: ", err)
					result <- 0
					break
				}

				// Check if the contract code exists at the given block number
				if len(code) > 0 {
					result <- int64(blockNumber)
					break
				}
			}
			wg.Done()
		}()
	}

	// Send requests to the channel
	for blockNumber := startBlock; blockNumber <= toBlock; blockNumber++ {

		requests <- blockNumber
	}
	close(requests)

	// Wait for the first result to be received or all goroutines to finish
	select {
	case blockNumber := <-result:
		// Cancel pending requests
		for len(requests) > 0 {
			<-requests
		}
		return blockNumber, nil
	case <-time.After(30 * time.Second):
		// Timeout
		return 0, fmt.Errorf("%s", "Timeout waiting for the contract code to be found")
	}

	// Wait for all goroutines to finish
	wg.Wait()

	return 0, fmt.Errorf("%s", "The contract does not exist")
}
