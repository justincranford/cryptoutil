// Copyright (c) 2025 Justin Cranford

package businesslogic

import (
	"testing"

	cryptoutilKmsServer "cryptoutil/api/kms/server"
	cryptoutilOpenapiModel "cryptoutil/api/model"

	googleUuid "github.com/google/uuid"
	testify "github.com/stretchr/testify/require"
)

func TestToOrmAlgorithms(t *testing.T) {
	t.Parallel()

	mapper := NewOamOrmMapper()

	validAlgorithms := []cryptoutilOpenapiModel.ElasticKeyAlgorithm{
		cryptoutilOpenapiModel.A128CBCHS256Dir,
		cryptoutilOpenapiModel.A256GCMDir,
	}
	emptyAlgorithm := cryptoutilOpenapiModel.ElasticKeyAlgorithm("")
	algorithmsWithEmpty := []cryptoutilOpenapiModel.ElasticKeyAlgorithm{
		cryptoutilOpenapiModel.A128CBCHS256Dir,
		emptyAlgorithm,
	}

	tests := []struct {
		name          string
		input         *[]cryptoutilOpenapiModel.ElasticKeyAlgorithm
		expectError   bool
		expectNil     bool
		errorContains string
	}{
		{"nil input", nil, false, true, ""},
		{"valid algorithms", &validAlgorithms, false, false, ""},
		{"algorithms with empty", &algorithmsWithEmpty, true, false, "algorithm cannot be empty"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := mapper.toOrmAlgorithms(tc.input)

			if tc.expectError {
				testify.Error(t, err)
				testify.Contains(t, err.Error(), tc.errorContains)
			} else {
				testify.NoError(t, err)

				if tc.expectNil {
					testify.Nil(t, result)
				} else {
					testify.NotNil(t, result)
					testify.Len(t, result, len(*tc.input))
				}
			}
		})
	}
}

func TestToOrmElasticKeySorts(t *testing.T) {
	t.Parallel()

	mapper := NewOamOrmMapper()

	validSorts := []cryptoutilOpenapiModel.ElasticKeySort{
		cryptoutilOpenapiModel.ElasticKeySortElasticKeyIDASC,
		cryptoutilOpenapiModel.ElasticKeySortNameASC,
	}
	emptySort := cryptoutilOpenapiModel.ElasticKeySort("")
	sortsWithEmpty := []cryptoutilOpenapiModel.ElasticKeySort{
		cryptoutilOpenapiModel.ElasticKeySortNameASC,
		emptySort,
	}

	tests := []struct {
		name          string
		input         *[]cryptoutilOpenapiModel.ElasticKeySort
		expectError   bool
		expectNil     bool
		errorContains string
	}{
		{"nil input", nil, false, true, ""},
		{"valid sorts", &validSorts, false, false, ""},
		{"sorts with empty", &sortsWithEmpty, true, false, "elastic key sort cannot be empty"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := mapper.toOrmElasticKeySorts(tc.input)

			if tc.expectError {
				testify.Error(t, err)
				testify.Contains(t, err.Error(), tc.errorContains)
			} else {
				testify.NoError(t, err)

				if tc.expectNil {
					testify.Nil(t, result)
				} else {
					testify.NotNil(t, result)
					testify.Len(t, result, len(*tc.input))
				}
			}
		})
	}
}

