package util

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/darchlabs/synchronizer-v2/pkg/transaction"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
)

type MissingDataCTX struct {
	Transactions []*transaction.Transaction
	Contract     *smartcontract.SmartContract
	Client       *ethclient.Client
	IdGen        func() string
}

// TODO(nb): Manage getting the data in goroutines
func CompleteContractTxsData(ctx *MissingDataCTX) ([]*transaction.Transaction, error) {
	id := ctx.Contract.ID
	address := ctx.Contract.Address
	chainID := fmt.Sprint(SupportedNetworks[string(ctx.Contract.Network)])
	whaleLimit := big.NewFloat(params.Ether * 10000)
	whaleLimitInt, _ := whaleLimit.Int64()
	isWhale := "0"

	for _, tx := range ctx.Transactions {
		blockNum, err := strconv.ParseInt(tx.BlockNumber, 10, 64)
		if err != nil {
			continue
		}

		fromBalance, err := ctx.Client.BalanceAt(context.Background(), common.HexToAddress(tx.From), big.NewInt(blockNum))
		if err != nil {
			continue
		}

		contractBalance, err := ctx.Client.BalanceAt(context.Background(), common.HexToAddress(address), big.NewInt(blockNum))
		if err != nil {
			continue
		}

		if contractBalance.Int64() > whaleLimitInt {
			isWhale = "1"
		}

		tx.ID = ctx.IdGen()
		tx.ContractID = id
		tx.UpdatedAt = time.Now()
		tx.CreatedAt = time.Now()
		tx.ContractBalance = contractBalance.String()
		tx.FromBalance = fromBalance.String()
		tx.FromIsWhale = isWhale
		tx.ChainID = chainID
	}

	return ctx.Transactions, nil
}

func CompleteContractTxsDataWithoutBalance(ctx *MissingDataCTX) ([]*transaction.Transaction, error) {
	id := ctx.Contract.ID
	chainID := fmt.Sprint(SupportedNetworks[string(ctx.Contract.Network)])

	for _, tx := range ctx.Transactions {
		tx.ID = ctx.IdGen()
		tx.ContractID = id
		tx.UpdatedAt = time.Now()
		tx.CreatedAt = time.Now()
		tx.ChainID = chainID
	}

	return ctx.Transactions, nil
}

// func CompleteContractTxsDataWithoutBalance(transactions []*transaction.Transaction, contract *smartcontract.SmartContract, client *ethclient.Client, idGen func() string) ([]*transaction.Transaction, error) {
// 	id := contract.ID
// 	address := contract.Address

// 	// Number of worker goroutines
// 	workerCount := 1000
// 	// Create channels for jobs and results
// 	jobs := make(chan *transaction.Transaction, len(transactions))
// 	results := make(chan *transaction.Transaction, len(transactions))

// 	// Start worker goroutines
// 	for w := 0; w < workerCount; w++ {
// 		go func() {
// 			for tx := range jobs {
// 				fmt.Println("wcount: ", w)

// 				blockNum, err := strconv.ParseInt(tx.BlockNumber, 10, 64)
// 				if err != nil {
// 					results <- nil
// 					continue
// 				}

// 				contractBalance, err := client.BalanceAt(context.Background(), common.HexToAddress(address), big.NewInt(blockNum))
// 				if err != nil {
// 					results <- nil
// 					continue
// 				}

// 				tx.ID = idGen()
// 				tx.ContractID = id
// 				tx.UpdatedAt = time.Now()
// 				tx.CreatedAt = time.Now()
// 				tx.ContractBalance = contractBalance.String()

// 				results <- tx
// 			}
// 		}()
// 	}

// 	// Send jobs to the worker goroutines
// 	go func() {
// 		for _, tx := range transactions {
// 			jobs <- tx
// 		}
// 		close(jobs)
// 	}()

// 	// Collect results
// 	completedTransactions := make([]*transaction.Transaction, 0, len(transactions))
// 	for i := 0; i < len(transactions); i++ {
// 		result := <-results
// 		if result != nil {
// 			completedTransactions = append(completedTransactions, result)
// 		}
// 	}

// 	return completedTransactions, nil
// }
