// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// PushedAuthorizationRequest represents a pushed authorization request (RFC 9126).
// PAR allows OAuth clients to push authorization request parameters directly to the
// authorization server before redirecting the user agent, providing request integrity,
// confidentiality, and protection against parameter tampering.
type PushedAuthorizationRequest struct {
	ID         googleUuid.UUID `gorm:"type:text;primaryKey"`
	RequestURI string          `gorm:"type:text;uniqueIndex;not null"` // urn:ietf:params:oauth:request_uri:xxx
	ClientID   googleUuid.UUID `gorm:"type:text;index;not null"`

	// Stored authorization parameters (from OAuth 2.1 /authorize request).
	ResponseType        string `gorm:"type:text;not null"`
	RedirectURI         string `gorm:"type:text;not null"`
	Scope               string `gorm:"type:text"`
	State               string `gorm:"type:text"`
	CodeChallenge       string `gorm:"type:text;not null"`
	CodeChallengeMethod string `gorm:"type:text;not null"`
	Nonce               string `gorm:"type:text"`

	// Additional parameters as JSON blob (for future extensibility).
	AdditionalParams string `gorm:"type:text;serializer:json"`

	// Lifecycle tracking.
	Used      bool       `gorm:"not null;default:false;index"`
	ExpiresAt time.Time  `gorm:"not null;index"`
	CreatedAt time.Time  `gorm:"not null"`
	UsedAt    *time.Time `gorm:"index"`
}

// IsExpired checks if the pushed authorization request has expired.
// PAR entries have a short lifetime (90 seconds by default) to reduce
// the window for attacks and prevent stale authorization requests.
func (p *PushedAuthorizationRequest) IsExpired() bool {
	return time.Now().UTC().After(p.ExpiresAt)
}

// IsUsed checks if the pushed authorization request has already been used.
// request_uri values are single-use only per RFC 9126 to prevent replay attacks.
func (p *PushedAuthorizationRequest) IsUsed() bool {
	return p.Used
}

// MarkAsUsed marks the pushed authorization request as used and records the timestamp.
// This enforces single-use semantics for request_uri values.
func (p *PushedAuthorizationRequest) MarkAsUsed() {
	p.Used = true
	now := time.Now().UTC()
	p.UsedAt = &now
}
