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

func CompleteContractTxsData(ctx *MissingDataCTX) ([]*transaction.Transaction, error) {
	id := ctx.Contract.ID
	chainID := fmt.Sprint(SupportedNetworks[string(ctx.Contract.Network)])
	address := ctx.Contract.Address
	whaleLimit := big.NewFloat(params.Ether * 10000)
	whaleLimitInt, _ := whaleLimit.Int64()
	isWhale := "0"

	// Number of worker goroutines
	workerCount := len(ctx.Transactions)
	if workerCount > 1000 {
		workerCount = 1000
	}
	// Create channels for jobs and results
	jobs := make(chan *transaction.Transaction, len(ctx.Transactions))
	results := make(chan *transaction.Transaction, len(ctx.Transactions))

	// Start worker goroutines
	for w := 0; w < workerCount; w++ {
		go func() {
			for tx := range jobs {
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
				tx.ChainID = chainID
				tx.ContractBalance = contractBalance.String()
				tx.FromBalance = fromBalance.String()
				tx.FromIsWhale = isWhale

				results <- tx
			}
		}()
	}

	// Send jobs to the worker goroutines
	go func() {
		for _, tx := range ctx.Transactions {
			jobs <- tx
		}
		close(jobs)
	}()

	// Collect results
	completedTransactions := make([]*transaction.Transaction, 0, len(ctx.Transactions))
	for i := 0; i < len(ctx.Transactions); i++ {
		result := <-results
		if result != nil {
			completedTransactions = append(completedTransactions, result)
		}
	}

	return completedTransactions, nil
}

func CompleteContractTxsDataWithoutBalance(ctx *MissingDataCTX) ([]*transaction.Transaction, error) {
	id := ctx.Contract.ID
	chainID := fmt.Sprint(SupportedNetworks[string(ctx.Contract.Network)])

	// Number of worker goroutines
	workerCount := len(ctx.Transactions)
	if workerCount > 1000 {
		workerCount = 1000
	}
	// Create channels for jobs and results
	jobs := make(chan *transaction.Transaction, len(ctx.Transactions))
	results := make(chan *transaction.Transaction, len(ctx.Transactions))

	// Start worker goroutines
	for w := 0; w < workerCount; w++ {
		go func() {
			for tx := range jobs {

				tx.ID = ctx.IdGen()
				tx.ContractID = id
				tx.UpdatedAt = time.Now()
				tx.CreatedAt = time.Now()
				tx.ChainID = chainID

				results <- tx
			}
		}()
	}

	// Send jobs to the worker goroutines
	go func() {
		for _, tx := range ctx.Transactions {
			jobs <- tx
		}
		close(jobs)
	}()

	// Collect results
	completedTransactions := make([]*transaction.Transaction, 0, len(ctx.Transactions))
	for i := 0; i < len(ctx.Transactions); i++ {
		result := <-results
		if result != nil {
			completedTransactions = append(completedTransactions, result)
		}
	}

	return completedTransactions, nil
}
