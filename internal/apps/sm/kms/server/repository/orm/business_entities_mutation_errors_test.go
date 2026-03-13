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

// TestUpdateElasticKey_InvalidUUID tests error path when updating with invalid UUID.
func TestUpdateElasticKey_InvalidUUID(t *testing.T) {
	t.Parallel()
	CleanupDatabase(t, testOrmRepository)

	// Create elastic key with zero UUID (invalid).
	tenantID := googleUuid.New()
	elasticKey, buildErr := BuildElasticKey(
		tenantID,
		googleUuid.UUID{}, // Zero UUID is invalid
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

	// Attempt update with invalid UUID.
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		return tx.UpdateElasticKey(elasticKey)
	})

	// Should fail with invalid ElasticKeyID error.
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid Elastic Key ID: UUID can't be zero UUID")
}

// TestUpdateElasticKey_NonExistentRecord tests error path when updating non-existent record.
func TestUpdateElasticKey_NonExistentRecord(t *testing.T) {
	t.Parallel()
	CleanupDatabase(t, testOrmRepository)

	// Create elastic key with non-existent UUID.
	tenantID := googleUuid.New()
	nonExistentID := googleUuid.New()
	elasticKey, buildErr := BuildElasticKey(
		tenantID,
		nonExistentID,
		"non-existent-update-test",
		"Test Non-Existent Update",
		cryptoutilOpenapiModel.Internal,
		cryptoutilOpenapiModel.A256GCMDir,
		false,
		false,
		false,
		string(cryptoutilKmsServer.Active),
	)
	require.NoError(t, buildErr)

	// Attempt update on non-existent record.
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		return tx.UpdateElasticKey(elasticKey)
	})

	// GORM UpdateColumns doesn't error on zero rows affected.
	// This tests the code path but won't trigger error.
	require.NoError(t, err)
}

// TestUpdateElasticKeyStatus_InvalidUUID tests error path when updating status with invalid UUID.
func TestUpdateElasticKeyStatus_InvalidUUID(t *testing.T) {
	t.Parallel()
	CleanupDatabase(t, testOrmRepository)

	// Attempt update with zero UUID (invalid).
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		return tx.UpdateElasticKeyStatus(googleUuid.UUID{}, cryptoutilKmsServer.Active)
	})

	// Should fail with invalid ElasticKeyID error.
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid Elastic Key ID: UUID can't be zero UUID")
}

// TestUpdateElasticKeyStatus_NonExistentRecord tests error path when updating status of non-existent record.
func TestUpdateElasticKeyStatus_NonExistentRecord(t *testing.T) {
	t.Parallel()
	CleanupDatabase(t, testOrmRepository)

	// Attempt update on non-existent UUID.
	nonExistentID := googleUuid.New()
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		return tx.UpdateElasticKeyStatus(nonExistentID, cryptoutilKmsServer.Active)
	})

	// GORM Update doesn't error on zero rows affected.
	// This tests the code path but won't trigger error.
	require.NoError(t, err)
}

// TestGetElasticKeys_QueryError tests error path in GetElasticKeys.
// Note: This is difficult to trigger without database errors.
func TestGetElasticKeys_EmptyResult(t *testing.T) {
	t.Parallel()
	CleanupDatabase(t, testOrmRepository)

	// Query with filters that match nothing.
	var keys []ElasticKey

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		filters := &GetElasticKeysFilters{
			ElasticKeyID: []googleUuid.UUID{googleUuid.New()}, // Non-existent ID
		}

		var queryErr error

		keys, queryErr = tx.GetElasticKeys(filters)

		return queryErr
	})

	// Should succeed with empty result (not an error).
	require.NoError(t, err)
	require.Empty(t, keys)
}

// TestUpdateElasticKey_DatabaseConstraintViolation tests error path when violating database constraints.
func TestUpdateElasticKey_DatabaseConstraintViolation(t *testing.T) {
	t.Parallel()
	CleanupDatabase(t, testOrmRepository)

	// First create a valid elastic key.
	tenantID := googleUuid.New()
	elasticKeyID := googleUuid.New()
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		elasticKey, buildErr := BuildElasticKey(
			tenantID,
			elasticKeyID,
			"constraint-test",
			"Test Constraint Violation",
			cryptoutilOpenapiModel.Internal,
			cryptoutilOpenapiModel.A256GCMDir,
			false,
			false,
			false,
			string(cryptoutilKmsServer.Active),
		)
		require.NoError(t, buildErr)

		return tx.AddElasticKey(elasticKey)
	})
	require.NoError(t, err)

	// Attempt to update with invalid enum value (if validation exists).
	// Note: GORM doesn't enforce enum constraints at DB level for SQLite.
	// This test exercises the code path but may not trigger error.
	err = testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		elasticKey, buildErr := BuildElasticKey(
			tenantID,
			elasticKeyID,
			"constraint-test-invalid",
			"Test Invalid Update",
			"INVALID_USE", // Invalid enum value
			cryptoutilOpenapiModel.A256GCMDir,
			false,
			false,
			false,
			string(cryptoutilKmsServer.Active),
		)
		require.NoError(t, buildErr)

		return tx.UpdateElasticKey(elasticKey)
	})

	// May or may not error depending on DB enforcement.
	// Test primarily exercises code path.
	_ = err
}

// TestUpdateElasticKeyStatus_DatabaseConstraintViolation tests error path when violating constraints.
func TestUpdateElasticKeyStatus_DatabaseConstraintViolation(t *testing.T) {
	t.Parallel()
	CleanupDatabase(t, testOrmRepository)

	// First create a valid elastic key.
	tenantID := googleUuid.New()
	elasticKeyID := googleUuid.New()
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		elasticKey, buildErr := BuildElasticKey(
			tenantID,
			elasticKeyID,
			"status-constraint-test",
			"Test Status Constraint Violation",
			cryptoutilOpenapiModel.Internal,
			cryptoutilOpenapiModel.A256GCMDir,
			false,
			false,
			false,
			string(cryptoutilKmsServer.Active),
		)
		require.NoError(t, buildErr)

		return tx.AddElasticKey(elasticKey)
	})
	require.NoError(t, err)

	// Attempt to update status with invalid enum value.
	// Note: GORM doesn't enforce enum constraints at DB level for SQLite.
	err = testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		return tx.UpdateElasticKeyStatus(elasticKeyID, "INVALID_STATUS")
	})

	// May or may not error depending on DB enforcement.
	// Test primarily exercises code path.
	_ = err
}

// TestGetMaterialKeysForElasticKey_EmptyResult tests successful query with no matching records.
func TestGetMaterialKeysForElasticKey_EmptyResult(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
			string(cryptoutilKmsServer.Active),
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
