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
	KeyPoolAlgorithm         KeyPoolAlgorithm `gorm:"size:38;not null;check:key_pool_algorithm IN ('A256GCM/A256GCMKW', 'A192GCM/A192GCMKW', 'A128GCM/A128GCMKW', 'A256GCM/Dir', 'A192GCM/Dir', 'A128GCM/Dir', 'A256CBC-HS512/A256GCMKW', 'A192CBC-HS384/A192GCMKW', 'A128CBC-HS256/A128GCMKW', 'A256CBC-HS512/Dir', 'A192CBC-HS384/Dir', 'A128CBC-HS256/Dir')"`
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
	A256GCM_A256GCMKW      KeyPoolAlgorithm = "A256GCM/A256GCMKW"
	A192GCM_A192GCMKW      KeyPoolAlgorithm = "A192GCM/A192GCMKW"
	A128GCM_A128GCMKW      KeyPoolAlgorithm = "A128GCM/A128GCMKW"
	A256GCM_Dir            KeyPoolAlgorithm = "A256GCM/Dir"
	A192GCM_Dir            KeyPoolAlgorithm = "A192GCM/Dir"
	A128GCM_Dir            KeyPoolAlgorithm = "A128GCM/Dir"
	A256CBCHS512_A256GCMKW KeyPoolAlgorithm = "A256CBC-HS512/A256GCMKW"
	A192CBCHS384_A192GCMKW KeyPoolAlgorithm = "A192CBC-HS384/A192GCMKW"
	A128CBCHS256_A128GCMKW KeyPoolAlgorithm = "A128CBC-HS256/A128GCMKW"
	A256CBCHS512_Dir       KeyPoolAlgorithm = "A256CBC-HS512/Dir"
	A192CBCHS384_Dir       KeyPoolAlgorithm = "A192CBC-HS384/Dir"
	A128CBCHS256_Dir       KeyPoolAlgorithm = "A128CBC-HS256/Dir"
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
