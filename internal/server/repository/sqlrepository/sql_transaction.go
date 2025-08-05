package sqlrepository

import (
	"context"
	"database/sql"
	"fmt"
	"runtime/debug"
	"sync"

	googleUuid "github.com/google/uuid"
)

type SqlTransaction struct {
	sqlRepository *SqlRepository
	guardState    sync.Mutex
	state         *SqlTransactionState
}

type SqlTransactionState struct {
	ctx           context.Context
	readOnly      bool
	transactionID googleUuid.UUID
	sqlTx         *sql.Tx
}

func (sqlRepository *SqlRepository) WithTransaction(ctx context.Context, readOnly bool, function func(sqlTransaction *SqlTransaction) error) error {
	if readOnly {
		switch sqlRepository.dbType {
		case DBTypeSQLite: // SQLite lacks support for read-only transactions
			sqlRepository.telemetryService.Slogger.Warn("database doesn't support read-only transactions", "dbType", string(sqlRepository.dbType))
			return fmt.Errorf("database %s doesn't support read-only transactions", string(sqlRepository.dbType))
		case DBTypePostgres:
			sqlRepository.telemetryService.Slogger.Debug("database supports read-only transactions", "dbType", string(sqlRepository.dbType))
		default:
			return fmt.Errorf("%w: %s", ErrUnsupportedDBType, string(sqlRepository.dbType))
		}
	}

	sqlTransaction, err := sqlRepository.newTransaction()
	if err != nil {
		sqlRepository.telemetryService.Slogger.Error("failed to create transaction", "error", err)
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	err = sqlTransaction.begin(ctx, readOnly)
	if err != nil {
		sqlRepository.telemetryService.Slogger.Error("failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if sqlTransaction.state != nil { // Avoid rollback if already committed or rolled back
			if err := sqlTransaction.rollback(); err != nil {
				sqlRepository.telemetryService.Slogger.Error("failed to rollback transaction", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly(), "error", err)
			}
		}
		if recover := recover(); recover != nil {
			sqlRepository.telemetryService.Slogger.Error("panic occurred during transaction", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly(), "panic", recover, "stack", string(debug.Stack()))
			panic(recover) // re-throw the panic after rollback
		}
	}()

	if err := function(sqlTransaction); err != nil {
		sqlRepository.telemetryService.Slogger.Error("transaction function failed", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly(), "error", err)
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	return sqlTransaction.commit()
}

func (sqlRepository *SqlRepository) newTransaction() (*SqlTransaction, error) {
	if sqlRepository.verboseMode {
		sqlRepository.telemetryService.Slogger.Debug("new transaction")
	}
	return &SqlTransaction{sqlRepository: sqlRepository}, nil
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

	if sqlTransaction.sqlRepository.verboseMode {
		sqlTransaction.sqlRepository.telemetryService.Slogger.Debug("beginning transaction", "readOnly", readOnly)
	}
	if sqlTransaction.state != nil {
		sqlTransaction.sqlRepository.telemetryService.Slogger.Error("transaction already started", "transactionID", sqlTransaction.TransactionID())
		return fmt.Errorf("transaction already started")
	}

	transactionID, err := googleUuid.NewV7()
	if err != nil {
		sqlTransaction.sqlRepository.telemetryService.Slogger.Error("failed to generate transaction ID", "error", err)
		return fmt.Errorf("failed to generate transaction ID: %w", err)
	}

	sqlTx, err := sqlTransaction.sqlRepository.sqlDB.BeginTx(ctx, &sql.TxOptions{ReadOnly: readOnly})
	if err != nil {
		sqlTransaction.sqlRepository.telemetryService.Slogger.Error("failed to begin transaction", "transactionID", transactionID, "readOnly", readOnly, "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	sqlTransaction.state = &SqlTransactionState{ctx: ctx, readOnly: readOnly, transactionID: transactionID, sqlTx: sqlTx}
	sqlTransaction.sqlRepository.telemetryService.Slogger.Debug("started transaction", "transactionID", transactionID, "readOnly", readOnly)
	return nil
}

func (sqlTransaction *SqlTransaction) commit() error {
	sqlTransaction.guardState.Lock()
	defer sqlTransaction.guardState.Unlock()

	if sqlTransaction.sqlRepository.verboseMode {
		sqlTransaction.sqlRepository.telemetryService.Slogger.Debug("committing transaction", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly())
	}
	if sqlTransaction.state == nil {
		sqlTransaction.sqlRepository.telemetryService.Slogger.Error("can't commit because transaction not active", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly())
		return fmt.Errorf("can't commit because transaction not active")
	}

	err := sqlTransaction.state.sqlTx.Commit()
	if err != nil {
		sqlTransaction.sqlRepository.telemetryService.Slogger.Error("failed to commit transaction", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly(), "error", err)
		return err
	}

	sqlTransaction.sqlRepository.telemetryService.Slogger.Debug("committed transaction", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly())
	sqlTransaction.state = nil
	return nil
}

func (sqlTransaction *SqlTransaction) rollback() error {
	sqlTransaction.guardState.Lock()
	defer sqlTransaction.guardState.Unlock()

	if sqlTransaction.sqlRepository.verboseMode {
		sqlTransaction.sqlRepository.telemetryService.Slogger.Warn("rolling back transaction", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly())
	}
	if sqlTransaction.state == nil {
		sqlTransaction.sqlRepository.telemetryService.Slogger.Error("can't rollback because transaction not active", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly())
		return fmt.Errorf("can't rollback because transaction not active")
	}

	err := sqlTransaction.state.sqlTx.Rollback()
	if err != nil {
		sqlTransaction.sqlRepository.telemetryService.Slogger.Error("failed to rollback transaction", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly(), "error", err)
		return err
	}

	sqlTransaction.sqlRepository.telemetryService.Slogger.Warn("rolled back transaction", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly())
	sqlTransaction.state = nil
	return nil
}
