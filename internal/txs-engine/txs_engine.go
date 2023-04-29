package txsengine

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/darchlabs/synchronizer-v2"
	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/darchlabs/synchronizer-v2/pkg/transaction"
	"github.com/darchlabs/synchronizer-v2/pkg/util"
	"github.com/ethereum/go-ethereum/ethclient"
)

type idGenerator func() string

type TxsEngine interface {
	Start(seconds int64)
	Run() error
	Halt()
	GetStatus() StatusEngine
	SetStatus(status StatusEngine)
	GetContractTransactions(contract *smartcontract.SmartContract) error
}

type T struct {
	SmartContractStorage    synchronizer.SmartContractStorage
	transactionStorage      synchronizer.TransactionStorage
	Status                  StatusEngine
	idGen                   idGenerator
	NetworksEtherscanURL    map[string]string
	NetworksEtherscanAPIKey map[string]string
	NetworksNodesURL        map[string]string
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

func New(ss *synchronizer.SmartContractStorage, ts *synchronizer.TransactionStorage, idGen idGenerator, etherscanUrlMap map[string]string, etherscanApiKeyMap map[string]string, nodesUrlMap map[string]string) *T {
	return &T{
		SmartContractStorage:    *ss,
		transactionStorage:      *ts,
		Status:                  StatusIdle,
		idGen:                   idGen,
		NetworksEtherscanURL:    etherscanUrlMap,
		NetworksEtherscanAPIKey: etherscanApiKeyMap,
		NetworksNodesURL:        nodesUrlMap,
	}
}

func (t *T) Start(seconds int64) {
	for t.GetStatus() == StatusIdle || t.GetStatus() == StatusRunning {
		log.Print("starting ...")
		err := t.Run()
		if err != nil {
			t.SetStatus(StatusError)
		}
		log.Print("finished!")

		log.Print("sleeping ...")
		time.Sleep(time.Duration(time.Duration(seconds) * time.Second))
		log.Print("sleept!")
	}
}

func (t *T) Run() error {
	if t.GetStatus() == StatusStopped || t.GetStatus() == StatusStopping || t.GetStatus() == StatusError {
		return nil
	}

	// Update enigne to running status when starting it
	t.SetStatus(StatusRunning)

	// Get all the current sc's
	scArr, err := t.SmartContractStorage.ListUniqueSmartContractsByNetwork()
	if err != nil {
		return err
	}

	// TODO(nb): use goroutines for executing the smart contracts at the same time
	// Iterate over contracts for getting their tx's
	for _, contract := range scArr {
		// If it is stopped, return err with the other contracts
		if contract.Status != smartcontract.StatusIdle && contract.Status != smartcontract.StatusRunning {
			continue
		}

		err = t.GetContractTransactions(contract)
		if err != nil {
			fmt.Printf("\nerr: %v on contract: %s", err, contract.Address)
			continue
		}

	}

	return nil
}

// Get status
func (t *T) GetStatus() StatusEngine {
	return t.Status
}

// Set status with mutex
func (t *T) SetStatus(status StatusEngine) {
	t.Status = status
}

// Stop the tx engine
func (t *T) Halt() {
	t.Status = StatusStopped
}

func (t *T) GetContractTransactions(contract *smartcontract.SmartContract) error {
	log.Println("contract started at: ", contract.Name)
	// Update contract status to synching
	t.SmartContractStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusSynching, nil)

	// Validate and get etherscan api keys
	etherscanApiUrl, etherscanApiKey, err := util.CheckAndGetApis(contract, t.NetworksEtherscanURL, t.NetworksEtherscanAPIKey)
	if err != nil {
		t.SmartContractStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusError, err)
		return err
	}

	// Validate and get the node url
	nodeUrl, err := util.CheckAndGetNodeURL(contract, t.NetworksNodesURL)
	if err != nil {
		t.SmartContractStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusError, err)
		return err
	}

	// Instance client with the node url
	client, err := ethclient.Dial(nodeUrl)
	if err != nil {
		t.SmartContractStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusError, err)
		return err
	}

	// Get tx's
	startBlock := contract.LastTxBlockSynced + 1
	transactions, err := util.GetTransactionsFromEtherscan(etherscanApiUrl, etherscanApiKey, contract.Address, startBlock)
	if err != nil {
		t.SmartContractStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusError, err)
		return err
	}

	// Manage when it reachs the 10.000 logs limit and get the missing ones
	apiResponseLimit := 10000
	numberOfTxs := len(transactions)

	if numberOfTxs == 0 {
		lastBlock, err := client.BlockNumber(context.Background())
		if err != nil {
			t.SmartContractStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusError, err)
			return err
		}

		t.SmartContractStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusRunning, nil)
		t.SmartContractStorage.UpdateLastBlockNumber(contract.ID, int64(lastBlock))
		return err
	}

	if numberOfTxs == apiResponseLimit {
		for numberOfTxs == apiResponseLimit {
			// Get the last block number
			startBlock, err := strconv.ParseInt(transactions[len(transactions)-1].BlockNumber, 10, 64)
			if err != nil {
				return err
			}

			// Get the transactions but starting from the last block number
			newTransactions, err := util.GetTransactionsFromEtherscan(etherscanApiUrl, etherscanApiKey, contract.Address, startBlock)
			if err != nil {
				return err
			}

			transactions = append(transactions, newTransactions...)
			numberOfTxs = len(newTransactions)
		}
	}

	txsWithBalances := transactions
	var txsWithoutBalances []*transaction.Transaction

	txNumberLimit := 5000
	if len(transactions) > txNumberLimit {
		txsWithBalances = transactions[0:5000]
		txsWithoutBalances = transactions[5000:]
	}

	// Complete the data (calculating also the balance for those which don't overpass the limit)
	missingDataCTX := &util.MissingDataCTX{
		Transactions: txsWithBalances,
		Contract:     contract,
		Client:       client,
		IdGen:        t.idGen,
	}
	txsWithBalances, err = util.CompleteContractTxsData(missingDataCTX)
	if err != nil {
		// If there is an error, update it
		t.SmartContractStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusError, err)
		// Update the txs without balances arr with the total transactions arr
		txsWithoutBalances = transactions
	}

	if len(txsWithoutBalances) > 0 {
		missingDataCTX = &util.MissingDataCTX{
			Transactions: txsWithoutBalances,
			Contract:     contract,
			Client:       nil,
			IdGen:        t.idGen,
		}
		txsWithoutBalances, err = util.CompleteContractTxsDataWithoutBalance(missingDataCTX)
		if err != nil {
			t.SmartContractStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusError, err)
			return err
		}
	}

	var completedTxs []*transaction.Transaction
	completedTxs = append(completedTxs, txsWithBalances...)
	completedTxs = append(completedTxs, txsWithoutBalances...)

	// Insert them in the storage
	err = t.transactionStorage.InsertTxsByContract(completedTxs)
	if err != nil {
		t.SmartContractStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusError, err)
		return err
	}

	log.Println("contract finished at: ", contract.Name)
	return nil
}
