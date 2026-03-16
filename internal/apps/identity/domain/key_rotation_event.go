// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// KeyRotationEvent represents an audit log entry for key management operations.
type KeyRotationEvent struct {
	// Primary identifier.
	ID googleUuid.UUID `gorm:"type:text;primaryKey" json:"id"`

	// Event metadata.
	EventType string    `gorm:"index;not null" json:"event_type"` // Event type (rotation, revocation, expiration).
	KeyType   string    `gorm:"index;not null" json:"key_type"`   // Key type (client_secret, jwk, api_key).
	KeyID     string    `gorm:"index;not null" json:"key_id"`     // Key identifier (client ID, JWK kid, API key ID).
	Timestamp time.Time `gorm:"index;not null" json:"timestamp"`  // Event timestamp.

	// Operation details.
	Initiator     string     `gorm:"index;not null" json:"initiator"`                   // Initiator (user ID or "system").
	OldKeyVersion *int       `json:"old_key_version,omitempty"`                         // Previous key version (if rotation).
	NewKeyVersion *int       `json:"new_key_version,omitempty"`                         // New key version (if rotation).
	GracePeriod   *string    `json:"grace_period,omitempty"`                            // Grace period duration (e.g., "24h", "7d").
	Reason        string     `json:"reason,omitempty"`                                  // Reason for operation (e.g., "scheduled", "emergency").
	Metadata      string     `gorm:"type:text" json:"metadata,omitempty"`               // Additional metadata (JSON blob).
	Success       *bool      `gorm:"type:boolean;not null;default:true" json:"success"` // Operation success status.
	ErrorMessage  string     `gorm:"type:text" json:"error_message,omitempty"`          // Error message (if failed).
	DeletedAt     *time.Time `gorm:"index" json:"deleted_at,omitempty"`                 // Soft delete timestamp.
}

// BeforeCreate generates UUID for new rotation events.
func (e *KeyRotationEvent) BeforeCreate(_ *gorm.DB) error {
	if e.ID == googleUuid.Nil {
		e.ID = googleUuid.Must(googleUuid.NewV7())
	}

	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}

	return nil
}

// TableName returns the table name for KeyRotationEvent entities.
func (KeyRotationEvent) TableName() string {
	return "key_rotation_events"
}

// Event type constants for KeyRotationEvent.
const (
	EventTypeRotation   = "rotation"   // Key rotation event.
	EventTypeRevocation = "revocation" // Key revocation event.
	EventTypeExpiration = "expiration" // Key expiration event.
)

// Key type constants for KeyRotationEvent.
const (
	KeyTypeClientSecret = "client_secret" // OAuth 2.1 client secret.
	KeyTypeJWK          = "jwk"           // JSON Web Key (signing/encryption).
	KeyTypeAPIKey       = "api_key"       // Service-to-service API key.
)
