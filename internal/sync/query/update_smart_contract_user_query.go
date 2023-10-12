package query

import (
	"fmt"
	"time"

	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

type UpdateSmartContractUserQueryInput struct {
	Address              *string
	UserID               *string
	SmartContractAddress *string
	Name                 *string
	Error                error
	WebhookURL           *string
	Status               *storage.SmartContractUserStatus
	UpdatedAt            *time.Time
}

func (sq *SmartContractUserQuerier) UpdateSmartContractUserQuery(
	tx storage.Transaction,
	input *UpdateSmartContractUserQueryInput,
) (*storage.SmartContractUserRecord, error) {
	var record storage.SmartContractUserRecord
	var em *string
	if input.Error != nil {
		s := input.Error.Error()
		em = &s
	}
	query := `
		UPDATE smartcontract_user
		SET
			sc_address = COALESCE($2, sc_address)
			name = COALESCE($3, name),
			error = $4,
			webhook = COALESCE($5, name),
			status = COALESCE($6, name),
			updated_at = COALESCE($7, name),
		WHERE address = $1`
	if input.UserID != nil {
		query = fmt.Sprintf("%s AND user_id = $1", *input.UserID)
	}

	err := tx.Get(
		&record,
		query,
		input.Address,
		input.SmartContractAddress,
		input.Name,
		em,
		input.WebhookURL,
		input.Status,
		input.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "query: SmartContractUserQuerier.UpdateSmartContractUserQuery tx.Get error")
	}

	return &record, nil
}
