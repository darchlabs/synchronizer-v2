package txsengine

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/darchlabs/synchronizer-v2"
	transactionstorage "github.com/darchlabs/synchronizer-v2/internal/storage/transaction"
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
}

type T struct {
	ScStorage               synchronizer.SmartContractStorage
	transactionStorage      *transactionstorage.Storage
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

func New(ss synchronizer.SmartContractStorage, ts *transactionstorage.Storage, idGen idGenerator, etherscanUrlMap map[string]string, etherscanApiKeyMap map[string]string, nodesUrlMap map[string]string) *T {
	return &T{
		ScStorage:               ss,
		transactionStorage:      ts,
		Status:                  StatusIdle,
		idGen:                   idGen,
		NetworksEtherscanURL:    etherscanUrlMap,
		NetworksEtherscanAPIKey: etherscanApiKeyMap,
		NetworksNodesURL:        nodesUrlMap,
	}
}

func (te *T) Start(seconds int64) {
	for te.GetStatus() == StatusIdle || te.GetStatus() == StatusRunning {
		err := te.Run()
		if err != nil {
			te.SetStatus(StatusError)
		}

		time.Sleep(time.Duration(time.Duration(seconds) * time.Second))
	}
}

func (te *T) Run() error {
	if te.GetStatus() == StatusStopped || te.GetStatus() == StatusStopping || te.GetStatus() == StatusError {
		return nil
	}

	// Update to running when starting it
	te.SetStatus(StatusRunning)

	// Get all the current sc's
	scArr, err := te.ScStorage.ListUniqueSmartContractsByNetwork()
	if err != nil {
		return err
	}

	// TODO(nb): use goroutines for executing the smart contracts at the same time
	// Iterate over contracts for getting their tx's
	for _, contract := range scArr {
		log.Println("contract started at: ", contract.Name)
		// If it is stopped, continue with the other contracts
		if contract.Status == smartcontract.StatusStopping || contract.Status == smartcontract.StatusStopped || contract.Status == smartcontract.StatusSynching {
			continue
		}

		etherscanApiUrl, etherscanApiKey, err := util.CheckAndGetApis(contract, te.NetworksEtherscanURL, te.NetworksEtherscanAPIKey)
		if err != nil {
			te.ScStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusError, err)
			continue
		}

		nodeUrl, err := util.CheckAndGetNodeURL(contract, te.NetworksNodesURL)
		if err != nil {
			te.ScStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusError, err)
			continue
		}

		client, err := ethclient.Dial(nodeUrl)
		if err != nil {
			return err
		}

		// Get tx's
		startBlock := contract.LastTxBlockSynced + 1
		transactions, err := GetTransactions(etherscanApiUrl, etherscanApiKey, contract.Address, startBlock)
		if err != nil {
			te.ScStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusError, err)
			continue
		}

		// Manage when it reachs the 10.000 logs limit and get the missing ones
		apiResponseLimit := 10000
		numberOfTxs := len(transactions)

		if numberOfTxs == 0 {
			lastBlock, err := client.BlockNumber(context.Background())
			if err != nil {
				te.ScStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusError, err)
				continue
			}

			te.ScStorage.UpdateLastBlockNumber(contract.ID, int64(lastBlock))
			continue
		}

		if numberOfTxs == apiResponseLimit {
			for numberOfTxs == apiResponseLimit {
				// Get the last block number
				startBlock, err := strconv.ParseInt(transactions[len(transactions)-1].BlockNumber, 10, 64)
				if err != nil {
					continue
				}

				// Get the transactions but starting from the last block number
				newTransactions, err := GetTransactions(etherscanApiUrl, etherscanApiKey, contract.Address, startBlock)
				if err != nil {
					continue
				}

				transactions = append(transactions, newTransactions...)
				numberOfTxs = len(newTransactions)
			}
		}

		// Create an id per txs item
		log.Println("len: txs", len(transactions))
		if len(transactions) < 25000 {
			missingDataCTX := &util.MissingDataCTX{
				Transactions: transactions,
				Contract:     contract,
				Client:       client,
				IdGen:        te.idGen,
			}

			transactions, err = util.CompleteContractTxsData(missingDataCTX)
			if err != nil {
				te.ScStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusError, err)
				continue
			}

		} else {
			missingDataCTX := &util.MissingDataCTX{
				Transactions: transactions,
				Contract:     contract,
				Client:       nil,
				IdGen:        te.idGen,
			}
			transactions, err = util.CompleteContractTxsDataWithoutBalance(missingDataCTX)
			if err != nil {
				te.ScStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusError, err)
				continue
			}

		}

		// Insert them in the storage
		err = te.transactionStorage.InsertTxsByContract(transactions)
		if err != nil {
			te.ScStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusError, err)
			continue
		}

		log.Println("contract finished at: ", contract.Name)
	}

	return nil
}

func GetTransactions(apiUrl string, apiKey string, address string, startBlock int64) ([]*transaction.Transaction, error) {
	var txs []*transaction.Transaction

	type Response struct {
		Result []*transaction.Transaction `json"result"`
	}

	var res Response

	url := fmt.Sprintf("%s?module=account&action=txlist&address=%s&startblock=%s&endblock=99999999&sort=asc&apikey=%s", apiUrl, address, fmt.Sprint(startBlock), apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	txs = res.Result
	return txs, err
}

// Get status
func (te *T) GetStatus() StatusEngine {
	return te.Status
}

// Set status with mutex
func (te *T) SetStatus(status StatusEngine) {
	te.Status = status
}

// Stop the tx engine
func (te *T) Halt() {
	te.Status = StatusStopped
}
