// Copyright (c) 2025 Justin Cranford

package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// Key represents a cryptographic key for signing or encryption.
type Key struct {
	// Primary identifier.
	ID googleUuid.UUID `gorm:"type:text;primaryKey" json:"id"`

	// Key metadata.
	Usage      string `gorm:"not null;index" json:"usage"`      // "signing" or "encryption".
	Algorithm  string `gorm:"not null" json:"algorithm"`        // Key algorithm (RS256, ES256, etc.).
	PrivateKey string `gorm:"type:text;not null" json:"-"`      // Private key (JWK format).
	PublicKey  string `gorm:"type:text" json:"public_key"`      // Public key (JWK format, asymmetric only).
	Active     bool   `gorm:"default:true;index" json:"active"` // Active for new operations.

	// Lifecycle timestamps.
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"` // Key creation time.
	ExpiresAt time.Time      `gorm:"index" json:"expires_at"`          // Key expiration time.
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"` // Last update time.
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName returns the table name for Key entity.
func (k *Key) TableName() string {
	return "keys"
}

// BeforeCreate sets ID if not already set.
func (k *Key) BeforeCreate(_ *gorm.DB) error {
	if k.ID == googleUuid.Nil {
		k.ID = googleUuid.Must(googleUuid.NewV7())
	}

	return nil
}

// IsExpired checks if the key has expired.
func (k *Key) IsExpired() bool {
	return time.Now().UTC().After(k.ExpiresAt)
}
