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

// TestGetElasticKeysWithVersioningAllowedFilter tests the VersioningAllowed filter path.
func TestGetElasticKeysWithVersioningAllowedFilter(t *testing.T) {
	t.Parallel()
	CleanupDatabase(t, testOrmRepository)

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
		if err := tx.AddElasticKey(keyWithVersioning); err != nil {
			return err
		}

		return tx.AddElasticKey(keyWithoutVersioning)
	})
	require.NoError(t, err)

	// Test with VersioningAllowed=true filter.
	versioningAllowedTrue := true
	err = testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		filters := &GetElasticKeysFilters{
			VersioningAllowed: &versioningAllowedTrue,
		}
		keys, err := tx.GetElasticKeys(filters)
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
			VersioningAllowed: &versioningAllowedFalse,
		}
		keys, err := tx.GetElasticKeys(filters)
		require.NoError(t, err)
		require.Len(t, keys, 1, "Expected 1 key with VersioningAllowed=false")
		require.Equal(t, keyWithoutVersioning.ElasticKeyID, keys[0].ElasticKeyID)

		return nil
	})
	require.NoError(t, err)
}
