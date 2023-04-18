package txsengine

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/darchlabs/synchronizer-v2"
	"github.com/darchlabs/synchronizer-v2/internal/blockchain"
	transactionstorage "github.com/darchlabs/synchronizer-v2/internal/storage/transaction"
	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/darchlabs/synchronizer-v2/pkg/transaction"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
)

type idGenerator func() string
type dateGenerator func() time.Time

type TxsEngine interface {
	Run() error
	Halt()
	GetStatus() StatusEngine
	SetStatus(status StatusEngine)
	ExecEngine()
	GetContractTxs(contract *smartcontract.SmartContract) error
}

type T struct {
	ScStorage          synchronizer.SmartContractStorage
	transactionStorage *transactionstorage.Storage
	Status             StatusEngine
	// mu                  sync.RWMutex

	idGen   idGenerator
	dateGen dateGenerator
}

// Define the enigne status
type StatusEngine string

const (
	StatusIdle     StatusEngine = "idle"
	StatusRunning  StatusEngine = "running"
	StatusStopping StatusEngine = "stopping"
	StatusStopped  StatusEngine = "stopped"
	StatusError    StatusEngine = "error"
)

func New(ss synchronizer.SmartContractStorage, ts *transactionstorage.Storage, idGen idGenerator, dateGen dateGenerator) *T {
	return &T{
		ScStorage:          ss,
		transactionStorage: ts,
		Status:             StatusIdle,
		// mu:                  sync.RWMutex{},
		idGen:   idGen,
		dateGen: dateGen,
	}
}

// Run txs engine process
func (te *T) Run() {
	// Check if the engine status is running
	if te.GetStatus() == StatusIdle || te.GetStatus() == StatusRunning {
		// Exec engine
		te.ExecEngine()
	}
}

func (te *T) ExecEngine() {
	// Update to running when starting it
	te.SetStatus(StatusRunning)

	// While the status is running, the engine'll execute
	for te.GetStatus() == StatusRunning {

		// Get all the current sc's
		scArr, err := te.ScStorage.ListSmartContracts("asc", 1, 0)
		if err != nil {
			log.Println("there was an error while getting the contracts from the engine")
			log.Panicf("the txs engine can't perform: %v", err)
		}

		// TODO(nb): Use go routines for managing concurrent sc's
		for _, contract := range scArr {
			// If it is stopped, continue with the other contracts
			if contract.Status == smartcontract.StatusStopping || contract.Status == smartcontract.StatusStopped {
				continue
			}

			// While retry (on error) is less than the limit, it will keep executing
			retry := 0
			if retry < 5 {
				// Get the contract txs
				err := te.GetContractTxs(contract)

				if err != nil {
					// Add one to retry in case of error
					retry += 1
				}

				// If there is no error, retry should be 0
				retry = 0
				continue
			}

			// If the retry on error limit was ovecome, update contract status and error
			errorMsg := fmt.Errorf("engine stopped syncing smart contract due to multiple consecutive fails. The last error was: %v", err)
			te.UpdateStateOnErrorByContract(contract.ID, nil, smartcontract.StatusStopped, errorMsg)
		}

		return
	}
}

