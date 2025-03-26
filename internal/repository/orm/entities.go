package orm

import (
	"log"
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

var ormTableStructs = []any{&KeyPool{}, &Key{}}

type KeyPool struct {
	KeyPoolID                  googleUuid.UUID      `gorm:"type:uuid;primaryKey"`
	KeyPoolName                string               `gorm:"size:63;not null;check:length(key_pool_name) >= 1;unique"`
	KeyPoolDescription         string               `gorm:"size:255;not null;check:length(key_pool_description) >= 1"`
	KeyPoolProvider            KeyPoolProviderEnum  `gorm:"size:8;not null;check:key_pool_provider IN ('Internal')"`
	KeyPoolAlgorithm           KeyPoolAlgorithmEnum `gorm:"size:15;not null;check:key_pool_algorithm IN ('AES-256', 'AES-192', 'AES-128')"`
	KeyPoolIsVersioningAllowed bool                 `gorm:"not null;check:key_pool_is_versioning_allowed IN (TRUE, FALSE)"`
	KeyPoolIsImportAllowed     bool                 `gorm:"not null;check:key_pool_is_import_allowed IN (TRUE, FALSE)"`
	KeyPoolIsExportAllowed     bool                 `gorm:"not null;check:key_pool_is_export_allowed IN (TRUE, FALSE)"`
	KeyPoolStatus              KeyPoolStatusEnum    `gorm:"size:16;not null;check:key_pool_status IN ('active', 'disabled', 'pending_generate', 'pending_import')"`
}

func (k *KeyPool) BeforeCreate(tx *gorm.DB) (err error) {
	if k.KeyPoolID == googleUuid.Nil {
		k.KeyPoolID, err = googleUuid.NewV7()
		if err != nil {
			log.Printf("failed to generate UUIDv7: %v", err)
		}
	}
	return
}

type Key struct {
	KeyPoolID         googleUuid.UUID `gorm:"type:uuid;primaryKey"`
	KeyID             int             `gorm:"primaryKey;autoIncrement:false;not null;check(key_id >= 0)"`
	KeyMaterial       []byte          `gorm:"not null;check(length(key_material) >= 1)"`
	KeyGenerateDate   *time.Time
	KeyImportDate     *time.Time
	KeyExpirationDate *time.Time
	KeyRevocationDate *time.Time
}

type KeyPoolCreate struct {
	Algorithm           KeyPoolAlgorithmEnum       `json:"algorithm,omitempty"`
	Description         KeyPoolDescription         `json:"description"`
	IsExportAllowed     KeyPoolIsExportAllowed     `json:"isExportAllowed,omitempty"`
	IsImportAllowed     KeyPoolIsImportAllowed     `json:"isImportAllowed,omitempty"`
	IsVersioningAllowed KeyPoolIsVersioningAllowed `json:"isVersioningAllowed,omitempty"`
	Name                KeyPoolName                `json:"name"`
	Provider            KeyPoolProviderEnum        `json:"provider,omitempty"`
}

type KeyPoolAlgorithmEnum string

const (
	AES128 KeyPoolAlgorithmEnum = "AES-128"
	AES192 KeyPoolAlgorithmEnum = "AES-192"
	AES256 KeyPoolAlgorithmEnum = "AES-256"
)

type KeyPoolProviderEnum string

const (
	Internal KeyPoolProviderEnum = "Internal"
)

type KeyPoolStatusEnum string

const (
	Active          KeyPoolStatusEnum = "active"
	Disabled        KeyPoolStatusEnum = "disabled"
	PendingGenerate KeyPoolStatusEnum = "pending_generate"
	PendingImport   KeyPoolStatusEnum = "pending_import"
)

type (
	KeyPoolDescription         string
	KeyPoolId                  string
	KeyPoolIsExportAllowed     bool
	KeyPoolIsImportAllowed     bool
	KeyPoolIsVersioningAllowed bool
	KeyPoolName                string
)
