package db

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"go-service-template/monitor"
	"go-service-template/repositories"
)

type TxDBContext struct {
	db *sql.DB
	tx *sql.Tx
}

func CreateTxDBContext(db *sql.DB) *TxDBContext {
	return &TxDBContext{db: db, tx: nil}
}

func (txDb *TxDBContext) StartTx() error {
	if txDb.tx != nil {
		return errors.New("transaction already started")
	}

	newTx, err := txDb.db.Begin()
	if err != nil {
		return fmt.Errorf("unable to create transaction: %w", err)
	}

	txDb.tx = newTx

	return nil
}

func (txDb *TxDBContext) Exec(ctx monitor.ApplicationContext, query string, args ...interface{}) (sql.Result, error) {
	var err error
	var res sql.Result
	var prepStmt *sql.Stmt

	if txDb.tx != nil {
		prepStmt, err = txDb.tx.PrepareContext(ctx, query)
	} else {
		prepStmt, err = txDb.db.PrepareContext(ctx, query)
	}

	if err != nil {
		return nil, fmt.Errorf("error preparing statement '%v'. Error: %w", query, err)
	}

	res, err = prepStmt.ExecContext(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query '%v'. Error: %w", query, err)
	}

	err = prepStmt.Close()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (txDb *TxDBContext) CommitTx() error {
	if txDb.tx == nil {
		return nil
	}

	if err := txDb.tx.Commit(); err != nil {
		return fmt.Errorf("tx commit failed: %w", err)
	}

	txDb.tx = nil

	return nil
}

func (txDb *TxDBContext) RollbackTx() error {
	if txDb.tx == nil {
		return nil
	}

	if err := txDb.tx.Rollback(); err != nil {
		return fmt.Errorf("tx rollback failed: %w", err)
	}

	txDb.tx = nil

	return nil
}

func (txDb *TxDBContext) Ping() error {
	if err := txDb.db.Ping(); err != nil {
		return fmt.Errorf("error pinging DB: %w", err)
	}

	return nil
}

func (txDb *TxDBContext) getDBReader() repositories.DBReader {
	if txDb.tx != nil {
		return txDb.tx
	}

	return txDb.db
}

func (txDb *TxDBContext) WithTx(ctx monitor.ApplicationContext, fn func(fnCtx monitor.ApplicationContext) error) error {
	var err error

	if err = txDb.StartTx(); err != nil {
		return err
	}

	if err = fn(ctx); err != nil {
		if rollbackErr := txDb.RollbackTx(); rollbackErr != nil {
			return fmt.Errorf("tx rollback failed: %w", rollbackErr)
		}

		return err
	}

	if err = txDb.CommitTx(); err != nil {
		return fmt.Errorf("tx commit failed: %w", err)
	}

	return nil
}
