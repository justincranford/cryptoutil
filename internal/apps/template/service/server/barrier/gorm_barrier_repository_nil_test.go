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

func TestGormTransaction_GetRootKey_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create isolated database for this test.
	db, cleanup := createIsolatedDB(t)
	defer cleanup()

	barrierRepo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { barrierRepo.Shutdown() })

	// Test: GetRootKey with non-existent UUID should return error.
	nonExistentUUID, _ := googleUuid.NewV7()
	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, err := tx.GetRootKey(&nonExistentUUID)

		return err
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get root key")
}

// TestGormTransaction_GetIntermediateKey_NotFound tests GetIntermediateKey with non-existent UUID.
func TestGormTransaction_GetIntermediateKey_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create isolated database for this test.
	db, cleanup := createIsolatedDB(t)
	defer cleanup()

	barrierRepo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { barrierRepo.Shutdown() })

	// Test: GetIntermediateKey with non-existent UUID should return error.
	nonExistentUUID, _ := googleUuid.NewV7()
	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, err := tx.GetIntermediateKey(&nonExistentUUID)

		return err
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get intermediate key")
}

// TestGormTransaction_GetContentKey_NotFound tests GetContentKey with non-existent UUID.
func TestGormTransaction_GetContentKey_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create isolated database for this test.
	db, cleanup := createIsolatedDB(t)
	defer cleanup()

	barrierRepo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { barrierRepo.Shutdown() })

	// Test: GetContentKey with non-existent UUID should return error.
	nonExistentUUID, _ := googleUuid.NewV7()
	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, err := tx.GetContentKey(&nonExistentUUID)

		return err
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get content key")
}

// TestGormTransaction_AddRootKey_DuplicateKey tests AddRootKey with duplicate UUID (UNIQUE constraint violation).
func TestGormTransaction_AddRootKey_DuplicateKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create isolated database for this test.
	db, cleanup := createIsolatedDB(t)
	defer cleanup()

	barrierRepo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { barrierRepo.Shutdown() })

	// Create first root key with a specific UUID.
	keyUUID, _ := googleUuid.NewV7()
	key1 := &cryptoutilAppsTemplateServiceServerBarrier.RootKey{
		UUID:      keyUUID,
		Encrypted: "encrypted_root_key_1",
		KEKUUID:   googleUuid.UUID{},
	}

	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		return tx.AddRootKey(key1)
	})
	require.NoError(t, err)

	// Try to add another root key with the same UUID - should fail with UNIQUE constraint violation.
	key2 := &cryptoutilAppsTemplateServiceServerBarrier.RootKey{
		UUID:      keyUUID, // Same UUID as key1
		Encrypted: "encrypted_root_key_2",
		KEKUUID:   googleUuid.UUID{},
	}

	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		return tx.AddRootKey(key2)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to add root key")
}

// TestGormTransaction_AddIntermediateKey_DuplicateKey tests AddIntermediateKey with duplicate UUID.
func TestGormTransaction_AddIntermediateKey_DuplicateKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create isolated database for this test.
	db, cleanup := createIsolatedDB(t)
	defer cleanup()

	barrierRepo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { barrierRepo.Shutdown() })

	// Create first intermediate key with a specific UUID.
	keyUUID, _ := googleUuid.NewV7()
	kekUUID, _ := googleUuid.NewV7()
	key1 := &cryptoutilAppsTemplateServiceServerBarrier.IntermediateKey{
		UUID:      keyUUID,
		Encrypted: "encrypted_intermediate_key_1",
		KEKUUID:   kekUUID,
	}

	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		return tx.AddIntermediateKey(key1)
	})
	require.NoError(t, err)

	// Try to add another intermediate key with the same UUID - should fail.
	key2 := &cryptoutilAppsTemplateServiceServerBarrier.IntermediateKey{
		UUID:      keyUUID, // Same UUID as key1
		Encrypted: "encrypted_intermediate_key_2",
		KEKUUID:   kekUUID,
	}

	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		return tx.AddIntermediateKey(key2)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to add intermediate key")
}

// TestGormTransaction_AddContentKey_DuplicateKey tests AddContentKey with duplicate UUID.
func TestGormTransaction_AddContentKey_DuplicateKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create isolated database for this test.
	db, cleanup := createIsolatedDB(t)
	defer cleanup()

	barrierRepo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { barrierRepo.Shutdown() })

	// Create first content key with a specific UUID.
	keyUUID, _ := googleUuid.NewV7()
	kekUUID, _ := googleUuid.NewV7()
	key1 := &cryptoutilAppsTemplateServiceServerBarrier.ContentKey{
		UUID:      keyUUID,
		Encrypted: "encrypted_content_key_1",
		KEKUUID:   kekUUID,
	}

	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		return tx.AddContentKey(key1)
	})
	require.NoError(t, err)

	// Try to add another content key with the same UUID - should fail.
	key2 := &cryptoutilAppsTemplateServiceServerBarrier.ContentKey{
		UUID:      keyUUID, // Same UUID as key1
		Encrypted: "encrypted_content_key_2",
		KEKUUID:   kekUUID,
	}

	err = barrierRepo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		return tx.AddContentKey(key2)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to add content key")
}
