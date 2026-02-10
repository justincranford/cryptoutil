//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford

package orm

import (
	"testing"

	cryptoutilKmsServer "cryptoutil/api/kms/server"
	cryptoutilOpenapiModel "cryptoutil/api/model"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestOrmTransaction_AddElasticKey_InvalidUUID tests validation error for invalid elastic key ID.
func TestOrmTransaction_AddElasticKey_InvalidUUID(t *testing.T) {
	t.Parallel()
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Create elastic key with zero UUID (invalid).
		tenantID := googleUuid.New()
		elasticKey, buildErr := BuildElasticKey(
			tenantID,
			googleUuid.UUID{}, // Zero UUID - invalid
			"invalid-uuid-test",
			"Test Invalid UUID",
			cryptoutilOpenapiModel.Internal,
			cryptoutilOpenapiModel.A256GCMDir,
			false,
			false,
			false,
			string(cryptoutilKmsServer.Active),
		)
		require.NoError(t, buildErr)

		// Try to add elastic key with invalid UUID.
		createErr := tx.AddElasticKey(elasticKey)
		require.Error(t, createErr, "Should fail with invalid UUID")
		require.Contains(t, createErr.Error(), ErrFailedToAddElasticKey)

		return nil
	})
	require.NoError(t, err)
}

// TestOrmTransaction_UpdateElasticKey_InvalidUUID tests validation error for invalid elastic key ID.
func TestOrmTransaction_UpdateElasticKey_InvalidUUID(t *testing.T) {
	t.Parallel()
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Create elastic key with zero UUID (invalid).
		tenantID := googleUuid.New()
		elasticKey, buildErr := BuildElasticKey(
			tenantID,
			googleUuid.UUID{}, // Zero UUID - invalid
			"invalid-uuid-update-test",
			"Test Invalid UUID Update",
			cryptoutilOpenapiModel.Internal,
			cryptoutilOpenapiModel.A256GCMDir,
			false,
			false,
			false,
			string(cryptoutilKmsServer.Active),
		)
		require.NoError(t, buildErr)

		// Try to update elastic key with invalid UUID.
		updateErr := tx.UpdateElasticKey(elasticKey)
		require.Error(t, updateErr, "Should fail with invalid UUID")
		require.Contains(t, updateErr.Error(), ErrFailedToUpdateElasticKeyByElasticKeyID)

		return nil
	})
	require.NoError(t, err)
}

// TestOrmTransaction_UpdateElasticKeyStatus_InvalidUUID tests validation error for invalid elastic key ID.
func TestOrmTransaction_UpdateElasticKeyStatus_InvalidUUID(t *testing.T) {
	t.Parallel()
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Try to update status with zero UUID (invalid).
		updateErr := tx.UpdateElasticKeyStatus(googleUuid.UUID{}, cryptoutilKmsServer.Inactive)
		require.Error(t, updateErr, "Should fail with invalid UUID")
		require.Contains(t, updateErr.Error(), ErrFailedToUpdateElasticKeyStatusByElasticKeyID)

		return nil
	})
	require.NoError(t, err)
}

// TestOrmTransaction_GetElasticKey_InvalidUUID tests validation error for invalid elastic key ID.
func TestOrmTransaction_GetElasticKey_InvalidUUID(t *testing.T) {
	t.Parallel()
	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		// Try to get elastic key with nil UUID.
		tenantID := googleUuid.New()
		_, getErr := tx.GetElasticKey(tenantID, nil)
		require.Error(t, getErr, "Should fail with nil UUID")
		require.Contains(t, getErr.Error(), ErrFailedToGetElasticKeyByElasticKeyID)

		return nil
	})
	require.NoError(t, err)
}

// TestOrmTransaction_AddElasticKeyMaterialKey_InvalidElasticKeyUUID tests validation error.
func TestOrmTransaction_AddElasticKeyMaterialKey_InvalidElasticKeyUUID(t *testing.T) {
	t.Parallel()
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Create material key with zero elastic key UUID (invalid).
		materialKey := &MaterialKey{
			ElasticKeyID:                  googleUuid.UUID{}, // Zero UUID - invalid
			MaterialKeyID:                 googleUuid.New(),
			MaterialKeyClearPublic:        []byte("public-key-data"),
			MaterialKeyEncryptedNonPublic: []byte("encrypted-private-key-data"),
		}

		// Try to add material key with invalid elastic key UUID.
		createErr := tx.AddElasticKeyMaterialKey(materialKey)
		require.Error(t, createErr, "Should fail with invalid elastic key UUID")
		require.Contains(t, createErr.Error(), ErrFailedToAddMaterialKey)

		return nil
	})
	require.NoError(t, err)
}

// TestOrmTransaction_AddElasticKeyMaterialKey_InvalidMaterialKeyUUID tests validation error.
func TestOrmTransaction_AddElasticKeyMaterialKey_InvalidMaterialKeyUUID(t *testing.T) {
	t.Parallel()
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		// Create material key with zero material key UUID (invalid).
		materialKey := &MaterialKey{
			ElasticKeyID:                  googleUuid.New(),
			MaterialKeyID:                 googleUuid.UUID{}, // Zero UUID - invalid
			MaterialKeyClearPublic:        []byte("public-key-data"),
			MaterialKeyEncryptedNonPublic: []byte("encrypted-private-key-data"),
		}

		// Try to add material key with invalid material key UUID.
		createErr := tx.AddElasticKeyMaterialKey(materialKey)
		require.Error(t, createErr, "Should fail with invalid material key UUID")
		require.Contains(t, createErr.Error(), ErrFailedToAddMaterialKey)

		return nil
	})
	require.NoError(t, err)
}

// TestOrmTransaction_GetMaterialKeysForElasticKey_InvalidUUID tests validation error.
func TestOrmTransaction_GetMaterialKeysForElasticKey_InvalidUUID(t *testing.T) {
	t.Parallel()
	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		// Try to get material keys with nil elastic key UUID.
		_, getErr := tx.GetMaterialKeysForElasticKey(nil, &GetElasticKeyMaterialKeysFilters{})
		require.Error(t, getErr, "Should fail with nil UUID")
		require.Contains(t, getErr.Error(), ErrFailedToGetMaterialKeysByElasticKeyID)

		return nil
	})
	require.NoError(t, err)
}
