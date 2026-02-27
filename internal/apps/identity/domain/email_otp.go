// Copyright (c) 2025 Justin Cranford

// Package domain defines core domain models for the identity service.
package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// EmailOTP represents an email-based one-time password for MFA.
type EmailOTP struct {
	ID        googleUuid.UUID `gorm:"type:text;primaryKey" json:"id"`
	UserID    googleUuid.UUID `gorm:"type:text;index;not null" json:"user_id"`
	CodeHash  string          `gorm:"type:text;not null" json:"-"` // PBKDF2-HMAC-SHA256 hash of OTP code (FIPS-compliant).
	Used      bool            `gorm:"default:false;not null" json:"used"`
	UsedAt    *time.Time      `gorm:"type:timestamp" json:"used_at,omitempty"`
	CreatedAt time.Time       `gorm:"type:timestamp;not null" json:"created_at"`
	ExpiresAt time.Time       `gorm:"type:timestamp;not null;index" json:"expires_at"`
}

// TableName specifies the database table name.
func (EmailOTP) TableName() string {
	return "email_otps"
}

// IsExpired checks if the OTP has expired.
func (e *EmailOTP) IsExpired() bool {
	return time.Now().UTC().After(e.ExpiresAt)
}

// IsUsed checks if the OTP has been used.
func (e *EmailOTP) IsUsed() bool {
	return e.Used
}

// MarkAsUsed marks the OTP as used.
func (e *EmailOTP) MarkAsUsed() {
	now := time.Now().UTC()
	e.Used = true
	e.UsedAt = &now
}
