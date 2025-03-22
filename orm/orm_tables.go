package orm

import (
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ormTableStructs = []any{&KEKPool{}, &KEK{}}

type KEKPool struct {
	KEKPoolID                  uuid.UUID `gorm:"type:uuid;primaryKey"`
	KEKPoolName                string    `gorm:"size:63;not null;check:length(kek_pool_name) >= 1"`
	KEKPoolDescription         string    `gorm:"size:255;not null;check:length(kek_pool_description) >= 1"`
	KEKPoolAlgorithm           string    `gorm:"size:15;not null;check:kek_pool_algorithm IN ('AES-256', 'AES-192', 'AES-128')"`
	KEKPoolStatus              string    `gorm:"size:16;not null;check:kek_pool_status IN ('active', 'disabled', 'pending_generate', 'pending_import')"`
	KEKPoolProvider            string    `gorm:"size:8;not null;check:kek_pool_provider IN ('Internal')"`
	KEKPoolIsVersioningAllowed bool      `gorm:"not null;check:kek_pool_is_versioning_allowed IN (TRUE, FALSE)"`
	KEKPoolIsImportAllowed     bool      `gorm:"not null;check:kek_pool_is_import_allowed IN (TRUE, FALSE)"`
	KEKPoolIsExportAllowed     bool      `gorm:"not null;check:kek_pool_is_export_allowed IN (TRUE, FALSE)"`
}

func (KEKPool) TableName() string {
	return "kek_pool"
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
	KEKGenerateDate   *string   `gorm:"size:20;check(length(kek_generate_date)   == 20)"` // ISO 8601
	KEKImportDate     *string   `gorm:"size:20;check(length(kek_import_date)     == 20)"` // ISO 8601
	KEKExpirationDate *string   `gorm:"size:20;check(length(kek_expiration_date) == 20)"` // ISO 8601
	KEKRevocationDate *string   `gorm:"size:20;check(length(kek_revocation_date) == 20)"` // ISO 8601
}

func (KEK) TableName() string {
	return "kek"
}

type KEKPoolCreate struct {
	Algorithm           KEKPoolAlgorithm           `json:"algorithm,omitempty"`
	Description         KEKPoolDescription         `json:"description"`
	IsExportAllowed     KEKPoolIsExportAllowed     `json:"isExportAllowed,omitempty"`
	IsImportAllowed     KEKPoolIsImportAllowed     `json:"isImportAllowed,omitempty"`
	IsVersioningAllowed KEKPoolIsVersioningAllowed `json:"isVersioningAllowed,omitempty"`
	Name                KEKPoolName                `json:"name"`
	Provider            KEKPoolProvider            `json:"provider,omitempty"`
}

type KEKPoolAlgorithm string

const (
	AES128 KEKPoolAlgorithm = "AES-128"
	AES192 KEKPoolAlgorithm = "AES-192"
	AES256 KEKPoolAlgorithm = "AES-256"
)

type KEKPoolProvider string

const (
	Internal KEKPoolProvider = "Internal"
)

type KEKPoolStatus string

const (
	Active          KEKPoolStatus = "active"
	Disabled        KEKPoolStatus = "disabled"
	PendingGenerate KEKPoolStatus = "pending_generate"
	PendingImport   KEKPoolStatus = "pending_import"
)

type (
	KEKPoolDescription         string
	KEKPoolId                  string
	KEKPoolIsExportAllowed     bool
	KEKPoolIsImportAllowed     bool
	KEKPoolIsVersioningAllowed bool
	KEKPoolName                string
)
