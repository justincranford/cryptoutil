// Copyright (c) 2025 Justin Cranford

package businesslogic

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"
	"time"

	cryptoutilKmsServer "cryptoutil/api/kms/server"
	cryptoutilOpenapiModel "cryptoutil/api/model"

	googleUuid "github.com/google/uuid"
	testify "github.com/stretchr/testify/require"
)

func TestToOrmGetElasticKeysQueryParams(t *testing.T) {
	t.Parallel()

	mapper := NewOamOrmMapper()

	validUUID := googleUuid.New()
	zeroUUID := googleUuid.UUID{}
	algorithm := cryptoutilOpenapiModel.A128CBCHS256Dir
	name := "test-key"
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
			"invalid sort",
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
		{
			"invalid material key UUID",
			&cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams{
				MaterialKeyID: &[]googleUuid.UUID{zeroMaterialKeyID},
			},
			true,
			false,
			"MaterialKeyID",
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
			"invalid material key UUID",
			&cryptoutilOpenapiModel.MaterialKeysQueryParams{
				MaterialKeyID: &[]googleUuid.UUID{zeroMatKeyID},
			},
			true,
			false,
			"MaterialKeyID",
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
