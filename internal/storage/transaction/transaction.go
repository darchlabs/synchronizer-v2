package transactionstorage

import (
	"errors"

	"github.com/darchlabs/synchronizer-v2/internal/storage"
)

var (
	ErrTransactionsEmpty = errors.New("transactions array empty error")
)

type Storage struct {
	storage *storage.S
}

func New(s *storage.S) *Storage {
	return &Storage{
		storage: s,
	}
}

// NOTE:
// 1. lListTxs Method moved to: internal/storage/transaction/select_transaction_query.go
// 2. GetTxsCount Method moved to: internal/storage/transaction/count_transaction_query.go
// 3. ListTxsById Method moved to: internal/storage/transaction/select_transaction_by_id_query.go
// 4. GetTxsCountById Method moved to: internal/storage/transaction/count_transaction_by_id_query.go
// 5. GetTvlById Method moved to: internal/storage/transaction/select_contract_balance_by_id_query.go
//    5.1 ListTvlsById Method moved to: internal/storage/transaction/select_contract_balance_by_id_query.go
// 6. GetAddressesCountById Method moved to: internal/storage/transaction/count_addresses_by_id_query.go
// 7. GetFailedTxsCountById Method moved to: internal/storage/transaction/count_failed_transactions_by_id_query.go
// 8. GetTotalGasSpentById Method moved to: internal/storage/transaction/sum_gas_used_by_contract_id_query.go
// 9. ListGasSpentById Method moved to: internal/storage/transaction/select_gas_used_by_id_query.go
// 10. GetValueTransferredById Method moved to: internal/storage/transaction/sum_value_transfered_by_contract_id_query.go
// 11. InsertTxs Method moved to: internal/storage/transaction/insert_transactions_query.go
// 12. DeleteTransactionsByContractId Method moved to: internal/storage/transaction/delete_transaction_by_contract_id_query.go
