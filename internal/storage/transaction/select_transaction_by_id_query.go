package transactionstorage

import (
	"fmt"

	"github.com/darchlabs/synchronizer-v2"
	"github.com/darchlabs/synchronizer-v2/pkg/transaction"
)

func (s *Storage) ListTxsById(id string, ctx *synchronizer.ListItemsInRangeCtx) ([]*transaction.Transaction, error) {
	// define events response
	var txs []*transaction.Transaction

	// get txs from db
	eventQuery := fmt.Sprintf("SELECT * FROM transactions WHERE contract_id = $1 AND timestamp BETWEEN $2 AND $3 ORDER BY block_number %s LIMIT $4 OFFSET $5", ctx.Sort)
	err := s.storage.DB.Select(&txs, eventQuery, id, ctx.StartTime, ctx.EndTime, ctx.Limit, ctx.Offset)
	if err != nil {
		return nil, err
	}

	// Return an empty array and not null in case there are no rows
	if len(txs) == 0 {
		return []*transaction.Transaction{}, nil
	}

	return txs, nil
}
