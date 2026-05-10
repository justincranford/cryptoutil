//go:build integration
// +build integration

// Copyright (c) 2025-2026 Justin Cranford.
package orm

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilOpenapiModel "cryptoutil/api/sm-kms/models"
	cryptoutilKmsServer "cryptoutil/api/sm-kms/server"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Sequential: uses shared package-level SQLite fixture state via CleanupDatabase.
// TestOrmTransaction_GetElasticKeyMaterialKeyVersion tests getting a specific material key version.
func TestOrmTransaction_GetElasticKeyMaterialKeyVersion(t *testing.T) {
	CleanupDatabase(t, testOrmRepository, KMSCleanupTables)

	// Create elastic key and multiple material keys.
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Create elastic key.
		tenantID := googleUuid.New()
		ekID := googleUuid.New()
		elasticKey, buildErr := BuildElasticKey(
			tenantID,
			ekID,
			"material-key-version-test",
			"Test Material Key Version",
			cryptoutilOpenapiModel.Internal,
			cryptoutilOpenapiModel.A256GCMDir,
			false,
			false,
			false,
			string(cryptoutilKmsServer.Active),
		)
		require.NoError(t, buildErr)

		createErr := AddElasticKey(tx.GormTx(), testTelemetryService.Slogger, elasticKey)
		require.NoError(t, createErr)

		// Create multiple material keys.
		mkID1 := googleUuid.New()
		materialKey1 := &MaterialKey{
			ElasticKeyID:                  ekID,
			MaterialKeyID:                 mkID1,
			MaterialKeyClearPublic:        []byte("public-key-data-1"),
			MaterialKeyEncryptedNonPublic: []byte("encrypted-private-key-data-1"),
		}

		createErr = AddElasticKeyMaterialKey(tx.GormTx(), testTelemetryService.Slogger, materialKey1)
		require.NoError(t, createErr)

		mkID2 := googleUuid.New()
		materialKey2 := &MaterialKey{
			ElasticKeyID:                  ekID,
			MaterialKeyID:                 mkID2,
			MaterialKeyClearPublic:        []byte("public-key-data-2"),
			MaterialKeyEncryptedNonPublic: []byte("encrypted-private-key-data-2"),
		}

		createErr = AddElasticKeyMaterialKey(tx.GormTx(), testTelemetryService.Slogger, materialKey2)
		require.NoError(t, createErr)

		return nil
	})
	require.NoError(t, err)

	// Test GetElasticKeyMaterialKeyVersion.
	err = testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		// Get all material keys.
		filters := GetMaterialKeysFilters{
			PageSize: cryptoutilSharedMagic.JoseJAMaxMaterials,
		}
		allKeys, getErr := GetMaterialKeys(tx.GormTx(), testTelemetryService.Slogger, &filters)
		require.NoError(t, getErr)
		require.Len(t, allKeys, 2, "Should have 2 material keys")

		// Get specific version by ekID and mkID.
		key1, getErr := GetElasticKeyMaterialKeyVersion(tx.GormTx(), testTelemetryService.Slogger, &allKeys[0].ElasticKeyID, &allKeys[0].MaterialKeyID)
		require.NoError(t, getErr)
		require.NotNil(t, key1)
		require.Equal(t, allKeys[0].MaterialKeyID, key1.MaterialKeyID)

		key2, getErr := GetElasticKeyMaterialKeyVersion(tx.GormTx(), testTelemetryService.Slogger, &allKeys[1].ElasticKeyID, &allKeys[1].MaterialKeyID)
		require.NoError(t, getErr)
		require.NotNil(t, key2)
		require.Equal(t, allKeys[1].MaterialKeyID, key2.MaterialKeyID)

		return nil
	})
	require.NoError(t, err)
}

