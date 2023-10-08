package transactionstorage

func (s *Storage) GetTotalGasSpentById(id string) (int64, error) {
	var totalGasSpent int64

	txs, err := s.GetTxsCountById(id)
	if err != nil {
		return 0, nil
	}

	// check if txs count is zero
	if txs == 0 {
		return 0, nil
	}

	// execute query and retrieve result
	query := "SELECT SUM(CAST(gas_used AS bigint)) FROM transactions where contract_id = $1;"
	err = s.storage.DB.Get(&totalGasSpent, query, id)
	if err != nil {
		return 0, err
	}

	return totalGasSpent, nil
}
