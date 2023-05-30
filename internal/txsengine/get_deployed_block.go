package txsengine

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func getDeployedBlockNumber(client *ethclient.Client, address common.Address, toBlock uint64) (int64, error) {
	startBlock := uint64(0) // Replace with a reasonable starting block number

	// Iterate through blocks
	for blockNumber := startBlock; blockNumber <= toBlock; blockNumber++ {
		blockNumberBigInt := big.NewInt(int64(blockNumber))

		// Fetch the contract code at the given block number
		code, err := client.CodeAt(context.Background(), address, blockNumberBigInt)
		if err != nil {
			return 0, err
		}

		// Check if the contract code exists at the given block number
		if len(code) > 0 {
			return int64(blockNumber), nil
		}
	}

	return 0, fmt.Errorf("%s", "The contract does not exist")
}
