//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford

package orm

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Root Key Tests

func TestOrmTransaction_AddRootKey(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		rootKeyID := googleUuid.New()
		kekID := googleUuid.New()

		rootKey := &RootKey{
			UUID:      rootKeyID,
			Encrypted: "encrypted-root-key-data",
			KEKUUID:   kekID,
		}

		createErr := tx.AddRootKey(rootKey)
		require.NoError(t, createErr, "Should successfully add root key")

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_GetRootKeys(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	var rootKeyID1, rootKeyID2 googleUuid.UUID

	// Create 2 root keys.
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		kekID := googleUuid.New()

		rootKeyID1 = googleUuid.New()
		rootKey1 := &RootKey{
			UUID:      rootKeyID1,
			Encrypted: "encrypted-root-key-1",
			KEKUUID:   kekID,
		}
		require.NoError(t, tx.AddRootKey(rootKey1))

		rootKeyID2 = googleUuid.New()
		rootKey2 := &RootKey{
			UUID:      rootKeyID2,
			Encrypted: "encrypted-root-key-2",
			KEKUUID:   kekID,
		}
		require.NoError(t, tx.AddRootKey(rootKey2))

		return nil
	})
	require.NoError(t, err)

	// Get all root keys.
	err = testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		rootKeys, getErr := tx.GetRootKeys()
		require.NoError(t, getErr, "Should successfully get root keys")
		require.Len(t, rootKeys, 2, "Should return 2 root keys")

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_GetRootKeyLatest(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	var rootKeyID1, rootKeyID2 googleUuid.UUID

	// Create 2 root keys.
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		kekID := googleUuid.New()

		rootKeyID1 = googleUuid.New()
		rootKey1 := &RootKey{
			UUID:      rootKeyID1,
			Encrypted: "encrypted-root-key-1",
			KEKUUID:   kekID,
		}
		require.NoError(t, tx.AddRootKey(rootKey1))

		rootKeyID2 = googleUuid.New()
		rootKey2 := &RootKey{
			UUID:      rootKeyID2,
			Encrypted: "encrypted-root-key-2",
			KEKUUID:   kekID,
		}
		require.NoError(t, tx.AddRootKey(rootKey2))

		return nil
	})
	require.NoError(t, err)

	// Get latest root key.
	err = testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		latestKey, getErr := tx.GetRootKeyLatest()
		require.NoError(t, getErr, "Should successfully get latest root key")

		// Latest should have highest UUID value (DESC order).
		expectedLatest := rootKeyID1
		if rootKeyID2.String() > rootKeyID1.String() {
			expectedLatest = rootKeyID2
		}

		require.Equal(t, expectedLatest, latestKey.UUID)

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_GetRootKey(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	var rootKeyID googleUuid.UUID

	// Create root key.
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		rootKeyID = googleUuid.New()
		kekID := googleUuid.New()

		rootKey := &RootKey{
			UUID:      rootKeyID,
			Encrypted: "encrypted-root-key-specific",
			KEKUUID:   kekID,
		}
		require.NoError(t, tx.AddRootKey(rootKey))

		return nil
	})
	require.NoError(t, err)

	// Get specific root key.
	err = testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		rootKey, getErr := tx.GetRootKey(&rootKeyID)
		require.NoError(t, getErr, "Should successfully get root key by UUID")
		require.Equal(t, rootKeyID, rootKey.UUID)
		require.Equal(t, "encrypted-root-key-specific", rootKey.Encrypted)

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_DeleteRootKey(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	var rootKeyID googleUuid.UUID

	// Create root key.
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		rootKeyID = googleUuid.New()
		kekID := googleUuid.New()

		rootKey := &RootKey{
			UUID:      rootKeyID,
			Encrypted: "encrypted-root-key-to-delete",
			KEKUUID:   kekID,
		}
		require.NoError(t, tx.AddRootKey(rootKey))

		return nil
	})
	require.NoError(t, err)

	// Delete root key.
	err = testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		deletedKey, deleteErr := tx.DeleteRootKey(&rootKeyID)
		require.NoError(t, deleteErr, "Should successfully delete root key")
		require.NotNil(t, deletedKey)

		return nil
	})
	require.NoError(t, err)

	// Verify deletion.
	err = testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		rootKeys, getErr := tx.GetRootKeys()
		require.NoError(t, getErr)
		require.Empty(t, rootKeys, "Root key should be deleted")

		return nil
	})

	require.NoError(t, err)
}

// Intermediate Key Tests

