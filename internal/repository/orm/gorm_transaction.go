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
	ctx        context.Context
	autoCommit bool
	readOnly   bool
	txID       googleUuid.UUID
	gormTx     *gorm.DB
}

// RepositoryProvider

func (r *RepositoryProvider) WithTransaction(ctx context.Context, autoCommit, readOnly bool, function func(repositoryTransaction *RepositoryTransaction) error) error {
	tx := &RepositoryTransaction{repositoryProvider: r}

	err := tx.begin(ctx, autoCommit, readOnly)
	if err != nil {
		r.telemetryService.Slogger.Error("failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if tx.state != nil && !tx.state.autoCommit { // ensure commit() or other rollback() calls didn't overwrite tx.state with nil
			if err := tx.rollback(); err != nil {
				r.telemetryService.Slogger.Error("failed to rollback transaction", "txID", tx.ID(), "autoCommit", tx.AutoCommit(), "readOnly", tx.ReadOnly(), "error", err)
			}
		}
		if recover := recover(); recover != nil {
			r.telemetryService.Slogger.Error("panic occurred during transaction", "txID", tx.ID(), "autoCommit", tx.AutoCommit(), "readOnly", tx.ReadOnly(), "panic", recover)
			panic(recover)
		}
	}()

	if err := function(tx); err != nil {
		r.telemetryService.Slogger.Error("transaction function failed", "txID", tx.ID(), "autoCommit", tx.AutoCommit(), "readOnly", tx.ReadOnly(), "error", err)
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	if !tx.state.autoCommit {
		if err := tx.commit(); err != nil { // clears state
			r.telemetryService.Slogger.Error("failed to commit transaction", "txID", tx.ID(), "autoCommit", tx.AutoCommit(), "readOnly", tx.ReadOnly(), "error", err)
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}
	return nil
}

// RepositoryTransaction

func (tx *RepositoryTransaction) ID() *googleUuid.UUID {
	if tx.state == nil {
		return nil
	}
	transactionIDCopy := googleUuid.UUID(tx.state.txID)
	return &transactionIDCopy
}

func (repositoryTransaction *RepositoryTransaction) Context() context.Context {
	if repositoryTransaction.state == nil {
		return nil
	}
	return repositoryTransaction.state.ctx
}

func (repositoryTransaction *RepositoryTransaction) AutoCommit() bool {
	if repositoryTransaction.state == nil {
		return false
	}
	return repositoryTransaction.state.autoCommit
}

func (repositoryTransaction *RepositoryTransaction) ReadOnly() bool {
	if repositoryTransaction.state == nil {
		return false
	}
	return repositoryTransaction.state.readOnly
}

// Helpers

func (tx *RepositoryTransaction) begin(ctx context.Context, autoCommit, readOnly bool) error {
	tx.guardState.Lock()
	defer tx.guardState.Unlock()

	tx.repositoryProvider.telemetryService.Slogger.Info("beginning transaction", "autoCommit", autoCommit, "readOnly", readOnly)

	if tx.state != nil {
		tx.repositoryProvider.telemetryService.Slogger.Error("transaction already started", "txID", tx.ID(), "autoCommit", tx.AutoCommit(), "readOnly", tx.ReadOnly())
		return fmt.Errorf("transaction already started")
	}

	txID, err := googleUuid.NewV7()
	if err != nil {
		tx.repositoryProvider.telemetryService.Slogger.Error("failed to generate transaction ID", "error", err)
		return fmt.Errorf("failed to generate transaction ID: %w", err)
	}

	gormTx, err := tx.beginImplementation(ctx, autoCommit, readOnly, txID)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	tx.state = &RepositoryTransactionState{ctx: ctx, autoCommit: autoCommit, readOnly: readOnly, txID: txID, gormTx: gormTx}
	tx.repositoryProvider.telemetryService.Slogger.Info("started transaction", "txID", txID, "autoCommit", autoCommit, "readOnly", readOnly)
	return nil
}

func (tx *RepositoryTransaction) commit() error {
	tx.guardState.Lock()
	defer tx.guardState.Unlock()

	tx.repositoryProvider.telemetryService.Slogger.Info("committing transaction", "txID", tx.ID(), "autoCommit", tx.AutoCommit(), "readOnly", tx.ReadOnly())

	if tx.state == nil {
		tx.repositoryProvider.telemetryService.Slogger.Error("can't commit because transaction not active", "txID", tx.ID(), "autoCommit", tx.AutoCommit(), "readOnly", tx.ReadOnly())
		return fmt.Errorf("can't commit because transaction not active")
	} else if tx.state.autoCommit {
		tx.repositoryProvider.telemetryService.Slogger.Error("can't commit because transaction is autocommit", "txID", tx.ID(), "autoCommit", tx.AutoCommit(), "readOnly", tx.ReadOnly())
		return fmt.Errorf("can't commit because transaction is autocommit")
	}

	if err := tx.commitImplementation(); err != nil {
		tx.repositoryProvider.telemetryService.Slogger.Error("failed to commit transaction", "txID", tx.ID(), "autoCommit", tx.AutoCommit(), "readOnly", tx.ReadOnly(), "error", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	tx.repositoryProvider.telemetryService.Slogger.Info("committed transaction", "txID", tx.ID(), "autoCommit", tx.AutoCommit(), "readOnly", tx.ReadOnly())
	tx.state = nil
	return nil
}

func (tx *RepositoryTransaction) rollback() error {
	tx.guardState.Lock()
	defer tx.guardState.Unlock()

	tx.repositoryProvider.telemetryService.Slogger.Info("rolling back transaction", "txID", tx.ID(), "autoCommit", tx.AutoCommit(), "readOnly", tx.ReadOnly())

	if tx.state == nil {
		tx.repositoryProvider.telemetryService.Slogger.Error("can't rollback because transaction not active", "txID", tx.ID(), "autoCommit", tx.AutoCommit(), "readOnly", tx.ReadOnly())
		return fmt.Errorf("can't rollback because transaction not active")
	} else if tx.state.autoCommit {
		tx.repositoryProvider.telemetryService.Slogger.Error("can't rollback because transaction is autocommit", "txID", tx.ID(), "autoCommit", tx.AutoCommit(), "readOnly", tx.ReadOnly())
		return fmt.Errorf("can't rollback because transaction is autocommit")
	}

	if err := tx.rollbackImplementation(); err != nil {
		tx.repositoryProvider.telemetryService.Slogger.Error("failed to rollback transaction", "txID", tx.ID(), "autoCommit", tx.AutoCommit(), "readOnly", tx.ReadOnly(), "error", err)
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	tx.repositoryProvider.telemetryService.Slogger.Info("rolled back transaction", "txID", tx.ID(), "autoCommit", tx.AutoCommit(), "readOnly", tx.ReadOnly())
	tx.state = nil
	return nil
}

// Implementation using Gorm

func (tx *RepositoryTransaction) beginImplementation(ctx context.Context, autoCommit bool, readOnly bool, txID googleUuid.UUID) (*gorm.DB, error) {
	gormTx := tx.repositoryProvider.gormDB.WithContext(ctx)
	if !autoCommit {
		gormTx = gormTx.Begin(&sql.TxOptions{ReadOnly: readOnly})
		if gormTx.Error != nil {
			return nil, fmt.Errorf("failed to begin gorm transaction: %w", gormTx.Error)
		}
	}
	return gormTx, nil
}

func (tx *RepositoryTransaction) commitImplementation() error {
	gormTx := tx.state.gormTx.Commit()
	if gormTx.Error != nil {
		return fmt.Errorf("failed to commit gorm transaction: %w", gormTx.Error)
	}
	return nil
}

func (tx *RepositoryTransaction) rollbackImplementation() error {
	gormTx := tx.state.gormTx.Rollback()
	if gormTx.Error != nil {
		return fmt.Errorf("failed to rollback gorm transaction: %w", gormTx.Error)
	}
	return nil
}
