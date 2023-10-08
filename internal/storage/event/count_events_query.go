package eventstorage

import "github.com/pkg/errors"

func (s *Storage) GetEventsCount() (int64, error) {
	var totalRows int64
	err := s.storage.DB.Get(&totalRows, "SELECT COUNT(*) FROM event")
	if err != nil {
		return 0, errors.Wrap(err, "eventstorage: Storage.GetEventsCount s.storage.DB.Get error")
	}

	return totalRows, nil
}
