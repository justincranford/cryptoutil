package orm

import (
	"time"

	googleUuid "github.com/google/uuid"
)

type KeyPool struct {
	KeyPoolID                googleUuid.UUID  `gorm:"type:uuid;primaryKey"`
	KeyPoolName              string           `gorm:"size:63;not null;check:length(key_pool_name) >= 1;unique"`
	KeyPoolDescription       string           `gorm:"size:255;not null;check:length(key_pool_description) >= 1"`
	KeyPoolProvider          KeyPoolProvider  `gorm:"size:8;not null;check:key_pool_provider IN ('Internal')"`
	KeyPoolAlgorithm         KeyPoolAlgorithm `gorm:"size:15;not null;check:key_pool_algorithm IN ('AES-256', 'AES-192', 'AES-128')"`
	KeyPoolVersioningAllowed bool             `gorm:"not null;check:key_pool_versioning_allowed IN (TRUE, FALSE)"`
	KeyPoolImportAllowed     bool             `gorm:"not null;check:key_pool_import_allowed IN (TRUE, FALSE)"`
	KeyPoolExportAllowed     bool             `gorm:"not null;check:key_pool_export_allowed IN (TRUE, FALSE)"`
	KeyPoolStatus            KeyPoolStatus    `gorm:"size:34;not null;check:key_pool_status IN ('creating', 'import_failed', 'pending_import', 'pending_generate', 'generate_failed', 'active', 'disabled', 'pending_delete_was_import_failed', 'pending_delete_was_pending_import', 'pending_delete_was_active', 'pending_delete_was_disabled', 'pending_delete_was_generate_failed', 'started_delete', 'finished_delete')"`
}

type Key struct {
	KeyPoolID         googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	KeyID             googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	KeyMaterial       []byte          `gorm:"not null;check(length(key_material) >= 1)"`
	KeyGenerateDate   *time.Time
	KeyImportDate     *time.Time
	KeyExpirationDate *time.Time
	KeyRevocationDate *time.Time
}

type KeyPoolAlgorithm string

const (
	AES128 KeyPoolAlgorithm = "AES-128"
	AES192 KeyPoolAlgorithm = "AES-192"
	AES256 KeyPoolAlgorithm = "AES-256"
)

type KeyPoolProvider string

const (
	Internal KeyPoolProvider = "Internal"
)

type KeyPoolStatus string

const (
	Creating                       KeyPoolStatus = "creating"
	ImportFailed                   KeyPoolStatus = "import_failed"
	PendingImport                  KeyPoolStatus = "pending_import"
	PendingGenerate                KeyPoolStatus = "pending_generate"
	GenerateFailed                 KeyPoolStatus = "generate_failed"
	Active                         KeyPoolStatus = "active"
	Disabled                       KeyPoolStatus = "disabled"
	PendingDeleteWasImportFailed   KeyPoolStatus = "pending_delete_was_import_failed"
	PendingDeleteWasPendingImport  KeyPoolStatus = "pending_delete_was_pending_import"
	PendingDeleteWasActive         KeyPoolStatus = "pending_delete_was_active"
	PendingDeleteWasDisabled       KeyPoolStatus = "pending_delete_was_disabled"
	PendingDeleteWasGenerateFailed KeyPoolStatus = "pending_delete_was_generate_failed"
	StartedDelete                  KeyPoolStatus = "started_delete"
	FinishedDelete                 KeyPoolStatus = "finished_delete"
)

type (
	KeyPoolDescription       string
	KeyPoolId                string
	KeyPoolExportAllowed     bool
	KeyPoolImportAllowed     bool
	KeyPoolVersioningAllowed bool
	KeyPoolName              string
)
