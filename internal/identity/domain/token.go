// Copyright (c) 2025 Justin Cranford
//
//

package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// TokenType represents the type of token.
type TokenType string

// Token type constants.
const (
	// TokenTypeAccess is an access token.
	TokenTypeAccess TokenType = "access"
	// TokenTypeRefresh is a refresh token.
	TokenTypeRefresh TokenType = "refresh"
	// TokenTypeID is an ID token (OIDC).
	TokenTypeID TokenType = "id"
)

// TokenFormat represents the format of the token.
type TokenFormat string

// Token format constants.
const (
	// TokenFormatJWS is JSON Web Signature (signed).
	TokenFormatJWS TokenFormat = "jws"
	// TokenFormatJWE is JSON Web Encryption (encrypted).
	TokenFormatJWE TokenFormat = "jwe"
	// TokenFormatUUID is an opaque UUID token.
	TokenFormatUUID TokenFormat = "uuid"
)

// Token represents an OAuth 2.1 / OIDC token.
type Token struct {
	// Primary identifier.
	ID googleUuid.UUID `gorm:"type:text;primaryKey" json:"id"`

	// Token identification.
	TokenValue  string      `gorm:"uniqueIndex;not null" json:"-"`    // Token value (JWT or UUID).
	TokenType   TokenType   `gorm:"index;not null" json:"token_type"` // Token type.
	TokenFormat TokenFormat `gorm:"not null" json:"token_format"`     // Token format.

	// Token associations.
	ClientID googleUuid.UUID `gorm:"type:text;primaryKey" json:"client_id"`    // Associated client.
	UserID   NullableUUID    `gorm:"type:text;index" json:"user_id,omitempty"` // Associated user (if applicable).

	// Token metadata.
	Scopes    []string  `gorm:"serializer:json" json:"scopes"`    // Granted scopes.
	IssuedAt  time.Time `gorm:"index;not null" json:"issued_at"`  // Token issuance time.
	ExpiresAt time.Time `gorm:"index;not null" json:"expires_at"` // Token expiration time.
	NotBefore time.Time `json:"not_before,omitempty"`             // Token not valid before time.

	// Token status.
	Revoked   IntBool    `gorm:"type:integer;index;default:0" json:"revoked"` // Token revocation status (INTEGER for cross-DB compatibility).
	RevokedAt *time.Time `json:"revoked_at,omitempty"`                        // Token revocation time.

	// Refresh token association (optional).
	RefreshTokenID NullableUUID `gorm:"type:text;index" json:"refresh_token_id,omitempty"` // Associated refresh token.

	// PKCE code challenge (for authorization codes).
	CodeChallenge       string `json:"-"`                               // PKCE code challenge.
	CodeChallengeMethod string `json:"code_challenge_method,omitempty"` // PKCE code challenge method.

	// GORM timestamps.
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate generates UUID for new tokens.
func (t *Token) BeforeCreate(_ *gorm.DB) error {
	if t.ID == googleUuid.Nil {
		t.ID = googleUuid.Must(googleUuid.NewV7())
	}

	return nil
}

// TableName returns the table name for Token entities.
func (Token) TableName() string {
	return "tokens"
}

// IsExpired checks if the token has expired.
func (t *Token) IsExpired() bool {
	return time.Now().UTC().After(t.ExpiresAt)
}

// IsValid checks if the token is valid (not expired and not revoked).
func (t *Token) IsValid() bool {
	return !t.IsExpired() && !t.Revoked.Bool()
}
