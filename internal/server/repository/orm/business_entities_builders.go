package orm

import (
	"time"

	googleUuid "github.com/google/uuid"
)

func BuildElasticKey(elasticKeyID googleUuid.UUID, name, description string, provider ElasticKeyProvider, algorithm ElasticKeyAlgorithm, versioningAllowed, importAllowed, exportAllowed bool, status string) (*ElasticKey, error) {
	elasticKey := ElasticKey{
		ElasticKeyID:                elasticKeyID,
		ElasticKeyName:              name,
		ElasticKeyDescription:       description,
		ElasticKeyProvider:          provider,
		ElasticKeyAlgorithm:         algorithm,
		ElasticKeyVersioningAllowed: versioningAllowed,
		ElasticKeyImportAllowed:     importAllowed,
		ElasticKeyExportAllowed:     exportAllowed,
		ElasticKeyStatus:            ElasticKeyStatus(status),
	}
	return &elasticKey, nil
}

func BuildKey(elasticKeyID googleUuid.UUID, materialKeyID googleUuid.UUID, keyMaterial []byte, generateDate, importDate, expirationDate, revocationDate *time.Time) *MaterialKey {
	key := MaterialKey{
		ElasticKeyID:              elasticKeyID,
		MaterialKeyID:             materialKeyID,
		KeyMaterial:               keyMaterial,
		MaterialKeyGenerateDate:   generateDate,
		MaterialKeyImportDate:     importDate,
		MaterialKeyExpirationDate: expirationDate,
		MaterialKeyRevocationDate: revocationDate,
	}
	return &key
}

func ElasticKeyStatusInitial(importAllowed bool) string {
	if importAllowed {
		return string(PendingImport)
	}
	return string(PendingGenerate)
}
