// Copyright (c) 2025 Justin Cranford
//

package barrier_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	cryptoutilTemplateBarrier "cryptoutil/internal/apps/template/service/server/barrier"
)

// createIsolatedDB creates an isolated in-memory SQLite database for repository tests.
func createIsolatedDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()

	dbUUID, err := googleUuid.NewV7()
	require.NoError(t, err)

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbUUID.String())
	sqlDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)

	ctx := context.Background()

	// Configure SQLite for concurrent operations.
	_, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)
	_, err = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(0) // In-memory: never close connections.

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	// Create barrier tables.
	err = createBarrierTables(sqlDB)
	require.NoError(t, err)

	cleanup := func() {
		if closeErr := sqlDB.Close(); closeErr != nil {
			panic("failed to close SQL DB: " + closeErr.Error())
		}
	}

	return db, cleanup
}

// TestGormBarrierRepository_RootKey_Lifecycle tests complete root key lifecycle.
func TestGormBarrierRepository_RootKey_Lifecycle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create isolated database for this test.
	db, cleanup := createIsolatedDB(t)
	defer cleanup()

	barrierRepo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { barrierRepo.Shutdown() })

	// Test: GetRootKeyLatest should return ErrNoRootKeyFound when no keys exist.
	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		latest, err := tx.GetRootKeyLatest()
		require.ErrorIs(t, err, cryptoutilTemplateBarrier.ErrNoRootKeyFound, "Should get ErrNoRootKeyFound when no root keys exist")
		require.Nil(t, latest, "Latest should be nil when error occurs")

		return nil
	})
	require.NoError(t, err)

	// Create first root key.
	key1UUID, _ := googleUuid.NewV7()
	key1 := &cryptoutilTemplateBarrier.BarrierRootKey{
		UUID:      key1UUID,
		Encrypted: "encrypted_root_key_1",
		KEKUUID:   googleUuid.UUID{},
	}

	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		return tx.AddRootKey(key1)
	})
	require.NoError(t, err)

	// Test: GetRootKeyLatest should return the first key.
	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		latest, err := tx.GetRootKeyLatest()
		require.NoError(t, err)
		require.NotNil(t, latest)
		require.Equal(t, key1.UUID, latest.UUID)
		require.Equal(t, key1.Encrypted, latest.Encrypted)

		return nil
	})
	require.NoError(t, err)

	// Test: GetRootKey by UUID should return the specific key.
	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		retrieved, err := tx.GetRootKey(&key1UUID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		require.Equal(t, key1.UUID, retrieved.UUID)
		require.Equal(t, key1.Encrypted, retrieved.Encrypted)

		return nil
	})
	require.NoError(t, err)

	// Create second root key (newer).
	key2UUID, _ := googleUuid.NewV7()
	key2 := &cryptoutilTemplateBarrier.BarrierRootKey{
		UUID:      key2UUID,
		Encrypted: "encrypted_root_key_2",
		KEKUUID:   googleUuid.UUID{},
	}

	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		return tx.AddRootKey(key2)
	})
	require.NoError(t, err)

	// Test: GetRootKeyLatest should return the second (newer) key.
	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		latest, err := tx.GetRootKeyLatest()
		require.NoError(t, err)
		require.NotNil(t, latest)
		require.Equal(t, key2.UUID, latest.UUID, "Latest key should be the most recently created")

		return nil
	})
	require.NoError(t, err)

	// Test: Both keys should still be retrievable by UUID.
	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		retrieved1, err := tx.GetRootKey(&key1UUID)
		require.NoError(t, err)
		require.NotNil(t, retrieved1)
		require.Equal(t, key1.UUID, retrieved1.UUID)

		retrieved2, err := tx.GetRootKey(&key2UUID)
		require.NoError(t, err)
		require.NotNil(t, retrieved2)
		require.Equal(t, key2.UUID, retrieved2.UUID)

		return nil
	})
	require.NoError(t, err)
}

