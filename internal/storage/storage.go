package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type S struct {
	DB *sqlx.DB
}

type Store struct {
	*sqlx.DB
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

func NewStorage(driver, dsn string) (*Store, error) {
	db, err := sqlx.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	// Maximum Idle Connections
	db.SetMaxIdleConns(20)
	// Idle Connection Timeout
	db.SetConnMaxIdleTime(1 * time.Second)
	// Connection Lifetime
	db.SetConnMaxLifetime(30 * time.Second)

	return &Store{db}, nil
}

func (st *Store) BeginTx(ctx context.Context) (Transactioner, error) {
	tx, err := st.DB.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return nil, errors.Wrap(err, "database: Store.BeginTx st.BeginTxx error")
	}
	return tx, nil
}
