// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// UserJWK represents a per-user encryption key stored as JWK.
// Algorithm: ECDH-ES (key agreement) + A256GCM (content encryption).
type UserJWK struct {
	ID         googleUuid.UUID `gorm:"type:text;primaryKey"`
	UserID     googleUuid.UUID `gorm:"type:text;not null;index"`
	JWKJson    string          `gorm:"type:text;not null"` // JWK in JSON format.
	Algorithm  string          `gorm:"type:text;not null;default:'ECDH-ES'"`
	Encryption string          `gorm:"type:text;not null;default:'A256GCM'"`
	KeyID      string          `gorm:"type:text;not null"` // kid claim from JWK.
	IsActive   bool            `gorm:"not null;default:true;index"`
	CreatedAt  time.Time       `gorm:"autoCreateTime"`
	UpdatedAt  time.Time       `gorm:"autoUpdateTime"`
}

// TableName returns the database table name.
func (UserJWK) TableName() string {
	return "users_jwks"
}

// UserMessageJWK represents a per-user/message encryption key stored as JWK.
// Algorithm: dir (direct encryption) + A256GCM (content encryption).
type UserMessageJWK struct {
	ID         googleUuid.UUID `gorm:"type:text;primaryKey"`
	UserID     googleUuid.UUID `gorm:"type:text;not null;index"`
	MessageID  googleUuid.UUID `gorm:"type:text;not null;index"`
	JWKJson    string          `gorm:"type:text;not null"` // JWK in JSON format.
	Algorithm  string          `gorm:"type:text;not null;default:'dir'"`
	Encryption string          `gorm:"type:text;not null;default:'A256GCM'"`
	KeyID      string          `gorm:"type:text;not null"` // kid claim from JWK.
	CreatedAt  time.Time       `gorm:"autoCreateTime"`
	UpdatedAt  time.Time       `gorm:"autoUpdateTime"`
}

// TableName returns the database table name.
func (UserMessageJWK) TableName() string {
	return "users_messages_jwks"
}

// MessageJWK represents a per-message encryption key stored as JWK.
// Algorithm: dir (direct encryption) + A256GCM (content encryption).
type MessageJWK struct {
	ID         googleUuid.UUID `gorm:"type:text;primaryKey"`
	MessageID  googleUuid.UUID `gorm:"type:text;not null;index"`
	JWKJson    string          `gorm:"type:text;not null"` // JWK in JSON format.
	Algorithm  string          `gorm:"type:text;not null;default:'dir'"`
	Encryption string          `gorm:"type:text;not null;default:'A256GCM'"`
	KeyID      string          `gorm:"type:text;not null"` // kid claim from JWK.
	IsActive   bool            `gorm:"not null;default:true;index"`
	CreatedAt  time.Time       `gorm:"autoCreateTime"`
	UpdatedAt  time.Time       `gorm:"autoUpdateTime"`
}

// TableName returns the database table name.
func (MessageJWK) TableName() string {
	return "messages_jwks"
}
