package orm

import (
	cryptoutilBusinessModel "cryptoutil/internal/common/businessmodel"

	"time"

	googleUuid "github.com/google/uuid"
)

type ElasticKey struct {
	ElasticKeyID                googleUuid.UUID                             `gorm:"type:uuid;primaryKey"`
	ElasticKeyName              string                                      `gorm:"size:63;not null;check:length(elastic_key_name) >= 1;unique"`
	ElasticKeyDescription       string                                      `gorm:"size:255;not null;check:length(elastic_key_description) >= 1"`
	ElasticKeyProvider          cryptoutilBusinessModel.ElasticKeyProvider  `gorm:"size:8;not null;check:elastic_key_provider IN ('Internal')"`
	ElasticKeyAlgorithm         cryptoutilBusinessModel.ElasticKeyAlgorithm `gorm:"size:26;not null"`
	ElasticKeyVersioningAllowed bool                                        `gorm:"not null;check:elastic_key_versioning_allowed IN (TRUE, FALSE)"`
	ElasticKeyImportAllowed     bool                                        `gorm:"not null;check:elastic_key_import_allowed IN (TRUE, FALSE)"`
	ElasticKeyExportAllowed     bool                                        `gorm:"not null;check:elastic_key_export_allowed IN (TRUE, FALSE)"`
	ElasticKeyStatus            cryptoutilBusinessModel.ElasticKeyStatus    `gorm:"size:34;not null;check:elastic_key_status IN ('creating', 'import_failed', 'pending_import', 'pending_generate', 'generate_failed', 'active', 'disabled', 'pending_delete_was_import_failed', 'pending_delete_was_pending_import', 'pending_delete_was_active', 'pending_delete_was_disabled', 'pending_delete_was_generate_failed', 'started_delete', 'finished_delete')"`
}

type MaterialKey struct {
	ElasticKeyID                  googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	MaterialKeyID                 googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	ClearPublicKeyMaterial        []byte          `gorm:"check(length(clear_public_key_material) >= 1)"`
	EncryptedNonPublicKeyMaterial []byte          `gorm:"not null;check(length(encrypted_non_public_key_material) >= 1)"`
	MaterialKeyGenerateDate       *time.Time
	MaterialKeyImportDate         *time.Time
	MaterialKeyExpirationDate     *time.Time
	MaterialKeyRevocationDate     *time.Time
}
