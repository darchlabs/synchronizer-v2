package transactionstorage

import "github.com/pkg/errors"

func (s *Storage) GetFailedTxsCountById(id string) (int64, error) {
	var totalFailedTxs []int64

	// execute query and retrieve result
	err := s.storage.DB.Select(
		&totalFailedTxs, `
		SELECT COUNT(*)
		FROM transactions
		WHERE contract_id = $1
		AND (is_error = '1' OR tx_receipt_status = '0');`,
		id,
	)
	if err != nil {
		return 0, errors.Wrap(err, "transactionstorage: Storage.GetFailedTxsCountById s.storage.DB.Select error")
	}

	return totalFailedTxs[0], nil
}
