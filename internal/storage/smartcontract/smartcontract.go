package smartcontractstorage

import (
	"fmt"
	"time"

	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
)

type Storage struct {
	storage *storage.S
}

func New(s *storage.S) *Storage {
	return &Storage{
		storage: s,
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
	query := "INSERT INTO smartcontracts ( id, name, network, node_url, address,last_tx_block_synced, status, error, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id"
	err := s.storage.DB.Get(&smartcontractId, query, sc.ID, sc.Name, sc.Network, sc.NodeURL, sc.Address,
		sc.LastTxBlockSynced, sc.Status, sc.Error, sc.CreatedAt, sc.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// get created smartcontract
	createdSmartcontract, err := s.GetSmartContractByID(smartcontractId)
	if err != nil {
		return nil, err
	}

	return createdSmartcontract, nil
}

func (s *Storage) UpdateLastBlockNumber(id string, blockNumber int64) error {
	// get current sc
	current, _ := s.GetSmartContractByID(id)
	if current == nil {
		return fmt.Errorf("smartcontract does not exist")
	}

	fmt.Println("id: ", id)
	fmt.Println("blockNumber: ", blockNumber)

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
	current, _ := s.GetSmartContractByID(id)
	if current == nil {
		return fmt.Errorf("smartcontract does not exist")
	}

	// update smartcontract status and error in database
	query := fmt.Sprintf("UPDATE smartcontracts SET status = $1, error = $2, updated_at = $3 WHERE id = %s", current.ID)
	_, err = s.storage.DB.Exec(query, status, err, time.Now())
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetSmartContractByID(id string) (*smartcontract.SmartContract, error) {
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
	// get smartcontract from db
	_, err := s.storage.DB.Exec("DELETE FROM smartcontracts WHERE address = $1", address)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) ListSmartContracts(sort string, limit int64, offset int64) ([]*smartcontract.SmartContract, error) {
	// define smartcontracts response
	smartcontracts := []*smartcontract.SmartContract{}

	fmt.Println("as: ")
	// get smartcontracts from db
	scQuery := fmt.Sprintf("SELECT * FROM smartcontracts ORDER BY created_at %s LIMIT $1 OFFSET $2", sort)
	err := s.storage.DB.Select(&smartcontracts, scQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	fmt.Println("scscs: ", smartcontracts)

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
