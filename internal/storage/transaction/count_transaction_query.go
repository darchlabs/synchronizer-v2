package transactionstorage

func (s *Storage) GetTxsCount() (int64, error) {
	// define events response
	var totalTxs []int64

	// get txs from db
	eventQuery := "SELECT COUNT(*) FROM transactions"
	err := s.storage.DB.Select(&totalTxs, eventQuery)
	if err != nil {
		return 0, err
	}

	return totalTxs[0], nil
}
