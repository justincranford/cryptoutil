// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"time"

	cryptoutilOpenapiModel "cryptoutil/api/model"

	googleUuid "github.com/google/uuid"
)

// ElasticKey represents a key envelope that can contain multiple material key versions.
type ElasticKey struct {
	ElasticKeyID                googleUuid.UUID                            `gorm:"type:uuid;primaryKey"`
	ElasticKeyName              string                                     `gorm:"size:63;not null;check:length(elastic_key_name) >= 1;unique"`
	ElasticKeyDescription       string                                     `gorm:"size:255;not null;check:length(elastic_key_description) >= 1"`
	ElasticKeyProvider          cryptoutilOpenapiModel.ElasticKeyProvider  `gorm:"size:8;not null;check:elastic_key_provider IN ('Internal')"`
	ElasticKeyAlgorithm         cryptoutilOpenapiModel.ElasticKeyAlgorithm `gorm:"size:26;not null"`
	ElasticKeyVersioningAllowed bool                                       `gorm:"not null;check:elastic_key_versioning_allowed IN (TRUE, FALSE)"`
	ElasticKeyImportAllowed     bool                                       `gorm:"not null;check:elastic_key_import_allowed IN (TRUE, FALSE)"`
	ElasticKeyStatus            cryptoutilOpenapiModel.ElasticKeyStatus    `gorm:"size:34;not null;check:elastic_key_status IN ('creating', 'import_failed', 'pending_import', 'pending_generate', 'generate_failed', 'active', 'disabled', 'pending_delete_was_import_failed', 'pending_delete_was_pending_import', 'pending_delete_was_active', 'pending_delete_was_disabled', 'pending_delete_was_generate_failed', 'started_delete', 'finished_delete')"`
}

// MaterialKey represents a specific key version within an elastic key.
type MaterialKey struct {
	ElasticKeyID                  googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	MaterialKeyID                 googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	MaterialKeyClearPublic        []byte          `gorm:""`
	MaterialKeyEncryptedNonPublic []byte          `gorm:"not null;check(length(material_key_encrypted_non_public) >= 1)"`
	MaterialKeyGenerateDate       *time.Time
	MaterialKeyImportDate         *time.Time
	MaterialKeyExpirationDate     *time.Time
	MaterialKeyRevocationDate     *time.Time
}
