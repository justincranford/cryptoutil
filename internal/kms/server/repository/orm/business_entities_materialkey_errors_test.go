// Copyright (c) 2025 Justin Cranford

package orm

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilOpenapiModel "cryptoutil/api/model"
)

// TestGetMaterialKeysForElasticKey_EmptyResult tests successful query with no matching records.
func TestGetMaterialKeysForElasticKey_EmptyResult(t *testing.T) {
	nonExistentID := googleUuid.New()

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		// Query for material keys with non-existent elastic key ID.
		keys, getErr := tx.GetMaterialKeysForElasticKey(&nonExistentID, &GetElasticKeyMaterialKeysFilters{})
		require.NoError(t, getErr, "Query should succeed even with no results")
		require.Empty(t, keys, "Should return empty slice for non-existent elastic key")

		return nil
	})
	require.NoError(t, err)
}

// TestGetMaterialKeys_EmptyResult tests successful query with filters returning no matches.
func TestGetMaterialKeys_EmptyResult(t *testing.T) {
	nonExistentID := googleUuid.New()

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		// Query with filter that matches no records.
		filters := &GetMaterialKeysFilters{
			ElasticKeyID: []googleUuid.UUID{nonExistentID},
		}
		keys, getErr := tx.GetMaterialKeys(filters)
		require.NoError(t, getErr, "Query should succeed even with no results")
		require.Empty(t, keys, "Should return empty slice for non-matching filter")

		return nil
	})
	require.NoError(t, err)
}

// TestAddElasticKeyMaterialKey_DuplicateConstraintViolation tests database constraint error for duplicate key.
func TestAddElasticKeyMaterialKey_DuplicateConstraintViolation(t *testing.T) {
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Create elastic key first.
		tenantID := googleUuid.New()
		elasticKey, buildErr := BuildElasticKey(
			tenantID,
			googleUuid.New(),
			"material-key-duplicate-test",
			"Test Material Key Duplicate",
			cryptoutilOpenapiModel.Internal,
			cryptoutilOpenapiModel.A256GCMDir,
			false,
			false,
			false,
			string(cryptoutilOpenapiModel.Creating),
		)
		require.NoError(t, buildErr, "Should build elastic key")

		createErr := tx.AddElasticKey(elasticKey)
		require.NoError(t, createErr, "Should create elastic key")

		// Create first material key successfully.
		materialKey1 := &MaterialKey{
			ElasticKeyID:                  elasticKey.ElasticKeyID,
			MaterialKeyID:                 googleUuid.New(),
			MaterialKeyClearPublic:        []byte("public-key-data-1"),
			MaterialKeyEncryptedNonPublic: []byte("encrypted-private-key-data-1"),
		}
		addErr1 := tx.AddElasticKeyMaterialKey(materialKey1)
		require.NoError(t, addErr1, "First material key creation should succeed")

		// Attempt to create duplicate material key (same ElasticKeyID + MaterialKeyID).
		materialKey2 := &MaterialKey{
			ElasticKeyID:                  elasticKey.ElasticKeyID,
			MaterialKeyID:                 materialKey1.MaterialKeyID, // DUPLICATE
			MaterialKeyClearPublic:        []byte("public-key-data-2"),
			MaterialKeyEncryptedNonPublic: []byte("encrypted-private-key-data-2"),
		}
		addErr2 := tx.AddElasticKeyMaterialKey(materialKey2)
		require.Error(t, addErr2, "Duplicate material key should fail")
		require.Contains(t, addErr2.Error(), "UNIQUE", "Error should mention UNIQUE constraint")

		return nil
	})
	require.NoError(t, err)
}
