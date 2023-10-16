package smartcontractstorage

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/darchlabs/synchronizer-v2"
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
)

type Storage struct {
	storage            *storage.S
	eventStorage       synchronizer.EventStorage
	transactionStorage synchronizer.TransactionStorage
}

func New(s *storage.S, e synchronizer.EventStorage, t synchronizer.TransactionStorage) *Storage {
	return &Storage{
		storage:            s,
		eventStorage:       e,
		transactionStorage: t,
	}
}

func (s *Storage) InsertSmartContract(sc *smartcontract.SmartContract) (*smartcontract.SmartContract, error) {
	// get current sc
	current, _ := s.GetSmartContractByAddress(sc.Address)
	if current != nil {
		return nil, fmt.Errorf("smartcontract already exists with address=%s", sc.Address)
	}

	// insert new smartcontract in database
	var smartcontractId string
	query := "INSERT INTO smartcontracts ( id, name, network, node_url, address,last_tx_block_synced, status, error, webhook, created_at, updated_at, initial_block_number) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id"
	err := s.storage.DB.Get(&smartcontractId, query, sc.ID, sc.Name, sc.Network, sc.NodeURL, sc.Address,
		sc.LastTxBlockSynced, sc.Status, sc.Error, sc.Webhook, sc.CreatedAt, sc.UpdatedAt, sc.InitialBlockNumber)
	if err != nil {
		return nil, err
	}

	// get created smartcontract
	createdSmartcontract, err := s.GetSmartContractById(smartcontractId)
	if err != nil {
		return nil, err
	}

	return createdSmartcontract, nil
}

func (s *Storage) UpdateSmartContract(sc *smartcontract.SmartContract) (*smartcontract.SmartContract, error) {
	// get current sc
	current, err := s.GetSmartContractByAddress(sc.Address)
	if err != nil {
		return nil, fmt.Errorf("smartcontract not found with address=%s", sc.Address)
	}

	// prepare dinamic sql
	query := "UPDATE smartcontracts SET "
	args := []interface{}{}

	// check if name is changed
	if sc.Name != current.Name && sc.Name != "" {
		query += "name = $" + strconv.Itoa(len(args)+1) + ", "
		args = append(args, sc.Name)
	}

	// check if nodeurl is changed
	if sc.NodeURL != current.NodeURL && sc.NodeURL != "" {
		query += "node_url = $" + strconv.Itoa(len(args)+1) + ", "
		args = append(args, sc.NodeURL)
	}

	// check if last_tx_block_synced is changed
	if sc.LastTxBlockSynced != current.LastTxBlockSynced && sc.LastTxBlockSynced != 0 {
		query += "last_tx_block_synced = $" + strconv.Itoa(len(args)+1) + ", "
		args = append(args, sc.LastTxBlockSynced)
	}

	// check if last_tx_block_synced is changed
	if sc.EngineLastTxBlockSynced != current.EngineLastTxBlockSynced && sc.EngineLastTxBlockSynced != 0 {
		query += "engine_last_tx_block_synced = $" + strconv.Itoa(len(args)+1) + ", "
		args = append(args, sc.EngineLastTxBlockSynced)
	}

	// check if status is changed
	if sc.Status != current.Status && sc.Status != "" {
		query += "status = $" + strconv.Itoa(len(args)+1) + ", "
		args = append(args, sc.Status)
	}

	// check if error is changed
	if sc.Error != nil {
		query += "error = $" + strconv.Itoa(len(args)+1) + ", "
		args = append(args, sc.Error)
	}

	// check if engine error is changed
	if sc.EngineError != nil {
		query += "engine_error = $" + strconv.Itoa(len(args)+1) + ", "
		args = append(args, sc.EngineError)
	}

	// check if initial block number is changed
	if sc.InitialBlockNumber != current.InitialBlockNumber && sc.InitialBlockNumber != 0 {
		query += "initial_block_number = $" + strconv.Itoa(len(args)+1) + ", "
		args = append(args, sc.InitialBlockNumber)
	}

	// check if engine status is changed
	if sc.EngineStatus != current.EngineStatus && sc.EngineStatus != "" {
		query += "engine_status = $" + strconv.Itoa(len(args)+1) + ", "
		args = append(args, sc.EngineStatus)
	}

	// check if webhook is changed
	if sc.Webhook != current.Webhook && sc.Webhook != "" {
		query += "webhook = $" + strconv.Itoa(len(args)+1) + ", "
		args = append(args, sc.Webhook)
	}

	// add where sql condition
	query = strings.TrimSuffix(query, ", ")
	query += " WHERE id = $" + strconv.Itoa(len(args)+1)
	args = append(args, current.ID)

	// execute query
	_, err = s.storage.DB.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	// get updated smartcontract
	updatedSmartContract, err := s.GetSmartContractById(current.ID)
	if err != nil {
		return nil, err
	}

	return updatedSmartContract, nil
}

