// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// SecretStatus represents the status of a client secret or API key.
type SecretStatus string

// Secret status constants.
const (
	// SecretStatusActive means the secret is active and usable.
	SecretStatusActive SecretStatus = "active"
	// SecretStatusExpired means the secret has expired (past grace period).
	SecretStatusExpired SecretStatus = "expired"
	// SecretStatusRevoked means the secret has been manually revoked.
	SecretStatusRevoked SecretStatus = "revoked"
)

// ClientSecretVersion represents a versioned client secret with lifecycle metadata.
type ClientSecretVersion struct {
	// Primary identifier.
	ID googleUuid.UUID `gorm:"type:text;primaryKey" json:"id"`

	// Secret relationship.
	ClientID googleUuid.UUID `gorm:"type:text;index;not null" json:"client_id"` // Associated client ID.

	// Secret data (hashed).
	SecretHash string `gorm:"not null" json:"-"` // Hashed secret (PBKDF2-HMAC-SHA256, FIPS-compliant).

	// Secret metadata.
	Version   int          `gorm:"not null" json:"version"`                 // Secret version number (monotonic).
	Status    SecretStatus `gorm:"not null;default:'active'" json:"status"` // Secret status.
	CreatedAt time.Time    `gorm:"not null" json:"created_at"`              // Creation timestamp.
	ExpiresAt *time.Time   `gorm:"index" json:"expires_at,omitempty"`       // Expiration timestamp (grace period end).
	RevokedAt *time.Time   `gorm:"index" json:"revoked_at,omitempty"`       // Revocation timestamp (manual revoke).
	RotatedAt *time.Time   `json:"rotated_at,omitempty"`                    // Rotation timestamp (new version created).
	DeletedAt *time.Time   `gorm:"index" json:"deleted_at,omitempty"`       // Soft delete timestamp.
	CreatedBy string       `gorm:"index" json:"created_by,omitempty"`       // Initiator (user ID or "system").
	RevokedBy string       `gorm:"index" json:"revoked_by,omitempty"`       // Revoker (user ID or "system").
}

// BeforeCreate generates UUID for new secret versions.
func (s *ClientSecretVersion) BeforeCreate(_ *gorm.DB) error {
	if s.ID == googleUuid.Nil {
		s.ID = googleUuid.Must(googleUuid.NewV7())
	}

	return nil
}

// TableName returns the table name for ClientSecretVersion entities.
func (ClientSecretVersion) TableName() string {
	return "client_secret_versions"
}

// IsValid checks if the secret is currently valid (active and not expired).
func (s *ClientSecretVersion) IsValid(now time.Time) bool {
	if s.Status != SecretStatusActive {
		return false
	}

	if s.ExpiresAt != nil && now.After(*s.ExpiresAt) {
		return false
	}

	return true
}

// MarkExpired marks the secret as expired.
func (s *ClientSecretVersion) MarkExpired() {
	s.Status = SecretStatusExpired
}

// MarkRevoked marks the secret as revoked with revoker metadata.
func (s *ClientSecretVersion) MarkRevoked(revokedBy string) {
	now := time.Now().UTC()
	s.Status = SecretStatusRevoked
	s.RevokedAt = &now
	s.RevokedBy = revokedBy
}
