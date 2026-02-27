// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// AuthorizationRequest represents a pending OAuth 2.1 authorization request.
// This domain model supports the authorization code flow with PKCE.
type AuthorizationRequest struct {
	// Primary key.
	ID googleUuid.UUID `gorm:"type:text;primaryKey" json:"id"`

	// Client information.
	ClientID    string `gorm:"type:text;not null;index" json:"client_id"`
	RedirectURI string `gorm:"type:text;not null" json:"redirect_uri"`

	// Request parameters.
	ResponseType string `gorm:"type:text;not null" json:"response_type"`
	Scope        string `gorm:"type:text" json:"scope"`
	State        string `gorm:"type:text" json:"state"`
	Nonce        string `gorm:"type:text" json:"nonce"`

	// PKCE parameters (OAuth 2.1 required).
	CodeChallenge       string `gorm:"type:text;not null" json:"code_challenge"`
	CodeChallengeMethod string `gorm:"type:text;not null;default:'S256'" json:"code_challenge_method"`

	// User information (populated after authentication).
	UserID NullableUUID `gorm:"type:text;index" json:"user_id"`

	// Authorization code (generated after consent).
	Code string `gorm:"type:text;uniqueIndex" json:"code"`

	// Request metadata.
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`

	// Consent status - use IntBool for cross-DB compatibility (INTEGER in both SQLite and PostgreSQL).
	ConsentGranted IntBool `gorm:"type:integer;not null;default:0" json:"consent_granted"`

	// Single-use enforcement - use IntBool for cross-DB compatibility.
	Used   IntBool    `gorm:"type:integer;not null;default:0;index" json:"used"`
	UsedAt *time.Time `gorm:"index" json:"used_at,omitempty"`
}

// TableName returns the database table name for AuthorizationRequest.
func (AuthorizationRequest) TableName() string {
	return "authorization_requests"
}

// IsExpired checks if the authorization request has expired.
func (a *AuthorizationRequest) IsExpired() bool {
	return time.Now().UTC().After(a.ExpiresAt)
}

// IsUsed checks if the authorization code has been used.
func (a *AuthorizationRequest) IsUsed() bool {
	return a.Used.Bool()
}
