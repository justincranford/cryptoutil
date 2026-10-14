// Copyright (c) 2025-2026 Justin Cranford.
package businesslogic

import (
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilOpenapiModel "cryptoutil/api/sm-kms/models"
	cryptoutilKmsServer "cryptoutil/api/sm-kms/server"

	googleUuid "github.com/google/uuid"
	testify "github.com/stretchr/testify/require"
)

func TestToOrmGetElasticKeysQueryParams(t *testing.T) {
	t.Parallel()

	mapper := NewOamOrmMapper()

	validUUID := googleUuid.New()
	zeroUUID := googleUuid.UUID{}
	algorithm := cryptoutilOpenapiModel.A128CBCHS256Dir
	name := testElasticKeyName
	versioningAllowed := true
	negativePage := cryptoutilKmsServer.PageNumber(-1)
	zeroPageSize := cryptoutilKmsServer.PageSize(0)
	emptyAlgorithm := cryptoutilOpenapiModel.ElasticKeyAlgorithm("")
	emptyString := ""
	emptySort := cryptoutilOpenapiModel.ElasticKeySort("")

	tests := []struct {
		name          string
		params        *cryptoutilOpenapiModel.ElasticKeysQueryParams
		expectError   bool
		expectNil     bool
		errorContains string
	}{
		{testCaseNilParams, nil, false, false, ""},
		{
			testCaseValidParams,
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
			testCaseInvalidPageNumber,
			&cryptoutilOpenapiModel.ElasticKeysQueryParams{
				Page: &negativePage,
			},
			true,
			false,
			testErrorPageNumber,
		},
		{
			testCaseInvalidPageSize,
			&cryptoutilOpenapiModel.ElasticKeysQueryParams{
				Size: &zeroPageSize,
			},
			true,
			false,
			testErrorPageSize,
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
		{
			"invalid UUID",
			&cryptoutilOpenapiModel.ElasticKeysQueryParams{
				ElasticKeyID: &[]googleUuid.UUID{zeroUUID},
			},
			true,
			false,
			"Elastic Key ID",
		},
		{
			testCaseInvalidSort,
			&cryptoutilOpenapiModel.ElasticKeysQueryParams{
				Sort: &[]cryptoutilOpenapiModel.ElasticKeySort{emptySort},
			},
			true,
			false,
			"Elastic Key Sort",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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
	t.Parallel()

	mapper := NewOamOrmMapper()

	materialKeyID := googleUuid.New()
	zeroMaterialKeyID := googleUuid.UUID{}
	minDate := time.Now().UTC().Add(-cryptoutilSharedMagic.HoursPerDay * time.Hour)
	maxDate := time.Now().UTC()
	futureDate := time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay * time.Hour)
	negativePage := cryptoutilKmsServer.PageNumber(-1)
	zeroPageSize := cryptoutilKmsServer.PageSize(0)
	invalidSort := cryptoutilOpenapiModel.MaterialKeySort(cryptoutilSharedMagic.DevSetupTestInvalidName)

	tests := []struct {
		name          string
		params        *cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams
		expectError   bool
		expectNil     bool
		errorContains string
	}{
		{testCaseNilParams, nil, false, false, ""},
		{
			testCaseValidParams,
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
			testCaseInvalidPageNumber,
			&cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams{
				Page: &negativePage,
			},
			true,
			false,
			testErrorPageNumber,
		},
		{
			testCaseInvalidPageSize,
			&cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams{
				Size: &zeroPageSize,
			},
			true,
			false,
			testErrorPageSize,
		},
		{
			testCaseInvalidDateRange,
			&cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams{
				MinGenerateDate: &futureDate,
			},
			true,
			false,
			testErrorGenerateDateRange,
		},
		{
			testCaseInvalidSort,
			&cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams{
				Sort: &[]cryptoutilOpenapiModel.MaterialKeySort{invalidSort},
			},
			true,
			false,
			testErrorKeySort,
		},
		{
			testCaseInvalidMaterialKeyUUID,
			&cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams{
				MaterialKeyID: &[]googleUuid.UUID{zeroMaterialKeyID},
			},
			true,
			false,
			testErrorMaterialKeyID,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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
	t.Parallel()

	mapper := NewOamOrmMapper()

	elasticKeyID := googleUuid.New()
	materialKeyID := googleUuid.New()
	zeroElasticKeyID := googleUuid.UUID{}
	zeroMatKeyID := googleUuid.UUID{}
	minDate := time.Now().UTC().Add(-cryptoutilSharedMagic.HoursPerDay * time.Hour)
	maxDate := time.Now().UTC()
	futureDate := time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay * time.Hour)
	negativePage := cryptoutilKmsServer.PageNumber(-1)
	zeroPageSize := cryptoutilKmsServer.PageSize(0)
	invalidSort := cryptoutilOpenapiModel.MaterialKeySort(cryptoutilSharedMagic.DevSetupTestInvalidName)

	tests := []struct {
		name          string
		params        *cryptoutilOpenapiModel.MaterialKeysQueryParams
		expectError   bool
		expectNil     bool
		errorContains string
	}{
		{testCaseNilParams, nil, false, false, ""},
		{
			testCaseValidParams,
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
			testCaseInvalidPageNumber,
			&cryptoutilOpenapiModel.MaterialKeysQueryParams{
				Page: &negativePage,
			},
			true,
			false,
			testErrorPageNumber,
		},
		{
			testCaseInvalidPageSize,
			&cryptoutilOpenapiModel.MaterialKeysQueryParams{
				Size: &zeroPageSize,
			},
			true,
			false,
			testErrorPageSize,
		},
		{
			testCaseInvalidDateRange,
			&cryptoutilOpenapiModel.MaterialKeysQueryParams{
				MinGenerateDate: &futureDate,
			},
			true,
			false,
			testErrorGenerateDateRange,
		},
		{
			testCaseInvalidSort,
			&cryptoutilOpenapiModel.MaterialKeysQueryParams{
				Sort: &[]cryptoutilOpenapiModel.MaterialKeySort{invalidSort},
			},
			true,
			false,
			testErrorKeySort,
		},
		{
			"invalid elastic key UUID",
			&cryptoutilOpenapiModel.MaterialKeysQueryParams{
				ElasticKeyID: &[]googleUuid.UUID{zeroElasticKeyID},
			},
			true,
			false,
			"ElasticKeyID",
		},
		{
			testCaseInvalidMaterialKeyUUID,
			&cryptoutilOpenapiModel.MaterialKeysQueryParams{
				MaterialKeyID: &[]googleUuid.UUID{zeroMatKeyID},
			},
			true,
			false,
			testErrorMaterialKeyID,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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
