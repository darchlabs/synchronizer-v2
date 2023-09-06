package smartcontractstorage

import (
	"time"

	"github.com/darchlabs/synchronizer-v2/pkg/smartcontract"
	"github.com/pkg/errors"
)

// TODO: Improve this function to avoid manipulate error received by parameter and messing with
// proper function call errrors
func (s *Storage) UpdateStatusAndError(id string, status smartcontract.SmartContractStatus, err error) error {
	// get current sc
	current, _ := s.GetSmartContractById(id)
	if current == nil {
		return ErrSmartcontractNotFound
	}

	// If the err is nil, the update err will be an empty string
	updateErr := ""
	if err != nil {
		updateErr = err.Error()
	}

	// update smartcontract status and error in database
	_, err = s.storage.DB.Exec(`
		UPDATE smartcontracts
		SET status = $1,
				error = $2,
				updated_at = $3
		WHERE id = $4;`,
		status,
		updateErr,
		time.Now(),
		current.ID,
	)
	if err != nil {
		return errors.Wrap(err, "smartcontractstorage: Storage.UpdateStatusAndError s.storage.DB.Exec error")
	}

	return nil
}
