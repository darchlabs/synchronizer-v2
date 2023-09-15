package storage

import "time"

type SmartContractUserRecord struct {
	ID              string     `db:"id"`
	UserID          string     `db:"user_id"`
	SmartContractID string     `db:"smartcontract_id"`
	DeletedAt       *time.Time `db:"deleted_at"`
}
