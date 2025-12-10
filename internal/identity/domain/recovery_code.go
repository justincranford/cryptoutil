// Copyright (c) 2025 Iwan van der Kleijn
// SPDX-License-Identifier: MIT

package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// RecoveryCode represents a single-use backup authentication code.
// Recovery codes provide emergency access when users lose their primary MFA factors.
type RecoveryCode struct {
	ID        googleUuid.UUID `gorm:"type:text;primaryKey"`
	UserID    googleUuid.UUID `gorm:"type:text;index;not null"`
	CodeHash  string          `gorm:"type:text;not null"` // bcrypt hash of code.
	Used      bool            `gorm:"not null;default:false;index"`
	UsedAt    *time.Time      `gorm:"index"`
	CreatedAt time.Time       `gorm:"not null"`
	ExpiresAt time.Time       `gorm:"not null;index"`
}

// IsExpired checks if the recovery code has expired.
func (r *RecoveryCode) IsExpired() bool {
	return time.Now().UTC().After(r.ExpiresAt)
}

// IsUsed checks if the recovery code has already been used.
func (r *RecoveryCode) IsUsed() bool {
	return r.Used
}

// MarkAsUsed marks the recovery code as used and sets the usage timestamp.
func (r *RecoveryCode) MarkAsUsed() {
	r.Used = true
	now := time.Now().UTC()
	r.UsedAt = &now
}
