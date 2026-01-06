// Copyright (c) 2025 Justin Cranford
//
//

// Package domain defines cipher-im domain models.
package domain

import (
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilTemplateRealms "cryptoutil/internal/template/server/realms"
)

// User represents a cipher-im user account.
//
// Simplified 3-table design:
// - Password stored as PBKDF2-HMAC-SHA256 hash
// - No ECDH keys stored in users table (keys are ephemeral per-message).
//
// UpdatedAt field tracks last modification time (GORM auto-updates on save).
// Future usage: Track last login time, password changes, profile updates.
type User struct {
	ID           googleUuid.UUID `gorm:"type:text;primaryKey"` // UUIDv7
	Username     string          `gorm:"type:text;uniqueIndex;not null"`
	PasswordHash string          `gorm:"type:text;not null"` // PBKDF2-HMAC-SHA256 hash
	CreatedAt    time.Time       `gorm:"autoCreateTime"`
}

// TableName returns the database table name for User.
func (User) TableName() string {
	return "users"
}

// GetID returns the user's unique identifier.
func (u *User) GetID() googleUuid.UUID {
	return u.ID
}

// GetUsername returns the user's username.
func (u *User) GetUsername() string {
	return u.Username
}

// GetPasswordHash returns the user's password hash.
func (u *User) GetPasswordHash() string {
	return u.PasswordHash
}

// SetID sets the user's unique identifier.
func (u *User) SetID(id googleUuid.UUID) {
	u.ID = id
}

// SetUsername sets the user's username.
func (u *User) SetUsername(username string) {
	u.Username = username
}

// SetPasswordHash sets the user's password hash.
func (u *User) SetPasswordHash(hash string) {
	u.PasswordHash = hash
}

// Compile-time check that User implements realms.UserModel interface.
var _ cryptoutilTemplateRealms.UserModel = (*User)(nil)
