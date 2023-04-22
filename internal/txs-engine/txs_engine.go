package txsengine

import (
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
)

type idGenerator func() string

type TxsEngine interface {
	Start(seconds int64) error
	Halt()
	GetStatus() StatusEngine
	SetStatus(status StatusEngine)
}

type T struct {
	ScStorage          synchronizer.SmartContractStorage
	transactionStorage *transactionstorage.Storage
	Status             StatusEngine
	idGen              idGenerator
	etherscanApiUrl    string
	etherscanApiKey    string
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

func New(ss synchronizer.SmartContractStorage, ts *transactionstorage.Storage, idGen idGenerator, apiUrl string, apiKey string) *T {
	return &T{
		ScStorage:          ss,
		transactionStorage: ts,
		Status:             StatusIdle,
		idGen:              idGen,
		etherscanApiUrl:    apiUrl,
		etherscanApiKey:    apiKey,
	}
}

func (te *T) Start(seconds int64) error {
	if te.GetStatus() == StatusStopped || te.GetStatus() == StatusStopping || te.GetStatus() == StatusError {
		fmt.Println("te.GetStatus(): ", te.GetStatus() != StatusIdle)
		fmt.Println(te.GetStatus())
		return nil
	}

	fmt.Println(1)
	// Update to running when starting it
	te.SetStatus(StatusRunning)

	fmt.Println(2)
	for te.GetStatus() == StatusRunning {

		// While the status is running, the engine'll execute

		fmt.Println(2)
		// Get all the current sc's
		scArr, err := te.ScStorage.ListSmartContracts("asc", 1, 0)
		if err != nil {
			log.Println("there was an error while getting the contracts from the engine")
			log.Panicf("the txs engine can't perform: %v", err)
		}

		// Iterate over contracts for getting their tx's
		for _, contract := range scArr {
			fmt.Println(3)
			// If it is stopped, continue with the other contracts
			if contract.Status == smartcontract.StatusStopping || contract.Status == smartcontract.StatusStopped || contract.Status == smartcontract.StatusSynching {
				continue
			}

			// TODO(nb): Manage some retry's in case the api fails before continuing?
			// Get tx's
			fmt.Println(4)
			startBlock := int64(0)
			transactions, err := GetTransactions(te.etherscanApiUrl, te.etherscanApiKey, contract.Address, startBlock)
			if err != nil {
				continue
			}

			fmt.Println(5)
			// Manage when it reachs the 10.000 logs limit and get the missing ones
			apiResponseLimit := 10000
			numberOfTxs := len(transactions)
			if numberOfTxs == apiResponseLimit {

				fmt.Println(6)
				for numberOfTxs == apiResponseLimit {

					// Get the last block number
					startBlock, err := strconv.ParseInt(transactions[len(transactions)-1].BlockNumber, 10, 64)
					if err != nil {
						continue
					}

					fmt.Println(6.2)
					// Get the transactions but starting from the last block number
					newTransactions, err := GetTransactions(te.etherscanApiUrl, te.etherscanApiKey, contract.Address, startBlock)
					if err != nil {
						continue
					}

					fmt.Println(6.3)
					transactions = append(transactions, newTransactions...)

					fmt.Println(6.5)
					numberOfTxs = len(newTransactions)
					fmt.Println(6.6)

				}
			}

			fmt.Println(7)
			// Create an id per txs item
			transactions, err = util.CompleteTxsDataByContract(transactions, contract, te.idGen)
			if err != nil {
				continue
			}

			fmt.Println(8)
			// Insert them in the storage
			err = te.transactionStorage.InsertTxsByContract(transactions)
			if err != nil {
				fmt.Println("err while inserting: ", err)
				continue
			}
			fmt.Println(9)
		}

		time.Sleep(time.Duration(time.Duration(seconds) * time.Second))
	}

	fmt.Println(10)
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
		fmt.Println("Error getting transactions:", err)
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		fmt.Println("Error unmarshalling response body:", err)
		return nil, err
	}

	txs = res.Result
	fmt.Println("resuilt: ", txs)
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
