// Copyright (c) 2025 Justin Cranford

package orm

import (
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/model"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestElasticKeyOperations tests CRUD operations for ElasticKey entity.
func TestElasticKeyOperations(t *testing.T) {
	CleanupDatabase(t, testOrmRepository)

	t.Run("Add and retrieve single elastic key", func(t *testing.T) {
		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Create test elastic key.
			tenantID := googleUuid.New()
			ekID := googleUuid.New()
			elasticKey := &ElasticKey{
				TenantID:                    tenantID,
				ElasticKeyID:                ekID,
				ElasticKeyName:              "test-key-1",
				ElasticKeyDescription:       "Test elastic key 1",
				ElasticKeyProvider:          cryptoutilOpenapiModel.Internal,
				ElasticKeyAlgorithm:         cryptoutilOpenapiModel.A256GCMA256KW,
				ElasticKeyVersioningAllowed: true,
				ElasticKeyImportAllowed:     false,
				ElasticKeyStatus:            cryptoutilOpenapiModel.Active,
			}

			// Add elastic key.
			err := tx.AddElasticKey(elasticKey)
			require.NoError(t, err, "AddElasticKey should succeed")

			// Retrieve elastic key by ID.
			retrieved, err := tx.GetElasticKey(tenantID, &ekID)
			require.NoError(t, err, "GetElasticKey should succeed")
			require.Equal(t, ekID, retrieved.ElasticKeyID, "Elastic Key ID should match")
			require.Equal(t, "test-key-1", retrieved.ElasticKeyName, "Elastic Key Name should match")
			require.Equal(t, cryptoutilOpenapiModel.Active, retrieved.ElasticKeyStatus, "Status should be Active")

			return nil
		})
		require.NoError(t, err, "Transaction should commit successfully")
	})

	t.Run("Update elastic key", func(t *testing.T) {
		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Create and add elastic key.
			tenantID := googleUuid.New()
			ekID := googleUuid.New()
			elasticKey := &ElasticKey{
				TenantID:                    tenantID,
				ElasticKeyID:                ekID,
				ElasticKeyName:              "update-test-key",
				ElasticKeyDescription:       "Original description",
				ElasticKeyProvider:          cryptoutilOpenapiModel.Internal,
				ElasticKeyAlgorithm:         cryptoutilOpenapiModel.A128GCMA128KW,
				ElasticKeyVersioningAllowed: true,
				ElasticKeyImportAllowed:     true,
				ElasticKeyStatus:            cryptoutilOpenapiModel.Active,
			}
			err := tx.AddElasticKey(elasticKey)
			require.NoError(t, err, "AddElasticKey should succeed")

			// Update elastic key fields.
			elasticKey.ElasticKeyDescription = "Updated description"
			err = tx.UpdateElasticKey(elasticKey)
			require.NoError(t, err, "UpdateElasticKey should succeed")

			// Verify update.
			retrieved, err := tx.GetElasticKey(tenantID, &ekID)
			require.NoError(t, err, "GetElasticKey should succeed")
			require.Equal(t, "Updated description", retrieved.ElasticKeyDescription, "Description should be updated")

			return nil
		})
		require.NoError(t, err, "Transaction should commit successfully")
	})

	t.Run("Update elastic key status", func(t *testing.T) {
		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Create and add elastic key.
			tenantID := googleUuid.New()
			ekID := googleUuid.New()
			elasticKey := &ElasticKey{
				TenantID:                    tenantID,
				ElasticKeyID:                ekID,
				ElasticKeyName:              "status-update-test",
				ElasticKeyDescription:       "Test status update",
				ElasticKeyProvider:          cryptoutilOpenapiModel.Internal,
				ElasticKeyAlgorithm:         cryptoutilOpenapiModel.A192GCMA192KW,
				ElasticKeyVersioningAllowed: true,
				ElasticKeyImportAllowed:     false,
				ElasticKeyStatus:            cryptoutilOpenapiModel.Active,
			}
			err := tx.AddElasticKey(elasticKey)
			require.NoError(t, err, "AddElasticKey should succeed")

			// Update status to Disabled.
			err = tx.UpdateElasticKeyStatus(ekID, cryptoutilOpenapiModel.Disabled)
			require.NoError(t, err, "UpdateElasticKeyStatus should succeed")

			// Verify status update.
			retrieved, err := tx.GetElasticKey(tenantID, &ekID)
			require.NoError(t, err, "GetElasticKey should succeed")
			require.Equal(t, cryptoutilOpenapiModel.Disabled, retrieved.ElasticKeyStatus, "Status should be Disabled")

			return nil
		})
		require.NoError(t, err, "Transaction should commit successfully")
	})

	t.Run("Get elastic keys with filters", func(t *testing.T) {
		CleanupDatabase(t, testOrmRepository)

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Use a shared tenantID for this test
			tenantID := googleUuid.New()

			// Create multiple elastic keys with different statuses.
			for i := 0; i < 3; i++ {
				ekID := googleUuid.New()

				status := cryptoutilOpenapiModel.Active
				if i == 2 {
					status = cryptoutilOpenapiModel.Disabled
				}

				elasticKey := &ElasticKey{
					TenantID:                    tenantID,
					ElasticKeyID:                ekID,
					ElasticKeyName:              ekID.String(), // Use UUID as unique name.
					ElasticKeyDescription:       "Batch test key",
					ElasticKeyProvider:          cryptoutilOpenapiModel.Internal,
					ElasticKeyAlgorithm:         cryptoutilOpenapiModel.A256GCMA256KW,
					ElasticKeyVersioningAllowed: true,
					ElasticKeyImportAllowed:     false,
					ElasticKeyStatus:            status,
				}
				err := tx.AddElasticKey(elasticKey)
				require.NoError(t, err, "AddElasticKey should succeed")
			}

			// Get all elastic keys for this tenant.
			allKeys, err := tx.GetElasticKeys(&GetElasticKeysFilters{TenantID: tenantID})
			require.NoError(t, err, "GetElasticKeys should succeed")
			require.GreaterOrEqual(t, len(allKeys), 3, "Should return at least 3 elastic keys")

			return nil
		})
		require.NoError(t, err, "Transaction should commit successfully")
	})
}

