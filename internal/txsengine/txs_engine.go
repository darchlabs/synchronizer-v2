package txsengine

import (
	"context"
	"log"
	"math"
	"strings"
	"time"

	"github.com/darchlabs/synchronizer-v2"
	ethclientrate "github.com/darchlabs/synchronizer-v2/internal/ethclient_rate"
	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/ethereum/go-ethereum/ethclient"
)

type idGenerator func() string

type TxsEngine interface {
	Start(seconds int64)
	Run() error
	Halt()
	GetStatus() StatusEngine
	SetStatus(status StatusEngine)
	GetContractTransactions(contractId string, apiUrl string, apiKey string) error
}

type T struct {
	smartContractStorage    synchronizer.SmartContractStorage
	transactionStorage      synchronizer.TransactionStorage
	status                  StatusEngine
	idGen                   idGenerator
	networksEtherscanURL    map[string]string
	networksEtherscanAPIKey map[string]string
	networksNodesURL        map[string]string
	maxTransactions         int

	client HTTPClient
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

const (
	SCAN_RESPONSE_LIMIT = 10000
	BATCH_TRANSACTIONS  = 10
)

type Config struct {
	ContractStorage    synchronizer.SmartContractStorage
	TransactionStorage synchronizer.TransactionStorage
	IdGen              idGenerator
	EtherscanUrlMap    map[string]string
	ApiKeyMap          map[string]string
	NodesUrlMap        map[string]string
	Client             HTTPClient
	MaxTransactions    int
}

func New(c Config) *T {
	return &T{
		smartContractStorage:    c.ContractStorage,
		transactionStorage:      c.TransactionStorage,
		idGen:                   c.IdGen,
		networksEtherscanURL:    c.EtherscanUrlMap,
		networksEtherscanAPIKey: c.ApiKeyMap,
		networksNodesURL:        c.NodesUrlMap,
		client:                  c.Client,
		maxTransactions:         c.MaxTransactions,

		status: StatusIdle,
	}
}

func (t *T) Start(seconds int64) {
	go func() {
		for t.GetStatus() == StatusIdle || t.GetStatus() == StatusRunning {
			// log.Print("starting ...")
			err := t.Run()
			if err != nil {
				t.SetStatus(StatusError)
			}
			// log.Print("finished!")

			// log.Print("sleeping ...")
			time.Sleep(time.Duration(time.Duration(seconds) * time.Second))
			// log.Print("sleept!")
		}
	}()
}

func (t *T) Run() error {
	if t.GetStatus() == StatusStopped || t.GetStatus() == StatusStopping || t.GetStatus() == StatusError {
		return nil
	}

	// Update enigne to running status when starting it
	t.SetStatus(StatusRunning)

	// Get all the current sc's
	scArr, err := t.smartContractStorage.ListUniqueSmartContractsByNetwork()
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

		// Validate and get etherscan api keys
		etherscanApiUrl, etherscanApiKey, err := checkAndGetApis(contract, t.networksEtherscanURL, t.networksEtherscanAPIKey)
		if err != nil {
			_ = t.smartContractStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusError, err)
			return err
		}

		err = t.GetContractTransactions(contract.ID, etherscanApiUrl, etherscanApiKey)
		if err != nil {
			log.Printf("\nerr: %v on contract: %s", err, contract.Address)
			continue
		}
	}

	return nil
}

// Get status
func (t *T) GetStatus() StatusEngine {
	return t.status
}

// Set status with mutex
func (t *T) SetStatus(status StatusEngine) {
	t.status = status
}

// Stop the tx engine
func (t *T) Halt() {
	t.status = StatusStopped
}

func (t *T) GetContractTransactions(contractId string, apiUrl string, apiKey string) error {
	log.Println("contract started at: ", contractId)

	// get smartcontract with latest data
	contract, err := t.smartContractStorage.GetSmartContractById(contractId)
	if err != nil {
		_ = t.smartContractStorage.UpdateStatusAndError(contractId, smartcontract.StatusError, err)
		return err
	}

	// get current transaction quota for calculate quota
	currentCount, err := t.transactionStorage.GetTxsCountById(contract.ID)
	if err != nil {
		_ = t.smartContractStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusError, err)
		return err
	}

	// check if smartcontract has limit and set Status
	if currentCount >= int64(t.maxTransactions) {
		if contract.Status != smartcontract.StatusQuotaExceeded {
			_ = t.smartContractStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusQuotaExceeded, nil)
		}

		return nil
	}

	// validate and get the node url if node url are not defined
	nodeURL := contract.NodeURL
	if contract.NodeURL == "" {
		nodeURL, err = checkAndGetNodeURL(contract, t.networksNodesURL)
		if err != nil {
			_ = t.smartContractStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusError, err)
			return err
		}
	}

	// create instance client with the node url
	client, err := ethclient.Dial(nodeURL)
	if err != nil {
		_ = t.smartContractStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusError, err)
		return err
	}

	// get last block number
	lastBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		_ = t.smartContractStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusError, err)
		return err
	}

	// Update contract status to synching
	_ = t.smartContractStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusSynching, nil)

	// get transaction from etherscan
	startBlock := contract.LastTxBlockSynced + 1
	transactions, err := t.getTransactionsFromEtherscan(apiUrl, apiKey, contract.Address, startBlock, int64(lastBlock))
	if err != nil && !strings.Contains(err.Error(), "No transactions found") {
		_ = t.smartContractStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusError, err)
		return err
	}

	// when the response from the scan does not have any transactions
	if len(transactions) == 0 {
		_ = t.smartContractStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusRunning, nil)
		t.smartContractStorage.UpdateLastBlockNumber(contract.ID, int64(lastBlock))

		return nil
	}

	// initialize eth client with rate limiter
	clientWithRateLimiter := ethclientrate.NewClient(&ethclientrate.Options{
		MaxRetry:        2,
		MaxRequest:      25,
		WindowInSeconds: 1,
	}, client)

	// prepare iterators
	var from, to, count int
	if BATCH_TRANSACTIONS > len(transactions) {
		to = len(transactions)
	} else {
		to = BATCH_TRANSACTIONS
	}
	to = int(math.Min(float64(t.maxTransactions), float64(to)))

	for {
		// TODO(ca): never enter here bc inside of "completeContractTxsData" only uses continue when have some error
		completedTransactions, err := completeContractTxsData(clientWithRateLimiter, contract, transactions[from:to], t.idGen)
		if err != nil {
			_ = t.smartContractStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusError, err)
			return err
		}

		// insert them in the storage
		err = t.transactionStorage.InsertTxs(completedTransactions)
		if err != nil {
			_ = t.smartContractStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusError, err)
			return err
		}

		// check if count plus transactions in the range exceeds the maximum limit
		count = count + len(transactions[from:to])
		if int(currentCount)+count >= t.maxTransactions {
			_ = t.smartContractStorage.UpdateStatusAndError(contract.ID, smartcontract.StatusQuotaExceeded, nil)
			break
		}

		// check if count exceeds the length of transactions
		if count >= len(transactions) {
			break
		}

		// update from value for iterator
		if from+BATCH_TRANSACTIONS > len(transactions) {
			from = len(transactions)
		} else {
			from = from + BATCH_TRANSACTIONS
		}
		from = int(math.Min(float64(t.maxTransactions), float64(from)))

		// update to value for iterator
		if to+BATCH_TRANSACTIONS > len(transactions) {
			to = len(transactions)
		} else {
			to = to + BATCH_TRANSACTIONS
		}
		to = int(math.Min(float64(t.maxTransactions), float64(to)))
	}

	log.Println("contract finished at: ", contract.Name)
	return nil
}
