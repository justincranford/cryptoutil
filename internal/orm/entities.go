package orm

import (
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ormTableStructs = []any{&KEKPool{}, &KEK{}}

type KEKPool struct {
	KEKPoolID                  uuid.UUID            `gorm:"type:uuid;primaryKey"`
	KEKPoolName                string               `gorm:"size:63;not null;check:length(kek_pool_name) >= 1"`
	KEKPoolDescription         string               `gorm:"size:255;not null;check:length(kek_pool_description) >= 1"`
	KEKPoolProvider            KEKPoolProviderEnum  `gorm:"size:8;not null;check:kek_pool_provider IN ('Internal')"`
	KEKPoolAlgorithm           KEKPoolAlgorithmEnum `gorm:"size:15;not null;check:kek_pool_algorithm IN ('AES-256', 'AES-192', 'AES-128')"`
	KEKPoolIsVersioningAllowed bool                 `gorm:"not null;check:kek_pool_is_versioning_allowed IN (TRUE, FALSE)"`
	KEKPoolIsImportAllowed     bool                 `gorm:"not null;check:kek_pool_is_import_allowed IN (TRUE, FALSE)"`
	KEKPoolIsExportAllowed     bool                 `gorm:"not null;check:kek_pool_is_export_allowed IN (TRUE, FALSE)"`
	KEKPoolStatus              KEKPoolStatusEnum    `gorm:"size:16;not null;check:kek_pool_status IN ('active', 'disabled', 'pending_generate', 'pending_import')"`
}

func (k *KEKPool) BeforeCreate(tx *gorm.DB) (err error) {
	if k.KEKPoolID == uuid.Nil {
		k.KEKPoolID, err = uuid.NewV7()
		if err != nil {
			log.Printf("failed to generate UUIDv7: %v", err)
		}
	}
	return
}

type KEK struct {
	KEKPoolID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	KEKID             int       `gorm:"primaryKey;autoIncrement:false;not null;check(kek_id >= 0)"`
	KEKMaterial       []byte    `gorm:"not null;check(length(kek_material) >= 1)"`
	KEKGenerateDate   *time.Time
	KEKImportDate     *time.Time
	KEKExpirationDate *time.Time
	KEKRevocationDate *time.Time
}

type KEKPoolCreate struct {
	Algorithm           KEKPoolAlgorithmEnum       `json:"algorithm,omitempty"`
	Description         KEKPoolDescription         `json:"description"`
	IsExportAllowed     KEKPoolIsExportAllowed     `json:"isExportAllowed,omitempty"`
	IsImportAllowed     KEKPoolIsImportAllowed     `json:"isImportAllowed,omitempty"`
	IsVersioningAllowed KEKPoolIsVersioningAllowed `json:"isVersioningAllowed,omitempty"`
	Name                KEKPoolName                `json:"name"`
	Provider            KEKPoolProviderEnum        `json:"provider,omitempty"`
}

type KEKPoolAlgorithmEnum string

const (
	AES128 KEKPoolAlgorithmEnum = "AES-128"
	AES192 KEKPoolAlgorithmEnum = "AES-192"
	AES256 KEKPoolAlgorithmEnum = "AES-256"
)

type KEKPoolProviderEnum string

const (
	Internal KEKPoolProviderEnum = "Internal"
)

type KEKPoolStatusEnum string

const (
	Active          KEKPoolStatusEnum = "active"
	Disabled        KEKPoolStatusEnum = "disabled"
	PendingGenerate KEKPoolStatusEnum = "pending_generate"
	PendingImport   KEKPoolStatusEnum = "pending_import"
)

type (
	KEKPoolDescription         string
	KEKPoolId                  string
	KEKPoolIsExportAllowed     bool
	KEKPoolIsImportAllowed     bool
	KEKPoolIsVersioningAllowed bool
	KEKPoolName                string
)
