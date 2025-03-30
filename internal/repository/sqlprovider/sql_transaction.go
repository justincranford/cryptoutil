package sqlprovider

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	googleUuid "github.com/google/uuid"
)

type SqlTransaction struct {
	sqlProvider *SqlProvider
	guardState  sync.Mutex
	state       *SqlTransactionState
}

type SqlTransactionState struct {
	ctx           context.Context
	readOnly      bool
	transactionID googleUuid.UUID
	sqlTx         *sql.Tx
}

func (sqlProvider *SqlProvider) WithTransaction(ctx context.Context, readOnly bool, function func(sqlTransaction *SqlTransaction) error) error {
	sqlTransaction, err := sqlProvider.newTransaction()
	if err != nil {
		sqlProvider.telemetryService.Slogger.Error("failed to create transaction", "error", err)
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	err = sqlTransaction.begin(ctx, readOnly)
	if err != nil {
		sqlProvider.telemetryService.Slogger.Error("failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if sqlTransaction.state != nil { // Avoid rollback if already committed or rolled back
			if err := sqlTransaction.rollback(); err != nil {
				sqlProvider.telemetryService.Slogger.Error("failed to rollback transaction", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly(), "error", err)
			}
		}
		if r := recover(); r != nil {
			sqlProvider.telemetryService.Slogger.Error("panic occurred during transaction", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly(), "panic", r)
			panic(r) // re-throw the panic after rollback
		}
	}()

	if err := function(sqlTransaction); err != nil {
		sqlProvider.telemetryService.Slogger.Error("transaction function failed", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly(), "error", err)
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	return sqlTransaction.commit()
}

func (sqlProvider *SqlProvider) newTransaction() (*SqlTransaction, error) {
	sqlProvider.telemetryService.Slogger.Info("new transaction")
	return &SqlTransaction{sqlProvider: sqlProvider}, nil
}

// TransactionID Transaction ID is valid (non-nil) only when a transaction is active
func (sqlTransaction *SqlTransaction) TransactionID() *googleUuid.UUID {
	if sqlTransaction.state == nil {
		return nil
	}
	transactionIDCopy := googleUuid.UUID(sqlTransaction.state.transactionID)
	return &transactionIDCopy
}

// Context Transaction context is valid (non-nil) only when a transaction is active
func (sqlTransaction *SqlTransaction) Context() context.Context {
	if sqlTransaction.state == nil {
		return nil
	}
	return sqlTransaction.state.ctx
}

// IsReadOnly Boolean for if a transaction read-only is valid only when a transaction is active
func (sqlTransaction *SqlTransaction) IsReadOnly() bool {
	if sqlTransaction.state == nil {
		return false
	}
	return sqlTransaction.state.readOnly
}

func (sqlTransaction *SqlTransaction) begin(ctx context.Context, readOnly bool) error {
	sqlTransaction.guardState.Lock()
	defer sqlTransaction.guardState.Unlock()

	sqlTransaction.sqlProvider.telemetryService.Slogger.Info("beginning transaction", "readOnly", readOnly)
	if sqlTransaction.state != nil {
		sqlTransaction.sqlProvider.telemetryService.Slogger.Error("transaction already started", "transactionID", sqlTransaction.TransactionID())
		return fmt.Errorf("transaction already started")
	}

	transactionID, err := googleUuid.NewV7()
	if err != nil {
		sqlTransaction.sqlProvider.telemetryService.Slogger.Error("failed to generate transaction ID", "error", err)
		return fmt.Errorf("failed to generate transaction ID: %w", err)
	}

	sqlTx, err := sqlTransaction.sqlProvider.sqlDB.BeginTx(ctx, &sql.TxOptions{ReadOnly: readOnly})
	if err != nil {
		sqlTransaction.sqlProvider.telemetryService.Slogger.Error("failed to begin transaction", "transactionID", transactionID, "readOnly", readOnly, "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	sqlTransaction.state = &SqlTransactionState{ctx: ctx, readOnly: readOnly, transactionID: transactionID, sqlTx: sqlTx}
	sqlTransaction.sqlProvider.telemetryService.Slogger.Info("started transaction", "transactionID", transactionID, "readOnly", readOnly)
	return nil
}

func (sqlTransaction *SqlTransaction) commit() error {
	sqlTransaction.guardState.Lock()
	defer sqlTransaction.guardState.Unlock()

	sqlTransaction.sqlProvider.telemetryService.Slogger.Info("committing transaction", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly())
	if sqlTransaction.state == nil {
		sqlTransaction.sqlProvider.telemetryService.Slogger.Error("can't commit because transaction not active", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly())
		return fmt.Errorf("can't commit because transaction not active")
	}

	err := sqlTransaction.state.sqlTx.Commit()
	if err != nil {
		sqlTransaction.sqlProvider.telemetryService.Slogger.Error("failed to commit transaction", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly(), "error", err)
		return err
	}

	sqlTransaction.sqlProvider.telemetryService.Slogger.Info("committed transaction", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly())
	sqlTransaction.state = nil
	return nil
}

func (sqlTransaction *SqlTransaction) rollback() error {
	sqlTransaction.guardState.Lock()
	defer sqlTransaction.guardState.Unlock()

	sqlTransaction.sqlProvider.telemetryService.Slogger.Info("rolling back transaction", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly())
	if sqlTransaction.state == nil {
		sqlTransaction.sqlProvider.telemetryService.Slogger.Error("can't rollback because transaction not active", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly())
		return fmt.Errorf("can't rollback because transaction not active")
	}

	err := sqlTransaction.state.sqlTx.Rollback()
	if err != nil {
		sqlTransaction.sqlProvider.telemetryService.Slogger.Error("failed to rollback transaction", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly(), "error", err)
		return err
	}

	sqlTransaction.sqlProvider.telemetryService.Slogger.Info("rolled back transaction", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly())
	sqlTransaction.state = nil
	return nil
}
