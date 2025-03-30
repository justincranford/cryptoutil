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
	guardState  sync.Mutex
	state       *SqlTransactionState
}

type SqlTransactionState struct {
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
		if sqlTransaction.state != nil { // Avoid rollback if already committed or rolled back
			if err := sqlTransaction.Rollback(); err != nil {
				sp.telemetryService.Slogger.Error("failed to rollback transaction", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly(), "error", err)
			}
		}
		if r := recover(); r != nil {
			sp.telemetryService.Slogger.Error("panic occurred during transaction", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly(), "panic", r)
			panic(r) // re-throw the panic after rollback
		}
	}()

	if err := function(sqlTransaction); err != nil {
		sp.telemetryService.Slogger.Error("transaction function failed", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly(), "error", err)
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	return sqlTransaction.Commit()
}

func (pr *SqlProvider) NewTransaction() (*SqlTransaction, error) {
	pr.telemetryService.Slogger.Info("new transaction")
	return &SqlTransaction{sqlProvider: pr}, nil
}

// TransactionID Transaction ID is valid (non-nil) only when a transaction is active
func (tx *SqlTransaction) TransactionID() *googleUuid.UUID {
	if tx.state == nil {
		return nil
	}
	transactionIDCopy := googleUuid.UUID(tx.state.transactionID)
	return &transactionIDCopy
}

// Context Transaction context is valid (non-nil) only when a transaction is active
func (tx *SqlTransaction) Context() context.Context {
	if tx.state == nil {
		return nil
	}
	return tx.state.ctx
}

// IsReadOnly Boolean for if a transaction read-only is valid only when a transaction is active
func (tx *SqlTransaction) IsReadOnly() bool {
	if tx.state == nil {
		return false
	}
	return tx.state.readOnly
}

func (tx *SqlTransaction) Begin(ctx context.Context, readOnly bool) error {
	tx.guardState.Lock()
	defer tx.guardState.Unlock()

	tx.sqlProvider.telemetryService.Slogger.Info("beginning transaction", "readOnly", readOnly)
	if tx.state != nil {
		tx.sqlProvider.telemetryService.Slogger.Error("transaction already started", "transactionID", tx.TransactionID())
		return fmt.Errorf("transaction already started")
	}

	transactionID, err := googleUuid.NewV7()
	if err != nil {
		tx.sqlProvider.telemetryService.Slogger.Error("failed to generate transaction ID", "error", err)
		return fmt.Errorf("failed to generate transaction ID: %w", err)
	}

	sqlTx, err := tx.sqlProvider.sqlDB.BeginTx(ctx, &sql.TxOptions{ReadOnly: readOnly})
	if err != nil {
		tx.sqlProvider.telemetryService.Slogger.Error("failed to begin transaction", "transactionID", transactionID, "readOnly", readOnly, "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	tx.state = &SqlTransactionState{ctx: ctx, readOnly: readOnly, transactionID: transactionID, sqlTx: sqlTx}
	tx.sqlProvider.telemetryService.Slogger.Info("started transaction", "transactionID", transactionID, "readOnly", readOnly)
	return nil
}

func (tx *SqlTransaction) Commit() error {
	tx.guardState.Lock()
	defer tx.guardState.Unlock()

	tx.sqlProvider.telemetryService.Slogger.Info("committing transaction", "transactionID", tx.TransactionID(), "readOnly", tx.IsReadOnly())
	if tx.state == nil {
		tx.sqlProvider.telemetryService.Slogger.Error("can't commit because transaction not active", "transactionID", tx.TransactionID(), "readOnly", tx.IsReadOnly())
		return fmt.Errorf("can't commit because transaction not active")
	}

	err := tx.state.sqlTx.Commit()
	if err != nil {
		tx.sqlProvider.telemetryService.Slogger.Error("failed to commit transaction", "transactionID", tx.TransactionID(), "readOnly", tx.IsReadOnly(), "error", err)
		return err
	}

	tx.sqlProvider.telemetryService.Slogger.Info("committed transaction", "transactionID", tx.TransactionID(), "readOnly", tx.IsReadOnly())
	tx.state = nil
	return nil
}

func (tx *SqlTransaction) Rollback() error {
	tx.guardState.Lock()
	defer tx.guardState.Unlock()

	tx.sqlProvider.telemetryService.Slogger.Info("rolling back transaction", "transactionID", tx.TransactionID(), "readOnly", tx.IsReadOnly())
	if tx.state == nil {
		tx.sqlProvider.telemetryService.Slogger.Error("can't rollback because transaction not active", "transactionID", tx.TransactionID(), "readOnly", tx.IsReadOnly())
		return fmt.Errorf("can't rollback because transaction not active")
	}

	err := tx.state.sqlTx.Rollback()
	if err != nil {
		tx.sqlProvider.telemetryService.Slogger.Error("failed to rollback transaction", "transactionID", tx.TransactionID(), "readOnly", tx.IsReadOnly(), "error", err)
		return err
	}

	tx.sqlProvider.telemetryService.Slogger.Info("rolled back transaction", "transactionID", tx.TransactionID(), "readOnly", tx.IsReadOnly())
	tx.state = nil
	return nil
}
