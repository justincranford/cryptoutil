//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford

package orm

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Test GetRootKeyLatest error path when no records exist.
func TestGetRootKeyLatest_NoRecordsError(t *testing.T) {
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		// Attempt to get latest root key when none exist.
		rootKey, err := tx.GetRootKeyLatest()
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to load latest root key")
		require.Nil(t, rootKey)
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)

		return nil
	})
	require.NoError(t, err)
}

// Test GetRootKey error path when key doesn't exist.
func TestGetRootKey_NotFoundError(t *testing.T) {
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	nonExistentUUID := googleUuid.New()

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		// Attempt to get non-existent root key.
		rootKey, err := tx.GetRootKey(&nonExistentUUID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to load key key with UUID")
		require.Nil(t, rootKey)
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)

		return nil
	})
	require.NoError(t, err)
}

// Test DeleteRootKey error path when key doesn't exist.
func TestDeleteRootKey_NotFoundError(t *testing.T) {
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	nonExistentUUID := googleUuid.New()

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Attempt to delete non-existent root key (GORM Delete doesn't error on missing record).
		rootKey, err := tx.DeleteRootKey(&nonExistentUUID)
		require.NoError(t, err) // Delete succeeds even if record doesn't exist.
		require.NotNil(t, rootKey)

		return nil
	})
	require.NoError(t, err)
}

// Test GetIntermediateKeyLatest error path when no records exist.
func TestGetIntermediateKeyLatest_NoRecordsError(t *testing.T) {
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		// Attempt to get latest intermediate key when none exist.
		intermediateKey, err := tx.GetIntermediateKeyLatest()
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to load latest intermediate key")
		require.Nil(t, intermediateKey)
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)

		return nil
	})
	require.NoError(t, err)
}

// Test GetIntermediateKey error path when key doesn't exist.
func TestGetIntermediateKey_NotFoundError(t *testing.T) {
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	nonExistentUUID := googleUuid.New()

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		// Attempt to get non-existent intermediate key.
		intermediateKey, err := tx.GetIntermediateKey(&nonExistentUUID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to load key key with UUID") // Note: Source has typo "key key".
		require.Nil(t, intermediateKey)
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)

		return nil
	})
	require.NoError(t, err)
}

// Test DeleteIntermediateKey error path when key doesn't exist.
func TestDeleteIntermediateKey_NotFoundError(t *testing.T) {
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	nonExistentUUID := googleUuid.New()

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Attempt to delete non-existent intermediate key (GORM Delete doesn't error on missing record).
		intermediateKey, err := tx.DeleteIntermediateKey(&nonExistentUUID)
		require.NoError(t, err) // Delete succeeds even if record doesn't exist.
		require.NotNil(t, intermediateKey)

		return nil
	})
	require.NoError(t, err)
}

// Test GetContentKeyLatest error path when no records exist.
func TestGetContentKeyLatest_NoRecordsError(t *testing.T) {
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		// Attempt to get latest content key when none exist.
		contentKey, err := tx.GetContentKeyLatest()
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to load latest content key")
		require.Nil(t, contentKey)
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)

		return nil
	})
	require.NoError(t, err)
}

// Test GetContentKey error path when key doesn't exist.
func TestGetContentKey_NotFoundError(t *testing.T) {
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	nonExistentUUID := googleUuid.New()

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		// Attempt to get non-existent content key.
		contentKey, err := tx.GetContentKey(&nonExistentUUID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to load key key with UUID") // Note: Source has typo "key key".
		require.Nil(t, contentKey)
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)

		return nil
	})
	require.NoError(t, err)
}

// Test DeleteContentKey error path when key doesn't exist.
func TestDeleteContentKey_NotFoundError(t *testing.T) {
	t.Cleanup(func() { CleanupDatabase(t, testOrmRepository) })

	nonExistentUUID := googleUuid.New()

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Attempt to delete non-existent content key (GORM Delete doesn't error on missing record).
		contentKey, err := tx.DeleteContentKey(&nonExistentUUID)
		require.NoError(t, err) // Delete succeeds even if record doesn't exist.
		require.NotNil(t, contentKey)

		return nil
	})
	require.NoError(t, err)
}
