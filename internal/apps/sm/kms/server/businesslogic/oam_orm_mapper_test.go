// Copyright (c) 2025 Justin Cranford

package businesslogic

import (
	"testing"
	"time"

	cryptoutilKmsServer "cryptoutil/api/kms/server"
	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilOrmRepository "cryptoutil/internal/apps/sm/kms/server/repository/orm"

	googleUuid "github.com/google/uuid"
	testify "github.com/stretchr/testify/require"
)

const testDescription = "test description"

func TestNewOamOrmMapper(t *testing.T) {
	mapper := NewOamOrmMapper()
	testify.NotNil(t, mapper, "mapper should not be nil")
}

func TestToOrmAddElasticKey(t *testing.T) {
	mapper := NewOamOrmMapper()
	elasticKeyID := googleUuid.New()
	tenantID := googleUuid.New()

	description := testDescription
	versioningAllowed := true
	importAllowed := false

	create := &cryptoutilKmsServer.ElasticKeyCreate{
		Name:              "test-key",
		Description:       &description,
		Provider:          string(cryptoutilOpenapiModel.Internal),
		Algorithm:         string(cryptoutilOpenapiModel.A128CBCHS256Dir),
		VersioningAllowed: &versioningAllowed,
		ImportAllowed:     &importAllowed,
	}

	result := mapper.toOrmAddElasticKey(&elasticKeyID, tenantID, create)

	testify.Equal(t, elasticKeyID, result.ElasticKeyID)
	testify.Equal(t, tenantID, result.TenantID)
	testify.Equal(t, "test-key", result.ElasticKeyName)
	testify.Equal(t, testDescription, result.ElasticKeyDescription)
	testify.Equal(t, cryptoutilOpenapiModel.Internal, result.ElasticKeyProvider)
	testify.Equal(t, cryptoutilOpenapiModel.A128CBCHS256Dir, result.ElasticKeyAlgorithm)
	testify.Equal(t, versioningAllowed, result.ElasticKeyVersioningAllowed)
	testify.Equal(t, importAllowed, result.ElasticKeyImportAllowed)
	testify.Equal(t, cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingGenerate), result.ElasticKeyStatus)
}

func TestToOrmAddMaterialKey(t *testing.T) {
	mapper := NewOamOrmMapper()
	elasticKeyID := googleUuid.New()
	materialKeyID := googleUuid.New()
	publicBytes := []byte("public-key-data")
	encryptedBytes := []byte("encrypted-private-key-data")
	generateDate := time.Now().UTC()
	generateDateMillis := generateDate.UnixMilli()

	result := mapper.toOrmAddMaterialKey(&elasticKeyID, &materialKeyID, publicBytes, encryptedBytes, generateDate)

	testify.Equal(t, elasticKeyID, result.ElasticKeyID)
	testify.Equal(t, materialKeyID, result.MaterialKeyID)
	testify.Equal(t, publicBytes, result.MaterialKeyClearPublic)
	testify.Equal(t, encryptedBytes, result.MaterialKeyEncryptedNonPublic)
	testify.NotNil(t, result.MaterialKeyGenerateDate)
	testify.Equal(t, generateDateMillis, *result.MaterialKeyGenerateDate)
}

func TestToOamElasticKeyStatus(t *testing.T) {
	tests := []struct {
		name           string
		importAllowed  bool
		expectedStatus cryptoutilKmsServer.ElasticKeyStatus
	}{
		{"import allowed", true, cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingImport)},
		{"import not allowed", false, cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingGenerate)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := toOamElasticKeyStatus(&tc.importAllowed)
			testify.NotNil(t, result)
			testify.Equal(t, tc.expectedStatus, *result)
		})
	}
}

