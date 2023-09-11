package transactionstorage

import "github.com/pkg/errors"

func (s *Storage) GetValueTransferredById(id string) (int64, error) {
	var totalValueTransferred []int64

	// execute query and retrieve result
	err := s.storage.DB.Select(
		&totalValueTransferred,
		"SELECT SUM(value::bigint) FROM transactions WHERE contract_id = $1",
		id,
	)
	if err != nil {
		return 0, errors.Wrap(err, "transactionstorage: Storage.GetValueTransferredById s.storage.DB.Select error")
	}

	return totalValueTransferred[0], nil
}