// TestGormBarrierRepository_IntermediateKey_Lifecycle tests complete intermediate key lifecycle.
func TestGormBarrierRepository_IntermediateKey_Lifecycle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create isolated database for this test.
	db, cleanup := createIsolatedDB(t)
	defer cleanup()

	barrierRepo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { barrierRepo.Shutdown() })

	// Create parent root key first.
	rootKeyUUID, _ := googleUuid.NewV7()
	rootKey := &cryptoutilTemplateBarrier.BarrierRootKey{
		UUID:      rootKeyUUID,
		Encrypted: "encrypted_root_key_1",
		KEKUUID:   googleUuid.UUID{},
	}

	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		return tx.AddRootKey(rootKey)
	})
	require.NoError(t, err)

	// Test: GetIntermediateKeyLatest should return ErrNoIntermediateKeyFound when no intermediate keys exist.
	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		latest, err := tx.GetIntermediateKeyLatest()
		require.ErrorIs(t, err, cryptoutilTemplateBarrier.ErrNoIntermediateKeyFound, "Should get ErrNoIntermediateKeyFound when no intermediate keys exist")
		require.Nil(t, latest, "Latest should be nil when error occurs")

		return nil
	})
	require.NoError(t, err)

	// Create first intermediate key.
	key1UUID, _ := googleUuid.NewV7()
	key1 := &cryptoutilTemplateBarrier.BarrierIntermediateKey{
		UUID:      key1UUID,
		Encrypted: "encrypted_intermediate_key_1",
		KEKUUID:   rootKeyUUID,
	}

	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		return tx.AddIntermediateKey(key1)
	})
	require.NoError(t, err)

	// Test: GetIntermediateKeyLatest should return the first key.
	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		latest, err := tx.GetIntermediateKeyLatest()
		require.NoError(t, err)
		require.NotNil(t, latest)
		require.Equal(t, key1.UUID, latest.UUID)
		require.Equal(t, key1.Encrypted, latest.Encrypted)
		require.Equal(t, rootKeyUUID, latest.KEKUUID)

		return nil
	})
	require.NoError(t, err)

	// Test: GetIntermediateKey by UUID should return the specific key.
	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		retrieved, err := tx.GetIntermediateKey(&key1UUID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		require.Equal(t, key1.UUID, retrieved.UUID)
		require.Equal(t, key1.Encrypted, retrieved.Encrypted)

		return nil
	})
	require.NoError(t, err)

	// Create second intermediate key (newer).
	key2UUID, _ := googleUuid.NewV7()
	key2 := &cryptoutilTemplateBarrier.BarrierIntermediateKey{
		UUID:      key2UUID,
		Encrypted: "encrypted_intermediate_key_2",
		KEKUUID:   rootKeyUUID,
	}

	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		return tx.AddIntermediateKey(key2)
	})
	require.NoError(t, err)

	// Test: GetIntermediateKeyLatest should return the second (newer) key.
	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		latest, err := tx.GetIntermediateKeyLatest()
		require.NoError(t, err)
		require.NotNil(t, latest)
		require.Equal(t, key2.UUID, latest.UUID, "Latest key should be the most recently created")

		return nil
	})
	require.NoError(t, err)
}

// TestGormBarrierRepository_ContentKey_Lifecycle tests complete content key lifecycle.
func TestGormBarrierRepository_ContentKey_Lifecycle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create isolated database for this test.
	db, cleanup := createIsolatedDB(t)
	defer cleanup()

	barrierRepo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { barrierRepo.Shutdown() })
	// Create parent root key.
	rootKeyUUID, _ := googleUuid.NewV7()
	rootKey := &cryptoutilTemplateBarrier.BarrierRootKey{
		UUID:      rootKeyUUID,
		Encrypted: "encrypted_root_key",
		KEKUUID:   googleUuid.UUID{},
	}

	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		return tx.AddRootKey(rootKey)
	})
	require.NoError(t, err)

	// Create parent intermediate key.
	intermediateKeyUUID, _ := googleUuid.NewV7()
	intermediateKey := &cryptoutilTemplateBarrier.BarrierIntermediateKey{
		UUID:      intermediateKeyUUID,
		Encrypted: "encrypted_intermediate_key",
		KEKUUID:   rootKeyUUID,
	}

	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		return tx.AddIntermediateKey(intermediateKey)
	})
	require.NoError(t, err)

	// Create first content key.
	key1UUID, _ := googleUuid.NewV7()
	key1 := &cryptoutilTemplateBarrier.BarrierContentKey{
		UUID:      key1UUID,
		Encrypted: "encrypted_content_key_1",
		KEKUUID:   intermediateKeyUUID,
	}

	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		return tx.AddContentKey(key1)
	})
	require.NoError(t, err)

	// Test: GetContentKey by UUID should return the specific key.
	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		retrieved, err := tx.GetContentKey(&key1UUID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		require.Equal(t, key1.UUID, retrieved.UUID)
		require.Equal(t, key1.Encrypted, retrieved.Encrypted)
		require.Equal(t, intermediateKeyUUID, retrieved.KEKUUID)

		return nil
	})
	require.NoError(t, err)

	// Create second content key.
	key2UUID, _ := googleUuid.NewV7()
	key2 := &cryptoutilTemplateBarrier.BarrierContentKey{
		UUID:      key2UUID,
		Encrypted: "encrypted_content_key_2",
		KEKUUID:   intermediateKeyUUID,
	}

	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		return tx.AddContentKey(key2)
	})
	require.NoError(t, err)

	// Test: GetContentKey by UUID should return the second key.
	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		retrieved, err := tx.GetContentKey(&key2UUID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		require.Equal(t, key2.UUID, retrieved.UUID)
		require.Equal(t, key2.Encrypted, retrieved.Encrypted)

		return nil
	})
	require.NoError(t, err)
}

