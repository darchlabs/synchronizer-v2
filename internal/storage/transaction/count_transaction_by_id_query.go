package transactionstorage

func (s *Storage) GetTxsCountById(id string) (int64, error) {
	// define events response
	var totalTxsNum []int64

	// get txs from db
	eventQuery := "SELECT COUNT(*) FROM transactions WHERE contract_id = $1"
	err := s.storage.DB.Select(&totalTxsNum, eventQuery, id)
	if err != nil {
		return 0, err
	}

	return totalTxsNum[0], nil
}
