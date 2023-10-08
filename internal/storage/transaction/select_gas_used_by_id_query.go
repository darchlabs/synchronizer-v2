package transactionstorage

import (
	"sort"
	"strconv"

	"github.com/darchlabs/synchronizer-v2"
	"github.com/pkg/errors"
)

func (s *Storage) ListGasSpentById(id string, startTs int64, endTs int64, interval int64) ([][]string, error) {
	// Define array for the query
	var gasUsedAndTimestamp []synchronizer.GasTimestamp

	// execute query and retrieve result
	err := s.storage.DB.Select(
		&gasUsedAndTimestamp, `
		SELECT gas_used, timestamp
		FROM transactions
		WHERE contract_id = $1
		AND timestamp
		BETWEEN $2 AND $3
		ORDER BY block_number DESC;`,
		id,
		startTs,
		endTs,
	)
	if err != nil {
		return nil, errors.Wrap(err, "transactionstorage: Storage.ListGasSpentById s.storage.DB.Select error")
	}

	// finish method when response is empty
	if len(gasUsedAndTimestamp) == 0 {
		return [][]string{}, nil
	}

	// Group the timestamps by interval and sum the gas used
	intervalEnd := endTs
	intervalStart := endTs - interval
	var result [][]string
	var sumGasUsed int64
	for _, data := range gasUsedAndTimestamp {
		if data.Timestamp >= intervalStart {
			sumGasUsed += data.GasUsed
		} else {
			// Add the sum of gas used during the interval to the result
			result = append(result, []string{strconv.FormatInt(sumGasUsed, 10), strconv.FormatInt(intervalEnd, 10)})
			// Start a new interval
			intervalEnd = intervalStart
			intervalStart -= interval
			sumGasUsed = data.GasUsed
		}
	}

	// Add the last interval if any
	if sumGasUsed > 0 {
		result = append(result, []string{strconv.FormatInt(sumGasUsed, 10), strconv.FormatInt(intervalEnd, 10)})
	}

	// sort slice ASC
	sort.Slice(result, func(i, j int) bool {
		left, _ := strconv.ParseInt(result[i][1], 10, 64)
		right, _ := strconv.ParseInt(result[j][1], 10, 64)
		return left < right
	})

	return result, nil
}
