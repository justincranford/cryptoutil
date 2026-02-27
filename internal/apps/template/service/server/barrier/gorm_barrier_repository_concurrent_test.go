// Copyright (c) 2025 Justin Cranford
//

package barrier_test

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
)

func TestGormRepository_ConcurrentTransactions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const numGoroutines = 5

	// Create isolated database for this test.
	db, cleanup := createIsolatedDB(t)
	defer cleanup()

	barrierRepo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { barrierRepo.Shutdown() })

	// Launch multiple concurrent transactions.
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			keyUUID, _ := googleUuid.NewV7()
			key := &cryptoutilAppsTemplateServiceServerBarrier.RootKey{
				UUID:      keyUUID,
				Encrypted: "encrypted_root_key_" + string(rune(id)),
				KEKUUID:   googleUuid.UUID{},
			}

			err := barrierRepo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
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

// TestGormRepository_NewWithNilDB tests NewGormRepository with nil db.
func TestGormRepository_NewWithNilDB(t *testing.T) {
	t.Parallel()

	repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(nil)
	require.Error(t, err)
	require.Nil(t, repo)
	require.Contains(t, err.Error(), "db must be non-nil")
}

// TestGormRepository_Shutdown tests Shutdown method.
func TestGormRepository_Shutdown(t *testing.T) {
	t.Parallel()

	// Create isolated database for this test.
	db, cleanup := createIsolatedDB(t)
	defer cleanup()

	barrierRepo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)

	// Shutdown should not panic and can be called multiple times safely.
	barrierRepo.Shutdown()
	barrierRepo.Shutdown() // Should be idempotent.
}

// TestGormTransaction_Context tests that Context returns correct context.
func TestGormTransaction_Context(t *testing.T) {
	t.Parallel()

	// Create isolated database for this test.
	db, cleanup := createIsolatedDB(t)
	defer cleanup()

	barrierRepo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { barrierRepo.Shutdown() })

	// Create a context with a custom value to verify it's passed through.
	type contextKey string

	const testKey contextKey = "test-key"

	ctx := context.WithValue(context.Background(), testKey, "test-value")

	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
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

// TestGormTransaction_AddRootKey_NilKey tests AddRootKey with nil key.
func TestGormTransaction_AddRootKey_NilKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create isolated database for this test.
	db, cleanup := createIsolatedDB(t)
	defer cleanup()

	barrierRepo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { barrierRepo.Shutdown() })

	// Test: AddRootKey with nil key should return error.
	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		return tx.AddRootKey(nil)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "key must be non-nil")
}

// TestGormTransaction_AddIntermediateKey_NilKey tests AddIntermediateKey with nil key.
func TestGormTransaction_AddIntermediateKey_NilKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create isolated database for this test.
	db, cleanup := createIsolatedDB(t)
	defer cleanup()

	barrierRepo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { barrierRepo.Shutdown() })

	// Test: AddIntermediateKey with nil key should return error.
	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		return tx.AddIntermediateKey(nil)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "key must be non-nil")
}

// TestGormTransaction_AddContentKey_NilKey tests AddContentKey with nil key.
func TestGormTransaction_AddContentKey_NilKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create isolated database for this test.
	db, cleanup := createIsolatedDB(t)
	defer cleanup()

	barrierRepo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { barrierRepo.Shutdown() })

	// Test: AddContentKey with nil key should return error.
	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		return tx.AddContentKey(nil)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "key must be non-nil")
}

// TestGormTransaction_GetRootKey_NilUUID tests GetRootKey with nil UUID.
func TestGormTransaction_GetRootKey_NilUUID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create isolated database for this test.
	db, cleanup := createIsolatedDB(t)
	defer cleanup()

	barrierRepo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { barrierRepo.Shutdown() })

	// Test: GetRootKey with nil UUID should return error.
	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, err := tx.GetRootKey(nil)

		return err
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "uuid must be non-nil")
}

// TestGormTransaction_GetIntermediateKey_NilUUID tests GetIntermediateKey with nil UUID.
func TestGormTransaction_GetIntermediateKey_NilUUID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create isolated database for this test.
	db, cleanup := createIsolatedDB(t)
	defer cleanup()

	barrierRepo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { barrierRepo.Shutdown() })

	// Test: GetIntermediateKey with nil UUID should return error.
	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, err := tx.GetIntermediateKey(nil)

		return err
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "uuid must be non-nil")
}

// TestGormTransaction_GetContentKey_NilUUID tests GetContentKey with nil UUID.
func TestGormTransaction_GetContentKey_NilUUID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create isolated database for this test.
	db, cleanup := createIsolatedDB(t)
	defer cleanup()

	barrierRepo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { barrierRepo.Shutdown() })

	// Test: GetContentKey with nil UUID should return error.
	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, err := tx.GetContentKey(nil)

		return err
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "uuid must be non-nil")
}

// TestGormTransaction_GetRootKey_NotFound tests GetRootKey with non-existent UUID.
