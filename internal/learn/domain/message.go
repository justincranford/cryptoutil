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
// Messages use hybrid encryption:
// - Sender's ECDH private key + Receiver's ECDH public key → shared secret
// - Shared secret + HKDF → AES-256 key
// - AES-256-GCM encrypts message content.
type Message struct {
	ID               googleUuid.UUID `gorm:"type:text;primaryKey"`
	SenderID         googleUuid.UUID `gorm:"type:text;not null;index"`
	EncryptedContent []byte          `gorm:"type:bytea;not null"` // AES-256-GCM ciphertext.
	Nonce            []byte          `gorm:"type:bytea;not null"` // GCM nonce.
	CreatedAt        time.Time       `gorm:"autoCreateTime"`

	// Relationships.
	Sender    User              `gorm:"foreignKey:SenderID"`
	Receivers []MessageReceiver `gorm:"foreignKey:MessageID"`
}

// TableName returns the database table name for Message.
func (Message) TableName() string {
	return "messages"
}

// MessageReceiver represents a message recipient with their encrypted key.
//
// Each receiver gets the sender's ephemeral ECDH public key to derive the shared secret.
type MessageReceiver struct {
	ID           googleUuid.UUID `gorm:"type:text;primaryKey"`
	MessageID    googleUuid.UUID `gorm:"type:text;not null;index"`
	ReceiverID   googleUuid.UUID `gorm:"type:text;not null;index"`
	SenderPubKey []byte          `gorm:"type:bytea;not null"` // Sender's ephemeral ECDH public key.
	ReceivedAt   *time.Time      `gorm:"default:null"`

	// Relationships.
	Message  Message `gorm:"foreignKey:MessageID"`
	Receiver User    `gorm:"foreignKey:ReceiverID"`
}

// TableName returns the database table name for MessageReceiver.
func (MessageReceiver) TableName() string {
	return "message_receivers"
}
