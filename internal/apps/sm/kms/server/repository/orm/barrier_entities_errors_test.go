//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford

package orm

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Root Key Error Path Tests

func TestOrmTransaction_GetRootKey_NotFound(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		nonExistentUUID := googleUuid.New()
		_, getErr := tx.GetRootKey(&nonExistentUUID)
		require.Error(t, getErr, "Should fail when root key not found")
		require.Contains(t, getErr.Error(), "failed to load key key with UUID")

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_GetRootKeyLatest_NotFound(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		_, getErr := tx.GetRootKeyLatest()
		require.Error(t, getErr, "Should fail when no root keys exist")
		require.Contains(t, getErr.Error(), "failed to load latest root key")

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_DeleteRootKey_NotFound(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		nonExistentUUID := googleUuid.New()
		deletedKey, deleteErr := tx.DeleteRootKey(&nonExistentUUID)
		// GORM Delete doesn't error when no rows affected - it's a no-op.
		require.NoError(t, deleteErr)
		require.NotNil(t, deletedKey)

		return nil
	})

	require.NoError(t, err)
}

// Intermediate Key Error Path Tests

func TestOrmTransaction_GetIntermediateKey_NotFound(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		nonExistentUUID := googleUuid.New()
		_, getErr := tx.GetIntermediateKey(&nonExistentUUID)
		require.Error(t, getErr, "Should fail when intermediate key not found")
		require.Contains(t, getErr.Error(), "failed to load key key with UUID")

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_GetIntermediateKeyLatest_NotFound(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		_, getErr := tx.GetIntermediateKeyLatest()
		require.Error(t, getErr, "Should fail when no intermediate keys exist")
		require.Contains(t, getErr.Error(), "failed to load latest intermediate key")

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_DeleteIntermediateKey_NotFound(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		nonExistentUUID := googleUuid.New()
		deletedKey, deleteErr := tx.DeleteIntermediateKey(&nonExistentUUID)
		// GORM Delete doesn't error when no rows affected - it's a no-op.
		require.NoError(t, deleteErr)
		require.NotNil(t, deletedKey)

		return nil
	})

	require.NoError(t, err)
}

// Content Key Error Path Tests

func TestOrmTransaction_GetContentKey_NotFound(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		nonExistentUUID := googleUuid.New()
		_, getErr := tx.GetContentKey(&nonExistentUUID)
		require.Error(t, getErr, "Should fail when content key not found")
		require.Contains(t, getErr.Error(), "failed to load key key with UUID")

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_GetContentKeyLatest_NotFound(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		_, getErr := tx.GetContentKeyLatest()
		require.Error(t, getErr, "Should fail when no content keys exist")
		require.Contains(t, getErr.Error(), "failed to load latest content key")

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_DeleteContentKey_NotFound(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		nonExistentUUID := googleUuid.New()
		deletedKey, deleteErr := tx.DeleteContentKey(&nonExistentUUID)
		// GORM Delete doesn't error when no rows affected - it's a no-op.
		require.NoError(t, deleteErr)
		require.NotNil(t, deletedKey)

		return nil
	})

	require.NoError(t, err)
}

// Database Constraint Error Tests

func TestOrmTransaction_AddRootKey_DuplicateUUID(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	rootKeyID := googleUuid.New()
	kekID := googleUuid.New()

	// Create first root key.
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		rootKey := &RootKey{
			UUID:      rootKeyID,
			Encrypted: "encrypted-root-key-1",
			KEKUUID:   kekID,
		}

		return tx.AddRootKey(rootKey)
	})
	require.NoError(t, err)

	// Try to create duplicate root key with same UUID.
	err = testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		duplicateKey := &RootKey{
			UUID:      rootKeyID, // Same UUID - violates PRIMARY KEY constraint.
			Encrypted: "encrypted-root-key-duplicate",
			KEKUUID:   kekID,
		}
		createErr := tx.AddRootKey(duplicateKey)
		require.Error(t, createErr, "Should fail with duplicate UUID")
		require.Contains(t, createErr.Error(), "failed to add root key")

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_AddIntermediateKey_DuplicateUUID(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	intermediateKeyID := googleUuid.New()
	kekID := googleUuid.New()

	// Create first intermediate key.
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		intermediateKey := &IntermediateKey{
			UUID:      intermediateKeyID,
			Encrypted: "encrypted-intermediate-key-1",
			KEKUUID:   kekID,
		}

		return tx.AddIntermediateKey(intermediateKey)
	})
	require.NoError(t, err)

	// Try to create duplicate intermediate key with same UUID.
	err = testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		duplicateKey := &IntermediateKey{
			UUID:      intermediateKeyID, // Same UUID - violates PRIMARY KEY constraint.
			Encrypted: "encrypted-intermediate-key-duplicate",
			KEKUUID:   kekID,
		}
		createErr := tx.AddIntermediateKey(duplicateKey)
		require.Error(t, createErr, "Should fail with duplicate UUID")
		require.Contains(t, createErr.Error(), "failed to add intermediate key")

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_AddContentKey_DuplicateUUID(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	contentKeyID := googleUuid.New()
	kekID := googleUuid.New()

	// Create first content key.
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		contentKey := &ContentKey{
			UUID:      contentKeyID,
			Encrypted: "encrypted-content-key-1",
			KEKUUID:   kekID,
		}

		return tx.AddContentKey(contentKey)
	})
	require.NoError(t, err)

	// Try to create duplicate content key with same UUID.
	err = testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		duplicateKey := &ContentKey{
			UUID:      contentKeyID, // Same UUID - violates PRIMARY KEY constraint.
			Encrypted: "encrypted-content-key-duplicate",
			KEKUUID:   kekID,
		}
		createErr := tx.AddContentKey(duplicateKey)
		require.Error(t, createErr, "Should fail with duplicate UUID")
		require.Contains(t, createErr.Error(), "failed to add content key")

		return nil
	})

	require.NoError(t, err)
}