func TestToOamElasticKey(t *testing.T) {
	mapper := NewOamOrmMapper()
	elasticKeyID := googleUuid.New()

	ormElasticKey := &cryptoutilOrmRepository.ElasticKey{
		ElasticKeyID:                elasticKeyID,
		ElasticKeyName:              "test-key",
		ElasticKeyDescription:       testDescription,
		ElasticKeyProvider:          cryptoutilOpenapiModel.Internal,
		ElasticKeyAlgorithm:         cryptoutilOpenapiModel.A128CBCHS256Dir,
		ElasticKeyVersioningAllowed: true,
		ElasticKeyImportAllowed:     false,
		ElasticKeyStatus:            cryptoutilKmsServer.Active,
	}

	result := mapper.toOamElasticKey(ormElasticKey)

	testify.NotNil(t, result.ElasticKeyID)
	testify.Equal(t, elasticKeyID, *result.ElasticKeyID)
	testify.Equal(t, "test-key", *result.Name)
	testify.Equal(t, testDescription, *result.Description)
	testify.Equal(t, string(cryptoutilOpenapiModel.Internal), *result.Provider)
	testify.Equal(t, string(cryptoutilOpenapiModel.A128CBCHS256Dir), *result.Algorithm)
	testify.Equal(t, true, *result.VersioningAllowed)
	testify.Equal(t, false, *result.ImportAllowed)
	testify.Equal(t, cryptoutilKmsServer.Active, *result.Status)
}

func TestToOamElasticKeys(t *testing.T) {
	mapper := NewOamOrmMapper()
	id1 := googleUuid.New()
	id2 := googleUuid.New()

	ormElasticKeys := []cryptoutilOrmRepository.ElasticKey{
		{
			ElasticKeyID:                id1,
			ElasticKeyName:              "key1",
			ElasticKeyDescription:       "desc1",
			ElasticKeyProvider:          cryptoutilOpenapiModel.Internal,
			ElasticKeyAlgorithm:         cryptoutilOpenapiModel.A128CBCHS256Dir,
			ElasticKeyVersioningAllowed: true,
			ElasticKeyImportAllowed:     false,
			ElasticKeyStatus:            cryptoutilKmsServer.Active,
		},
		{
			ElasticKeyID:                id2,
			ElasticKeyName:              "key2",
			ElasticKeyDescription:       "desc2",
			ElasticKeyProvider:          cryptoutilOpenapiModel.Internal,
			ElasticKeyAlgorithm:         cryptoutilOpenapiModel.A128GCMDir,
			ElasticKeyVersioningAllowed: false,
			ElasticKeyImportAllowed:     true,
			ElasticKeyStatus:            cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingImport),
		},
	}

	results := mapper.toOamElasticKeys(ormElasticKeys)

	testify.Len(t, results, 2)
	testify.Equal(t, id1, *results[0].ElasticKeyID)
	testify.Equal(t, "key1", *results[0].Name)
	testify.Equal(t, id2, *results[1].ElasticKeyID)
	testify.Equal(t, "key2", *results[1].Name)
}

