package eventstorage

func (s *Storage) GetEventDataCount(address string, eventName string) (int64, error) {
	var totalRows int64
	err := s.storage.DB.Get(
		&totalRows, `
			SELECT COUNT(event_data.*)
			FROM event_data
			JOIN event ON event_data.event_id = event.id
			JOIN abi ON event.abi_id = abi.id
			WHERE event.address = $1
			AND abi.name = $2;`,
		address,
		eventName,
	)
	if err != nil {
		return 0, err
	}

	return totalRows, nil
}
