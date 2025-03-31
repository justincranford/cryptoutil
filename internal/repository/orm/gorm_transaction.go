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
	ctx           context.Context
	autoCommit    bool
	readOnly      bool
	transactionID googleUuid.UUID
	gormTx        *gorm.DB
}

// RepositoryProvider

func (r *RepositoryProvider) WithTransaction(ctx context.Context, autoCommit, readOnly bool, function func(repositoryTransaction *RepositoryTransaction) error) error {
	repositoryTransaction := &RepositoryTransaction{repositoryProvider: r}

	err := repositoryTransaction.begin(ctx, autoCommit, readOnly)
	if err != nil {
		r.telemetryService.Slogger.Error("failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if repositoryTransaction.state != nil && !repositoryTransaction.state.autoCommit {
			if err := repositoryTransaction.rollback(); err != nil {
				r.telemetryService.Slogger.Error("failed to rollback transaction", "transactionID", repositoryTransaction.TransactionID(), "readOnly", repositoryTransaction.IsReadOnly(), "error", err)
			}
		}
		telemetryService := r.telemetryService
		if r := recover(); r != nil {
			telemetryService.Slogger.Error("panic occurred during transaction", "transactionID", repositoryTransaction.TransactionID(), "readOnly", repositoryTransaction.IsReadOnly(), "panic", r)
			panic(r)
		}
	}()

	if err := function(repositoryTransaction); err != nil {
		r.telemetryService.Slogger.Error("transaction function failed", "transactionID", repositoryTransaction.TransactionID(), "readOnly", repositoryTransaction.IsReadOnly(), "error", err)
		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	if repositoryTransaction.state.autoCommit {
		return nil
	}
	return repositoryTransaction.commit()
}

// RepositoryTransaction

func (repositoryTransaction *RepositoryTransaction) TransactionID() *googleUuid.UUID {
	if repositoryTransaction.state == nil {
		return nil
	}
	transactionIDCopy := googleUuid.UUID(repositoryTransaction.state.transactionID)
	return &transactionIDCopy
}

func (repositoryTransaction *RepositoryTransaction) Context() context.Context {
	if repositoryTransaction.state == nil {
		return nil
	}
	return repositoryTransaction.state.ctx
}

func (repositoryTransaction *RepositoryTransaction) IsAutoCommit() bool {
	if repositoryTransaction.state == nil {
		return false
	}
	return repositoryTransaction.state.autoCommit
}

func (repositoryTransaction *RepositoryTransaction) IsReadOnly() bool {
	if repositoryTransaction.state == nil {
		return false
	}
	return repositoryTransaction.state.readOnly
}

// Helpers

func (repositoryTransaction *RepositoryTransaction) begin(ctx context.Context, autoCommit, readOnly bool) error {
	repositoryTransaction.guardState.Lock()
	defer repositoryTransaction.guardState.Unlock()

	repositoryTransaction.repositoryProvider.telemetryService.Slogger.Info("beginning transaction", "autoCommit", autoCommit, "readOnly", readOnly)

	if repositoryTransaction.state != nil {
		repositoryTransaction.repositoryProvider.telemetryService.Slogger.Error("transaction already started", "transactionID", repositoryTransaction.TransactionID(), "autoCommit", repositoryTransaction.IsAutoCommit(), "readOnly", repositoryTransaction.IsReadOnly())
		return fmt.Errorf("transaction already started")
	}

	transactionID, err := googleUuid.NewV7()
	if err != nil {
		repositoryTransaction.repositoryProvider.telemetryService.Slogger.Error("failed to generate transaction ID", "error", err)
		return fmt.Errorf("failed to generate transaction ID: %w", err)
	}

	gormTx := repositoryTransaction.repositoryProvider.gormDB.WithContext(ctx)
	if !autoCommit {
		gormTx = gormTx.Begin(&sql.TxOptions{ReadOnly: readOnly}) // begin TX
		if gormTx.Error != nil {
			repositoryTransaction.repositoryProvider.telemetryService.Slogger.Error("failed to begin transaction", "transactionID", transactionID, "autoCommit", autoCommit, "readOnly", readOnly, "error", gormTx.Error)
			return fmt.Errorf("failed to begin transaction: %w", gormTx.Error)
		}
	}

	repositoryTransaction.state = &RepositoryTransactionState{ctx: ctx, autoCommit: autoCommit, readOnly: readOnly, transactionID: transactionID, gormTx: gormTx}
	repositoryTransaction.repositoryProvider.telemetryService.Slogger.Info("started transaction", "transactionID", transactionID, "autoCommit", autoCommit, "readOnly", readOnly)
	return nil
}

func (repositoryTransaction *RepositoryTransaction) commit() error {
	repositoryTransaction.guardState.Lock()
	defer repositoryTransaction.guardState.Unlock()

	repositoryTransaction.repositoryProvider.telemetryService.Slogger.Info("committing transaction", "transactionID", repositoryTransaction.TransactionID(), "autoCommit", repositoryTransaction.IsAutoCommit(), "readOnly", repositoryTransaction.IsReadOnly())

	if repositoryTransaction.state == nil {
		repositoryTransaction.repositoryProvider.telemetryService.Slogger.Error("can't commit because transaction not active", "transactionID", repositoryTransaction.TransactionID(), "autoCommit", repositoryTransaction.IsAutoCommit(), "readOnly", repositoryTransaction.IsReadOnly())
		return fmt.Errorf("can't commit because transaction not active")
	} else if repositoryTransaction.state.autoCommit {
		repositoryTransaction.repositoryProvider.telemetryService.Slogger.Error("can't commit because transaction is auto committed", "transactionID", repositoryTransaction.TransactionID(), "autoCommit", repositoryTransaction.IsAutoCommit(), "readOnly", repositoryTransaction.IsReadOnly())
		return fmt.Errorf("can't commit because transaction is auto committed")
	}

	if err := repositoryTransaction.state.gormTx.Commit().Error; err != nil {
		repositoryTransaction.repositoryProvider.telemetryService.Slogger.Error("failed to commit transaction", "transactionID", repositoryTransaction.TransactionID(), "autoCommit", repositoryTransaction.IsAutoCommit(), "readOnly", repositoryTransaction.IsReadOnly(), "error", err)
		return err
	}

	repositoryTransaction.repositoryProvider.telemetryService.Slogger.Info("committed transaction", "transactionID", repositoryTransaction.TransactionID(), "autoCommit", repositoryTransaction.IsAutoCommit(), "readOnly", repositoryTransaction.IsReadOnly())
	repositoryTransaction.state = nil
	return nil
}

func (repositoryTransaction *RepositoryTransaction) rollback() error {
	repositoryTransaction.guardState.Lock()
	defer repositoryTransaction.guardState.Unlock()

	repositoryTransaction.repositoryProvider.telemetryService.Slogger.Info("rolling back transaction", "transactionID", repositoryTransaction.TransactionID(), "autoCommit", repositoryTransaction.IsAutoCommit(), "readOnly", repositoryTransaction.IsReadOnly())

	if repositoryTransaction.state == nil {
		repositoryTransaction.repositoryProvider.telemetryService.Slogger.Error("can't rollback because transaction not active", "transactionID", repositoryTransaction.TransactionID(), "autoCommit", repositoryTransaction.IsAutoCommit(), "readOnly", repositoryTransaction.IsReadOnly())
		return fmt.Errorf("can't rollback because transaction not active")
	} else if repositoryTransaction.state.autoCommit {
		repositoryTransaction.repositoryProvider.telemetryService.Slogger.Error("can't rollback because transaction is auto committed", "transactionID", repositoryTransaction.TransactionID(), "autoCommit", repositoryTransaction.IsAutoCommit(), "readOnly", repositoryTransaction.IsReadOnly())
		return fmt.Errorf("can't rollback because transaction is auto committed")
	}

	if err := repositoryTransaction.state.gormTx.Rollback().Error; err != nil {
		repositoryTransaction.repositoryProvider.telemetryService.Slogger.Error("failed to rollback transaction", "transactionID", repositoryTransaction.TransactionID(), "autoCommit", repositoryTransaction.IsAutoCommit(), "readOnly", repositoryTransaction.IsReadOnly(), "error", err)
		return err
	}

	repositoryTransaction.repositoryProvider.telemetryService.Slogger.Info("rolled back transaction", "transactionID", repositoryTransaction.TransactionID(), "autoCommit", repositoryTransaction.IsAutoCommit(), "readOnly", repositoryTransaction.IsReadOnly())
	repositoryTransaction.state = nil
	return nil
}
