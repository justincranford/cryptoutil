package sqlprovider

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	googleUuid "github.com/google/uuid"
)

var (
	uuidZero = googleUuid.UUID{}
)

type SqlTransaction struct {
	sqlProvider *SqlProvider

	guardState    sync.Mutex
	isActive      bool
	ctx           context.Context
	readOnly      bool
	transactionID googleUuid.UUID
	sqlTx         *sql.Tx
}

func (sp *SqlProvider) WithTransaction(ctx context.Context, readOnly bool, function func(sqlTransaction *SqlTransaction) error) error {
	sqlTransaction, err := sp.NewTransaction()
	if err != nil {
		sp.telemetryService.Slogger.Error("failed to create transaction", "error", err)
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	err = sqlTransaction.Begin(ctx, readOnly)
	if err != nil {
		sp.telemetryService.Slogger.Error("failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if sqlTransaction.sqlTx != nil { // Avoid rollback if already committed or rolled back
			if err := sqlTransaction.Rollback(); err != nil {
				sp.telemetryService.Slogger.Error("failed to rollback transaction", "transactionID", sqlTransaction.transactionID, "error", err)
			}
		}
		if r := recover(); r != nil {
			sp.telemetryService.Slogger.Error("panic occurred during transaction", "transactionID", sqlTransaction.transactionID, "panic", r)
			panic(r) // re-throw the panic after rollback
		}
	}()

	if err := function(sqlTransaction); err != nil {
		sp.telemetryService.Slogger.Error("transaction function failed", "transactionID", sqlTransaction.transactionID, "error", err)
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	return sqlTransaction.Commit()
}

func (pr *SqlProvider) NewTransaction() (*SqlTransaction, error) {
	pr.telemetryService.Slogger.Info("new transaction")
	return &SqlTransaction{sqlProvider: pr, isActive: false, transactionID: uuidZero}, nil
}

func (tx *SqlTransaction) Begin(ctx context.Context, readOnly bool) error {
	tx.guardState.Lock()
	defer tx.guardState.Unlock()

	tx.sqlProvider.telemetryService.Slogger.Info("beginning transaction", "transactionID", tx.transactionID, "readOnly", tx.readOnly)
	if tx.isActive {
		tx.sqlProvider.telemetryService.Slogger.Error("transaction already started", "transactionID", tx.transactionID)
		return fmt.Errorf("transaction already started")
	}

	transactionID, err := googleUuid.NewV7()
	if err != nil {
		tx.sqlProvider.telemetryService.Slogger.Error("failed to generate transaction ID", "error", err)
		return fmt.Errorf("failed to generate transaction ID: %w", err)
	}

	sqlTx, err := tx.sqlProvider.sqlDB.BeginTx(ctx, &sql.TxOptions{ReadOnly: tx.readOnly})
	if err != nil {
		tx.sqlProvider.telemetryService.Slogger.Error("failed to begin transaction", "transactionID", transactionID, "readOnly", tx.readOnly, "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	tx.setState(ctx, readOnly, transactionID, sqlTx)
	tx.sqlProvider.telemetryService.Slogger.Info("started transaction", "transactionID", tx.transactionID, "readOnly", tx.readOnly)
	return nil
}

func (tx *SqlTransaction) Commit() error {
	tx.guardState.Lock()
	defer tx.guardState.Unlock()

	tx.sqlProvider.telemetryService.Slogger.Info("committing transaction", "transactionID", tx.transactionID, "readOnly", tx.readOnly)
	if !tx.isActive {
		tx.sqlProvider.telemetryService.Slogger.Error("can't commit because transaction not active", "transactionID", tx.transactionID)
		return fmt.Errorf("can't commit because transaction not active")
	}

	err := tx.sqlTx.Commit()
	if err != nil {
		tx.sqlProvider.telemetryService.Slogger.Error("failed to commit transaction", "transactionID", tx.transactionID, "error", err)
		return err
	}

	tx.sqlProvider.telemetryService.Slogger.Info("committed transaction", "transactionID", tx.transactionID)

	tx.clearState()
	return nil
}

func (tx *SqlTransaction) Rollback() error {
	tx.guardState.Lock()
	defer tx.guardState.Unlock()

	tx.sqlProvider.telemetryService.Slogger.Info("rolling back transaction", "transactionID", tx.transactionID)
	if !tx.isActive {
		tx.sqlProvider.telemetryService.Slogger.Error("can't rollback because transaction not active", "transactionID", tx.transactionID)
		return fmt.Errorf("can't rollback because transaction not active")
	}

	err := tx.sqlTx.Rollback()
	if err != nil {
		tx.sqlProvider.telemetryService.Slogger.Error("failed to rollback transaction", "transactionID", tx.transactionID, "error", err)
		return err
	}

	tx.sqlProvider.telemetryService.Slogger.Info("rolled back transaction", "transactionID", tx.transactionID)

	tx.clearState()
	return nil
}

func (tx *SqlTransaction) setState(ctx context.Context, readOnly bool, transactionID googleUuid.UUID, sqlTx *sql.Tx) {
	tx.ctx = ctx
	tx.readOnly = readOnly
	tx.isActive = true
	tx.transactionID = transactionID
	tx.sqlTx = sqlTx
}

func (tx *SqlTransaction) clearState() {
	tx.ctx = nil
	tx.readOnly = true
	tx.isActive = false
	tx.transactionID = uuidZero
	tx.sqlTx = nil
}
