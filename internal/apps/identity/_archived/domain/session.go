// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// Session represents a user authentication session.
type Session struct {
	// Primary identifier.
	ID googleUuid.UUID `gorm:"type:text;primaryKey" json:"id"`

	// Session identification.
	SessionID string `gorm:"uniqueIndex;not null" json:"session_id"` // Session identifier.

	// Session associations.
	UserID   googleUuid.UUID `gorm:"type:text;index;not null" json:"user_id"`    // Associated user.
	ClientID NullableUUID    `gorm:"type:text;index" json:"client_id,omitempty"` // Associated client (if applicable).

	// Session metadata.
	IPAddress string `json:"ip_address,omitempty"` // Client IP address.
	UserAgent string `json:"user_agent,omitempty"` // Client user agent.

	// Session lifetime.
	IssuedAt   time.Time `gorm:"index;not null" json:"issued_at"`  // Session creation time.
	ExpiresAt  time.Time `gorm:"index;not null" json:"expires_at"` // Session expiration time.
	LastSeenAt time.Time `json:"last_seen_at"`                     // Last activity time.

	// Session status.
	Active    *bool      `gorm:"type:boolean;default:true;not null" json:"active"` // Session active status.
	RevokedAt *time.Time `gorm:"column:revoked_at" json:"revoked_at,omitempty"`    // Session revocation time.

	AuthenticationMethods []string  `gorm:"serializer:json" json:"authentication_methods"` // Used authentication methods.
	AuthenticationTime    time.Time `json:"authentication_time"`                           // Authentication completion time.

	// OAuth 2.1 / OIDC metadata.
	Nonce         string   `gorm:"column:nonce" json:"nonce,omitempty"` // OIDC nonce value.
	CodeChallenge string   `gorm:"column:code_challenge" json:"-"`      // PKCE code challenge.
	GrantedScopes []string `gorm:"serializer:json" json:"scopes"`       // Granted scopes.

	// GORM timestamps.
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate generates UUID for new sessions.
func (s *Session) BeforeCreate(_ *gorm.DB) error {
	if s.ID == googleUuid.Nil {
		s.ID = googleUuid.Must(googleUuid.NewV7())
	}

	if s.SessionID == "" {
		s.SessionID = googleUuid.Must(googleUuid.NewV7()).String()
	}

	return nil
}

// TableName returns the table name for Session entities.
func (Session) TableName() string {
	return "sessions"
}

// IsExpired checks if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().UTC().After(s.ExpiresAt)
}

// IsValid checks if the session is valid (not expired and active).
func (s *Session) IsValid() bool {
	return !s.IsExpired() && s.Active != nil && *s.Active
}
