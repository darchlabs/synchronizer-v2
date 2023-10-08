package transactionstorage

import (
	"fmt"

	"github.com/darchlabs/synchronizer-v2"
	"github.com/pkg/errors"
)

func (s *Storage) ListUniqueAddresses(id string, ctx *synchronizer.ListItemsInRangeCtx) ([]string, error) {
	var uniqueAddresses []string

	// execute query and retrieve result
	err := s.storage.DB.Select(
		&uniqueAddresses, fmt.Sprintf(`
		SELECT DISTINCT t.from
		FROM (
			SELECT t.from, t.block_number
			FROM transactions AS t
			WHERE contract_id = $1
			AND timestamp
			BETWEEN $2 AND $3
			ORDER BY t.block_number %s
		) t
		LIMIT $4
		OFFSET $5;`, ctx.Sort),
		id,
		ctx.StartTime,
		ctx.EndTime,
		ctx.Limit,
		ctx.Offset,
	)
	if err != nil {
		return nil, errors.Wrap(err, "transactionstorage: Storage.ListUniqueAddresses s.storage.DB.Select error")
	}

	// Return an empty array and not null in case there are no rows
	if len(uniqueAddresses) == 0 {
		return []string{}, nil
	}

	return uniqueAddresses, nil
}
