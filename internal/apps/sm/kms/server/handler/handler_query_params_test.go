// Copyright (c) 2025 Justin Cranford

package handler

import (
	"testing"
	"time"

	cryptoutilKmsServer "cryptoutil/api/kms/server"
	cryptoutilOpenapiModel "cryptoutil/api/model"

	googleUuid "github.com/google/uuid"
	openapiTypes "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
)

func TestOamOasMapper_ToOamGetElasticKeyMaterialKeysQueryParams_AllPopulated(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()

	materialKeyID1 := openapiTypes.UUID(googleUuid.New())
	materialKeyID2 := openapiTypes.UUID(googleUuid.New())
	materialKeyIDs := cryptoutilKmsServer.MaterialKeyIDs{materialKeyID1, materialKeyID2}

	pageNum := 2
	pageSize := 50
	now := time.Now().UTC()

	params := &cryptoutilKmsServer.GetElastickeyElasticKeyIDMaterialkeysParams{
		MaterialKeyIds:    &materialKeyIDs,
		PageNumber:        &pageNum,
		PageSize:          &pageSize,
		MinGenerateDate:   &now,
		MaxGenerateDate:   &now,
		MinImportDate:     &now,
		MaxImportDate:     &now,
		MinExpirationDate: &now,
		MaxExpirationDate: &now,
		MinRevocationDate: &now,
		MaxRevocationDate: &now,
	}

	result := mapper.toOamGetElasticKeyMaterialKeysQueryParams(params)
	require.NotNil(t, result)
	require.NotNil(t, result.MaterialKeyID)
	require.Len(t, *result.MaterialKeyID, 2)
	require.Equal(t, cryptoutilOpenapiModel.MaterialKeyID(materialKeyID1), (*result.MaterialKeyID)[0])
	require.Equal(t, cryptoutilOpenapiModel.MaterialKeyID(materialKeyID2), (*result.MaterialKeyID)[1])
	require.NotNil(t, result.Page)
	require.Equal(t, cryptoutilOpenapiModel.PageNumber(2), *result.Page)
	require.NotNil(t, result.Size)
	require.Equal(t, cryptoutilOpenapiModel.PageSize(50), *result.Size)
	require.NotNil(t, result.MinGenerateDate)
	require.NotNil(t, result.MaxGenerateDate)
	require.NotNil(t, result.MinImportDate)
	require.NotNil(t, result.MaxImportDate)
	require.NotNil(t, result.MinExpirationDate)
	require.NotNil(t, result.MaxExpirationDate)
	require.NotNil(t, result.MinRevocationDate)
	require.NotNil(t, result.MaxRevocationDate)
}

