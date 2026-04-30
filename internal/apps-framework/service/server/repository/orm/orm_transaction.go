// Copyright (c) 2025-2026 Justin Cranford.
//
//

package orm

import (
	"context"
	"database/sql"
	"fmt"
	"runtime/debug"
	"sync"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// OrmTransaction represents a database transaction with explicit lifecycle management.
type OrmTransaction struct {
	ormRepository *OrmRepository
	guardState    sync.Mutex
	state         *OrmTransactionState
}

// OrmTransactionState holds the internal state of an ORM transaction.
type OrmTransactionState struct {
	ctx    context.Context
	txMode TransactionMode
	txID   googleUuid.UUID
	gormTx *gorm.DB
}

// TransactionMode specifies the isolation and commit behavior of a transaction.
type TransactionMode string

// AutoCommit, ReadWrite, ReadOnly are the supported transaction modes.
var (
	AutoCommit TransactionMode = "AutoCommit"
	// ReadWrite represents a read-write transaction.
	ReadWrite TransactionMode = "ReadWrite"
	// ReadOnly represents a read-only transaction.
	ReadOnly TransactionMode = "ReadOnly"
)

// NewOrmTransactionWithRepository creates an OrmTransaction bound to the given repository.
// Intended for testing error-path behaviour (e.g. commit/rollback without a prior Begin).
func NewOrmTransactionWithRepository(repo *OrmRepository) *OrmTransaction {
	return &OrmTransaction{ormRepository: repo}
}

// GormTx returns the underlying GORM transaction (or base DB for AutoCommit mode).
// Returns nil when no transaction is active.
func (tx *OrmTransaction) GormTx() *gorm.DB {
	if tx.state == nil {
		return nil
	}

	return tx.state.gormTx
}

// WithTransaction executes fn within a managed database transaction.
func (r *OrmRepository) WithTransaction(ctx context.Context, transactionMode TransactionMode, fn func(ormTransaction *OrmTransaction) error) error {
	tx := &OrmTransaction{ormRepository: r}

	err := tx.Begin(ctx, transactionMode)
	if err != nil {
		r.telemetryService.Slogger.Error("failed to begin transaction", cryptoutilSharedMagic.StringError, err)

		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if tx.state != nil && tx.state.txMode != AutoCommit {
			if err := tx.Rollback(); err != nil {
				r.telemetryService.Slogger.Error("failed to rollback transaction", "txID", tx.ID(), "mode", tx.Mode(), cryptoutilSharedMagic.StringError, err)
			}
		}

		if txRecover := recover(); txRecover != nil {
			r.telemetryService.Slogger.Error("panic occurred during transaction", "txID", tx.ID(), "mode", tx.Mode(), "panic", txRecover, "stack", string(debug.Stack()))
			panic(txRecover)
		}
	}()

	if err := fn(tx); err != nil {
		r.telemetryService.Slogger.Error("transaction function failed", "txID", tx.ID(), "mode", tx.Mode(), cryptoutilSharedMagic.StringError, err)

		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	if tx.state.txMode != AutoCommit {
		if err := tx.Commit(); err != nil {
			r.telemetryService.Slogger.Error("failed to commit transaction", "txID", tx.ID(), "mode", tx.Mode(), cryptoutilSharedMagic.StringError, err)

			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	return nil
}

// ID returns the unique identifier of the transaction.
func (tx *OrmTransaction) ID() *googleUuid.UUID {
	if tx.state == nil {
		return nil
	}

	return &tx.state.txID
}

// Context returns the context associated with the transaction.
func (tx *OrmTransaction) Context() context.Context {
	if tx.state == nil {
		return nil
	}

	return tx.state.ctx
}

// Mode returns the transaction mode (AutoCommit, ReadWrite, or ReadOnly).
func (tx *OrmTransaction) Mode() *TransactionMode {
	if tx.state == nil {
		return nil
	}

	return &tx.state.txMode
}

// Begin starts the transaction with the given mode.
// Returns an error if the transaction is already active.
func (tx *OrmTransaction) Begin(ctx context.Context, transactionMode TransactionMode) error {
	tx.guardState.Lock()
	defer tx.guardState.Unlock()

	if tx.ormRepository.verboseMode {
		tx.ormRepository.telemetryService.Slogger.Debug("beginning transaction", "mode", transactionMode)
	}

	if tx.state != nil {
		tx.ormRepository.telemetryService.Slogger.Error("transaction already started", "txID", tx.ID(), "mode", tx.Mode())

		return fmt.Errorf("transaction already started")
	}

	txID := tx.ormRepository.jwkGenService.GenerateUUIDv7()

	gormTx, err := tx.beginImplementation(ctx, transactionMode)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	tx.state = &OrmTransactionState{ctx: ctx, txMode: transactionMode, txID: *txID, gormTx: gormTx}
	tx.ormRepository.telemetryService.Slogger.Debug("started transaction", "txID", txID, "mode", transactionMode)

	return nil
}

// Commit commits the active transaction and clears the state.
// Returns an error if the transaction is not active or is in AutoCommit mode.
func (tx *OrmTransaction) Commit() error {
	tx.guardState.Lock()
	defer tx.guardState.Unlock()

	if tx.ormRepository.verboseMode {
		tx.ormRepository.telemetryService.Slogger.Debug("committing transaction", "txID", tx.ID(), "mode", tx.Mode())
	}

	if tx.state == nil {
		tx.ormRepository.telemetryService.Slogger.Error("can't commit because transaction not active", "txID", tx.ID(), "mode", tx.Mode())

		return fmt.Errorf("can't commit because transaction not active")
	} else if tx.state.txMode == AutoCommit {
		tx.ormRepository.telemetryService.Slogger.Error("can't commit because transaction is autocommit", "txID", tx.ID(), "mode", tx.Mode())

		return fmt.Errorf("can't commit because transaction is autocommit")
	}

	if _, err := tx.commitImplementation(); err != nil {
		tx.ormRepository.telemetryService.Slogger.Error("failed to commit transaction", "txID", tx.ID(), "mode", tx.Mode(), cryptoutilSharedMagic.StringError, err)

		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	tx.ormRepository.telemetryService.Slogger.Debug("committed transaction", "txID", tx.ID(), "mode", tx.Mode())
	tx.state = nil

	return nil
}

// Rollback rolls back the active transaction and clears the state.
// Returns an error if the transaction is not active or is in AutoCommit mode.
func (tx *OrmTransaction) Rollback() error {
	tx.guardState.Lock()
	defer tx.guardState.Unlock()

	if tx.ormRepository.verboseMode {
		tx.ormRepository.telemetryService.Slogger.Debug("rolling back transaction", "txID", tx.ID(), "mode", tx.Mode())
	}

	if tx.state == nil {
		tx.ormRepository.telemetryService.Slogger.Error("can't rollback because transaction not active", "txID", tx.ID(), "mode", tx.Mode())

		return fmt.Errorf("can't rollback because transaction not active")
	} else if tx.state.txMode == AutoCommit {
		tx.ormRepository.telemetryService.Slogger.Error("can't rollback because transaction is autocommit", "txID", tx.ID(), "mode", tx.Mode())

		return fmt.Errorf("can't rollback because transaction is autocommit")
	}

	if _, err := tx.rollbackImplementation(); err != nil {
		tx.ormRepository.telemetryService.Slogger.Error("failed to rollback transaction", "txID", tx.ID(), "mode", tx.Mode(), cryptoutilSharedMagic.StringError, err)

		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	tx.ormRepository.telemetryService.Slogger.Warn("rolled back transaction", "txID", tx.ID(), "mode", tx.Mode())
	tx.state = nil

	return nil
}

func (tx *OrmTransaction) beginImplementation(ctx context.Context, transactionMode TransactionMode) (*gorm.DB, error) {
	gormTx := tx.ormRepository.gormDB.WithContext(ctx)

	switch transactionMode {
	case AutoCommit:
		return gormTx, nil
	case ReadWrite:
		gormTx = gormTx.Begin(&sql.TxOptions{Isolation: sql.LevelReadCommitted})
	default:
		gormTx = gormTx.Begin(&sql.TxOptions{Isolation: sql.LevelReadCommitted, ReadOnly: true})
	}

	if gormTx.Error != nil {
		return nil, fmt.Errorf("failed to begin gorm transaction: %w", gormTx.Error)
	}

	return gormTx, nil
}

func (tx *OrmTransaction) commitImplementation() (*gorm.DB, error) {
	gormTx := tx.state.gormTx.Commit()
	if gormTx.Error != nil {
		return nil, fmt.Errorf("failed to commit gorm transaction: %w", gormTx.Error)
	}

	return gormTx, nil
}

func (tx *OrmTransaction) rollbackImplementation() (*gorm.DB, error) {
	gormTx := tx.state.gormTx.Rollback()
	if gormTx.Error != nil {
		return nil, fmt.Errorf("failed to rollback gorm transaction: %w", gormTx.Error)
	}

	return gormTx, nil
}
