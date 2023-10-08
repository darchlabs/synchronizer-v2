package transactionstorage

import (
	"fmt"
	"strconv"

	"github.com/darchlabs/synchronizer-v2"
)

func (s *Storage) GetTvlById(id string) (int64, error) {
	// define events response
	var lastTVL []string

	// get txs from db
	eventQuery := "SELECT contract_balance FROM transactions WHERE contract_id = $1 ORDER BY block_number DESC LIMIT 1"
	err := s.storage.DB.Select(&lastTVL, eventQuery, id)
	if err != nil {
		return 0, err
	}

	// Return an empty value in case there are no rows
	if len(lastTVL) == 0 {
		return 0, nil
	}

	currentTVL, err := strconv.ParseInt(lastTVL[0], 10, 64)
	if err != nil {
		return 0, err
	}

	return currentTVL, nil
}

func (s *Storage) ListTvlsById(id string, ctx *synchronizer.ListItemsInRangeCtx) ([][]string, error) {
	// create an arr of ContractBalanceTimestamp
	var balanceTimestamps []synchronizer.ContractBalanceTimestamp

	// get txs from db
	eventQuery := fmt.Sprintf("SELECT contract_balance, timestamp FROM transactions WHERE contract_id = $1 AND timestamp BETWEEN $2 AND $3 ORDER BY block_number %s LIMIT $4 OFFSET $5", ctx.Sort)
	err := s.storage.DB.Select(&balanceTimestamps, eventQuery, id, ctx.StartTime, ctx.EndTime, ctx.Limit, ctx.Offset)
	if err != nil {
		return nil, err
	}

	// define method response
	var tvlWithTimestampArr [][]string
	// Iterate over balanceTimestamps and create tvlWithTimestampArr
	for _, item := range balanceTimestamps {
		tvlWithTimestampArr = append(tvlWithTimestampArr, []string{item.ContractBalance, item.Timestamp})
	}

	// Return an empty array and not null in case there are no rows
	if len(tvlWithTimestampArr) == 0 {
		return [][]string{}, nil
	}

	return tvlWithTimestampArr, nil
}
