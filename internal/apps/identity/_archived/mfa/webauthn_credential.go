// Copyright (c) 2025 Justin Cranford
//
//

package mfa

import (
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// WebAuthnCredential represents a WebAuthn credential stored in the database.
// Each credential is associated with a user and contains the public key
// and metadata needed for authentication ceremonies.
type WebAuthnCredential struct {
	ID              googleUuid.UUID `gorm:"type:text;primaryKey"`
	UserID          googleUuid.UUID `gorm:"type:text;index;not null"`
	CredentialID    []byte          `gorm:"type:blob;not null"`
	PublicKey       []byte          `gorm:"type:blob;not null"`
	AttestationType string          `gorm:"type:text;not null"`
	Transports      string          `gorm:"type:text"` // JSON array of transport types
	SignCount       uint32          `gorm:"not null;default:0"`
	AAGUID          []byte          `gorm:"type:blob"`
	CloneWarning    bool            `gorm:"not null;default:false"`
	DisplayName     string          `gorm:"type:text;not null"` // User-friendly name
	CreatedAt       time.Time       `gorm:"not null"`
	UpdatedAt       time.Time       `gorm:"not null"`
	LastUsedAt      *time.Time      `gorm:"index"`
	DeletedAt       gorm.DeletedAt  `gorm:"index"`
}

// TableName returns the table name for GORM.
func (WebAuthnCredential) TableName() string {
	return "webauthn_credentials"
}

// BeforeCreate sets default values before creating a credential.
func (c *WebAuthnCredential) BeforeCreate(_ *gorm.DB) error {
	if c.ID == googleUuid.Nil {
		c.ID = googleUuid.Must(googleUuid.NewV7())
	}

	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now

	return nil
}

// BeforeUpdate updates the UpdatedAt timestamp.
func (c *WebAuthnCredential) BeforeUpdate(_ *gorm.DB) error {
	c.UpdatedAt = time.Now().UTC()

	return nil
}

// WebAuthnSession represents a temporary session for WebAuthn ceremonies.
// Sessions store challenge data during registration and authentication flows.
type WebAuthnSession struct {
	ID           googleUuid.UUID `gorm:"type:text;primaryKey"`
	UserID       googleUuid.UUID `gorm:"type:text;index;not null"`
	SessionData  []byte          `gorm:"type:blob;not null"` // JSON-encoded session data
	CeremonyType string          `gorm:"type:text;not null"` // registration or authentication
	CreatedAt    time.Time       `gorm:"not null"`
	ExpiresAt    time.Time       `gorm:"not null;index"`
}

// TableName returns the table name for GORM.
func (WebAuthnSession) TableName() string {
	return "webauthn_sessions"
}

// BeforeCreate sets default values before creating a session.
func (s *WebAuthnSession) BeforeCreate(_ *gorm.DB) error {
	if s.ID == googleUuid.Nil {
		s.ID = googleUuid.Must(googleUuid.NewV7())
	}

	s.CreatedAt = time.Now().UTC()

	return nil
}

// IsExpired checks if the session has expired.
func (s *WebAuthnSession) IsExpired() bool {
	return time.Now().UTC().After(s.ExpiresAt)
}

// WebAuthnCeremonyType represents the type of WebAuthn ceremony.
type WebAuthnCeremonyType string

const (
	// WebAuthnCeremonyRegistration represents a registration ceremony.
	WebAuthnCeremonyRegistration WebAuthnCeremonyType = "registration"
	// WebAuthnCeremonyAuthentication represents an authentication ceremony.
	WebAuthnCeremonyAuthentication WebAuthnCeremonyType = "authentication"
)
