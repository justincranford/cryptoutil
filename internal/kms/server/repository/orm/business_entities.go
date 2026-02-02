// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	cryptoutilOpenapiModel "cryptoutil/api/model"

	googleUuid "github.com/google/uuid"
)

// ElasticKey represents a key envelope that can contain multiple material key versions.
type ElasticKey struct {
	ElasticKeyID                googleUuid.UUID                            `gorm:"type:text;primaryKey"`
	TenantID                    googleUuid.UUID                            `gorm:"type:text;not null;index"`
	ElasticKeyName              string                                     `gorm:"type:text;not null;check:length(elastic_key_name) >= 1;uniqueIndex:idx_elastic_keys_tenant_name"`
	ElasticKeyDescription       string                                     `gorm:"type:text;not null;check:length(elastic_key_description) >= 1"`
	ElasticKeyProvider          cryptoutilOpenapiModel.ElasticKeyProvider  `gorm:"type:text;not null;check:elastic_key_provider IN ('Internal')"`
	ElasticKeyAlgorithm         cryptoutilOpenapiModel.ElasticKeyAlgorithm `gorm:"type:text;not null"`
	ElasticKeyVersioningAllowed bool                                       `gorm:"type:integer;not null"`
	ElasticKeyImportAllowed     bool                                       `gorm:"type:integer;not null"`
	ElasticKeyStatus            cryptoutilOpenapiModel.ElasticKeyStatus    `gorm:"type:text;not null"`
}

// MaterialKey represents a specific key version within an elastic key.
// Date fields are stored as Unix epoch milliseconds (BIGINT) for cross-database compatibility.
type MaterialKey struct {
	ElasticKeyID                  googleUuid.UUID `gorm:"type:text;primaryKey"`
	MaterialKeyID                 googleUuid.UUID `gorm:"type:text;primaryKey"`
	MaterialKeyClearPublic        []byte          `gorm:"type:blob"`
	MaterialKeyEncryptedNonPublic []byte          `gorm:"type:blob;not null"`
	MaterialKeyGenerateDate       *int64          `gorm:"type:bigint"` // Unix epoch milliseconds
	MaterialKeyImportDate         *int64          `gorm:"type:bigint"` // Unix epoch milliseconds
	MaterialKeyExpirationDate     *int64          `gorm:"type:bigint"` // Unix epoch milliseconds
	MaterialKeyRevocationDate     *int64          `gorm:"type:bigint"` // Unix epoch milliseconds
}