func TestToOamMaterialKey(t *testing.T) {
	mapper := NewOamOrmMapper()

	elasticKeyID := googleUuid.New()
	materialKeyID := googleUuid.New()
	generateDateMillis := time.Now().UTC().UnixMilli()
	importDateMillis := time.Now().UTC().Add(-time.Hour).UnixMilli()
	expirationDateMillis := time.Now().UTC().Add(time.Hour).UnixMilli()
	revocationDateMillis := time.Now().UTC().Add(-30 * time.Minute).UnixMilli()
	publicBytes := []byte(`{"kty":"RSA"}`)

	tests := []struct {
		name          string
		ormKey        *cryptoutilOrmRepository.MaterialKey
		expectError   bool
		errorContains string
	}{
		{
			name: "valid material key with public",
			ormKey: &cryptoutilOrmRepository.MaterialKey{
				ElasticKeyID:                  elasticKeyID,
				MaterialKeyID:                 materialKeyID,
				MaterialKeyClearPublic:        publicBytes,
				MaterialKeyEncryptedNonPublic: []byte("encrypted"),
				MaterialKeyGenerateDate:       &generateDateMillis,
			},
			expectError: false,
		},
		{
			name: "valid material key without public",
			ormKey: &cryptoutilOrmRepository.MaterialKey{
				ElasticKeyID:                  elasticKeyID,
				MaterialKeyID:                 materialKeyID,
				MaterialKeyClearPublic:        nil,
				MaterialKeyEncryptedNonPublic: []byte("encrypted"),
				MaterialKeyGenerateDate:       &generateDateMillis,
			},
			expectError: false,
		},
		{
			name:          "nil material key",
			ormKey:        nil,
			expectError:   true,
			errorContains: "material key cannot be nil",
		},
		{
			name: "missing elastic key ID",
			ormKey: &cryptoutilOrmRepository.MaterialKey{
				ElasticKeyID:  googleUuid.UUID{},
				MaterialKeyID: materialKeyID,
			},
			expectError:   true,
			errorContains: "material key missing required elastic key ID",
		},
		{
			name: "missing material key ID",
			ormKey: &cryptoutilOrmRepository.MaterialKey{
				ElasticKeyID:  elasticKeyID,
				MaterialKeyID: googleUuid.UUID{},
			},
			expectError:   true,
			errorContains: "material key missing required material key ID",
		},
		{
			name: "valid material key with all dates",
			ormKey: &cryptoutilOrmRepository.MaterialKey{
				ElasticKeyID:                  elasticKeyID,
				MaterialKeyID:                 materialKeyID,
				MaterialKeyClearPublic:        publicBytes,
				MaterialKeyEncryptedNonPublic: []byte("encrypted"),
				MaterialKeyGenerateDate:       &generateDateMillis,
				MaterialKeyImportDate:         &importDateMillis,
				MaterialKeyExpirationDate:     &expirationDateMillis,
				MaterialKeyRevocationDate:     &revocationDateMillis,
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := mapper.toOamMaterialKey(tc.ormKey)

			if tc.expectError {
				testify.Error(t, err)
				testify.Contains(t, err.Error(), tc.errorContains)
				testify.Nil(t, result)
			} else {
				testify.NoError(t, err)
				testify.NotNil(t, result)
				testify.Equal(t, elasticKeyID, *result.ElasticKeyID)
				testify.Equal(t, materialKeyID, *result.MaterialKeyID)

				if tc.ormKey.MaterialKeyClearPublic != nil {
					testify.NotNil(t, result.ClearPublic)
					testify.Equal(t, string(publicBytes), *result.ClearPublic)
				} else {
					testify.Nil(t, result.ClearPublic)
				}
			}
		})
	}
}

func TestToOamMaterialKeys(t *testing.T) {
	mapper := NewOamOrmMapper()

	elasticKeyID := googleUuid.New()
	materialKeyID1 := googleUuid.New()
	materialKeyID2 := googleUuid.New()
	generateDateMillis := time.Now().UTC().UnixMilli()

	tests := []struct {
		name        string
		ormKeys     []cryptoutilOrmRepository.MaterialKey
		expectError bool
	}{
		{
			name: "valid material keys",
			ormKeys: []cryptoutilOrmRepository.MaterialKey{
				{
					ElasticKeyID:                  elasticKeyID,
					MaterialKeyID:                 materialKeyID1,
					MaterialKeyClearPublic:        []byte("public1"),
					MaterialKeyEncryptedNonPublic: []byte("encrypted1"),
					MaterialKeyGenerateDate:       &generateDateMillis,
				},
				{
					ElasticKeyID:                  elasticKeyID,
					MaterialKeyID:                 materialKeyID2,
					MaterialKeyClearPublic:        []byte("public2"),
					MaterialKeyEncryptedNonPublic: []byte("encrypted2"),
					MaterialKeyGenerateDate:       &generateDateMillis,
				},
			},
			expectError: false,
		},
		{
			name: "invalid material key in slice",
			ormKeys: []cryptoutilOrmRepository.MaterialKey{
				{
					ElasticKeyID:  googleUuid.UUID{},
					MaterialKeyID: materialKeyID1,
				},
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			results, err := mapper.toOamMaterialKeys(tc.ormKeys)

			if tc.expectError {
				testify.Error(t, err)
				testify.Nil(t, results)
			} else {
				testify.NoError(t, err)
				testify.Len(t, results, len(tc.ormKeys))
			}
		})
	}
}

func TestToOptionalOrmUUIDs(t *testing.T) {
	mapper := NewOamOrmMapper()

	validUUID1 := googleUuid.New()
	validUUID2 := googleUuid.New()
	validUUIDs := []googleUuid.UUID{validUUID1, validUUID2}
	emptyUUIDs := []googleUuid.UUID{}

	tests := []struct {
		name        string
		input       *[]googleUuid.UUID
		expectError bool
		expectNil   bool
	}{
		{"nil input", nil, false, true},
		{"empty slice", &emptyUUIDs, false, true},
		{"valid UUIDs", &validUUIDs, false, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := mapper.toOptionalOrmUUIDs(tc.input)

			if tc.expectError {
				testify.Error(t, err)
			} else {
				testify.NoError(t, err)

				if tc.expectNil {
					testify.Nil(t, result)
				} else {
					testify.NotNil(t, result)
					testify.Equal(t, *tc.input, result)
				}
			}
		})
	}
}

