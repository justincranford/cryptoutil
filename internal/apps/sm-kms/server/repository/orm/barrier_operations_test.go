//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford

package orm

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestRootKeyOperations tests root key CRUD operations.
func TestRootKeyOperations(t *testing.T) {
	t.Parallel()
	t.Run("Add and retrieve multiple root keys", func(t *testing.T) {
		CleanupDatabase(t, testOrmRepository)

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add 3 root keys.
			const numKeys = 3

			addedKeys := make([]*RootKey, numKeys)

			for i := 0; i < numKeys; i++ {
				key := &RootKey{}
				key.SetUUID(googleUuid.New())
				key.SetEncrypted("encrypted-root-key-data")
				key.SetKEKUUID(googleUuid.New())

				err := tx.AddRootKey(key)
				require.NoError(t, err)

				addedKeys[i] = key
			}

			// Retrieve all keys.
			keys, err := tx.GetRootKeys()
			require.NoError(t, err)
			require.Len(t, keys, numKeys)

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("Get latest root key", func(t *testing.T) {
		CleanupDatabase(t, testOrmRepository)

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add 5 root keys.
			const numKeys = 5

			addedKeys := make([]*RootKey, numKeys)

			for i := 0; i < numKeys; i++ {
				key := &RootKey{}
				key.SetUUID(googleUuid.New())
				key.SetEncrypted("encrypted-root-key-data")
				key.SetKEKUUID(googleUuid.New())

				err := tx.AddRootKey(key)
				require.NoError(t, err)

				addedKeys[i] = key
			}

			// Get latest key (latest by UUID DESC ordering, not insertion order).
			latestKey, err := tx.GetRootKeyLatest()
			require.NoError(t, err)
			require.NotNil(t, latestKey)

			// Verify it's one of the keys we added.
			found := false

			for _, k := range addedKeys {
				if k.GetUUID() == latestKey.GetUUID() {
					found = true

					break
				}
			}

			require.True(t, found, "Latest key should be one of the added keys")

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("Get root key by UUID", func(t *testing.T) {
		CleanupDatabase(t, testOrmRepository)

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add key.
			key := &RootKey{}
			key.SetUUID(googleUuid.New())
			key.SetEncrypted("encrypted-root-key-data")
			key.SetKEKUUID(googleUuid.New())

			err := tx.AddRootKey(key)
			require.NoError(t, err)

			// Retrieve by UUID.
			targetUUID := key.GetUUID()
			retrievedKey, err := tx.GetRootKey(&targetUUID)
			require.NoError(t, err)
			require.NotNil(t, retrievedKey)
			require.Equal(t, targetUUID, retrievedKey.GetUUID())

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("Delete root key", func(t *testing.T) {
		CleanupDatabase(t, testOrmRepository)

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add 3 keys.
			const numKeys = 3

			addedKeys := make([]*RootKey, numKeys)

			for i := 0; i < numKeys; i++ {
				key := &RootKey{}
				key.SetUUID(googleUuid.New())
				key.SetEncrypted("encrypted-root-key-data")
				key.SetKEKUUID(googleUuid.New())

				err := tx.AddRootKey(key)
				require.NoError(t, err)

				addedKeys[i] = key
			}

			// Delete one key.
			targetUUID := addedKeys[1].GetUUID()
			deletedKey, err := tx.DeleteRootKey(&targetUUID)
			require.NoError(t, err)
			require.NotNil(t, deletedKey)

			// Verify deletion.
			remainingKeys, err := tx.GetRootKeys()
			require.NoError(t, err)
			require.Len(t, remainingKeys, numKeys-1)

			return nil
		})

		require.NoError(t, err)
	})
}

// TestIntermediateKeyOperations tests intermediate key CRUD operations.
func TestIntermediateKeyOperations(t *testing.T) {
	t.Parallel()
	t.Run("Add and retrieve multiple intermediate keys", func(t *testing.T) {
		CleanupDatabase(t, testOrmRepository)

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add 4 intermediate keys.
			const numKeys = 4

			addedKeys := make([]*IntermediateKey, numKeys)

			for i := 0; i < numKeys; i++ {
				key := &IntermediateKey{}
				key.SetUUID(googleUuid.New())
				key.SetEncrypted("encrypted-intermediate-key-data")
				key.SetKEKUUID(googleUuid.New())

				err := tx.AddIntermediateKey(key)
				require.NoError(t, err)

				addedKeys[i] = key
			}

			// Retrieve all keys.
			keys, err := tx.GetIntermediateKeys()
			require.NoError(t, err)
			require.Len(t, keys, numKeys)

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("Get latest intermediate key", func(t *testing.T) {
		CleanupDatabase(t, testOrmRepository)

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add 6 intermediate keys.
			const numKeys = 6

			addedKeys := make([]*IntermediateKey, numKeys)

			for i := 0; i < numKeys; i++ {
				key := &IntermediateKey{}
				key.SetUUID(googleUuid.New())
				key.SetEncrypted("encrypted-intermediate-key-data")
				key.SetKEKUUID(googleUuid.New())

				err := tx.AddIntermediateKey(key)
				require.NoError(t, err)

				addedKeys[i] = key
			}

			// Get latest key (latest by UUID DESC ordering, not insertion order).
			latestKey, err := tx.GetIntermediateKeyLatest()
			require.NoError(t, err)
			require.NotNil(t, latestKey)

			// Verify it's one of the keys we added.
			found := false

			for _, k := range addedKeys {
				if k.GetUUID() == latestKey.GetUUID() {
					found = true

					break
				}
			}

			require.True(t, found, "Latest key should be one of the added keys")

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("Get intermediate key by UUID", func(t *testing.T) {
		CleanupDatabase(t, testOrmRepository)

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add key.
			key := &IntermediateKey{}
			key.SetUUID(googleUuid.New())
			key.SetEncrypted("encrypted-intermediate-key-data")
			key.SetKEKUUID(googleUuid.New())

			err := tx.AddIntermediateKey(key)
			require.NoError(t, err)

			// Retrieve by UUID.
			targetUUID := key.GetUUID()
			retrievedKey, err := tx.GetIntermediateKey(&targetUUID)
			require.NoError(t, err)
			require.NotNil(t, retrievedKey)
			require.Equal(t, targetUUID, retrievedKey.GetUUID())

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("Delete intermediate key", func(t *testing.T) {
		CleanupDatabase(t, testOrmRepository)

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add 5 keys.
			const numKeys = 5

			addedKeys := make([]*IntermediateKey, numKeys)

			for i := 0; i < numKeys; i++ {
				key := &IntermediateKey{}
				key.SetUUID(googleUuid.New())
				key.SetEncrypted("encrypted-intermediate-key-data")
				key.SetKEKUUID(googleUuid.New())

				err := tx.AddIntermediateKey(key)
				require.NoError(t, err)

				addedKeys[i] = key
			}

			// Delete one key.
			targetUUID := addedKeys[2].GetUUID()
			deletedKey, err := tx.DeleteIntermediateKey(&targetUUID)
			require.NoError(t, err)
			require.NotNil(t, deletedKey)

			// Verify deletion.
			remainingKeys, err := tx.GetIntermediateKeys()
			require.NoError(t, err)
			require.Len(t, remainingKeys, numKeys-1)

			return nil
		})

		require.NoError(t, err)
	})
}

// TestContentKeyOperations tests content key CRUD operations.
func TestContentKeyOperations(t *testing.T) {
	t.Parallel()
	t.Run("Add and retrieve multiple content keys", func(t *testing.T) {
		CleanupDatabase(t, testOrmRepository)

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add 5 content keys.
			const numKeys = 5

			addedKeys := make([]*ContentKey, numKeys)

			for i := 0; i < numKeys; i++ {
				key := &ContentKey{}
				key.SetUUID(googleUuid.New())
				key.SetEncrypted("encrypted-content-key-data")
				key.SetKEKUUID(googleUuid.New())

				err := tx.AddContentKey(key)
				require.NoError(t, err)

				addedKeys[i] = key
			}

			// Retrieve all keys.
			keys, err := tx.GetContentKeys()
			require.NoError(t, err)
			require.Len(t, keys, numKeys)

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("Get latest content key", func(t *testing.T) {
		CleanupDatabase(t, testOrmRepository)

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add 7 content keys.
			const numKeys = 7

			addedKeys := make([]*ContentKey, numKeys)

			for i := 0; i < numKeys; i++ {
				key := &ContentKey{}
				key.SetUUID(googleUuid.New())
				key.SetEncrypted("encrypted-content-key-data")
				key.SetKEKUUID(googleUuid.New())

				err := tx.AddContentKey(key)
				require.NoError(t, err)

				addedKeys[i] = key
			}

			// Get latest key (latest by UUID DESC ordering, not insertion order).
			latestKey, err := tx.GetContentKeyLatest()
			require.NoError(t, err)
			require.NotNil(t, latestKey)

			// Verify it's one of the keys we added.
			found := false

			for _, k := range addedKeys {
				if k.GetUUID() == latestKey.GetUUID() {
					found = true

					break
				}
			}

			require.True(t, found, "Latest key should be one of the added keys")

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("Get content key by UUID", func(t *testing.T) {
		CleanupDatabase(t, testOrmRepository)

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add key.
			key := &ContentKey{}
			key.SetUUID(googleUuid.New())
			key.SetEncrypted("encrypted-content-key-data")
			key.SetKEKUUID(googleUuid.New())

			err := tx.AddContentKey(key)
			require.NoError(t, err)

			// Retrieve by UUID.
			targetUUID := key.GetUUID()
			retrievedKey, err := tx.GetContentKey(&targetUUID)
			require.NoError(t, err)
			require.NotNil(t, retrievedKey)
			require.Equal(t, targetUUID, retrievedKey.GetUUID())

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("Delete content key", func(t *testing.T) {
		CleanupDatabase(t, testOrmRepository)

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add 6 keys.
			const numKeys = 6

			addedKeys := make([]*ContentKey, numKeys)

			for i := 0; i < numKeys; i++ {
				key := &ContentKey{}
				key.SetUUID(googleUuid.New())
				key.SetEncrypted("encrypted-content-key-data")
				key.SetKEKUUID(googleUuid.New())

				err := tx.AddContentKey(key)
				require.NoError(t, err)

				addedKeys[i] = key
			}

			// Delete one key.
			targetUUID := addedKeys[3].GetUUID()
			deletedKey, err := tx.DeleteContentKey(&targetUUID)
			require.NoError(t, err)
			require.NotNil(t, deletedKey)

			// Verify deletion.
			remainingKeys, err := tx.GetContentKeys()
			require.NoError(t, err)
			require.Len(t, remainingKeys, numKeys-1)

			return nil
		})

		require.NoError(t, err)
	})
}
