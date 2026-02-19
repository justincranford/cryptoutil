//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford

package orm

import (
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilKmsServer "cryptoutil/api/kms/server"
	cryptoutilOpenapiModel "cryptoutil/api/model"
)

// TestAddElasticKey_DuplicateConstraintViolation tests AddElasticKey duplicate key error.
func TestAddElasticKey_DuplicateConstraintViolation(t *testing.T) {
	t.Parallel()

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Create elastic key successfully.
		tenantID := googleUuid.New()
		elasticKey, buildErr := BuildElasticKey(
			tenantID,
			googleUuid.New(),
			"duplicate-elastic-key-test",
			"Test Duplicate Elastic Key",
			cryptoutilOpenapiModel.Internal,
			cryptoutilOpenapiModel.A256GCMDir,
			false,
			false,
			false,
			string(cryptoutilKmsServer.Active),
		)
		require.NoError(t, buildErr, "Should build elastic key")

		createErr := tx.AddElasticKey(elasticKey)
		require.NoError(t, createErr, "First elastic key creation should succeed")

		// Attempt to create duplicate elastic key (same ID).
		duplicateKey := &ElasticKey{
			ElasticKeyID:                elasticKey.ElasticKeyID, // DUPLICATE
			ElasticKeyName:              "duplicate-name",
			ElasticKeyDescription:       "Duplicate description",
			ElasticKeyProvider:          cryptoutilOpenapiModel.Internal,
			ElasticKeyAlgorithm:         cryptoutilOpenapiModel.A256GCMDir,
			ElasticKeyVersioningAllowed: false,
			ElasticKeyImportAllowed:     false,
			ElasticKeyStatus:            cryptoutilKmsServer.Active,
		}
		addErr := tx.AddElasticKey(duplicateKey)
		require.Error(t, addErr, "Duplicate elastic key should fail")
		require.Contains(t, addErr.Error(), "UNIQUE", "Error should mention UNIQUE constraint")

		return nil
	})
	require.NoError(t, err)
}

// TestGetElasticKeyMaterialKeyLatest_NotFoundError tests GetElasticKeyMaterialKeyLatest when no material keys exist.
func TestGetElasticKeyMaterialKeyLatest_NotFoundError(t *testing.T) {
	t.Parallel()

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Create elastic key without any material keys.
		tenantID := googleUuid.New()
		elasticKey, buildErr := BuildElasticKey(
			tenantID,
			googleUuid.New(),
			"no-material-keys-test",
			"Test No Material Keys",
			cryptoutilOpenapiModel.Internal,
			cryptoutilOpenapiModel.A256GCMDir,
			false,
			false,
			false,
			string(cryptoutilKmsServer.Active),
		)
		require.NoError(t, buildErr, "Should build elastic key")

		createErr := tx.AddElasticKey(elasticKey)
		require.NoError(t, createErr, "Elastic key creation should succeed")

		// Attempt to get latest material key when none exist.
		_, getErr := tx.GetElasticKeyMaterialKeyLatest(elasticKey.ElasticKeyID)
		require.Error(t, getErr, "Should fail when no material keys exist")
		require.Contains(t, getErr.Error(), ErrFailedToGetLatestMaterialKeyByElasticKeyID, "Error should indicate get latest failure")

		return nil
	})
	require.NoError(t, err)
}

// TestGetElasticKeyMaterialKeyVersion_NotFoundError tests GetElasticKeyMaterialKeyVersion when material key version does not exist.
func TestGetElasticKeyMaterialKeyVersion_NotFoundError(t *testing.T) {
	t.Parallel()

	nonExistentElasticKeyID := googleUuid.New()
	nonExistentMaterialKeyID := googleUuid.New()

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		// Attempt to get non-existent material key version.
		_, getErr := tx.GetElasticKeyMaterialKeyVersion(&nonExistentElasticKeyID, &nonExistentMaterialKeyID)
		require.Error(t, getErr, "Should fail when material key version not found")
		require.Contains(t, getErr.Error(), ErrFailedToGetMaterialKeyByElasticKeyIDAndMaterialKeyID, "Error should indicate get failure")

		return nil
	})
	require.NoError(t, err)
}
