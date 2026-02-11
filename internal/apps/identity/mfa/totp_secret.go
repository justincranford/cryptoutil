// Copyright (c) 2025 Justin Cranford
//
//

package mfa

import (
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// TOTPSecret stores TOTP MFA configuration for a user.
type TOTPSecret struct {
	ID              googleUuid.UUID `gorm:"type:text;primaryKey"`
	UserID          googleUuid.UUID `gorm:"type:text;uniqueIndex;not null"`
	Secret          string          `gorm:"type:text;not null" json:"-"`
	Algorithm       string          `gorm:"type:text;not null;default:SHA1"`
	Digits          int             `gorm:"not null;default:6"`
	Period          int             `gorm:"not null;default:30"`
	Verified        bool            `gorm:"not null;default:false"`
	RecoveryEnabled bool            `gorm:"not null;default:true"`
	LastUsedAt      time.Time       `gorm:"type:timestamp"`
	FailedAttempts  int             `gorm:"not null;default:0"`
	LockedUntil     time.Time       `gorm:"type:timestamp"`
	CreatedAt       time.Time       `gorm:"type:timestamp;not null"`
	UpdatedAt       time.Time       `gorm:"type:timestamp;not null"`
	DeletedAt       gorm.DeletedAt  `gorm:"type:timestamp;index"`
}

// BeforeCreate hook generates UUID if not set.
func (t *TOTPSecret) BeforeCreate(tx *gorm.DB) error {
	if t.ID == googleUuid.Nil {
		t.ID = googleUuid.New()
	}

	return nil
}

// TableName returns the table name for TOTPSecret.
func (TOTPSecret) TableName() string {
	return "totp_secrets"
}
