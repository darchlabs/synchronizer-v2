package storage

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type S struct {
	DB *sqlx.DB
}

func New(dsn string) (*S, error) {
	// create connection to database
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	return &S{
		DB: db,
	}, nil
}
