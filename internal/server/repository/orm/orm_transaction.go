package orm

import (
	"context"
	"database/sql"
	"fmt"
	"runtime/debug"
	"sync"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

type OrmTransaction struct {
	ormRepository *OrmRepository
	guardState    sync.Mutex
	state         *OrmTransactionState
}

type OrmTransactionState struct {
	ctx    context.Context
	txMode TransactionMode
	txID   googleUuid.UUID
	gormTx *gorm.DB
}

type TransactionMode string

var (
	AutoCommit TransactionMode = "AutoCommit"
	ReadWrite  TransactionMode = "ReadWrite"
	ReadOnly   TransactionMode = "ReadOnly"
)

// OrmRepository

func (r *OrmRepository) WithTransaction(ctx context.Context, transactionMode TransactionMode, function func(ormTransaction *OrmTransaction) error) error {
	tx := &OrmTransaction{ormRepository: r}

	err := tx.begin(ctx, transactionMode)
	if err != nil {
		r.telemetryService.Slogger.Error("failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if tx.state != nil && tx.state.txMode != AutoCommit { // watch out for other commit() or rollback() calls that set tx.state as nil
			if err := tx.rollback(); err != nil {
				r.telemetryService.Slogger.Error("failed to rollback transaction", "txID", tx.ID(), "mode", tx.Mode(), "error", err)
			}
		}
		if recover := recover(); recover != nil {
			r.telemetryService.Slogger.Error("panic occurred during transaction", "txID", tx.ID(), "mode", tx.Mode(), "panic", recover, "stack", string(debug.Stack()))
			panic(recover)
		}
	}()

	if err := function(tx); err != nil {
		r.telemetryService.Slogger.Error("transaction function failed", "txID", tx.ID(), "mode", tx.Mode(), "error", err)
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	if tx.state.txMode != AutoCommit {
		if err := tx.commit(); err != nil { // clears state
			r.telemetryService.Slogger.Error("failed to commit transaction", "txID", tx.ID(), "mode", tx.Mode(), "error", err)
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}
	return nil
}

// RepositoryTransaction

func (tx *OrmTransaction) ID() *googleUuid.UUID {
	// tx.guardState.Lock()
	// defer tx.guardState.Unlock()
	if tx.state == nil {
		return nil
	}
	return &tx.state.txID
}

func (tx *OrmTransaction) Context() context.Context {
	// tx.guardState.Lock()
	// defer tx.guardState.Unlock()
	if tx.state == nil {
		return nil
	}
	return tx.state.ctx
}

func (tx *OrmTransaction) Mode() *TransactionMode {
	// tx.guardState.Lock()
	// defer tx.guardState.Unlock()
	if tx.state == nil {
		return nil
	}
	return &tx.state.txMode
}

// Helpers

func (tx *OrmTransaction) begin(ctx context.Context, transactionMode TransactionMode) error {
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

func (tx *OrmTransaction) commit() error {
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
		tx.ormRepository.telemetryService.Slogger.Error("failed to commit transaction", "txID", tx.ID(), "mode", tx.Mode(), "error", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	tx.ormRepository.telemetryService.Slogger.Debug("committed transaction", "txID", tx.ID(), "mode", tx.Mode())
	tx.state = nil
	return nil
}

func (tx *OrmTransaction) rollback() error {
	tx.guardState.Lock()
	defer tx.guardState.Unlock()

	tx.ormRepository.telemetryService.Slogger.Warn("rolling back transaction", "txID", tx.ID(), "mode", tx.Mode())

	if tx.state == nil {
		tx.ormRepository.telemetryService.Slogger.Error("can't rollback because transaction not active", "txID", tx.ID(), "mode", tx.Mode())
		return fmt.Errorf("can't rollback because transaction not active")
	} else if tx.state.txMode == AutoCommit {
		tx.ormRepository.telemetryService.Slogger.Error("can't rollback because transaction is autocommit", "txID", tx.ID(), "mode", tx.Mode())
		return fmt.Errorf("can't rollback because transaction is autocommit")
	}

	if _, err := tx.rollbackImplementation(); err != nil {
		tx.ormRepository.telemetryService.Slogger.Error("failed to rollback transaction", "txID", tx.ID(), "mode", tx.Mode(), "error", err)
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	tx.ormRepository.telemetryService.Slogger.Warn("rolled back transaction", "txID", tx.ID(), "mode", tx.Mode())
	tx.state = nil
	return nil
}

// Implementation using Gorm

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