// TestGormBarrierRepository_Transaction_Rollback tests transaction rollback behavior.
func TestGormBarrierRepository_Transaction_Rollback(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create isolated database for this test.
	db, cleanup := createIsolatedDB(t)
	defer cleanup()

	barrierRepo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { barrierRepo.Shutdown() })

	// Create a root key inside a transaction that will be rolled back.
	keyUUID, _ := googleUuid.NewV7()
	key := &cryptoutilTemplateBarrier.BarrierRootKey{
		UUID:      keyUUID,
		Encrypted: "encrypted_root_key",
		KEKUUID:   googleUuid.UUID{},
	}

	// Transaction that returns an error (should rollback).
	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		if addErr := tx.AddRootKey(key); addErr != nil {
			return fmt.Errorf("failed to add root key: %w", addErr)
		}
		// Force rollback by returning error.
		return context.Canceled
	})
	require.Error(t, err, "Transaction should fail")

	// Verify key was NOT persisted (transaction rolled back).
	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		latest, err := tx.GetRootKeyLatest()
		require.ErrorIs(t, err, cryptoutilTemplateBarrier.ErrNoRootKeyFound, "Should return ErrNoRootKeyFound when no keys exist")
		require.Nil(t, latest, "Key should not exist after rollback")

		return nil
	})
	require.NoError(t, err)
}

// TestGormBarrierRepository_ConcurrentTransactions tests concurrent transaction safety.
func TestGormBarrierRepository_ConcurrentTransactions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const numGoroutines = 5

	// Create isolated database for this test.
	db, cleanup := createIsolatedDB(t)
	defer cleanup()

	barrierRepo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { barrierRepo.Shutdown() })

	// Launch multiple concurrent transactions.
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			keyUUID, _ := googleUuid.NewV7()
			key := &cryptoutilTemplateBarrier.BarrierRootKey{
				UUID:      keyUUID,
				Encrypted: "encrypted_root_key_" + string(rune(id)),
				KEKUUID:   googleUuid.UUID{},
			}

			err := barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
				return tx.AddRootKey(key)
			})
			errors <- err
		}(i)
	}

	// Collect results.
	for i := 0; i < numGoroutines; i++ {
		err := <-errors
		require.NoError(t, err, "Concurrent transactions should succeed")
	}
}

// TestGormBarrierRepository_NewWithNilDB tests NewGormBarrierRepository with nil db.
func TestGormBarrierRepository_NewWithNilDB(t *testing.T) {
	t.Parallel()

	repo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(nil)
	require.Error(t, err)
	require.Nil(t, repo)
	require.Contains(t, err.Error(), "db must be non-nil")
}

// TestGormBarrierRepository_Shutdown tests Shutdown method.
func TestGormBarrierRepository_Shutdown(t *testing.T) {
	t.Parallel()

	// Create isolated database for this test.
	db, cleanup := createIsolatedDB(t)
	defer cleanup()

	barrierRepo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(db)
	require.NoError(t, err)

	// Shutdown should not panic and can be called multiple times safely.
	barrierRepo.Shutdown()
	barrierRepo.Shutdown() // Should be idempotent.
}

// TestGormBarrierTransaction_Context tests that Context returns correct context.
func TestGormBarrierTransaction_Context(t *testing.T) {
	t.Parallel()

	// Create isolated database for this test.
	db, cleanup := createIsolatedDB(t)
	defer cleanup()

	barrierRepo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { barrierRepo.Shutdown() })

	// Create a context with a custom value to verify it's passed through.
	type contextKey string

	const testKey contextKey = "test-key"

	ctx := context.WithValue(context.Background(), testKey, "test-value")

	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		// Get the context from transaction.
		txCtx := tx.Context()
		require.NotNil(t, txCtx)

		// Verify the context value is preserved.
		value := txCtx.Value(testKey)
		require.Equal(t, "test-value", value)

		return nil
	})
	require.NoError(t, err)
}
