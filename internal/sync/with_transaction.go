package sync

import (
	"context"
	"database/sql"

	"github.com/darchlabs/synchronizer-v2/internal/storage"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func (ng *Engine) WithTransaction(db storage.Transaction, fn func(txx *sqlx.Tx) error) (err error) {
	ctx := context.Background()
	var tx *sqlx.Tx

	defer func() {
		if err != nil {
			if txErr := tx.Rollback(); txErr != nil {
				err = errors.Wrapf(txErr, "sync: Engine.WithTransaction error %s", err)
				ng.logger.Infof("%s", err)
			}
		}
	}()

	switch store := db.(type) {
	case *sqlx.DB:
		tx, err = store.BeginTxx(ctx, &sql.TxOptions{
			Isolation: sql.LevelSerializable,
		})
		if err != nil {
			err = errors.Wrap(err, "sync: Engine.WithTransaction error")
			ng.logger.Infof("%s", err)
			return err
		}
	case *sqlx.Tx:
		tx = store

	default:
		Tx, err := ng.database.BeginTx(ctx)
		if err != nil {
			err = errors.Wrap(err, "sync: Engine.WithTransaction w.db.BeginTx error")
			ng.logger.Infof("%s", err)
			return err
		}

		tx = Tx.(*sqlx.Tx)
	}

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}

func (ng *Engine) InTransaction(fn func(txx *sqlx.Tx) error) error {
	return ng.WithTransaction(ng.database, fn)
}
