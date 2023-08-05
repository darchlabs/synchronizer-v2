package eventstorage

import "github.com/pkg/errors"

func (s *Storage) GetEventCountByAddress(address string) (int64, error) {
	var totalRows int64
	err := s.storage.DB.Get(&totalRows, "SELECT COUNT(*) FROM event WHERE address = $1", address)
	if err != nil {
		return 0, errors.Wrap(err, "eventstorage: Storage.GetEventCountByAddress s.storage.DB.Get error")
	}

	return totalRows, nil
}
