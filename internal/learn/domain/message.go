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
//
// NOTE: EncryptedContent and Nonce are stored per-receiver in MessageReceiver table,
// not in Message table. This is because each receiver gets a different encrypted copy
// based on their unique public key (ECDH produces different shared secret per receiver).
type Message struct {
	ID        googleUuid.UUID `gorm:"type:text;primaryKey"`
	SenderID  googleUuid.UUID `gorm:"type:text;not null;index"`
	CreatedAt time.Time       `gorm:"autoCreateTime"`

	// Relationships.
	Sender    User              `gorm:"foreignKey:SenderID"`
	Receivers []MessageReceiver `gorm:"foreignKey:MessageID"`
}

// TableName returns the database table name for Message.
func (Message) TableName() string {
	return "messages"
}

// MessageReceiver represents a message recipient with their encrypted copy.
//
// Each receiver gets:
// 1. Sender's ephemeral ECDH public key (to derive shared secret)
// 2. Their own encrypted copy of the message (unique per receiver)
// 3. Their own GCM nonce (unique per receiver)
//
// This enables multi-receiver encryption where each receiver can decrypt
// independently using their private key + sender's ephemeral public key.
type MessageReceiver struct {
	ID               googleUuid.UUID `gorm:"type:text;primaryKey"`
	MessageID        googleUuid.UUID `gorm:"type:text;not null;index"`
	ReceiverID       googleUuid.UUID `gorm:"type:text;not null;index"`
	SenderPubKey     []byte          `gorm:"type:bytea;not null"` // Sender's ephemeral ECDH public key.
	EncryptedContent []byte          `gorm:"type:bytea;not null"` // AES-256-GCM ciphertext (unique per receiver).
	Nonce            []byte          `gorm:"type:bytea;not null"` // GCM nonce (unique per receiver).
	ReceivedAt       *time.Time      `gorm:"default:null"`

	// Relationships.
	Message  Message `gorm:"foreignKey:MessageID"`
	Receiver User    `gorm:"foreignKey:ReceiverID"`
}

// TableName returns the database table name for MessageReceiver.
func (MessageReceiver) TableName() string {
	return "message_receivers"
}
