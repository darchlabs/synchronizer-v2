package util

import (
	"context"
	"math/big"
	"strconv"
	"time"

	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/darchlabs/synchronizer-v2/pkg/transaction"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
)

// TODO(nb): Manage getting the data in goroutines if there is a lot of data
func CompleteTxsDataByContract(transactions []*transaction.Transaction, contract *smartcontract.SmartContract, idGen func() string) ([]*transaction.Transaction, error) {
	id := contract.ID
	address := contract.Address

	client, err := ethclient.Dial(contract.NodeURL)
	if err != nil {
		return nil, err
	}

	whaleLimit := big.NewFloat(params.Ether * 10000)
	whaleLimitInt, _ := whaleLimit.Int64()
	isWhale := "0"

	for _, tx := range transactions {
		blockNum, err := strconv.ParseInt(tx.BlockNumber, 10, 64)
		if err != nil {
			continue
		}

		fromBalance, err := client.BalanceAt(context.Background(), common.HexToAddress(tx.From), big.NewInt(blockNum))
		if err != nil {
			continue
		}

		contractBalance, err := client.BalanceAt(context.Background(), common.HexToAddress(address), big.NewInt(blockNum))
		if err != nil {
			continue
		}

		if contractBalance.Int64() > whaleLimitInt {
			isWhale = "1"
		}

		tx.ID = idGen()
		tx.ContractID = id
		tx.UpdatedAt = time.Now()
		tx.CreatedAt = time.Now()
		tx.ContractBalance = contractBalance.String()
		tx.FromBalance = fromBalance.String()
		tx.FromIsWhale = isWhale
	}

	return transactions, nil
}
