package transactionstorage

import "github.com/pkg/errors"

func (s *Storage) GetAddressesCountById(id string) (int64, error) {
	// define events response
	var count int64

	// execute query and retrieve result
	err := s.storage.DB.Get(
		&count,
		"SELECT COUNT(DISTINCT t.from) FROM transactions as T WHERE contract_id = $1",
		id,
	)
	if err != nil {
		return 0, errors.Wrap(err, "transactionstorage: Storage.GetAddressesCountById s.storage.DB.Get error")
	}

	return count, nil
}