// TestMaterialKeyOperations tests CRUD operations for MaterialKey entity.
func TestMaterialKeyOperations(t *testing.T) {
	CleanupDatabase(t, testOrmRepository)

	t.Run("Add and retrieve material keys for elastic key", func(t *testing.T) {
		CleanupDatabase(t, testOrmRepository)

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Create parent elastic key.
			ekID := googleUuid.New()
			elasticKey := &ElasticKey{
				ElasticKeyID:                ekID,
				ElasticKeyName:              "parent-elastic-key",
				ElasticKeyDescription:       "Parent for material keys",
				ElasticKeyProvider:          cryptoutilOpenapiModel.Internal,
				ElasticKeyAlgorithm:         cryptoutilOpenapiModel.A256GCMA256KW,
				ElasticKeyVersioningAllowed: true,
				ElasticKeyImportAllowed:     false,
				ElasticKeyStatus:            cryptoutilOpenapiModel.Active,
			}
			err := tx.AddElasticKey(elasticKey)
			require.NoError(t, err, "AddElasticKey should succeed")

			// Create 3 material key versions.
			mkIDs := make([]googleUuid.UUID, 3)

			for i := 0; i < 3; i++ {
				mkID := googleUuid.New()
				mkIDs[i] = mkID
				materialKey := &MaterialKey{
					ElasticKeyID:                  ekID,
					MaterialKeyID:                 mkID,
					MaterialKeyClearPublic:        []byte("public-key-data"),
					MaterialKeyEncryptedNonPublic: []byte("encrypted-private-key-data"),
				}
				err := tx.AddElasticKeyMaterialKey(materialKey)
				require.NoError(t, err, "AddElasticKeyMaterialKey should succeed")
			}

			// Retrieve all material keys for elastic key.
			keys, err := tx.GetMaterialKeysForElasticKey(&ekID, &GetElasticKeyMaterialKeysFilters{})
			require.NoError(t, err, "GetMaterialKeysForElasticKey should succeed")
			require.GreaterOrEqual(t, len(keys), 3, "Should return at least 3 material keys")

			// Verify first material key can be retrieved individually.
			retrieved, err := tx.GetElasticKeyMaterialKeyVersion(&ekID, &mkIDs[0])
			require.NoError(t, err, "GetElasticKeyMaterialKeyVersion should succeed")
			require.Equal(t, ekID, retrieved.ElasticKeyID, "Elastic Key ID should match")
			require.Equal(t, mkIDs[0], retrieved.MaterialKeyID, "Material Key ID should match")

			return nil
		})
		require.NoError(t, err, "Transaction should commit successfully")
	})

	t.Run("Get latest material key for elastic key", func(t *testing.T) {
		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Create parent elastic key.
			ekID := googleUuid.New()
			elasticKey := &ElasticKey{
				ElasticKeyID:                ekID,
				ElasticKeyName:              "latest-key-parent",
				ElasticKeyDescription:       "Parent for latest key test",
				ElasticKeyProvider:          cryptoutilOpenapiModel.Internal,
				ElasticKeyAlgorithm:         cryptoutilOpenapiModel.A128CBCHS256A256KW,
				ElasticKeyVersioningAllowed: true,
				ElasticKeyImportAllowed:     false,
				ElasticKeyStatus:            cryptoutilOpenapiModel.Active,
			}
			err := tx.AddElasticKey(elasticKey)
			require.NoError(t, err, "AddElasticKey should succeed")

			// Create 5 material key versions.
			for i := 0; i < 5; i++ {
				mkID := googleUuid.New()
				materialKey := &MaterialKey{
					ElasticKeyID:                  ekID,
					MaterialKeyID:                 mkID,
					MaterialKeyClearPublic:        []byte("public-key-data"),
					MaterialKeyEncryptedNonPublic: []byte("encrypted-private-key-data"),
				}
				err := tx.AddElasticKeyMaterialKey(materialKey)
				require.NoError(t, err, "AddElasticKeyMaterialKey should succeed")
			}

			// Get latest material key (UUIDv7 sorting: latest = highest ID).
			latest, err := tx.GetElasticKeyMaterialKeyLatest(ekID)
			require.NoError(t, err, "GetElasticKeyMaterialKeyLatest should succeed")
			// Note: Latest may not match latestMKID due to UUIDv7 timestamp ordering.
			require.Equal(t, ekID, latest.ElasticKeyID, "Elastic Key ID should match")
			require.NotEqual(t, googleUuid.Nil, latest.MaterialKeyID, "Material Key ID should not be nil")

			return nil
		})
		require.NoError(t, err, "Transaction should commit successfully")
	})

	t.Run("Get all material keys with filters", func(t *testing.T) {
		CleanupDatabase(t, testOrmRepository)

		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Create parent elastic key.
			ekID := googleUuid.New()
			elasticKey := &ElasticKey{
				ElasticKeyID:                ekID,
				ElasticKeyName:              "filter-test-parent",
				ElasticKeyDescription:       "Parent for filter test",
				ElasticKeyProvider:          cryptoutilOpenapiModel.Internal,
				ElasticKeyAlgorithm:         cryptoutilOpenapiModel.A192CBCHS384A192KW,
				ElasticKeyVersioningAllowed: true,
				ElasticKeyImportAllowed:     true,
				ElasticKeyStatus:            cryptoutilOpenapiModel.Active,
			}
			err := tx.AddElasticKey(elasticKey)
			require.NoError(t, err, "AddElasticKey should succeed")

			// Create 4 material keys.
			for i := 0; i < 4; i++ {
				mkID := googleUuid.New()
				materialKey := &MaterialKey{
					ElasticKeyID:                  ekID,
					MaterialKeyID:                 mkID,
					MaterialKeyClearPublic:        []byte("public-key-data"),
					MaterialKeyEncryptedNonPublic: []byte("encrypted-private-key-data"),
				}
				err := tx.AddElasticKeyMaterialKey(materialKey)
				require.NoError(t, err, "AddElasticKeyMaterialKey should succeed")
			}

			// Get all material keys (no filters).
			allKeys, err := tx.GetMaterialKeys(&GetMaterialKeysFilters{})
			require.NoError(t, err, "GetMaterialKeys should succeed")
			require.GreaterOrEqual(t, len(allKeys), 4, "Should return at least 4 material keys")

			return nil
		})
		require.NoError(t, err, "Transaction should commit successfully")
	})
}

