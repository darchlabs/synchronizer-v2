package transactionstorage

import (
	"fmt"

	"github.com/darchlabs/synchronizer-v2"
	"github.com/darchlabs/synchronizer-v2/pkg/transaction"
	"github.com/pkg/errors"
)

func (s *Storage) ListFailedTxsById(id string, ctx *synchronizer.ListItemsInRangeCtx) ([]*transaction.Transaction, error) {
	var failedTxs []*transaction.Transaction

	// execute query and retrieve result
	err := s.storage.DB.Select(
		&failedTxs, fmt.Sprintf(`
		SELECT *
		FROM transactions
		WHERE contract_id = $1
		AND (is_error = '1' OR tx_receipt_status = '0')
		AND timestamp
		BETWEEN $2 AND $3
		ORDER BY block_number %s
		LIMIT $4
		OFFSET $5
		`, ctx.Sort),
		id,
		ctx.StartTime,
		ctx.EndTime,
		ctx.Limit,
		ctx.Offset,
	)
	if err != nil {
		return nil, errors.Wrap(err, "transactionstorage: Storage.ListFailedTxsById s.storage.DB.Select error")
	}

	// Return an empty array and not null in case there are no rows
	if len(failedTxs) == 0 {
		return []*transaction.Transaction{}, nil
	}

	return failedTxs, nil
}
