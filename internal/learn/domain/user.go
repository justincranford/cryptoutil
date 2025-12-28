// Copyright (c) 2025 Justin Cranford
//
//

// Package domain defines learn-im domain models.
package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// User represents a learn-im user account.
//
// Simplified 3-table design:
// - Password stored as PBKDF2-HMAC-SHA256 hash
// - No ECDH keys stored in users table (keys are ephemeral per-message).
type User struct {
	ID           googleUuid.UUID `gorm:"type:text;primaryKey"` // UUIDv7
	Username     string          `gorm:"type:text;uniqueIndex;not null"`
	PasswordHash string          `gorm:"type:text;not null"` // PBKDF2-HMAC-SHA256 hash
	CreatedAt    time.Time       `gorm:"autoCreateTime"`
	UpdatedAt    time.Time       `gorm:"autoUpdateTime"`
}

// TableName returns the database table name for User.
func (User) TableName() string {
	return "users"
}
