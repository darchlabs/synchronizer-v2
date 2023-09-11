package smartcontractstorage

import (
	"github.com/darchlabs/synchronizer-v2"
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

var (
	ErrSmartcontractNotFound = errors.New("smartcontract not found error")
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

func (s *Storage) Stop() error {
	err := s.storage.DB.Close()
	if err != nil {
		return err
	}

	return nil
}

// NOTE:
// 1. InsertSmartContract Method moved to: internal/storage/smartcontract/insert_smartcontract_query.go
// 2. UpdateLastBlockNumber Method moved to: internal/storage/smartcontract/update_last_block_number_query.go
// 2. UpdateStatusAndError Method moved to: internal/storage/smartcontract/update_status_and_error_query.go
// 4. GetSmartContractByID Method moved to: internal/storage/smartcontract/select_smartcontract_by_id.go
// 5. GetSmartContractByAddress Method moved to: internal/storage/smartcontract/select_smartcontract_by_address.go
// 6. DeleteSmartContractByAddress Method moved to: internal/storage/smartcontract/delete_smartcontract_by_address_query.go
// 7. ListAllSmartContracts Method moved to: internal/storage/smartcontract/select_all_smart_contracts_query.go
// 8. ListSmartContracts Method moved to: internal/storage/smartcontract/select_smart_contracts_query.go
// 9. ListUniqueSmartContractsByNetwork Method moved to: internal/storage/smartcontract/select_unique_smartcontracts_by_network.go
// 10. GetSmartContractsCount Method moved to: internal/storage/smartcontract/count_smartcontracts.go
// 10. UpdateSmartContract Method moved to: internal/storage/smartcontract/update_smart_contract_query.go
