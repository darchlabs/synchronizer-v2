package test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/darchlabs/synchronizer-v2/internal/storage"
	_ "github.com/darchlabs/synchronizer-v2/migrations"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

const (
	defaultTestDBDriver = "postgres"
	defaultTestDBDSN    = "postgres://postgres:postgres@120.0.0.1"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

func getDB() (*sqlx.DB, error) {
	driver := "postgres"
	dsn := os.Getenv("DATABASE_DSN")
	migrations := os.Getenv("DATABASE_MIGRATIONS_DIR")

	db, err := sqlx.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	db.Exec("drop database if exists darchlabs-test")
	db.Exec("create darchlabs-test")

	db.SetConnMaxLifetime(-1)
	err = goose.Up(db.DB, fmt.Sprintf("%s/%s", basepath, migrations))
	if err != nil {
		return nil, err
	}
	fmt.Printf("BasePath %s\n Migrations %s\n PATH %s\n", basepath, migrations, b)

	return db, nil
}

type TestDB struct {
	tx *sqlx.Tx
}

func (tdb *TestDB) BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	return tdb.tx, nil
}

func (tdb *TestDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	return tdb.tx, nil
}

func (tdb *TestDB) Exec(query string, params ...interface{}) (sql.Result, error) {
	return tdb.tx.Exec(query, params)
}

func (tdb *TestDB) Query(query string, params ...interface{}) (*sql.Rows, error) {
	return tdb.tx.Query(query, params)
}

func (tdb *TestDB) Get(dest interface{}, q string, args ...interface{}) error {
	return tdb.tx.Get(dest, q, args)
}

func (tdb *TestDB) Select(dest interface{}, q string, args ...interface{}) error {
	return tdb.tx.Select(dest, q, args)
}

func (tdb *TestDB) Rollback() error {
	return tdb.tx.Rollback()
}

func GetTxCall(t *testing.T, call func(tx *sqlx.Tx, testData interface{})) {
	t.Helper()
	conn, err := getDB()
	require.NoError(t, err)

	ctx := context.TODO()
	tx, err := conn.BeginTxx(ctx, nil)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	ctx = context.TODO()
	tx1, err := conn.BeginTxx(ctx, nil)
	require.NoError(t, err)

	call(tx1, nil)

	tx1.Commit()

	ctx = context.TODO()
	tx, err = conn.BeginTxx(ctx, nil)
	require.NoError(t, err)

	err = CleanDB(tx)
	require.NoError(t, err)

	tx.Commit()

	err = conn.Close()
	require.NoError(t, err)
}

func GetDBCall(t *testing.T, call func(db *sqlx.DB, testData interface{})) {
	t.Helper()
	conn, err := getDB()
	require.NoError(t, err)

	CleanDBConn(t, conn)

	ctx := context.TODO()
	tx, err := conn.BeginTxx(ctx, nil)
	require.NoError(t, err)

	tx.Commit()

	call(conn, nil)

	CleanDBConn(t, conn)

	conn.Close()
}

func CleanDBConn(t *testing.T, st *sqlx.DB) error {
	queries := []string{
		"DELETE FROM events;",
		"DELETE FROM inputs;",
		"DELETE FROM abi;",
		"DELETE FROM smartcontracts;",
	}

	for _, query := range queries {
		err := PrepareDeleteFromDB(st, query)
		if err != nil {
			return err
		}
	}

	return nil
}
func PrepareDeleteFrom(st *sqlx.Tx, query string) error {
	stmt, err := st.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	return nil
}

func PrepareDeleteFromDB(st *sqlx.DB, query string) error {
	stmt, err := st.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return err
	}
	return nil
}

func CleanDB(st *sqlx.Tx) (err error) {
	dropQueries := []string{
		"DELETE FROM event CASCADE;",
		"DELETE FROM abi CASCADE;",
		"DELETE FROM smartcontract_users CASCADE;",
		"DELETE FROM smartcontracts CASCADE;",
	}

	for _, query := range dropQueries {
		_, err := st.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetTestDB(db *sqlx.DB) *storage.Store {
	return &storage.Store{db}
}
