package orm

import (
	cryptoutilBusinessModel "cryptoutil/internal/common/businessmodel"
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"

	googleUuid "github.com/google/uuid"
)

func BuildElasticKey(elasticKeyID googleUuid.UUID, name, description string, provider cryptoutilBusinessModel.ElasticKeyProvider, algorithm cryptoutilOpenapiModel.ElasticKeyAlgorithm, versioningAllowed, importAllowed, exportAllowed bool, status string) (*ElasticKey, error) {
	elasticKey := ElasticKey{
		ElasticKeyID:                elasticKeyID,
		ElasticKeyName:              name,
		ElasticKeyDescription:       description,
		ElasticKeyProvider:          provider,
		ElasticKeyAlgorithm:         algorithm,
		ElasticKeyVersioningAllowed: versioningAllowed,
		ElasticKeyImportAllowed:     importAllowed,
		ElasticKeyExportAllowed:     exportAllowed,
		ElasticKeyStatus:            cryptoutilBusinessModel.ElasticKeyStatus(status),
	}
	return &elasticKey, nil
}

func ElasticKeyStatusInitial(importAllowed bool) string {
	if importAllowed {
		return string(cryptoutilBusinessModel.PendingImport)
	}
	return string(cryptoutilBusinessModel.PendingGenerate)
}