// Sequential: uses shared package-level SQLite fixture state via CleanupDatabase.
// TestOrmTransaction_GetElasticKeyMaterialKeyLatest tests getting the latest material key.
func TestOrmTransaction_GetElasticKeyMaterialKeyLatest(t *testing.T) {
	CleanupDatabase(t, testOrmRepository, KMSCleanupTables)

	var ekID googleUuid.UUID

	var mkID1, mkID2 googleUuid.UUID

	// Create elastic key and multiple material keys.
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Create elastic key.
		tenantID := googleUuid.New()
		ekID = googleUuid.New()
		elasticKey, buildErr := BuildElasticKey(
			tenantID,
			ekID,
			"latest-material-key-test",
			"Test Latest Material Key",
			cryptoutilOpenapiModel.Internal,
			cryptoutilOpenapiModel.A256GCMDir,
			false,
			false,
			false,
			string(cryptoutilKmsServer.Active),
		)
		require.NoError(t, buildErr)

		createErr := AddElasticKey(tx.GormTx(), testTelemetryService.Slogger, elasticKey)
		require.NoError(t, createErr)

		// Create first material key (older).
		mkID1 = googleUuid.New()
		materialKey1 := &MaterialKey{
			ElasticKeyID:                  ekID,
			MaterialKeyID:                 mkID1,
			MaterialKeyClearPublic:        []byte("public-key-data-1"),
			MaterialKeyEncryptedNonPublic: []byte("encrypted-private-key-data-1"),
		}

		createErr = AddElasticKeyMaterialKey(tx.GormTx(), testTelemetryService.Slogger, materialKey1)
		require.NoError(t, createErr)

		// Create second material key (newer).
		mkID2 = googleUuid.New()
		materialKey2 := &MaterialKey{
			ElasticKeyID:                  ekID,
			MaterialKeyID:                 mkID2,
			MaterialKeyClearPublic:        []byte("public-key-data-2"),
			MaterialKeyEncryptedNonPublic: []byte("encrypted-private-key-data-2"),
		}

		createErr = AddElasticKeyMaterialKey(tx.GormTx(), testTelemetryService.Slogger, materialKey2)
		require.NoError(t, createErr)

		return nil
	})
	require.NoError(t, err)

	// Test GetElasticKeyMaterialKeyLatest.
	err = testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		// Get latest material key.
		latestKey, getErr := GetElasticKeyMaterialKeyLatest(tx.GormTx(), testTelemetryService.Slogger, ekID)
		require.NoError(t, getErr)
		require.NotNil(t, latestKey)

		// Latest should be the material key with the highest material_key_id value (DESC order).
		// Since UUIDv7 is time-ordered and both were created in the same transaction,
		// we just verify that the latest key is one of the two we created.
		require.Contains(t, []googleUuid.UUID{mkID1, mkID2}, latestKey.MaterialKeyID,
			"Latest key should be one of the created keys")

		// The latest should be the one with the highest UUID value.
		expectedLatest := mkID1
		if mkID2.String() > mkID1.String() {
			expectedLatest = mkID2
		}

		require.Equal(t, expectedLatest, latestKey.MaterialKeyID,
			"Latest key should have the highest material_key_id value (DESC order)")

		return nil
	})
	require.NoError(t, err)
}

// TestOrmTransaction_GetElasticKeyMaterialKeyVersion_InvalidElasticKeyID tests validation errors.
func TestOrmTransaction_GetElasticKeyMaterialKeyVersion_InvalidElasticKeyID(t *testing.T) {
	t.Parallel()

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		// Test nil elasticKeyID.
		_, getErr := GetElasticKeyMaterialKeyVersion(tx.GormTx(), testTelemetryService.Slogger, nil, &googleUuid.UUID{})
		require.Error(t, getErr)
		require.Contains(t, getErr.Error(), ErrFailedToGetMaterialKeyByElasticKeyIDAndMaterialKeyID)

		return nil
	})
	require.NoError(t, err)
}

// TestOrmTransaction_GetElasticKeyMaterialKeyVersion_InvalidMaterialKeyID tests validation errors.
func TestOrmTransaction_GetElasticKeyMaterialKeyVersion_InvalidMaterialKeyID(t *testing.T) {
	t.Parallel()

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		ekID := googleUuid.New()

		// Test nil materialKeyID.
		_, getErr := GetElasticKeyMaterialKeyVersion(tx.GormTx(), testTelemetryService.Slogger, &ekID, nil)
		require.Error(t, getErr)
		require.Contains(t, getErr.Error(), ErrFailedToGetMaterialKeyByElasticKeyIDAndMaterialKeyID)

		return nil
	})
	require.NoError(t, err)
}

// TestOrmTransaction_GetElasticKeyMaterialKeyLatest_InvalidElasticKeyID tests validation errors.
func TestOrmTransaction_GetElasticKeyMaterialKeyLatest_InvalidElasticKeyID(t *testing.T) {
	t.Parallel()

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		// Test zero UUID (invalid).
		_, getErr := GetElasticKeyMaterialKeyLatest(tx.GormTx(), testTelemetryService.Slogger, googleUuid.UUID{})
		require.Error(t, getErr)
		require.Contains(t, getErr.Error(), ErrFailedToGetLatestMaterialKeyByElasticKeyID)

		return nil
	})
	require.NoError(t, err)
}

// TestOrmTransaction_GetElasticKeyMaterialKeyVersion_NotFound tests record not found error.
func TestOrmTransaction_GetElasticKeyMaterialKeyVersion_NotFound(t *testing.T) {
	t.Parallel()

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		ekID := googleUuid.New()
		mkID := googleUuid.New()

		// Get non-existent material key.
		_, getErr := GetElasticKeyMaterialKeyVersion(tx.GormTx(), testTelemetryService.Slogger, &ekID, &mkID)
		require.Error(t, getErr)
		require.Contains(t, getErr.Error(), ErrFailedToGetMaterialKeyByElasticKeyIDAndMaterialKeyID)

		return nil
	})
	require.NoError(t, err)
}

// TestOrmTransaction_GetElasticKeyMaterialKeyLatest_NotFound tests record not found error.
func TestOrmTransaction_GetElasticKeyMaterialKeyLatest_NotFound(t *testing.T) {
	t.Parallel()

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		ekID := googleUuid.New()

		// Get latest material key for non-existent elastic key.
		_, getErr := GetElasticKeyMaterialKeyLatest(tx.GormTx(), testTelemetryService.Slogger, ekID)
		require.Error(t, getErr)
		require.Contains(t, getErr.Error(), ErrFailedToGetLatestMaterialKeyByElasticKeyID)

		return nil
	})
	require.NoError(t, err)
}
