package txsengine

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/darchlabs/synchronizer-v2"
	"github.com/darchlabs/synchronizer-v2/internal/blockchain"
	transactionstorage "github.com/darchlabs/synchronizer-v2/internal/storage/transaction"
	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/darchlabs/synchronizer-v2/pkg/transaction"
	"github.com/darchlabs/synchronizer-v2/pkg/util"
	"github.com/ethereum/go-ethereum/ethclient"
)

type idGenerator func() string

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

	idGen idGenerator
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

func New(ss synchronizer.SmartContractStorage, ts *transactionstorage.Storage, idGen idGenerator) *T {
	return &T{
		ScStorage:          ss,
		transactionStorage: ts,
		Status:             StatusIdle,
		// mu:                  sync.RWMutex{},
		idGen: idGen,
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

		// TODO(nb): Use go routines for managing concurrent sc's ???
		// Iterate over contracts for getting their tx's
		for _, contract := range scArr {
			// If it is stopped, continue with the other contracts
			if contract.Status == smartcontract.StatusStopping || contract.Status == smartcontract.StatusStopped || contract.Status == smartcontract.StatusSynching {
				continue
			}
			var latestBlockNumber int64

			// While retry (on error) is less than the limit, it will keep executing
			retry := 0
			if retry < 5 {
				// Get the contract txs
				// TODO: should also return the latest block num or manage it
				timeBef := time.Now()
				latestBlockNumber, err = te.GetContractTxs(contract)
				if err != nil {
					// Add one to retry in case of error
					retry += 1
				}

				timeAf := time.Now()
				fmt.Println("finished exec time IN SECS: ", timeAf.Second()-timeBef.Second())

				// If there is no error, retry should be 0
				retry = 0
				continue
			}

			// If the retry on error limit was ovecome, update contract status and error
			errorMsg := fmt.Errorf("engine stopped syncing smart contract due to multiple consecutive fails. The last error was: %v", err)
			te.UpdateStateOnErrorByContract(contract.ID, nil, smartcontract.StatusStopped, latestBlockNumber, errorMsg)
		}

		return
	}

}

// It should get the last synce tx, get all of the tx made since then, insert them and update the last synced tx again
func (te *T) GetContractTxs(contract *smartcontract.SmartContract) (int64, error) {

	// TODO(nb): Manage map for the dial client
	client, err := ethclient.Dial("https://patient-delicate-pine.quiknode.pro/4200300eae9e45c661df02030bac8bc34f8b618e/")
	if err != nil {
		return 0, err
	}

	// Get the latest block mined in the blockchain
	toBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		return 0, err
	}

	// Get the last tx from the transactions table FILTER BY CONTRACT ID AND GET THE LAST
	sc, err := te.ScStorage.GetSmartContractByID(contract.ID)
	if err != nil {
		return 0, err
	}

	// Get the last synced tx block number
	lastSyncedTxBlock := sc.LastTxBlockSynced

	// TODO(nb): Get the deployed block number
	// If the last block number of the smart contract table is zero, get the block number from the first emitted log
	if lastSyncedTxBlock == 0 {
		// Get the block number from the first emitted logs of the contract (probably 1st event)
		var maxRetry int64 = 10
		firstEventBlock, err := blockchain.GetFirstLogBlockNum(client, contract.Address, maxRetry)
		if err != nil {
			return int64(firstEventBlock), err
		}

		lastSyncedTxBlock = int64(firstEventBlock)
		// Update the deployment tx on smart contracts table
		err = te.ScStorage.UpdateLastBlockNumber(contract.ID, lastSyncedTxBlock)
		if err != nil {
			return int64(firstEventBlock), err
		}
	}

	// get all of the new txs, starting from the last one we have
	var numberBatches int64 = 10
	missingBlocks := int64(toBlock) - lastSyncedTxBlock
	if missingBlocks == 0 {
		return lastSyncedTxBlock, nil
	}
	batchSize := missingBlocks / numberBatches
	if batchSize == 0 {
		batchSize = 1
	}

	if batchSize > 300 {
		batchSize = 300
	}

	for blockNum := lastSyncedTxBlock; blockNum < int64(toBlock); blockNum += batchSize {

		// Exec a batch of goroutines over the blocks
		var wg sync.WaitGroup

		// If there are not misss
		fmt.Println("-----------------")
		// If there are not misssing blocks, break the loop
		missingBlocks := int64(toBlock) - blockNum
		if missingBlocks == 0 {
			break
		}

		routinesNum := batchSize
		if routinesNum > missingBlocks {
			routinesNum = missingBlocks
		}

		wg.Add(int(routinesNum))
		for i := int64(0); i < int64(routinesNum); i++ {

			currentBlockNum := blockNum + i
			fmt.Println("block n~: ", currentBlockNum)
			if currentBlockNum >= int64(toBlock) {
				continue
			}

			go func() {
				defer wg.Done()

				block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(currentBlockNum)))
				if err != nil {
					te.UpdateStateOnErrorByContract(contract.ID, nil, smartcontract.StatusError, currentBlockNum, err)
					// return blockNum, err
					fmt.Println("error on go routine: ", err)
					return
				}

				blockCTX := &util.BlockCTX{
					IdGen:    util.IdGenerator(te.idGen),
					Client:   client,
					Block:    block,
					Contract: contract,
				}

				transactions, err := util.GetContractTxsByBlock(blockCTX)
				if err != nil {
					te.UpdateStateOnErrorByContract(contract.ID, transactions, smartcontract.StatusError, currentBlockNum, err)
					// return blockNum, err
					return
				}

				// insert them if there are
				if len(transactions) == 0 {
					te.ScStorage.UpdateLastBlockNumber(contract.ID, blockNum)
					return
				}

				err = te.transactionStorage.InsertTxsByContract(transactions, uint64(blockNum))
				if err != nil {
					fmt.Println("here-- err: ", err)
					return
					// return lastSyncedTxBlock, err
				}

			}()

		}
		wg.Wait()

	}

	return lastSyncedTxBlock, nil
}

// It will receive different items as tx of a single contract
func (te *T) UpdateStateOnErrorByContract(id string, transactions []*transaction.Transaction, status smartcontract.SmartContractStatus, block int64, err error) {
	// Insert the current transactions obtained before failling
	if len(transactions) > 0 {
		err := te.transactionStorage.InsertTxsByContract(transactions, uint64(block))
		if err != nil {
			fmt.Println("err in txs: ", err)
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