// It should get the last synce tx, get all of the tx made since then, insert them and update the last synced tx again
func (te *T) GetContractTxs(contract *smartcontract.SmartContract) error {
	// Define the tx array for those that will be inserted in the table
	var transactions []*transaction.Transaction

	// TODO(nb): Manage memory for the dial client
	client, err := ethclient.Dial(contract.NodeURL)
	if err != nil {
		return err
	}

	// Get the latest block mined in the blockchain
	toBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		return err
	}

	// Get the last tx from the transactions table FILTER BY CONTRACT ID AND GET THE LAST
	txs, err := te.transactionStorage.ListTxs("DESC", 1, 0)
	if err != nil {
		return err
	}

	//	TODO(nb): refactor this
	// Get the last synced tx block number
	var lastSyncedTxBlock int64
	// If the last block number of the smart contract table is zero, get it's deployed block number
	if len(txs) != 0 {
		// Get the last block from the transactions table
		lastSyncedTxBlock = txs[len(txs)-1].BlockNumber

	} else if contract.LastTxBlockSynced != 0 {
		lastSyncedTxBlock = contract.LastTxBlockSynced
	} else {
		//TODO(nb): Get the deployed block number
		// Get the block number from the first emitted logs of the contract (probably 1st event)
		var maxRetry int64 = 10
		firstEventBlock, err := blockchain.GetFirstLogBlockNum(client, contract.Address, maxRetry)
		if err != nil {
			return err
		}

		lastSyncedTxBlock = int64(firstEventBlock)
		// Update the deployment tx on smart contracts table
		err = te.ScStorage.UpdateLastBlockNumber(contract.ID, lastSyncedTxBlock)
		if err != nil {
			return err
		}
	}

	// get all of the new txs, starting from the last one we have
	//todo: Manage retry
	for block := lastSyncedTxBlock; block < int64(toBlock); block++ {
		fmt.Println("block n~: ", block)
		block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(block)))
		if err != nil {
			te.UpdateStateOnErrorByContract(contract.ID, transactions, smartcontract.StatusError, err)
			return err
		}

		// TODO(nb): Improve efficiency, evaluate goroutines usage
		for _, tx := range block.Transactions() {

			// TODO(nb): refactor it dividing in another function
			if tx.To() != nil && *tx.To() == common.HexToAddress(contract.Address) {
				fromAddress, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)

				if err != nil {

					if len(transactions) > 0 {
						te.UpdateStateOnErrorByContract(contract.ID, transactions, smartcontract.StatusError, err)
					}

					return err
				}

				fromBalance, err := client.BalanceAt(context.Background(), fromAddress, nil)
				if err != nil {
					te.UpdateStateOnErrorByContract(contract.ID, transactions, smartcontract.StatusError, err)
					return err
				}

				// get contract balance
				contractBalance, err := client.BalanceAt(context.Background(), *tx.To(), nil)
				if err != nil {
					te.UpdateStateOnErrorByContract(contract.ID, transactions, smartcontract.StatusError, err)
					return err
				}

				// Check if the sender is a whale
				oneEther := params.Ether

				whaleLimit := oneEther * 10000
				isWhale := fromBalance.Int64() > int64(whaleLimit)

				// Get the transaction receipt and checkk if the t succeded or not
				receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
				if err != nil {
					te.UpdateStateOnErrorByContract(contract.ID, transactions, smartcontract.StatusError, err)
					return err
				}
				txSucceded := receipt.Status == 1

				newTx := &transaction.Transaction{
					ID:              te.idGen(),
					ContractID:      contract.ID,
					Hash:            tx.Hash().String(),
					FromAddr:        fromAddress.String(),
					FromBalance:     fromBalance.String(),
					FromIsWhale:     isWhale,
					ContractBalance: contractBalance.String(),
					GasPaid:         fmt.Sprint(tx.Gas()),
					GasPrice:        tx.GasPrice().String(),
					GasCost:         tx.Cost().String(),
					Succeded:        txSucceded,
					BlockNumber:     block.Number().Int64(),
					CreatedAt:       te.dateGen(),
					UpdatedAt:       te.dateGen(),
				}

				transactions = append(transactions, newTx)
			}
		}
	}

	fmt.Println("here--")
	// insert them if there are
	if len(transactions) == 0 {
		te.ScStorage.UpdateLastBlockNumber(contract.ID, int64(toBlock))
		return nil
	}

	err = te.transactionStorage.InsertTxsByContract(transactions)
	if err != nil {
		fmt.Println("here-- err: ", err)
		return err
	}

	return nil
}

// It will receive different items as tx of a single contract
func (te *T) UpdateStateOnErrorByContract(id string, transactions []*transaction.Transaction, status smartcontract.SmartContractStatus, err error) {
	// Insert the current transactions obtained before failling
	if len(transactions) > 0 {
		err := te.transactionStorage.InsertTxsByContract(transactions)
		if err != nil {
		}
	}

	contractID := id

	// Update smart contract status and error fields
	te.ScStorage.UpdateStatusAndError(contractID, status, err)
}

// Get status
func (te *T) GetStatus() StatusEngine {
	// te.mu.RLock()
	// defer te.mu.RUnlock()
	return te.Status
}

// Set status with mutex
func (te *T) SetStatus(status StatusEngine) {
	// te.mu.Lock()
	// defer te.mu.Unlock()
	te.Status = status
}

// Stop the tx engine
func (te *T) Halt() {
	// te.mu.Lock()
	// defer te.mu.Unlock()
	te.Status = StatusStopped
}
