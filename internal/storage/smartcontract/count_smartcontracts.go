package smartcontractstorage

func (s *Storage) GetSmartContractsCount() (int64, error) {
	var totalRows int64
	query := "SELECT COUNT(*) FROM smartcontracts"
	err := s.storage.DB.Get(&totalRows, query)
	if err != nil {
		return 0, err
	}

	return totalRows, nil
}
