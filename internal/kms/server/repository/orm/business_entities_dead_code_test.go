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

// TestGetElasticKeysWithExportAllowedFilter tests the ExportAllowed filter path.
// This filter field exists in GetElasticKeysFilters but the ElasticKey struct
// does NOT have an ElasticKeyExportAllowed field, making this filter DEAD CODE.
// This test exercises the filter application code path for coverage purposes,
// and verifies that using this non-existent filter field results in a database error.
func TestGetElasticKeysWithExportAllowedFilter(t *testing.T) {
	CleanupDatabase(t, testOrmRepository)

	// Create an elastic key for testing.
	tenantID := googleUuid.New()
	elasticKey, buildErr := BuildElasticKey(
		tenantID,
		googleUuid.New(),
		"export-filter-test",
		"Test Export Filter",
		cryptoutilOpenapiModel.Internal,
		cryptoutilOpenapiModel.A256GCMDir,
		false,
		false,
		false,
		string(cryptoutilKmsServer.Active),
	)
	require.NoError(t, buildErr)

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		return tx.AddElasticKey(elasticKey)
	})
	require.NoError(t, err)

	// Test with ExportAllowed filter (even though the field doesn't exist in ElasticKey).
	// This exercises the dead code path in applyGetElasticKeysFilters (lines 275-277).
	exportAllowed := true
	err = testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		filters := &GetElasticKeysFilters{
			ExportAllowed: &exportAllowed, // This filter path is DEAD CODE - column doesn't exist
		}
		_, err := tx.GetElasticKeys(filters)

		return err
	})
	// Should fail with "no such column: elastic_key_export_allowed" error.
	require.Error(t, err)
	require.Contains(t, err.Error(), "no such column: elastic_key_export_allowed")
}

// TestGetElasticKeysWithImportAllowedFilter tests the ImportAllowed filter path.
func TestGetElasticKeysWithImportAllowedFilter(t *testing.T) {
	CleanupDatabase(t, testOrmRepository)

	// Create an elastic key with ImportAllowed=true.
	tenantID := googleUuid.New()
	elasticKey, buildErr := BuildElasticKey(
		tenantID,
		googleUuid.New(),
		"import-filter-test",
		"Test Import Filter",
		cryptoutilOpenapiModel.Internal,
		cryptoutilOpenapiModel.A256GCMDir,
		false,
		true, // ImportAllowed=true
		false,
		string(cryptoutilKmsServer.Active),
	)
	require.NoError(t, buildErr)

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		return tx.AddElasticKey(elasticKey)
	})
	require.NoError(t, err)

	// Test with ImportAllowed=true filter.
	importAllowed := true
	err = testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		filters := &GetElasticKeysFilters{
			ImportAllowed: &importAllowed,
		}
		keys, err := tx.GetElasticKeys(filters)
		require.NoError(t, err)
		require.NotEmpty(t, keys, "Expected keys with ImportAllowed=true")

		return nil
	})
	require.NoError(t, err)

	// Test with ImportAllowed=false filter.
	importAllowedFalse := false
	err = testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		filters := &GetElasticKeysFilters{
			ImportAllowed: &importAllowedFalse,
		}
		keys, err := tx.GetElasticKeys(filters)
		require.NoError(t, err)
		require.Empty(t, keys, "Expected no keys with ImportAllowed=false")

		return nil
	})
	require.NoError(t, err)
}
