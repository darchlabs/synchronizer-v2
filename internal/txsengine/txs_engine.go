package txsengine

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/darchlabs/synchronizer-v2"
	customlogger "github.com/darchlabs/synchronizer-v2/internal/custom-logger"
	ethclientrate "github.com/darchlabs/synchronizer-v2/internal/ethclient_rate"
	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/ethereum/go-ethereum/ethclient"
)

var EMPTY_ERROR_MESSAGE = ""

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

	log *customlogger.CustomLogger
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
	Log                *customlogger.CustomLogger
}

func New(c Config) *T {
	// initialize custom logger
	customLogger, err := customlogger.NewCustomLogger("purple", os.Stdout)
	if err != nil {
		panic(err)
	}

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
		log:    customLogger,
	}
}

func (t *T) Start(seconds int64) {
	t.log.Printf("transactions - running ticker each %d seconds", seconds)
	go func() {
		for t.GetStatus() == StatusIdle || t.GetStatus() == StatusRunning {
			err := t.Run()
			if err != nil {
				t.SetStatus(StatusError)
			}

			time.Sleep(time.Duration(time.Duration(seconds) * time.Second))
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
		if contract.EngineStatus != smartcontract.StatusIdle && contract.EngineStatus != smartcontract.StatusRunning {
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
			t.log.Printf("transactions - error getting transactions for contract %s", contract.ID)
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
	var err error

	// get smartcontract with latest data
	contract, err := t.smartContractStorage.GetSmartContractById(contractId)
	if err != nil {
		t.log.Printf("transactions - error getting smartcontract %s", err.Error())
		return err
	}

	// defer update status and error
	defer func() {
		if err != nil {
			t.log.Printf("transactions - defer error: %s", err.Error())

			errStr := err.Error()
			_, err = t.smartContractStorage.UpdateSmartContract(&smartcontract.SmartContract{
				ID:           contract.ID,
				Address:      contract.Address,
				EngineError:  &errStr,
				EngineStatus: smartcontract.StatusError,
			})
			if err != nil {
				t.log.Printf("transactions - defer updating smartcontract error: %s", err.Error())
			}
		}
	}()

	// get current transaction quota for calculate quota
	currentCount, err := t.transactionStorage.GetTxsCountById(contract.ID)
	if err != nil {
		return err
	}

	// check if smartcontract has limit and set Status
	if currentCount >= int64(t.maxTransactions) {
		if contract.EngineStatus != smartcontract.StatusQuotaExceeded {
			return fmt.Errorf("transactions - quota exceeded for contract %s", contract.ID)
		}

		return nil
	}

	// validate and get the node url if node url are not defined
	nodeURL := contract.NodeURL
	if contract.NodeURL == "" {
		nodeURL, err = checkAndGetNodeURL(contract, t.networksNodesURL)
		if err != nil {
			return err
		}
	}

	// create instance client with the node url
	client, err := ethclient.Dial(nodeURL)
	if err != nil {
		return err
	}

	// get last block number from etherscan
	// note: Before, the current block number of
	// the node was used, but we realized that the
	// etherscan will always be out of date
	lastBlock, err := t.GetCurrentBlockNumberFromEtherscan(apiUrl, apiKey)
	if err != nil {
		return err
	}

	// Update contract status to synching
	_, err = t.smartContractStorage.UpdateSmartContract(&smartcontract.SmartContract{
		ID:           contract.ID,
		Address:      contract.Address,
		EngineError:  &EMPTY_ERROR_MESSAGE,
		EngineStatus: smartcontract.StatusSynching,
	})
	if err != nil {
		return err
	}

	// BORRE EL + 1 POR CASO BORDE SI QUEDAN TXS RESTANTES EN UN BLOQUE POR LIMITE
	// get transaction from etherscan
	startBlock := contract.EngineLastTxBlockSynced
	transactions, err := t.getTransactionsFromEtherscan(apiUrl, apiKey, contract.Address, startBlock, int64(lastBlock))
	if err != nil && !strings.Contains(err.Error(), "No transactions found") {
		return err
	}

	t.log.Printf("transactions - address: %s from: %d to: %d txs: %d", contract.Address[:6]+"..."+contract.Address[len(contract.Address)-5:], startBlock, lastBlock, len(transactions))

	// SI ANDA CON TX, EVENTOS CAGA, SE DUPLICA TRIPLICA
	// HAY QUE IGNORAR LAS QUE YA ESTAN EN LA BASE DE DATOS
	// - event datas salen duplicados
	// - similar en webhooks

	// check if there are no transactions and update smartcontract
	if len(transactions) == 0 {
		contract.EngineError = nil
		contract.EngineStatus = smartcontract.StatusRunning
		contract.EngineLastTxBlockSynced = int64(lastBlock)
		_, err = t.smartContractStorage.UpdateSmartContract(contract)
		if err != nil {
			return err
		}

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

	for i := 0; ; i++ {
		// TODO(ca): never enter here bc inside of "completeContractTxsData" only uses continue when have some error
		completedTransactions, err := completeContractTxsData(clientWithRateLimiter, contract, transactions[from:to], t.idGen)
		if err != nil {
			return err
		}

		// insert them in the storage
		err = t.transactionStorage.InsertTxs(completedTransactions)
		if err != nil {
			return err
		}

		// check if count plus transactions in the range exceeds the maximum limit
		count = count + len(transactions[from:to])
		if int(currentCount)+count >= t.maxTransactions {

			_, err = t.smartContractStorage.UpdateSmartContract(&smartcontract.SmartContract{
				ID:           contract.ID,
				Address:      contract.Address,
				EngineError:  &EMPTY_ERROR_MESSAGE,
				EngineStatus: smartcontract.StatusQuotaExceeded,
			})
			if err != nil {
				return err
			}

			break
		}

		// check if count exceeds the length of transactions
		if count >= len(transactions) {
			break
		}

		// update from value for iterator
		if from+BATCH_TRANSACTIONS >= len(transactions) {
			from = len(transactions)
		} else {
			from = from + BATCH_TRANSACTIONS
		}
		from = int(math.Min(float64(t.maxTransactions-count), float64(from)))

		// update to value for iterator
		if to+BATCH_TRANSACTIONS >= len(transactions) {
			to = len(transactions)
		} else {
			to = to + BATCH_TRANSACTIONS
		}
		to = int(math.Min(float64(t.maxTransactions-count), float64(to)))
	}

	// update smartcontract when finished successfully

	_, err = t.smartContractStorage.UpdateSmartContract(&smartcontract.SmartContract{
		ID:                      contract.ID,
		Address:                 contract.Address,
		EngineError:             &EMPTY_ERROR_MESSAGE,
		EngineStatus:            smartcontract.StatusRunning,
		EngineLastTxBlockSynced: int64(lastBlock),
	})

	return nil
}
