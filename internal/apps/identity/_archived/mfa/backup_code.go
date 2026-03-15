// Copyright (c) 2025 Justin Cranford
//
//

package mfa

import (
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// BackupCode stores MFA recovery codes.
type BackupCode struct {
	ID        googleUuid.UUID `gorm:"type:text;primaryKey"`
	UserID    googleUuid.UUID `gorm:"type:text;index;not null"`
	CodeHash  string          `gorm:"type:text;uniqueIndex;not null" json:"-"`
	Used      bool            `gorm:"not null;default:false"`
	UsedAt    *time.Time      `gorm:"type:timestamp"`
	CreatedAt time.Time       `gorm:"type:timestamp;not null"`
	UpdatedAt time.Time       `gorm:"type:timestamp;not null"`
	DeletedAt gorm.DeletedAt  `gorm:"type:timestamp;index"`
}

// BeforeCreate hook generates UUID if not set.
func (b *BackupCode) BeforeCreate(tx *gorm.DB) error {
	if b.ID == googleUuid.Nil {
		b.ID = googleUuid.New()
	}

	return nil
}

// TableName returns the table name for BackupCode.
func (BackupCode) TableName() string {
	return "backup_codes"
}
