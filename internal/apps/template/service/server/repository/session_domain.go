// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// SessionJWK represents an encrypted JWK for session token signing/encryption.
// Used by both browser and service session management.
//
// Algorithm determines the JWK type:
//   - JWE algorithms: Generate symmetric encryption JWK
//   - JWS algorithms: Generate asymmetric signing JWK
//   - OPAQUE: Not used (no JWKs stored for hash-based tokens)
//
// Selection: The latest active JWK is deterministically selected using max(CreatedAt).
// This eliminates race conditions in multi-instance deployments where multiple
// instances share the same database.
type SessionJWK struct {
	ID           googleUuid.UUID `gorm:"type:text;primaryKey"`
	EncryptedJWK string          `gorm:"type:text;not null"` // JWK encrypted with barrier layer.
	CreatedAt    time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP;index"`
	Algorithm    string          `gorm:"type:text;not null"`          // JWE or JWS algorithm identifier.
	Active       bool            `gorm:"not null;default:true;index"` // Active key for signing, historical keys for verification.
}

// BrowserSessionJWK represents a JWK for browser session tokens.
// Table: browser_session_jwks.
type BrowserSessionJWK struct {
	SessionJWK
}

// ServiceSessionJWK represents a JWK for service session tokens.
// Table: service_session_jwks.
type ServiceSessionJWK struct {
	SessionJWK
}

// TableName returns the database table name for BrowserSessionJWK.
func (BrowserSessionJWK) TableName() string {
	return "browser_session_jwks"
}

// TableName returns the database table name for ServiceSessionJWK.
func (ServiceSessionJWK) TableName() string {
	return "service_session_jwks"
}

// Session represents a session with expiration and metadata.
// Used by both browser and service session management.
//
// Session token types:
//   - JWE: TokenHash is NULL, session identified by jti claim in JWT
//   - JWS: TokenHash is NULL, session identified by jti claim in JWT
//   - OPAQUE: TokenHash stores hashed UUIDv7 token, session identified by hash lookup
type Session struct {
	ID           googleUuid.UUID `gorm:"type:text;primaryKey"`
	TenantID     googleUuid.UUID `gorm:"type:text;not null;index"` // Tenant identifier for multi-tenancy isolation.
	RealmID      googleUuid.UUID `gorm:"type:text;not null;index"` // Realm identifier within tenant.
	TokenHash    *string         `gorm:"type:text;index"`          // Hashed token (OPAQUE only), NULL for JWE/JWS.
	Expiration   time.Time       `gorm:"not null;index"`
	CreatedAt    time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP"`
	LastActivity *time.Time      // Last activity timestamp for idle timeout.
}

// BrowserSession represents a browser user session.
// Table: browser_sessions.
type BrowserSession struct {
	Session
	UserID *string `gorm:"type:text;index"` // User identifier (optional, depends on service implementation).
}

// ServiceSession represents a service-to-service session.
// Table: service_sessions.
type ServiceSession struct {
	Session
	ClientID *string `gorm:"type:text;index"` // Client identifier (optional, depends on service implementation).
}

// TableName returns the database table name for BrowserSession.
func (BrowserSession) TableName() string {
	return "browser_sessions"
}

// TableName returns the database table name for ServiceSession.
func (ServiceSession) TableName() string {
	return "service_sessions"
}

// IsExpired checks if the session has expired.
func (s *Session) IsExpired() bool {
	return time.Now().UTC().After(s.Expiration)
}

// UpdateLastActivity updates the last activity timestamp to the current time.
func (s *Session) UpdateLastActivity() {
	now := time.Now().UTC()
	s.LastActivity = &now
}