func TestOrmTransaction_AddIntermediateKey(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		intermediateKeyID := googleUuid.New()
		kekID := googleUuid.New()

		intermediateKey := &IntermediateKey{
			UUID:      intermediateKeyID,
			Encrypted: "encrypted-intermediate-key-data",
			KEKUUID:   kekID,
		}

		createErr := tx.AddIntermediateKey(intermediateKey)
		require.NoError(t, createErr, "Should successfully add intermediate key")

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_GetIntermediateKeys(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	// Create 2 intermediate keys.
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		kekID := googleUuid.New()

		intermediateKeyID1 := googleUuid.New()
		intermediateKey1 := &IntermediateKey{
			UUID:      intermediateKeyID1,
			Encrypted: "encrypted-intermediate-key-1",
			KEKUUID:   kekID,
		}
		require.NoError(t, tx.AddIntermediateKey(intermediateKey1))

		intermediateKeyID2 := googleUuid.New()
		intermediateKey2 := &IntermediateKey{
			UUID:      intermediateKeyID2,
			Encrypted: "encrypted-intermediate-key-2",
			KEKUUID:   kekID,
		}
		require.NoError(t, tx.AddIntermediateKey(intermediateKey2))

		return nil
	})
	require.NoError(t, err)

	// Get all intermediate keys.
	err = testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		intermediateKeys, getErr := tx.GetIntermediateKeys()
		require.NoError(t, getErr, "Should successfully get intermediate keys")
		require.Len(t, intermediateKeys, 2, "Should return 2 intermediate keys")

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_GetIntermediateKeyLatest(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	var intermediateKeyID1, intermediateKeyID2 googleUuid.UUID

	// Create 2 intermediate keys.
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		kekID := googleUuid.New()

		intermediateKeyID1 = googleUuid.New()
		intermediateKey1 := &IntermediateKey{
			UUID:      intermediateKeyID1,
			Encrypted: "encrypted-intermediate-key-1",
			KEKUUID:   kekID,
		}
		require.NoError(t, tx.AddIntermediateKey(intermediateKey1))

		intermediateKeyID2 = googleUuid.New()
		intermediateKey2 := &IntermediateKey{
			UUID:      intermediateKeyID2,
			Encrypted: "encrypted-intermediate-key-2",
			KEKUUID:   kekID,
		}
		require.NoError(t, tx.AddIntermediateKey(intermediateKey2))

		return nil
	})
	require.NoError(t, err)

	// Get latest intermediate key.
	err = testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		latestKey, getErr := tx.GetIntermediateKeyLatest()
		require.NoError(t, getErr, "Should successfully get latest intermediate key")

		// Latest should have highest UUID value (DESC order).
		expectedLatest := intermediateKeyID1
		if intermediateKeyID2.String() > intermediateKeyID1.String() {
			expectedLatest = intermediateKeyID2
		}

		require.Equal(t, expectedLatest, latestKey.UUID)

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_GetIntermediateKey(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	var intermediateKeyID googleUuid.UUID

	// Create intermediate key.
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		intermediateKeyID = googleUuid.New()
		kekID := googleUuid.New()

		intermediateKey := &IntermediateKey{
			UUID:      intermediateKeyID,
			Encrypted: "encrypted-intermediate-key-specific",
			KEKUUID:   kekID,
		}
		require.NoError(t, tx.AddIntermediateKey(intermediateKey))

		return nil
	})
	require.NoError(t, err)

	// Get specific intermediate key.
	err = testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		intermediateKey, getErr := tx.GetIntermediateKey(&intermediateKeyID)
		require.NoError(t, getErr, "Should successfully get intermediate key by UUID")
		require.Equal(t, intermediateKeyID, intermediateKey.UUID)
		require.Equal(t, "encrypted-intermediate-key-specific", intermediateKey.Encrypted)

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_DeleteIntermediateKey(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	var intermediateKeyID googleUuid.UUID

	// Create intermediate key.
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		intermediateKeyID = googleUuid.New()
		kekID := googleUuid.New()

		intermediateKey := &IntermediateKey{
			UUID:      intermediateKeyID,
			Encrypted: "encrypted-intermediate-key-to-delete",
			KEKUUID:   kekID,
		}
		require.NoError(t, tx.AddIntermediateKey(intermediateKey))

		return nil
	})
	require.NoError(t, err)

	// Delete intermediate key.
	err = testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		deletedKey, deleteErr := tx.DeleteIntermediateKey(&intermediateKeyID)
		require.NoError(t, deleteErr, "Should successfully delete intermediate key")
		require.NotNil(t, deletedKey)

		return nil
	})
	require.NoError(t, err)

	// Verify deletion.
	err = testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		intermediateKeys, getErr := tx.GetIntermediateKeys()
		require.NoError(t, getErr)
		require.Empty(t, intermediateKeys, "Intermediate key should be deleted")

		return nil
	})

	require.NoError(t, err)
}

// Content Key Tests
