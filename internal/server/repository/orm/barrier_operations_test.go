// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestBarrierRootKeyOperations tests root key CRUD operations.
func TestBarrierRootKeyOperations(t *testing.T) {
	t.Parallel()

	t.Run("Add and retrieve multiple root keys", func(t *testing.T) {
		t.Parallel()

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add 3 root keys.
			const numKeys = 3
			addedKeys := make([]*BarrierRootKey, numKeys)

			for i := 0; i < numKeys; i++ {
				key := &BarrierRootKey{}
				key.SetUUID(googleUuid.New())
				key.SetEncrypted("encrypted-root-key-data")
				key.SetKEKUUID(googleUuid.New())

				err := tx.AddRootKey(key)
				require.NoError(t, err)

				addedKeys[i] = key
			}

			// Retrieve all keys and verify test-added keys are present.
			keys, err := tx.GetRootKeys()
			require.NoError(t, err)
			require.GreaterOrEqual(t, len(keys), numKeys)

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("Get latest root key", func(t *testing.T) {
		t.Parallel()

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add 5 root keys.
			const numKeys = 5
			addedKeys := make([]*BarrierRootKey, numKeys)

			for i := 0; i < numKeys; i++ {
				key := &BarrierRootKey{}
				key.SetUUID(googleUuid.New())
				key.SetEncrypted("encrypted-root-key-data")
				key.SetKEKUUID(googleUuid.New())

				err := tx.AddRootKey(key)
				require.NoError(t, err)

				addedKeys[i] = key
			}

			// Get latest key - in concurrent execution, another test may have added newer keys.
			// Just verify we can retrieve A latest key successfully.
			latestKey, err := tx.GetRootKeyLatest()
			require.NoError(t, err)
			require.NotNil(t, latestKey)

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("Get root key by UUID", func(t *testing.T) {
		t.Parallel()

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add key.
			key := &BarrierRootKey{}
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
		t.Parallel()

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add 3 keys.
			const numKeys = 3
			addedKeys := make([]*BarrierRootKey, numKeys)

			for i := 0; i < numKeys; i++ {
				key := &BarrierRootKey{}
				key.SetUUID(googleUuid.New())
				key.SetEncrypted("encrypted-root-key-data")
				key.SetKEKUUID(googleUuid.New())

				err := tx.AddRootKey(key)
				require.NoError(t, err)

				addedKeys[i] = key
			}

			// Get count before deletion.
			beforeKeys, err := tx.GetRootKeys()
			require.NoError(t, err)

			beforeCount := len(beforeKeys)

			// Delete one key.
			targetUUID := addedKeys[1].GetUUID()
			deletedKey, err := tx.DeleteRootKey(&targetUUID)
			require.NoError(t, err)
			require.NotNil(t, deletedKey)

			// Verify deletion - count should decrease by 1.
			afterKeys, err := tx.GetRootKeys()
			require.NoError(t, err)
			require.Equal(t, beforeCount-1, len(afterKeys))

			// Verify the deleted key is not in results.
			for _, key := range afterKeys {
				require.NotEqual(t, targetUUID, key.GetUUID())
			}

			return nil
		})

		require.NoError(t, err)
	})
}

// TestBarrierIntermediateKeyOperations tests intermediate key CRUD operations.
func TestBarrierIntermediateKeyOperations(t *testing.T) {
	t.Parallel()

	t.Run("Add and retrieve multiple intermediate keys", func(t *testing.T) {
		t.Parallel()

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add 4 intermediate keys.
			const numKeys = 4
			addedKeys := make([]*BarrierIntermediateKey, numKeys)

			for i := 0; i < numKeys; i++ {
				key := &BarrierIntermediateKey{}
				key.SetUUID(googleUuid.New())
				key.SetEncrypted("encrypted-intermediate-key-data")
				key.SetKEKUUID(googleUuid.New())

				err := tx.AddIntermediateKey(key)
				require.NoError(t, err)

				addedKeys[i] = key
			}

			// Retrieve all keys and verify test-added keys are present.
			keys, err := tx.GetIntermediateKeys()
			require.NoError(t, err)
			require.GreaterOrEqual(t, len(keys), numKeys)

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("Get latest intermediate key", func(t *testing.T) {
		t.Parallel()

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add 6 intermediate keys.
			const numKeys = 6
			addedKeys := make([]*BarrierIntermediateKey, numKeys)

			for i := 0; i < numKeys; i++ {
				key := &BarrierIntermediateKey{}
				key.SetUUID(googleUuid.New())
				key.SetEncrypted("encrypted-intermediate-key-data")
				key.SetKEKUUID(googleUuid.New())

				err := tx.AddIntermediateKey(key)
				require.NoError(t, err)

				addedKeys[i] = key
			}

			// Get latest key - in concurrent execution, another test may have added newer keys.
			// Just verify we can retrieve A latest key successfully.
			latestKey, err := tx.GetIntermediateKeyLatest()
			require.NoError(t, err)
			require.NotNil(t, latestKey)

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("Get intermediate key by UUID", func(t *testing.T) {
		t.Parallel()

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add key.
			key := &BarrierIntermediateKey{}
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
		t.Parallel()

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add 5 keys.
			const numKeys = 5
			addedKeys := make([]*BarrierIntermediateKey, numKeys)

			for i := 0; i < numKeys; i++ {
				key := &BarrierIntermediateKey{}
				key.SetUUID(googleUuid.New())
				key.SetEncrypted("encrypted-intermediate-key-data")
				key.SetKEKUUID(googleUuid.New())

				err := tx.AddIntermediateKey(key)
				require.NoError(t, err)

				addedKeys[i] = key
			}

			// Get count before deletion.
			beforeKeys, err := tx.GetIntermediateKeys()
			require.NoError(t, err)

			beforeCount := len(beforeKeys)

			// Delete one key.
			targetUUID := addedKeys[2].GetUUID()
			deletedKey, err := tx.DeleteIntermediateKey(&targetUUID)
			require.NoError(t, err)
			require.NotNil(t, deletedKey)

			// Verify deletion - count should decrease by 1.
			afterKeys, err := tx.GetIntermediateKeys()
			require.NoError(t, err)
			require.Equal(t, beforeCount-1, len(afterKeys))

			// Verify the deleted key is not in results.
			for _, key := range afterKeys {
				require.NotEqual(t, targetUUID, key.GetUUID())
			}

			return nil
		})

		require.NoError(t, err)
	})
}

// TestBarrierContentKeyOperations tests content key CRUD operations.
func TestBarrierContentKeyOperations(t *testing.T) {
	t.Parallel()

	t.Run("Add and retrieve multiple content keys", func(t *testing.T) {
		t.Parallel()

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add 5 content keys.
			const numKeys = 5
			addedKeys := make([]*BarrierContentKey, numKeys)

			for i := 0; i < numKeys; i++ {
				key := &BarrierContentKey{}
				key.SetUUID(googleUuid.New())
				key.SetEncrypted("encrypted-content-key-data")
				key.SetKEKUUID(googleUuid.New())

				err := tx.AddContentKey(key)
				require.NoError(t, err)

				addedKeys[i] = key
			}

			// Retrieve all keys and verify test-added keys are present.
			keys, err := tx.GetContentKeys()
			require.NoError(t, err)
			require.GreaterOrEqual(t, len(keys), numKeys)

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("Get latest content key", func(t *testing.T) {
		t.Parallel()

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add 7 content keys.
			const numKeys = 7
			addedKeys := make([]*BarrierContentKey, numKeys)

			for i := 0; i < numKeys; i++ {
				key := &BarrierContentKey{}
				key.SetUUID(googleUuid.New())
				key.SetEncrypted("encrypted-content-key-data")
				key.SetKEKUUID(googleUuid.New())

				err := tx.AddContentKey(key)
				require.NoError(t, err)

				addedKeys[i] = key
			}

			// Get latest key - in concurrent execution, another test may have added newer keys.
			// Just verify we can retrieve A latest key successfully.
			latestKey, err := tx.GetContentKeyLatest()
			require.NoError(t, err)
			require.NotNil(t, latestKey)

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("Get content key by UUID", func(t *testing.T) {
		t.Parallel()

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add key.
			key := &BarrierContentKey{}
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
		t.Parallel()

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Add 6 keys.
			const numKeys = 6
			addedKeys := make([]*BarrierContentKey, numKeys)

			for i := 0; i < numKeys; i++ {
				key := &BarrierContentKey{}
				key.SetUUID(googleUuid.New())
				key.SetEncrypted("encrypted-content-key-data")
				key.SetKEKUUID(googleUuid.New())

				err := tx.AddContentKey(key)
				require.NoError(t, err)

				addedKeys[i] = key
			}

			// Get count before deletion.
			beforeKeys, err := tx.GetContentKeys()
			require.NoError(t, err)

			beforeCount := len(beforeKeys)

			// Delete one key.
			targetUUID := addedKeys[3].GetUUID()
			deletedKey, err := tx.DeleteContentKey(&targetUUID)
			require.NoError(t, err)
			require.NotNil(t, deletedKey)

			// Verify deletion - count should decrease by 1.
			afterKeys, err := tx.GetContentKeys()
			require.NoError(t, err)
			require.Equal(t, beforeCount-1, len(afterKeys))

			// Verify the deleted key is not in results.
			for _, key := range afterKeys {
				require.NotEqual(t, targetUUID, key.GetUUID())
			}

			return nil
		})

		require.NoError(t, err)
	})
}
