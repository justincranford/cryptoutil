package orm

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

type RepositoryTransaction struct {
	repositoryProvider *RepositoryProvider
	guardState         sync.Mutex
	state              *RepositoryTransactionState
}

type RepositoryTransactionState struct {
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

// RepositoryProvider

func (r *RepositoryProvider) WithTransaction(ctx context.Context, transactionMode TransactionMode, function func(repositoryTransaction *RepositoryTransaction) error) error {
	tx := &RepositoryTransaction{repositoryProvider: r}

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
			r.telemetryService.Slogger.Error("panic occurred during transaction", "txID", tx.ID(), "mode", tx.Mode(), "panic", recover)
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

func (tx *RepositoryTransaction) ID() *googleUuid.UUID {
	// tx.guardState.Lock()
	// defer tx.guardState.Unlock()
	if tx.state == nil {
		return nil
	}
	return &tx.state.txID
}

func (tx *RepositoryTransaction) Context() context.Context {
	// tx.guardState.Lock()
	// defer tx.guardState.Unlock()
	if tx.state == nil {
		return nil
	}
	return tx.state.ctx
}

func (tx *RepositoryTransaction) Mode() *TransactionMode {
	// tx.guardState.Lock()
	// defer tx.guardState.Unlock()
	if tx.state == nil {
		return nil
	}
	return &tx.state.txMode
}

// Helpers

func (tx *RepositoryTransaction) begin(ctx context.Context, transactionMode TransactionMode) error {
	tx.guardState.Lock()
	defer tx.guardState.Unlock()

	tx.repositoryProvider.telemetryService.Slogger.Info("beginning transaction", "mode", transactionMode)

	if tx.state != nil {
		tx.repositoryProvider.telemetryService.Slogger.Error("transaction already started", "txID", tx.ID(), "mode", tx.Mode())
		return fmt.Errorf("transaction already started")
	}

	txID := tx.repositoryProvider.uuidV7Pool.Get().Private.(googleUuid.UUID)
	gormTx, err := tx.beginImplementation(ctx, transactionMode, txID)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	tx.state = &RepositoryTransactionState{ctx: ctx, txMode: transactionMode, txID: txID, gormTx: gormTx}
	tx.repositoryProvider.telemetryService.Slogger.Info("started transaction", "txID", txID, "mode", transactionMode)
	return nil
}

func (tx *RepositoryTransaction) commit() error {
	tx.guardState.Lock()
	defer tx.guardState.Unlock()

	tx.repositoryProvider.telemetryService.Slogger.Info("committing transaction", "txID", tx.ID(), "mode", tx.Mode())

	if tx.state == nil {
		tx.repositoryProvider.telemetryService.Slogger.Error("can't commit because transaction not active", "txID", tx.ID(), "mode", tx.Mode())
		return fmt.Errorf("can't commit because transaction not active")
	} else if tx.state.txMode == AutoCommit {
		tx.repositoryProvider.telemetryService.Slogger.Error("can't commit because transaction is autocommit", "txID", tx.ID(), "mode", tx.Mode())
		return fmt.Errorf("can't commit because transaction is autocommit")
	}

	if _, err := tx.commitImplementation(); err != nil {
		tx.repositoryProvider.telemetryService.Slogger.Error("failed to commit transaction", "txID", tx.ID(), "mode", tx.Mode(), "error", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	tx.repositoryProvider.telemetryService.Slogger.Info("committed transaction", "txID", tx.ID(), "mode", tx.Mode())
	tx.state = nil
	return nil
}

func (tx *RepositoryTransaction) rollback() error {
	tx.guardState.Lock()
	defer tx.guardState.Unlock()

	tx.repositoryProvider.telemetryService.Slogger.Info("rolling back transaction", "txID", tx.ID(), "mode", tx.Mode())

	if tx.state == nil {
		tx.repositoryProvider.telemetryService.Slogger.Error("can't rollback because transaction not active", "txID", tx.ID(), "mode", tx.Mode())
		return fmt.Errorf("can't rollback because transaction not active")
	} else if tx.state.txMode == AutoCommit {
		tx.repositoryProvider.telemetryService.Slogger.Error("can't rollback because transaction is autocommit", "txID", tx.ID(), "mode", tx.Mode())
		return fmt.Errorf("can't rollback because transaction is autocommit")
	}

	if _, err := tx.rollbackImplementation(); err != nil {
		tx.repositoryProvider.telemetryService.Slogger.Error("failed to rollback transaction", "txID", tx.ID(), "mode", tx.Mode(), "error", err)
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	tx.repositoryProvider.telemetryService.Slogger.Info("rolled back transaction", "txID", tx.ID(), "mode", tx.Mode())
	tx.state = nil
	return nil
}

// Implementation using Gorm

func (tx *RepositoryTransaction) beginImplementation(ctx context.Context, transactionMode TransactionMode, txID googleUuid.UUID) (*gorm.DB, error) {
	gormTx := tx.repositoryProvider.gormDB.WithContext(ctx)
	if transactionMode == AutoCommit {
		return gormTx, nil
	} else if transactionMode == ReadWrite {
		gormTx = gormTx.Begin()
	} else {
		gormTx = gormTx.Begin(&sql.TxOptions{ReadOnly: true})
	}
	if gormTx.Error != nil {
		return nil, fmt.Errorf("failed to begin gorm transaction: %w", gormTx.Error)
	}
	return gormTx, nil
}

func (tx *RepositoryTransaction) commitImplementation() (*gorm.DB, error) {
	gormTx := tx.state.gormTx.Commit()
	if gormTx.Error != nil {
		return nil, fmt.Errorf("failed to commit gorm transaction: %w", gormTx.Error)
	}
	return gormTx, nil
}

func (tx *RepositoryTransaction) rollbackImplementation() (*gorm.DB, error) {
	gormTx := tx.state.gormTx.Rollback()
	if gormTx.Error != nil {
		return nil, fmt.Errorf("failed to rollback gorm transaction: %w", gormTx.Error)
	}
	return gormTx, nil
}