func (s *Storage) UpdateLastBlockNumber(id string, blockNumber int64) error {
	// get current sc
	current, _ := s.GetSmartContractById(id)
	if current == nil {
		return fmt.Errorf("smartcontract does not exist")
	}

	// insert new smartcontract in database
	query := `UPDATE smartcontracts SET last_tx_block_synced = $1, updated_at = $2  WHERE id = $3 RETURNING *`
	_, err := s.storage.DB.Exec(query, blockNumber, time.Now(), id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) UpdateStatusAndError(id string, status smartcontract.SmartContractStatus, err error) error {
	// get current sc
	current, _ := s.GetSmartContractById(id)
	if current == nil {
		return fmt.Errorf("smartcontract does not exist")
	}

	// If the err is nil, the update err will be an empty string
	updateErr := ""
	if err != nil {
		updateErr = err.Error()
	}

	// update smartcontract status and error in database
	query := "UPDATE smartcontracts SET status = $1, error = $2, updated_at = $3 WHERE id = $4"
	_, err = s.storage.DB.Exec(query, status, updateErr, time.Now(), current.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetSmartContractById(id string) (*smartcontract.SmartContract, error) {
	// get smartcontract from db
	sc := &smartcontract.SmartContract{}
	err := s.storage.DB.Get(sc, "SELECT * FROM smartcontracts WHERE id = $1", id)
	if err != nil {
		return nil, err
	}

	return sc, nil
}

func (s *Storage) GetSmartContractByAddress(address string) (*smartcontract.SmartContract, error) {
	// get smartcontract from db
	sc := &smartcontract.SmartContract{}
	err := s.storage.DB.Get(sc, "SELECT * FROM smartcontracts WHERE address = $1", address)
	if err != nil {
		return nil, err
	}

	return sc, nil
}

func (s *Storage) DeleteSmartContractByAddress(address string) error {
	// list events by address from storage
	events, err := s.eventStorage.ListAllEvents()
	if err != nil {
		return nil
	}

	// delete events from storage
	for _, ev := range events {
		if ev.Address == address {
			err = s.eventStorage.DeleteEvent(address, ev.Abi.Name)
			if err != nil {
				return nil
			}
		}
	}

	// get smartcontract using the address
	sc, err := s.GetSmartContractByAddress(address)
	if err != nil {
		return nil
	}

	// delete transactions from storage
	err = s.transactionStorage.DeleteTransactionsByContractId(sc.ID)
	if err != nil {
		return nil
	}

	// delete smartcontract from db
	_, err = s.storage.DB.Exec("DELETE FROM smartcontracts WHERE address = $1", address)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) ListAllSmartContracts() ([]*smartcontract.SmartContract, error) {
	// define smartcontracts response
	smartcontracts := []*smartcontract.SmartContract{}

	// get smartcontracts from db
	scQuery := "SELECT * FROM smartcontracts"
	err := s.storage.DB.Select(&smartcontracts, scQuery)
	if err != nil {
		return nil, err
	}

	return smartcontracts, nil
}

func (s *Storage) ListSmartContracts(sort string, limit int64, offset int64) ([]*smartcontract.SmartContract, error) {
	// define smartcontracts response
	smartcontracts := []*smartcontract.SmartContract{}

	// get smartcontracts from db
	scQuery := fmt.Sprintf("SELECT * FROM smartcontracts ORDER BY created_at %s LIMIT $1 OFFSET $2", sort)

	err := s.storage.DB.Select(&smartcontracts, scQuery, limit, offset)
	if err != nil {
		return nil, err
	}

	return smartcontracts, nil
}

func (s *Storage) ListUniqueSmartContractsByNetwork() ([]*smartcontract.SmartContract, error) {
	// define smartcontracts response
	smartcontracts := []*smartcontract.SmartContract{}

	// get unique smartcontracts by network from db
	/* @dev: It creates a sub table with a partition with only address and network fields.
	 * This partition makes a counter for each row of smart contracts that has the same address
	 * and network. Then from that partition we only get the first row, ensuring that we are not
	 * getting any smart contract with this repeated info using the row number counter.
	 */
	scQuery := `SELECT id, name, network, node_url, address, engine_last_tx_block_synced, engine_status, engine_error, created_at, updated_at FROM (
					SELECT *, ROW_NUMBER() OVER (PARTITION BY address, network) AS rn
					FROM smartcontracts
				) AS sq
			WHERE sq.rn = 1`
	err := s.storage.DB.Select(&smartcontracts, scQuery)
	if err != nil {
		return nil, err
	}

	return smartcontracts, nil
}

func (s *Storage) GetSmartContractsCount() (int64, error) {
	var totalRows int64
	query := "SELECT COUNT(*) FROM smartcontracts"
	err := s.storage.DB.Get(&totalRows, query)
	if err != nil {
		return 0, err
	}

	return totalRows, nil
}

func (s *Storage) Stop() error {
	err := s.storage.DB.Close()
	if err != nil {
		return err
	}

	return nil
}
