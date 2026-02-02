// Copyright (c) 2025 Justin Cranford

package businesslogic

import (
	"testing"
	"time"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"

	googleUuid "github.com/google/uuid"
	testify "github.com/stretchr/testify/require"
)

func TestNewOamOrmMapper(t *testing.T) {
	mapper := NewOamOrmMapper()
	testify.NotNil(t, mapper, "mapper should not be nil")
}

func TestToOrmAddElasticKey(t *testing.T) {
	mapper := NewOamOrmMapper()
	elasticKeyID := googleUuid.New()
	tenantID := googleUuid.New()

	provider := cryptoutilOpenapiModel.Internal
	algorithm := cryptoutilOpenapiModel.A128CBCHS256Dir
	versioningAllowed := true
	importAllowed := false

	create := &cryptoutilOpenapiModel.ElasticKeyCreate{
		Name:              "test-key",
		Description:       "test description",
		Provider:          &provider,
		Algorithm:         &algorithm,
		VersioningAllowed: &versioningAllowed,
		ImportAllowed:     &importAllowed,
	}

	result := mapper.toOrmAddElasticKey(&elasticKeyID, tenantID, create)

	testify.Equal(t, elasticKeyID, result.ElasticKeyID)
	testify.Equal(t, tenantID, result.TenantID)
	testify.Equal(t, "test-key", result.ElasticKeyName)
	testify.Equal(t, "test description", result.ElasticKeyDescription)
	testify.Equal(t, provider, result.ElasticKeyProvider)
	testify.Equal(t, algorithm, result.ElasticKeyAlgorithm)
	testify.Equal(t, versioningAllowed, result.ElasticKeyVersioningAllowed)
	testify.Equal(t, importAllowed, result.ElasticKeyImportAllowed)
	testify.Equal(t, cryptoutilOpenapiModel.PendingGenerate, result.ElasticKeyStatus)
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
		expectedStatus cryptoutilOpenapiModel.ElasticKeyStatus
	}{
		{"import allowed", true, cryptoutilOpenapiModel.PendingImport},
		{"import not allowed", false, cryptoutilOpenapiModel.PendingGenerate},
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
		ElasticKeyDescription:       "test description",
		ElasticKeyProvider:          cryptoutilOpenapiModel.Internal,
		ElasticKeyAlgorithm:         cryptoutilOpenapiModel.A128CBCHS256Dir,
		ElasticKeyVersioningAllowed: true,
		ElasticKeyImportAllowed:     false,
		ElasticKeyStatus:            cryptoutilOpenapiModel.Active,
	}

	result := mapper.toOamElasticKey(ormElasticKey)

	testify.NotNil(t, result.ElasticKeyID)
	testify.Equal(t, elasticKeyID, *result.ElasticKeyID)
	testify.Equal(t, "test-key", *result.Name)
	testify.Equal(t, "test description", *result.Description)
	testify.Equal(t, cryptoutilOpenapiModel.Internal, *result.Provider)
	testify.Equal(t, cryptoutilOpenapiModel.A128CBCHS256Dir, *result.Algorithm)
	testify.Equal(t, true, *result.VersioningAllowed)
	testify.Equal(t, false, *result.ImportAllowed)
	testify.Equal(t, cryptoutilOpenapiModel.Active, *result.Status)
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
			ElasticKeyStatus:            cryptoutilOpenapiModel.Active,
		},
		{
			ElasticKeyID:                id2,
			ElasticKeyName:              "key2",
			ElasticKeyDescription:       "desc2",
			ElasticKeyProvider:          cryptoutilOpenapiModel.Internal,
			ElasticKeyAlgorithm:         cryptoutilOpenapiModel.A128GCMDir,
			ElasticKeyVersioningAllowed: false,
			ElasticKeyImportAllowed:     true,
			ElasticKeyStatus:            cryptoutilOpenapiModel.PendingImport,
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
				testify.Equal(t, elasticKeyID, result.ElasticKeyID)
				testify.Equal(t, materialKeyID, result.MaterialKeyID)

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

	zero := cryptoutilOpenapiModel.PageNumber(0)
	positive := cryptoutilOpenapiModel.PageNumber(5)
	negative := cryptoutilOpenapiModel.PageNumber(-1)

	tests := []struct {
		name        string
		input       *cryptoutilOpenapiModel.PageNumber
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

	one := cryptoutilOpenapiModel.PageSize(1)
	ten := cryptoutilOpenapiModel.PageSize(10)
	zero := cryptoutilOpenapiModel.PageSize(0)

	tests := []struct {
		name        string
		input       *cryptoutilOpenapiModel.PageSize
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

func TestToOrmGetElasticKeysQueryParams(t *testing.T) {
	mapper := NewOamOrmMapper()

	validUUID := googleUuid.New()
	algorithm := cryptoutilOpenapiModel.A128CBCHS256Dir
	name := "test-key"
	versioningAllowed := true
	negativePage := cryptoutilOpenapiModel.PageNumber(-1)
	zeroPageSize := cryptoutilOpenapiModel.PageSize(0)
	emptyAlgorithm := cryptoutilOpenapiModel.ElasticKeyAlgorithm("")
	emptyString := ""

	tests := []struct {
		name          string
		params        *cryptoutilOpenapiModel.ElasticKeysQueryParams
		expectError   bool
		expectNil     bool
		errorContains string
	}{
		{"nil params", nil, false, false, ""},
		{
			"valid params",
			&cryptoutilOpenapiModel.ElasticKeysQueryParams{
				ElasticKeyID:      &[]googleUuid.UUID{validUUID},
				Name:              &[]string{name},
				Algorithm:         &[]cryptoutilOpenapiModel.ElasticKeyAlgorithm{algorithm},
				VersioningAllowed: &versioningAllowed,
			},
			false,
			false,
			"",
		},
		{
			"invalid page number",
			&cryptoutilOpenapiModel.ElasticKeysQueryParams{
				Page: &negativePage,
			},
			true,
			false,
			"Page Number",
		},
		{
			"invalid page size",
			&cryptoutilOpenapiModel.ElasticKeysQueryParams{
				Size: &zeroPageSize,
			},
			true,
			false,
			"Page Size",
		},
		{
			"invalid algorithm",
			&cryptoutilOpenapiModel.ElasticKeysQueryParams{
				Algorithm: &[]cryptoutilOpenapiModel.ElasticKeyAlgorithm{emptyAlgorithm},
			},
			true,
			false,
			"Elastic Key Algorithm",
		},
		{
			"invalid name",
			&cryptoutilOpenapiModel.ElasticKeysQueryParams{
				Name: &[]string{emptyString},
			},
			true,
			false,
			"Elastic Key Name",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tenantID := googleUuid.New()
			result, err := mapper.toOrmGetElasticKeysQueryParams(tenantID, tc.params)

			if tc.expectError {
				testify.Error(t, err)
				testify.Contains(t, err.Error(), tc.errorContains)
			} else {
				testify.NoError(t, err)

				if tc.expectNil {
					testify.Nil(t, result)
				} else {
					testify.NotNil(t, result)
				}
			}
		})
	}
}

func TestToOrmGetMaterialKeysForElasticKeyQueryParams(t *testing.T) {
	mapper := NewOamOrmMapper()

	materialKeyID := googleUuid.New()
	minDate := time.Now().UTC().Add(-24 * time.Hour)
	maxDate := time.Now().UTC()
	futureDate := time.Now().UTC().Add(24 * time.Hour)
	negativePage := cryptoutilOpenapiModel.PageNumber(-1)
	zeroPageSize := cryptoutilOpenapiModel.PageSize(0)
	invalidSort := cryptoutilOpenapiModel.MaterialKeySort("invalid")

	tests := []struct {
		name          string
		params        *cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams
		expectError   bool
		expectNil     bool
		errorContains string
	}{
		{"nil params", nil, false, false, ""},
		{
			"valid params",
			&cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams{
				MaterialKeyID:   &[]googleUuid.UUID{materialKeyID},
				MinGenerateDate: &minDate,
				MaxGenerateDate: &maxDate,
			},
			false,
			false,
			"",
		},
		{
			"invalid page number",
			&cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams{
				Page: &negativePage,
			},
			true,
			false,
			"Page Number",
		},
		{
			"invalid page size",
			&cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams{
				Size: &zeroPageSize,
			},
			true,
			false,
			"Page Size",
		},
		{
			"invalid date range",
			&cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams{
				MinGenerateDate: &futureDate,
			},
			true,
			false,
			"Generate Date range",
		},
		{
			"invalid sort",
			&cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams{
				Sort: &[]cryptoutilOpenapiModel.MaterialKeySort{invalidSort},
			},
			true,
			false,
			"Key Sort",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := mapper.toOrmGetMaterialKeysForElasticKeyQueryParams(tc.params)

			if tc.expectError {
				testify.Error(t, err)
				testify.Contains(t, err.Error(), tc.errorContains)
			} else {
				testify.NoError(t, err)

				if tc.expectNil {
					testify.Nil(t, result)
				} else {
					testify.NotNil(t, result)
				}
			}
		})
	}
}

func TestToOrmGetMaterialKeysQueryParams(t *testing.T) {
	mapper := NewOamOrmMapper()

	elasticKeyID := googleUuid.New()
	materialKeyID := googleUuid.New()
	minDate := time.Now().UTC().Add(-24 * time.Hour)
	maxDate := time.Now().UTC()
	futureDate := time.Now().UTC().Add(24 * time.Hour)
	negativePage := cryptoutilOpenapiModel.PageNumber(-1)
	zeroPageSize := cryptoutilOpenapiModel.PageSize(0)
	invalidSort := cryptoutilOpenapiModel.MaterialKeySort("invalid")

	tests := []struct {
		name          string
		params        *cryptoutilOpenapiModel.MaterialKeysQueryParams
		expectError   bool
		expectNil     bool
		errorContains string
	}{
		{"nil params", nil, false, false, ""},
		{
			"valid params",
			&cryptoutilOpenapiModel.MaterialKeysQueryParams{
				ElasticKeyID:    &[]googleUuid.UUID{elasticKeyID},
				MaterialKeyID:   &[]googleUuid.UUID{materialKeyID},
				MinGenerateDate: &minDate,
				MaxGenerateDate: &maxDate,
			},
			false,
			false,
			"",
		},
		{
			"invalid page number",
			&cryptoutilOpenapiModel.MaterialKeysQueryParams{
				Page: &negativePage,
			},
			true,
			false,
			"Page Number",
		},
		{
			"invalid page size",
			&cryptoutilOpenapiModel.MaterialKeysQueryParams{
				Size: &zeroPageSize,
			},
			true,
			false,
			"Page Size",
		},
		{
			"invalid date range",
			&cryptoutilOpenapiModel.MaterialKeysQueryParams{
				MinGenerateDate: &futureDate,
			},
			true,
			false,
			"Generate Date range",
		},
		{
			"invalid sort",
			&cryptoutilOpenapiModel.MaterialKeysQueryParams{
				Sort: &[]cryptoutilOpenapiModel.MaterialKeySort{invalidSort},
			},
			true,
			false,
			"Key Sort",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := mapper.toOrmGetMaterialKeysQueryParams(tc.params)

			if tc.expectError {
				testify.Error(t, err)
				testify.Contains(t, err.Error(), tc.errorContains)
			} else {
				testify.NoError(t, err)

				if tc.expectNil {
					testify.Nil(t, result)
				} else {
					testify.NotNil(t, result)
				}
			}
		})
	}
}

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
		expectedStatus cryptoutilOpenapiModel.ElasticKeyStatus
	}{
		{"import allowed returns pending import", true, cryptoutilOpenapiModel.PendingImport},
		{"import not allowed returns pending generate", false, cryptoutilOpenapiModel.PendingGenerate},
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
	create := &cryptoutilOpenapiModel.ElasticKeyCreate{
		Name:        "test-key",
		Description: "test description",
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
	testify.Equal(t, cryptoutilOpenapiModel.PendingGenerate, result.ElasticKeyStatus)
}

func TestToOrmAddElasticKeyImportAllowed(t *testing.T) {
	t.Parallel()

	mapper := NewOamOrmMapper()
	elasticKeyID := googleUuid.New()
	tenantID := googleUuid.New()

	// Test with import allowed = true.
	importAllowed := true
	create := &cryptoutilOpenapiModel.ElasticKeyCreate{
		Name:          "test-key",
		Description:   "test description",
		ImportAllowed: &importAllowed,
	}

	result := mapper.toOrmAddElasticKey(&elasticKeyID, tenantID, create)

	testify.True(t, result.ElasticKeyImportAllowed)
	testify.Equal(t, cryptoutilOpenapiModel.PendingImport, result.ElasticKeyStatus)
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
