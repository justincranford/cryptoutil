// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository

import (
	"context"
	"database/sql"
	"fmt"
	"runtime/debug"
	"sync"

	googleUuid "github.com/google/uuid"
)

// SQLTransaction represents a database transaction.
type SQLTransaction struct {
	sqlRepository *SQLRepository
	guardState    sync.Mutex
	state         *SQLTransactionState
}

// SQLTransactionState represents the internal state of a transaction.
// SQLTransactionState represents the internal state of a transaction.
type SQLTransactionState struct {
	ctx           context.Context
	readOnly      bool
	transactionID googleUuid.UUID
	sqlTx         *sql.Tx
}

// WithTransaction executes a function within a database transaction.
func (s *SQLRepository) WithTransaction(ctx context.Context, readOnly bool, function func(sqlTransaction *SQLTransaction) error) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	} else if function == nil {
		return fmt.Errorf("function cannot be nil")
	}

	if function == nil {
		return fmt.Errorf("function cannot be nil")
	}

	if readOnly {
		switch s.dbType {
		case DBTypeSQLite: // SQLite lacks support for read-only transactions
			s.telemetryService.Slogger.Warn("database doesn't support read-only transactions", "dbType", string(s.dbType))

			return fmt.Errorf("database %s doesn't support read-only transactions", string(s.dbType))
		case DBTypePostgres:
			s.telemetryService.Slogger.Debug("database supports read-only transactions", "dbType", string(s.dbType))
		default:
			return fmt.Errorf("%w: %s", ErrUnsupportedDBType, string(s.dbType))
		}
	}

	sqlTransaction, err := s.newTransaction()
	if err != nil {
		s.telemetryService.Slogger.Error("failed to create transaction", "error", err)

		return fmt.Errorf("failed to create transaction: %w", err)
	}

	err = sqlTransaction.begin(ctx, readOnly)
	if err != nil {
		s.telemetryService.Slogger.Error("failed to begin transaction", "error", err)

		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if sqlTransaction.state != nil { // Avoid rollback if already committed or rolled back
			if err := sqlTransaction.rollback(); err != nil {
				s.telemetryService.Slogger.Error("failed to rollback transaction", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly(), "error", err)
			}
		}

		if r := recover(); r != nil {
			s.telemetryService.Slogger.Error("panic occurred during transaction", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly(), "panic", r, "stack", string(debug.Stack()))
			panic(r) // re-throw the panic after rollback
		}
	}()

	if err := function(sqlTransaction); err != nil {
		s.telemetryService.Slogger.Error("transaction function failed", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly(), "error", err)

		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	return sqlTransaction.commit()
}

func (s *SQLRepository) newTransaction() (*SQLTransaction, error) {
	if s == nil {
		return nil, fmt.Errorf("SQL repository cannot be nil")
	} else if s.sqlDB == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	if s.verboseMode {
		s.telemetryService.Slogger.Debug("new transaction")
	}

	return &SQLTransaction{sqlRepository: s}, nil
}

// TransactionID Transaction ID is valid (non-nil) only when a transaction is active.
func (sqlTransaction *SQLTransaction) TransactionID() *googleUuid.UUID {
	if sqlTransaction.state == nil {
		return nil
	}

	transactionIDCopy := sqlTransaction.state.transactionID

	return &transactionIDCopy
}

// Context Transaction context is valid (non-nil) only when a transaction is active.
func (sqlTransaction *SQLTransaction) Context() context.Context {
	if sqlTransaction.state == nil {
		return nil
	}

	return sqlTransaction.state.ctx
}

// IsReadOnly Boolean for if a transaction read-only is valid only when a transaction is active.
func (sqlTransaction *SQLTransaction) IsReadOnly() bool {
	if sqlTransaction.state == nil {
		return false
	}

	return sqlTransaction.state.readOnly
}

func (sqlTransaction *SQLTransaction) begin(ctx context.Context, readOnly bool) error {
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

	// NOTE: PostgreSQL does not support ReadOnly transactions, so always use default (ReadWrite)
	// See: 01-04.database.instructions.md - "SQLite does NOT support read-only transactions - NEVER use them"
	sqlTx, err := sqlTransaction.sqlRepository.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		sqlTransaction.sqlRepository.telemetryService.Slogger.Error("failed to begin transaction", "transactionID", transactionID, "readOnly", readOnly, "error", err)

		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	sqlTransaction.state = &SQLTransactionState{ctx: ctx, readOnly: readOnly, transactionID: transactionID, sqlTx: sqlTx}
	sqlTransaction.sqlRepository.telemetryService.Slogger.Debug("started transaction", "transactionID", transactionID, "readOnly", readOnly)

	return nil
}

func (sqlTransaction *SQLTransaction) commit() error {
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

		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	sqlTransaction.sqlRepository.telemetryService.Slogger.Debug("committed transaction", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly())
	sqlTransaction.state = nil

	return nil
}

func (sqlTransaction *SQLTransaction) rollback() error {
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

		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	sqlTransaction.sqlRepository.telemetryService.Slogger.Warn("rolled back transaction", "transactionID", sqlTransaction.TransactionID(), "readOnly", sqlTransaction.IsReadOnly())
	sqlTransaction.state = nil

	return nil
}
