// Copyright (c) 2025 Justin Cranford
//
//

// Package domain provides domain models for the cipher-im service.
package domain

import (
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// Message represents an encrypted message in the cipher-im system.
//
// Uses JWE JSON format (NOT Compact Serialization):
// - Multi-recipient encryption: N recipient AES256 JWKs (one per RecipientUserID)
// - Each recipient gets their own encrypted key copy in the JWE "recipients" array
// - Encryption: EncryptBytesWithContext(plaintext, []RecipientJWK) → JWE JSON
// - Decryption: DecryptBytesWithContext(jweJSON, recipientJWK) → plaintext
//
// Algorithm: enc=A256GCM (content encryption), alg=A256GCMKW (key wrapping per recipient).
//
// UpdatedAt field tracks last modification time (GORM auto-updates on save).
// Future usage: Track message edit history, read status changes.
type Message struct {
	ID        googleUuid.UUID `gorm:"type:text;primaryKey"`     // UUIDv7
	SenderID  googleUuid.UUID `gorm:"type:text;not null;index"` // UUIDv7
	JWE       string          `gorm:"type:text;not null"`       // JWE JSON format (multi-recipient)
	CreatedAt time.Time       `gorm:"autoCreateTime"`
	ReadAt    *time.Time      `gorm:"default:null;index"`

	// Relationships.
	Sender cryptoutilAppsTemplateServiceServerRepository.User `gorm:"foreignKey:SenderID"` // UUIDv7
}

// TableName returns the database table name for Message.
func (Message) TableName() string {
	return "messages"
}
