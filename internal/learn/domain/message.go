// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// Message represents an encrypted message in the learn-im system.
//
// Messages use JWE Compact Serialization format (eyJ...):
// - Each message is encrypted with a per-message JWK
// - JWK is stored in messages_jwks table
// - JWE compact format: BASE64URL(header).BASE64URL(key).BASE64URL(iv).BASE64URL(ciphertext).BASE64URL(tag)
//
// Algorithm: dir (direct encryption) + A256GCM (content encryption).
type Message struct {
	ID            googleUuid.UUID `gorm:"type:text;primaryKey"`
	SenderID      googleUuid.UUID `gorm:"type:text;not null;index"`
	RecipientID   googleUuid.UUID `gorm:"type:text;not null;index"`
	JWECompact    string          `gorm:"type:text;not null"` // JWE Compact Serialization format.
	KeyID         string          `gorm:"type:text;not null;index"` // References messages_jwks.key_id.
	CreatedAt     time.Time       `gorm:"autoCreateTime"`
	ReadAt        *time.Time      `gorm:"default:null;index"`

	// Relationships.
	Sender    User `gorm:"foreignKey:SenderID"`
	Recipient User `gorm:"foreignKey:RecipientID"`
	MessageJWK MessageJWK `gorm:"foreignKey:KeyID;references:KeyID"`
}

// TableName returns the database table name for Message.
func (Message) TableName() string {
	return "messages"
}
