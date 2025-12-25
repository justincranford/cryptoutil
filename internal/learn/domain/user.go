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
type User struct {
	ID           googleUuid.UUID `gorm:"type:text;primaryKey"`
	Username     string          `gorm:"type:text;uniqueIndex;not null"`
	PasswordHash string          `gorm:"type:text;not null"`
	PublicKey    []byte          `gorm:"type:bytea;not null"` // ECDH public key for message encryption.
	CreatedAt    time.Time       `gorm:"autoCreateTime"`
	UpdatedAt    time.Time       `gorm:"autoUpdateTime"`
}

// TableName returns the database table name for User.
func (User) TableName() string {
	return "users"
}
