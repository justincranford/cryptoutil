// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/model"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestBuildElasticKey tests the BuildElasticKey builder function.
func TestBuildElasticKey(t *testing.T) {
	t.Run("Build elastic key with import allowed (pending import status)", func(t *testing.T) {
		tenantID := googleUuid.New()
		ekID := googleUuid.New()
		elasticKey, err := BuildElasticKey(
			tenantID,
			ekID,
			"test-key",
			"Test Description",
			cryptoutilOpenapiModel.Internal,
			cryptoutilOpenapiModel.A256GCMA256KW,
			true,  // versioningAllowed
			true,  // importAllowed
			false, // exportAllowed
			"active",
		)

		require.NoError(t, err, "BuildElasticKey should succeed")
		require.NotNil(t, elasticKey, "ElasticKey should not be nil")
		require.Equal(t, ekID, elasticKey.ElasticKeyID, "Elastic Key ID should match")
		require.Equal(t, "test-key", elasticKey.ElasticKeyName, "Elastic Key Name should match")
		require.Equal(t, "Test Description", elasticKey.ElasticKeyDescription, "Description should match")
		require.Equal(t, cryptoutilOpenapiModel.Internal, elasticKey.ElasticKeyProvider, "Provider should match")
		require.Equal(t, cryptoutilOpenapiModel.A256GCMA256KW, elasticKey.ElasticKeyAlgorithm, "Algorithm should match")
		require.True(t, elasticKey.ElasticKeyVersioningAllowed, "Versioning should be allowed")
		require.True(t, elasticKey.ElasticKeyImportAllowed, "Import should be allowed")
		require.Equal(t, cryptoutilOpenapiModel.PendingImport, elasticKey.ElasticKeyStatus, "Status should be PendingImport when import allowed")
	})

	t.Run("Build elastic key with import not allowed (pending generate status)", func(t *testing.T) {
		tenantID := googleUuid.New()
		ekID := googleUuid.New()
		elasticKey, err := BuildElasticKey(
			tenantID,
			ekID,
			"gen-key",
			"Generated Key",
			cryptoutilOpenapiModel.Internal,
			cryptoutilOpenapiModel.A128GCMA128KW,
			false, // versioningAllowed
			false, // importAllowed
			false, // exportAllowed
			"active",
		)

		require.NoError(t, err, "BuildElasticKey should succeed")
		require.NotNil(t, elasticKey, "ElasticKey should not be nil")
		require.Equal(t, ekID, elasticKey.ElasticKeyID, "Elastic Key ID should match")
		require.Equal(t, "gen-key", elasticKey.ElasticKeyName, "Elastic Key Name should match")
		require.Equal(t, "Generated Key", elasticKey.ElasticKeyDescription, "Description should match")
		require.Equal(t, cryptoutilOpenapiModel.Internal, elasticKey.ElasticKeyProvider, "Provider should match")
		require.Equal(t, cryptoutilOpenapiModel.A128GCMA128KW, elasticKey.ElasticKeyAlgorithm, "Algorithm should match")
		require.False(t, elasticKey.ElasticKeyVersioningAllowed, "Versioning should not be allowed")
		require.False(t, elasticKey.ElasticKeyImportAllowed, "Import should not be allowed")
		require.Equal(t, cryptoutilOpenapiModel.PendingGenerate, elasticKey.ElasticKeyStatus, "Status should be PendingGenerate when import not allowed")
	})

	t.Run("Build elastic key with various algorithm types", func(t *testing.T) {
		algorithms := []cryptoutilOpenapiModel.ElasticKeyAlgorithm{
			cryptoutilOpenapiModel.A256GCMA256KW,
			cryptoutilOpenapiModel.A192GCMA192KW,
			cryptoutilOpenapiModel.A128GCMA128KW,
			cryptoutilOpenapiModel.A128CBCHS256A128KW,
			cryptoutilOpenapiModel.A192CBCHS384A192KW,
			cryptoutilOpenapiModel.A256CBCHS512A256KW,
		}

		for _, algo := range algorithms {
			tenantID := googleUuid.New()
			ekID := googleUuid.New()
			elasticKey, err := BuildElasticKey(
				tenantID,
				ekID,
				"algo-key",
				"Algorithm Test",
				cryptoutilOpenapiModel.Internal,
				algo,
				true,
				false,
				false,
				"active",
			)

			require.NoError(t, err, "BuildElasticKey should succeed for algorithm %s", algo)
			require.NotNil(t, elasticKey, "ElasticKey should not be nil for algorithm %s", algo)
			require.Equal(t, algo, elasticKey.ElasticKeyAlgorithm, "Algorithm should match for %s", algo)
		}
	})
}

// TestElasticKeyStatusInitial tests the ElasticKeyStatusInitial helper function.
func TestElasticKeyStatusInitial(t *testing.T) {
	t.Run("Import allowed returns PendingImport", func(t *testing.T) {
		status := ElasticKeyStatusInitial(true)
		require.Equal(t, cryptoutilOpenapiModel.PendingImport, status, "Status should be PendingImport when import allowed")
	})

	t.Run("Import not allowed returns PendingGenerate", func(t *testing.T) {
		status := ElasticKeyStatusInitial(false)
		require.Equal(t, cryptoutilOpenapiModel.PendingGenerate, status, "Status should be PendingGenerate when import not allowed")
	})
}