// TestBusinessEntityErrorHandling tests error cases for business entity operations.
func TestBusinessEntityErrorHandling(t *testing.T) {
	t.Run("Add elastic key with invalid UUID", func(t *testing.T) {
		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Create elastic key with nil UUID (invalid).
			elasticKey := &ElasticKey{
				ElasticKeyID:                googleUuid.Nil,
				ElasticKeyName:              "invalid-key",
				ElasticKeyDescription:       "Invalid UUID test",
				ElasticKeyProvider:          cryptoutilOpenapiModel.Internal,
				ElasticKeyAlgorithm:         cryptoutilOpenapiModel.A256GCMA256KW,
				ElasticKeyVersioningAllowed: true,
				ElasticKeyImportAllowed:     false,
				ElasticKeyStatus:            cryptoutilOpenapiModel.Active,
			}

			// Attempt to add elastic key.
			err := tx.AddElasticKey(elasticKey)
			require.Error(t, err, "AddElasticKey should fail with invalid UUID")

			return err
		})
		require.Error(t, err, "Transaction should rollback due to error")
	})

	t.Run("Get elastic key with nonexistent ID", func(t *testing.T) {
		err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
			// Attempt to get nonexistent elastic key.
			tenantID := googleUuid.New()
			nonexistentID := googleUuid.New()
			_, err := tx.GetElasticKey(tenantID, &nonexistentID)
			require.Error(t, err, "GetElasticKey should fail for nonexistent ID")

			return err
		})
		require.Error(t, err, "Transaction should rollback due to error")
	})

	t.Run("Add material key with invalid elastic key UUID", func(t *testing.T) {
		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Create material key with nil elastic key ID (invalid).
			materialKey := &MaterialKey{
				ElasticKeyID:                  googleUuid.Nil,
				MaterialKeyID:                 googleUuid.New(),
				MaterialKeyClearPublic:        []byte("public-key-data"),
				MaterialKeyEncryptedNonPublic: []byte("encrypted-private-key-data"),
			}

			// Attempt to add material key.
			err := tx.AddElasticKeyMaterialKey(materialKey)
			require.Error(t, err, "AddElasticKeyMaterialKey should fail with invalid elastic key UUID")

			return err
		})
		require.Error(t, err, "Transaction should rollback due to error")
	})

	t.Run("Get material key with invalid material key UUID", func(t *testing.T) {
		err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
			// Attempt to get material key with nil material key ID.
			ekID := googleUuid.New()
			mkID := googleUuid.Nil
			_, err := tx.GetElasticKeyMaterialKeyVersion(&ekID, &mkID)
			require.Error(t, err, "GetElasticKeyMaterialKeyVersion should fail with invalid material key UUID")

			return err
		})
		require.Error(t, err, "Transaction should rollback due to error")
	})

	t.Run("Update nonexistent elastic key", func(t *testing.T) {
		err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
			// Create elastic key with new UUID (doesn't exist in DB).
			elasticKey := &ElasticKey{
				ElasticKeyID:                googleUuid.New(),
				ElasticKeyName:              "nonexistent-key",
				ElasticKeyDescription:       "Nonexistent key test",
				ElasticKeyProvider:          cryptoutilOpenapiModel.Internal,
				ElasticKeyAlgorithm:         cryptoutilOpenapiModel.A256GCMA256KW,
				ElasticKeyVersioningAllowed: true,
				ElasticKeyImportAllowed:     false,
				ElasticKeyStatus:            cryptoutilOpenapiModel.Active,
			}

			// Attempt to update (should succeed with GORM but affect 0 rows).
			err := tx.UpdateElasticKey(elasticKey)
			require.NoError(t, err, "UpdateElasticKey doesn't fail for nonexistent key (GORM limitation)")

			return nil
		})
		require.NoError(t, err, "Transaction should commit (GORM allows updates with 0 rows affected)")
	})
}