func TestOamOasMapper_ToOamGetElasticKeyQueryParams_AllPopulated(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()

	elasticKeyID1 := openapiTypes.UUID(googleUuid.New())
	elasticKeyIDs := cryptoutilKmsServer.ElasticKeyIDs{elasticKeyID1}

	names := cryptoutilKmsServer.Names{"key-1", "key-2"}
	providers := cryptoutilKmsServer.Providers{"local", "hsm"}
	algorithms := cryptoutilKmsServer.Algorithms{"AES-256-GCM", "RSA-2048"}
	statuses := cryptoutilKmsServer.Statuses{"active", "disabled"}
	sorts := cryptoutilKmsServer.Sorts{"name:asc", "created_at:desc"}

	versioningAllowed := true
	importAllowed := false
	pageNum := 1
	pageSize := 25

	params := &cryptoutilKmsServer.GetElastickeysParams{
		ElasticKeyIds:     &elasticKeyIDs,
		Names:             &names,
		Providers:         &providers,
		Algorithms:        &algorithms,
		Statuses:          &statuses,
		Sorts:             &sorts,
		VersioningAllowed: &versioningAllowed,
		ImportAllowed:     &importAllowed,
		PageNumber:        &pageNum,
		PageSize:          &pageSize,
	}

	result := mapper.toOamGetElasticKeyQueryParams(params)
	require.NotNil(t, result)
	require.NotNil(t, result.ElasticKeyID)
	require.Len(t, *result.ElasticKeyID, 1)
	require.Equal(t, cryptoutilOpenapiModel.ElasticKeyID(elasticKeyID1), (*result.ElasticKeyID)[0])
	require.NotNil(t, result.Name)
	require.Len(t, *result.Name, 2)
	require.NotNil(t, result.Provider)
	require.Len(t, *result.Provider, 2)
	require.NotNil(t, result.Algorithm)
	require.Len(t, *result.Algorithm, 2)
	require.NotNil(t, result.Status)
	require.Len(t, *result.Status, 2)
	require.NotNil(t, result.Sort)
	require.Len(t, *result.Sort, 2)
	require.NotNil(t, result.VersioningAllowed)
	require.True(t, bool(*result.VersioningAllowed))
	require.NotNil(t, result.ImportAllowed)
	require.False(t, bool(*result.ImportAllowed))
	require.NotNil(t, result.Page)
	require.Equal(t, cryptoutilOpenapiModel.PageNumber(1), *result.Page)
	require.NotNil(t, result.Size)
	require.Equal(t, cryptoutilOpenapiModel.PageSize(25), *result.Size)
}

func TestOamOasMapper_ToOamGetMaterialKeysQueryParams_AllPopulated(t *testing.T) {
	t.Parallel()

	mapper := NewOasOamMapper()

	elasticKeyID1 := openapiTypes.UUID(googleUuid.New())
	elasticKeyIDs := cryptoutilKmsServer.ElasticKeyIDs{elasticKeyID1}

	materialKeyID1 := openapiTypes.UUID(googleUuid.New())
	materialKeyIDs := cryptoutilKmsServer.MaterialKeyIDs{materialKeyID1}

	now := time.Now().UTC()
	pageNum := 3
	pageSize := 100

	params := &cryptoutilKmsServer.GetMaterialkeysParams{
		ElasticKeyIds:     &elasticKeyIDs,
		MaterialKeyIds:    &materialKeyIDs,
		MinGenerateDate:   &now,
		MaxGenerateDate:   &now,
		MinImportDate:     &now,
		MaxImportDate:     &now,
		MinExpirationDate: &now,
		MaxExpirationDate: &now,
		MinRevocationDate: &now,
		MaxRevocationDate: &now,
		PageNumber:        &pageNum,
		PageSize:          &pageSize,
	}

	result := mapper.toOamGetMaterialKeysQueryParams(params)
	require.NotNil(t, result)
	require.NotNil(t, result.ElasticKeyID)
	require.Len(t, *result.ElasticKeyID, 1)
	require.Equal(t, cryptoutilOpenapiModel.ElasticKeyID(elasticKeyID1), (*result.ElasticKeyID)[0])
	require.NotNil(t, result.MaterialKeyID)
	require.Len(t, *result.MaterialKeyID, 1)
	require.Equal(t, cryptoutilOpenapiModel.MaterialKeyID(materialKeyID1), (*result.MaterialKeyID)[0])
	require.NotNil(t, result.MinGenerateDate)
	require.NotNil(t, result.MaxGenerateDate)
	require.NotNil(t, result.MinImportDate)
	require.NotNil(t, result.MaxImportDate)
	require.NotNil(t, result.MinExpirationDate)
	require.NotNil(t, result.MaxExpirationDate)
	require.NotNil(t, result.MinRevocationDate)
	require.NotNil(t, result.MaxRevocationDate)
	require.NotNil(t, result.Page)
	require.Equal(t, cryptoutilOpenapiModel.PageNumber(3), *result.Page)
	require.NotNil(t, result.Size)
	require.Equal(t, cryptoutilOpenapiModel.PageSize(100), *result.Size)
	require.Nil(t, result.Sort)
}
