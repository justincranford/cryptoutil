//go:build integration
// +build integration

// Copyright (c) 2025-2026 Justin Cranford.
package orm

import (
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/sm-kms/models"
	cryptoutilKmsServer "cryptoutil/api/sm-kms/server"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Sequential: uses shared package-level SQLite fixture state via CleanupDatabase.
// TestGetElasticKeysWithVersioningAllowedFilter tests the VersioningAllowed filter path.
func TestGetElasticKeysWithVersioningAllowedFilter(t *testing.T) {
	CleanupDatabase(t, testOrmRepository, KMSCleanupTables)

	// Create two elastic keys: one with versioning allowed, one without.
	tenantID := googleUuid.New()
	keyWithVersioning, buildErr := BuildElasticKey(
		tenantID,
		googleUuid.New(),
		"versioning-allowed-test",
		"Test Versioning Allowed",
		cryptoutilOpenapiModel.Internal,
		cryptoutilOpenapiModel.A256GCMDir,
		true, // VersioningAllowed=true
		false,
		false,
		string(cryptoutilKmsServer.Active),
	)
	require.NoError(t, buildErr)

	keyWithoutVersioning, buildErr := BuildElasticKey(
		tenantID,
		googleUuid.New(),
		"versioning-not-allowed-test",
		"Test Versioning Not Allowed",
		cryptoutilOpenapiModel.Internal,
		cryptoutilOpenapiModel.A256GCMDir,
		false, // VersioningAllowed=false
		false,
		false,
		string(cryptoutilKmsServer.Active),
	)
	require.NoError(t, buildErr)

	// Insert both keys.
	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		if err := AddElasticKey(tx.GormTx(), testTelemetryService.Slogger, keyWithVersioning); err != nil {
			return err
		}

		return AddElasticKey(tx.GormTx(), testTelemetryService.Slogger, keyWithoutVersioning)
	})
	require.NoError(t, err)

	// Test with VersioningAllowed=true filter.
	versioningAllowedTrue := true
	err = testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		filters := &GetElasticKeysFilters{
			TenantID:          tenantID,
			VersioningAllowed: &versioningAllowedTrue,
		}
		keys, err := GetElasticKeys(tx.GormTx(), testTelemetryService.Slogger, filters)
		require.NoError(t, err)
		require.Len(t, keys, 1, "Expected 1 key with VersioningAllowed=true")
		require.Equal(t, keyWithVersioning.ElasticKeyID, keys[0].ElasticKeyID)

		return nil
	})
	require.NoError(t, err)

	// Test with VersioningAllowed=false filter.
	versioningAllowedFalse := false
	err = testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		filters := &GetElasticKeysFilters{
			TenantID:          tenantID,
			VersioningAllowed: &versioningAllowedFalse,
		}
		keys, err := GetElasticKeys(tx.GormTx(), testTelemetryService.Slogger, filters)
		require.NoError(t, err)
		require.Len(t, keys, 1, "Expected 1 key with VersioningAllowed=false")
		require.Equal(t, keyWithoutVersioning.ElasticKeyID, keys[0].ElasticKeyID)

		return nil
	})
	require.NoError(t, err)
}
