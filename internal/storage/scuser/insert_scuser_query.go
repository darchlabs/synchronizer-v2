package scuserstorage

import (
	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/pkg/errors"
)

func (st *Storage) InsertSmartContractUserQuery(tx storage.Transaction, input *storage.SmartContractUserRecord) error {
	_, err := tx.Exec(`
		INSERT INTO smartcontract_users (id, user_id, sc_address)
		VALUES ($1, $2, $3);`,
		input.ID,
		input.UserID,
		input.SmartContractAddress,
	)
	if err != nil {
		return errors.Wrap(err, "scuserstorage: Storage.InsertSmartContractUserQuery tx.Exec error")
	}

	return nil
}
