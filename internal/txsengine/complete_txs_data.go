package txsengine

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"time"

	"github.com/darchlabs/synchronizer-v2/internal/sync/query"
	"github.com/darchlabs/synchronizer-v2/pkg/transaction"
	"github.com/darchlabs/synchronizer-v2/pkg/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

type EthClient interface {
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
}

const (
	WORKERS_COUNT = 1000
)

// TODO(ca): Implement limit below code because has request to the node
// TODO(ca): Implement logic to manage all "continue" cases
func completeContractTxsData(client EthClient, contract *query.SelectSmartContractQueryOutput, transactions []*transaction.Transaction, idGen func() string) ([]*transaction.Transaction, error) {
	// define channels
	jobs := make(chan *transaction.Transaction, len(transactions))
	results := make(chan *transaction.Transaction, len(transactions))

	// start worker goroutines
	for w := 0; w < WORKERS_COUNT; w++ {
		go func() {
			for tx := range jobs {
				tx.ID = idGen()
				tx.ContractID = contract.ID
				tx.UpdatedAt = time.Now()
				tx.CreatedAt = time.Now()
				tx.ChainID = fmt.Sprint(util.SupportedNetworks[string(contract.Network)])

				// parse block number
				blockNum, err := strconv.ParseInt(tx.BlockNumber, 10, 64)
				if err != nil {
					log.Printf("WARNING: Failed to parse block number for transaction %s: %v", tx.Hash, err)
				}

				// get tx fromBalance from node
				fromBalance, err := client.BalanceAt(context.Background(), common.HexToAddress(tx.From), big.NewInt(blockNum))
				if err != nil {
					log.Printf("WARNING: Failed to get fromBalance for transaction %s: %v", tx.Hash, err)
				}
				tx.FromBalance = fromBalance.String()

				// get contract balance from node
				contractBalance, err := client.BalanceAt(context.Background(), common.HexToAddress(contract.Address), big.NewInt(blockNum))
				if err != nil {
					log.Printf("WARNING: Failed to get contract balance for transaction %s: %v", tx.Hash, err)
				} else if contractBalance == nil {
					log.Printf("WARNING: Failed because contractBalance is nil, hash=%s error=%v", tx.Hash, err)
					tx.ContractBalance = "0"
					tx.FromIsWhale = "0"
				} else {
					tx.ContractBalance = contractBalance.String()

					// define if from balance is whale
					whaleLimit := big.NewFloat(params.Ether * 10000)
					whaleLimitInt, _ := whaleLimit.Int64()
					if contractBalance.Int64() > whaleLimitInt {
						tx.FromIsWhale = "1"
					} else {
						tx.FromIsWhale = "0"
					}
				}

				results <- tx
			}
		}()
	}

	// use go routine for send transactions to channel
	go func() {
		for _, tx := range transactions {
			jobs <- tx
		}

		close(jobs)
	}()

	// Wait for all goroutines to complete
	txs := make([]*transaction.Transaction, 0, len(transactions))
	for i := 0; i < len(transactions); i++ {
		tx := <-results
		txs = append(txs, tx)
	}

	return txs, nil
}
