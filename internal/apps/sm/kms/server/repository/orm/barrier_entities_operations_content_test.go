//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford

package orm

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestOrmTransaction_AddContentKey(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		contentKeyID := googleUuid.New()
		kekID := googleUuid.New()

		contentKey := &ContentKey{
			UUID:      contentKeyID,
			Encrypted: "encrypted-content-key-data",
			KEKUUID:   kekID,
		}

		createErr := tx.AddContentKey(contentKey)
		require.NoError(t, createErr, "Should successfully add content key")

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_GetContentKeys(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	// Create 2 content keys.
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		kekID := googleUuid.New()

		contentKeyID1 := googleUuid.New()
		contentKey1 := &ContentKey{
			UUID:      contentKeyID1,
			Encrypted: "encrypted-content-key-1",
			KEKUUID:   kekID,
		}
		require.NoError(t, tx.AddContentKey(contentKey1))

		contentKeyID2 := googleUuid.New()
		contentKey2 := &ContentKey{
			UUID:      contentKeyID2,
			Encrypted: "encrypted-content-key-2",
			KEKUUID:   kekID,
		}
		require.NoError(t, tx.AddContentKey(contentKey2))

		return nil
	})
	require.NoError(t, err)

	// Get all content keys.
	err = testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		contentKeys, getErr := tx.GetContentKeys()
		require.NoError(t, getErr, "Should successfully get content keys")
		require.Len(t, contentKeys, 2, "Should return 2 content keys")

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_GetContentKeyLatest(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	var contentKeyID1, contentKeyID2 googleUuid.UUID

	// Create 2 content keys.
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		kekID := googleUuid.New()

		contentKeyID1 = googleUuid.New()
		contentKey1 := &ContentKey{
			UUID:      contentKeyID1,
			Encrypted: "encrypted-content-key-1",
			KEKUUID:   kekID,
		}
		require.NoError(t, tx.AddContentKey(contentKey1))

		contentKeyID2 = googleUuid.New()
		contentKey2 := &ContentKey{
			UUID:      contentKeyID2,
			Encrypted: "encrypted-content-key-2",
			KEKUUID:   kekID,
		}
		require.NoError(t, tx.AddContentKey(contentKey2))

		return nil
	})
	require.NoError(t, err)

	// Get latest content key.
	err = testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		latestKey, getErr := tx.GetContentKeyLatest()
		require.NoError(t, getErr, "Should successfully get latest content key")

		// Latest should have highest UUID value (DESC order).
		expectedLatest := contentKeyID1
		if contentKeyID2.String() > contentKeyID1.String() {
			expectedLatest = contentKeyID2
		}

		require.Equal(t, expectedLatest, latestKey.UUID)

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_GetContentKey(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	var contentKeyID googleUuid.UUID

	// Create content key.
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		contentKeyID = googleUuid.New()
		kekID := googleUuid.New()

		contentKey := &ContentKey{
			UUID:      contentKeyID,
			Encrypted: "encrypted-content-key-specific",
			KEKUUID:   kekID,
		}
		require.NoError(t, tx.AddContentKey(contentKey))

		return nil
	})
	require.NoError(t, err)

	// Get specific content key.
	err = testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		contentKey, getErr := tx.GetContentKey(&contentKeyID)
		require.NoError(t, getErr, "Should successfully get content key by UUID")
		require.Equal(t, contentKeyID, contentKey.UUID)
		require.Equal(t, "encrypted-content-key-specific", contentKey.Encrypted)

		return nil
	})

	require.NoError(t, err)
}

func TestOrmTransaction_DeleteContentKey(t *testing.T) {
	t.Parallel()
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	var contentKeyID googleUuid.UUID

	// Create content key.
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		contentKeyID = googleUuid.New()
		kekID := googleUuid.New()

		contentKey := &ContentKey{
			UUID:      contentKeyID,
			Encrypted: "encrypted-content-key-to-delete",
			KEKUUID:   kekID,
		}
		require.NoError(t, tx.AddContentKey(contentKey))

		return nil
	})
	require.NoError(t, err)

	// Delete content key.
	err = testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		deletedKey, deleteErr := tx.DeleteContentKey(&contentKeyID)
		require.NoError(t, deleteErr, "Should successfully delete content key")
		require.NotNil(t, deletedKey)

		return nil
	})
	require.NoError(t, err)

	// Verify deletion.
	err = testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		contentKeys, getErr := tx.GetContentKeys()
		require.NoError(t, getErr)
		require.Empty(t, contentKeys, "Content key should be deleted")

		return nil
	})

	require.NoError(t, err)
}
