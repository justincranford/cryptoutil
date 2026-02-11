// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// ConsentDecision represents a user's consent to grant a client access to their data.
// Consent decisions can be reused for subsequent authorization requests to the same client/scope combination.
type ConsentDecision struct {
	// Primary key.
	ID googleUuid.UUID `gorm:"type:text;primaryKey" json:"id"`

	// User and client information.
	UserID   googleUuid.UUID `gorm:"type:text;not null;index" json:"user_id"`
	ClientID string          `gorm:"type:text;not null;index" json:"client_id"`

	// Granted scopes.
	Scope string `gorm:"type:text;not null" json:"scope"`

	// Consent metadata.
	GrantedAt time.Time `gorm:"not null" json:"granted_at"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`

	// Revocation tracking.
	RevokedAt *time.Time `gorm:"index" json:"revoked_at,omitempty"`
}

// TableName returns the database table name for ConsentDecision.
func (ConsentDecision) TableName() string {
	return "consent_decisions"
}

// IsExpired checks if the consent decision has expired.
func (c *ConsentDecision) IsExpired() bool {
	return time.Now().UTC().After(c.ExpiresAt)
}

// IsRevoked checks if the consent decision has been revoked.
func (c *ConsentDecision) IsRevoked() bool {
	return c.RevokedAt != nil
}
