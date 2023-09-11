package transactionstorage

import (
	"fmt"

	"github.com/darchlabs/synchronizer-v2/pkg/transaction"
)

func (s *Storage) ListTxs(sort string, limit int64, offset int64) ([]*transaction.Transaction, error) {
	// define events response
	txs := []*transaction.Transaction{}

	// get txs from db
	eventQuery := fmt.Sprintf("SELECT * FROM transactions ORDER BY block_number %s LIMIT $1 OFFSET $2", sort)
	err := s.storage.DB.Select(&txs, eventQuery, limit, offset)
	if err != nil {
		return nil, err
	}

	// Return an empty array and not null in case there are no rows
	if len(txs) == 0 {
		return []*transaction.Transaction{}, nil
	}

	return txs, nil
}