func TestToOptionalOrmStrings(t *testing.T) {
	mapper := NewOamOrmMapper()

	validStrings := []string{"value1", "value2"}
	emptyStrings := []string{}
	stringsWithEmpty := []string{"valid", ""}

	tests := []struct {
		name          string
		input         *[]string
		expectError   bool
		expectNil     bool
		errorContains string
	}{
		{"nil input", nil, false, true, ""},
		{"empty slice", &emptyStrings, false, true, ""},
		{"valid strings", &validStrings, false, false, ""},
		{"strings with empty value", &stringsWithEmpty, true, false, "value must not be empty string"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := mapper.toOptionalOrmStrings(tc.input)

			if tc.expectError {
				testify.Error(t, err)
				testify.Contains(t, err.Error(), tc.errorContains)
			} else {
				testify.NoError(t, err)

				if tc.expectNil {
					testify.Nil(t, result)
				} else {
					testify.NotNil(t, result)
					testify.Equal(t, *tc.input, result)
				}
			}
		})
	}
}

func TestToOrmDateRange(t *testing.T) {
	mapper := NewOamOrmMapper()

	now := time.Now().UTC()
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)
	farPast := now.Add(-48 * time.Hour)

	tests := []struct {
		name          string
		minDate       *time.Time
		maxDate       *time.Time
		expectError   bool
		errorContains string
	}{
		{"both nil", nil, nil, false, ""},
		{"valid past range", &farPast, &past, false, ""},
		{"min in future", &future, nil, true, "min date can't be in the future"},
		{"min after max", &past, &farPast, true, "min date must be before max date"},
		{"min equal max", &past, &past, false, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resultMin, resultMax, err := mapper.toOrmDateRange(tc.minDate, tc.maxDate)

			if tc.expectError {
				testify.Error(t, err)
				testify.Contains(t, err.Error(), tc.errorContains)
			} else {
				testify.NoError(t, err)
				testify.Equal(t, tc.minDate, resultMin)
				testify.Equal(t, tc.maxDate, resultMax)
			}
		})
	}
}

func TestToOrmPageNumber(t *testing.T) {
	mapper := NewOamOrmMapper()

	zero := cryptoutilKmsServer.PageNumber(0)
	positive := cryptoutilKmsServer.PageNumber(5)
	negative := cryptoutilKmsServer.PageNumber(-1)

	tests := []struct {
		name        string
		input       *cryptoutilKmsServer.PageNumber
		expected    int
		expectError bool
	}{
		{"nil returns default", nil, 0, false},
		{"zero page number", &zero, 0, false},
		{"positive page number", &positive, 5, false},
		{"negative page number", &negative, 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := mapper.toOrmPageNumber(tc.input)

			if tc.expectError {
				testify.Error(t, err)
				testify.Contains(t, err.Error(), "page number must be zero or higher")
			} else {
				testify.NoError(t, err)
				testify.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestToOrmPageSize(t *testing.T) {
	mapper := NewOamOrmMapper()

	one := cryptoutilKmsServer.PageSize(1)
	ten := cryptoutilKmsServer.PageSize(10)
	zero := cryptoutilKmsServer.PageSize(0)

	tests := []struct {
		name        string
		input       *cryptoutilKmsServer.PageSize
		expectError bool
		minValue    int
	}{
		{"nil returns default", nil, false, 1},
		{"size of one", &one, false, 1},
		{"size of ten", &ten, false, 10},
		{"zero size", &zero, true, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := mapper.toOrmPageSize(tc.input)

			if tc.expectError {
				testify.Error(t, err)
				testify.Contains(t, err.Error(), "page size must be one or higher")
			} else {
				testify.NoError(t, err)
				testify.GreaterOrEqual(t, result, tc.minValue)
			}
		})
	}
}
