// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	cryptoutilKmsServer "cryptoutil/api/kms/server"
	cryptoutilOpenapiModel "cryptoutil/api/model"

	googleUuid "github.com/google/uuid"
)

// BuildElasticKey constructs a new ElasticKey entity with the specified properties.
func BuildElasticKey(tenantID, elasticKeyID googleUuid.UUID, name, description string, provider cryptoutilOpenapiModel.ElasticKeyProvider, algorithm cryptoutilOpenapiModel.ElasticKeyAlgorithm, versioningAllowed, importAllowed, _ bool, _ string) (*ElasticKey, error) {
	elasticKey := ElasticKey{
		TenantID:                    tenantID,
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

// ElasticKeyStatusInitial returns the initial status for an elastic key based on import configuration.
func ElasticKeyStatusInitial(importAllowed bool) cryptoutilKmsServer.ElasticKeyStatus {
	if importAllowed {
		return cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingImport)
	}

	return cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingGenerate)
}
