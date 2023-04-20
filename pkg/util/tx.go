package util

import (
	"context"
	"fmt"
	"time"

	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/darchlabs/synchronizer-v2/pkg/transaction"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
)

type IdGenerator func() string

type BlockCTX struct {
	IdGen    IdGenerator
	Client   *ethclient.Client
	Block    *types.Block
	Contract *smartcontract.SmartContract
}

func GetContractTxsByBlock(ctx *BlockCTX) ([]*transaction.Transaction, error) {
	var contractTransactions []*transaction.Transaction
	for _, tx := range ctx.Block.Transactions() {

		// TODO(nb): refactor it dividing in another function
		if tx.To() != nil && *tx.To() == common.HexToAddress(ctx.Contract.Address) {
			fromAddress, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)

			if err != nil {

				if len(contractTransactions) > 0 {
					return contractTransactions, err
				}

			}

			fromBalance, err := ctx.Client.BalanceAt(context.Background(), fromAddress, nil)
			if err != nil {
				return contractTransactions, err
			}

			// get contract balance
			contractBalance, err := ctx.Client.BalanceAt(context.Background(), *tx.To(), nil)
			if err != nil {
				return contractTransactions, err
			}

			// Check if the sender is a whale
			tenThousandEther := params.Ether * 100000
			isWhale := fromBalance.Int64() > int64(tenThousandEther)

			// Get the transaction receipt and checkk if the t succeded or not
			receipt, err := ctx.Client.TransactionReceipt(context.Background(), tx.Hash())
			if err != nil {
				return contractTransactions, err
			}
			txSucceded := receipt.Status == 1

			newTx := &transaction.Transaction{
				ID:              ctx.IdGen(),
				ContractID:      ctx.Contract.ID,
				Hash:            tx.Hash().String(),
				FromAddr:        fromAddress.String(),
				FromBalance:     fromBalance.String(),
				FromIsWhale:     isWhale,
				ContractBalance: contractBalance.String(),
				GasPaid:         fmt.Sprint(tx.Gas()),
				GasPrice:        tx.GasPrice().String(),
				GasCost:         tx.Cost().String(),
				Succeded:        txSucceded,
				BlockNumber:     ctx.Block.Number().Int64(),
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}

			fmt.Println("newTX: ", newTx)
			contractTransactions = append(contractTransactions, newTx)

		}

	}
	return contractTransactions, nil
}
