// Copyright (c) 2025 Justin Cranford
//

package realms

import (
	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims for authentication.
//
// Structure:
// - UserID: UUIDv7 string identifying the authenticated user
// - Username: Username for logging/debugging
// - RegisteredClaims: Standard JWT fields (ExpiresAt, IssuedAt, etc.)
//
// Example Token Payload:
//
//	{
//	  "user_id": "01JGE123456789ABCDEFGHIJK",
//	  "username": "testuser",
//	  "exp": 1735948800,
//	  "iat": 1735948700
//	}
type Claims struct {
	// UserID is the UUIDv7 string identifying the authenticated user.
	// Used by handlers to retrieve user context from c.Locals(ContextKeyUserID).
	UserID string `json:"user_id"`

	// Username is the username of the authenticated user (for logging/debugging).
	Username string `json:"username"`

	// RegisteredClaims contains standard JWT fields.
	// - ExpiresAt: Token expiration time (Unix timestamp)
	// - IssuedAt: Token issuance time (Unix timestamp)
	jwt.RegisteredClaims
}