func TestToOrmMaterialKeySorts(t *testing.T) {
	t.Parallel()

	mapper := NewOamOrmMapper()

	validSorts := []cryptoutilOpenapiModel.MaterialKeySort{
		cryptoutilOpenapiModel.MaterialKeySort("elastic_key_id"),
		cryptoutilOpenapiModel.MaterialKeySort("elastic_key_id:ASC"),
		cryptoutilOpenapiModel.MaterialKeySort("elastic_key_id:DESC"),
		cryptoutilOpenapiModel.MaterialKeySort("material_key_id"),
		cryptoutilOpenapiModel.MaterialKeySort("generate_date"),
		cryptoutilOpenapiModel.MaterialKeySort("import_date"),
		cryptoutilOpenapiModel.MaterialKeySort("expiration_date"),
		cryptoutilOpenapiModel.MaterialKeySort("revocation_date"),
	}
	emptySort := cryptoutilOpenapiModel.MaterialKeySort("")
	sortsWithEmpty := []cryptoutilOpenapiModel.MaterialKeySort{
		cryptoutilOpenapiModel.MaterialKeySort("elastic_key_id"),
		emptySort,
	}
	invalidSort := cryptoutilOpenapiModel.MaterialKeySort("invalid_field")
	sortsWithInvalid := []cryptoutilOpenapiModel.MaterialKeySort{
		invalidSort,
	}

	tests := []struct {
		name          string
		input         *[]cryptoutilOpenapiModel.MaterialKeySort
		expectError   bool
		expectNil     bool
		errorContains string
	}{
		{"nil input", nil, false, true, ""},
		{"valid sorts", &validSorts, false, false, ""},
		{"sorts with empty", &sortsWithEmpty, true, false, "material key sort cannot be empty"},
		{"sorts with invalid", &sortsWithInvalid, true, false, "invalid material key sort value"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := mapper.toOrmMaterialKeySorts(tc.input)

			if tc.expectError {
				testify.Error(t, err)
				testify.Contains(t, err.Error(), tc.errorContains)
			} else {
				testify.NoError(t, err)

				if tc.expectNil {
					testify.Nil(t, result)
				} else {
					testify.NotNil(t, result)
					testify.Len(t, result, len(*tc.input))
				}
			}
		})
	}
}

func TestToElasticKeyStatusFromImportAllowed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		importAllowed  bool
		expectedStatus cryptoutilKmsServer.ElasticKeyStatus
	}{
		{"import allowed returns pending import", true, cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingImport)},
		{"import not allowed returns pending generate", false, cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingGenerate)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := toElasticKeyStatusFromImportAllowed(tc.importAllowed)
			testify.Equal(t, tc.expectedStatus, result)
		})
	}
}

func TestToOrmAddElasticKeyDefaults(t *testing.T) {
	t.Parallel()

	mapper := NewOamOrmMapper()
	elasticKeyID := googleUuid.New()
	tenantID := googleUuid.New()

	// Test with minimal input - all optional fields nil.
	descStr := testDescription
	create := &cryptoutilKmsServer.ElasticKeyCreate{
		Name:        "test-key",
		Description: &descStr,
	}

	result := mapper.toOrmAddElasticKey(&elasticKeyID, tenantID, create)

	// Verify defaults are applied.
	testify.Equal(t, elasticKeyID, result.ElasticKeyID)
	testify.Equal(t, tenantID, result.TenantID)
	testify.Equal(t, "test-key", result.ElasticKeyName)
	testify.Equal(t, cryptoutilOpenapiModel.Internal, result.ElasticKeyProvider)
	testify.Equal(t, cryptoutilOpenapiModel.A256GCMA256KW, result.ElasticKeyAlgorithm)
	testify.True(t, result.ElasticKeyVersioningAllowed)
	testify.False(t, result.ElasticKeyImportAllowed)
	testify.Equal(t, cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingGenerate), result.ElasticKeyStatus)
}

func TestToOrmAddElasticKeyImportAllowed(t *testing.T) {
	t.Parallel()

	mapper := NewOamOrmMapper()
	elasticKeyID := googleUuid.New()
	tenantID := googleUuid.New()

	// Test with import allowed = true.
	importAllowed := true
	descStr2 := testDescription
	create := &cryptoutilKmsServer.ElasticKeyCreate{
		Name:          "test-key",
		Description:   &descStr2,
		ImportAllowed: &importAllowed,
	}

	result := mapper.toOrmAddElasticKey(&elasticKeyID, tenantID, create)

	testify.True(t, result.ElasticKeyImportAllowed)
	testify.Equal(t, cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingImport), result.ElasticKeyStatus)
}

func TestToStrings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		items     *[]string
		expectNil bool
	}{
		{"nil input", nil, true},
		{"empty slice", &[]string{}, true},
		{"valid strings", &[]string{"a", "b", "c"}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := toStrings(tc.items, func(s string) string { return s })

			if tc.expectNil {
				testify.Nil(t, result)
			} else {
				testify.NotNil(t, result)
				testify.Len(t, result, len(*tc.items))
			}
		})
	}
}
