package orm

import (
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"

	googleUuid "github.com/google/uuid"
)

func BuildElasticKey(elasticKeyID googleUuid.UUID, name, description string, provider cryptoutilOpenapiModel.ElasticKeyProvider, algorithm cryptoutilOpenapiModel.ElasticKeyAlgorithm, versioningAllowed, importAllowed, exportAllowed bool, status string) (*ElasticKey, error) {
	elasticKey := ElasticKey{
		ElasticKeyID:                elasticKeyID,
		ElasticKeyName:              name,
		ElasticKeyDescription:       description,
		ElasticKeyProvider:          provider,
		ElasticKeyAlgorithm:         algorithm,
		ElasticKeyVersioningAllowed: versioningAllowed,
		ElasticKeyImportAllowed:     importAllowed,
		ElasticKeyStatus:            ElasticKeyStatusInitial(importAllowed),
	}
	return &elasticKey, nil
}

func ElasticKeyStatusInitial(importAllowed bool) cryptoutilOpenapiModel.ElasticKeyStatus {
	if importAllowed {
		return cryptoutilOpenapiModel.PendingImport
	}
	return cryptoutilOpenapiModel.PendingGenerate
}
