// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// MessageRecipientJWK represents a per-recipient decryption key for a message.
//
// Multi-recipient pattern:
// - Each message can have N recipients
// - Each recipient gets their own encrypted JWK for decrypting the message
// - JWK encrypted with alg=dir (direct encryption), enc=A256GCM
// - Stored in JSON format (NOT Compact Serialization).
type MessageRecipientJWK struct {
	ID           googleUuid.UUID `gorm:"type:text;primaryKey"`     // UUIDv7
	RecipientID  googleUuid.UUID `gorm:"type:text;not null;index"` // UUIDv7
	MessageID    googleUuid.UUID `gorm:"type:text;not null;index"` // UUIDv7
	EncryptedJWK string          `gorm:"type:text;not null"`       // Encrypted JWK in JSON format (enc=A256GCM, alg=dir)
	CreatedAt    time.Time       `gorm:"autoCreateTime"`

	// Relationships.
	Recipient cryptoutilAppsTemplateServiceServerRepository.User `gorm:"foreignKey:RecipientID"` // UUIDv7
	Message   Message                                            `gorm:"foreignKey:MessageID"`   // UUIDv7
}

// TableName returns the database table name.
func (MessageRecipientJWK) TableName() string {
	return "messages_recipient_jwks"
}
