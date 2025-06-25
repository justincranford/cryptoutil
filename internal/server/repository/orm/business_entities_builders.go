package orm

import (
	"cryptoutil/internal/common/constant"

	googleUuid "github.com/google/uuid"
)

func BuildElasticKey(elasticKeyID googleUuid.UUID, name, description string, provider constant.ElasticKeyProvider, algorithm constant.ElasticKeyAlgorithm, versioningAllowed, importAllowed, exportAllowed bool, status string) (*ElasticKey, error) {
	elasticKey := ElasticKey{
		ElasticKeyID:                elasticKeyID,
		ElasticKeyName:              name,
		ElasticKeyDescription:       description,
		ElasticKeyProvider:          provider,
		ElasticKeyAlgorithm:         algorithm,
		ElasticKeyVersioningAllowed: versioningAllowed,
		ElasticKeyImportAllowed:     importAllowed,
		ElasticKeyExportAllowed:     exportAllowed,
		ElasticKeyStatus:            constant.ElasticKeyStatus(status),
	}
	return &elasticKey, nil
}

func ElasticKeyStatusInitial(importAllowed bool) string {
	if importAllowed {
		return string(constant.PendingImport)
	}
	return string(constant.PendingGenerate)
}
