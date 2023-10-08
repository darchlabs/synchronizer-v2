package transactionstorage

import "github.com/pkg/errors"

func (s *Storage) DeleteTransactionsByContractId(id string) error {
	// delete transaction using id
	_, err := s.storage.DB.Exec("DELETE FROM transactions WHERE contract_id = $1", id)
	if err != nil {
		return errors.Wrap(err, "transactionstorage: Storage.DeleteTransactionsByContractId s.storage.DB.Exec error")
	}

	return nil
}
