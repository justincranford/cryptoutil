// Copyright (c) 2025 Justin Cranford

package magic

import "time"

// Session Management Magic Constants.

// SessionAlgorithmType represents the type of session token algorithm.
type SessionAlgorithmType string

const (
	// SessionAlgorithmOPAQUE uses hashed UUIDv7 tokens stored in database.
	// Token format: Base64-encoded UUIDv7, hashed with HighEntropyFixedRegistry.
	// Storage: Token hash stored in browser_sessions or service_sessions table.
	SessionAlgorithmOPAQUE SessionAlgorithmType = "OPAQUE"

	// SessionAlgorithmJWS uses signed JWT tokens (JSON Web Signature).
	// Token format: JWS compact serialization (header.payload.signature).
	// Storage: JWK stored encrypted in browser_session_jwks or service_session_jwks table.
	SessionAlgorithmJWS SessionAlgorithmType = "JWS"

	// SessionAlgorithmJWE uses encrypted JWT tokens (JSON Web Encryption).
	// Token format: JWE compact serialization (header.key.iv.ciphertext.tag).
	// Storage: JWK stored encrypted in browser_session_jwks or service_session_jwks table.
	SessionAlgorithmJWE SessionAlgorithmType = "JWE"
)

// JWS algorithm identifiers for session tokens.
const (
	// SessionJWSAlgorithmRS256 is RSA-SHA256 signature algorithm.
	SessionJWSAlgorithmRS256 = "RS256"

	// SessionJWSAlgorithmRS384 is RSA-SHA384 signature algorithm.
	SessionJWSAlgorithmRS384 = "RS384"

	// SessionJWSAlgorithmRS512 is RSA-SHA512 signature algorithm.
	SessionJWSAlgorithmRS512 = "RS512"

	// SessionJWSAlgorithmES256 is ECDSA-SHA256 signature algorithm (P-256 curve).
	SessionJWSAlgorithmES256 = "ES256"

	// SessionJWSAlgorithmES384 is ECDSA-SHA384 signature algorithm (P-384 curve).
	SessionJWSAlgorithmES384 = "ES384"

	// SessionJWSAlgorithmES512 is ECDSA-SHA512 signature algorithm (P-521 curve).
	SessionJWSAlgorithmES512 = "ES512"

	// SessionJWSAlgorithmEdDSA is Ed25519 signature algorithm.
	SessionJWSAlgorithmEdDSA = "EdDSA"

	// DefaultSessionJWSAlgorithm is the default JWS algorithm for session tokens.
	DefaultSessionJWSAlgorithm = SessionJWSAlgorithmRS256
)

// JWE algorithm identifiers for session tokens.
const (
	// SessionJWEAlgorithmDirA256GCM is direct encryption with AES-256-GCM.
	// Key agreement: dir (direct use of shared symmetric key).
	// Content encryption: A256GCM (AES-256-GCM).
	SessionJWEAlgorithmDirA256GCM = "dir+A256GCM"

	// SessionJWEAlgorithmA256GCMKW is AES-256-GCM key wrapping.
	// Key agreement: A256GCMKW (AES-256-GCM key wrap).
	// Content encryption: A256GCM (AES-256-GCM).
	SessionJWEAlgorithmA256GCMKWA256GCM = "A256GCMKW+A256GCM"

	// DefaultSessionJWEAlgorithm is the default JWE algorithm for session tokens.
	DefaultSessionJWEAlgorithm = SessionJWEAlgorithmDirA256GCM
)

// Session expiration and timeout defaults.
const (
	// DefaultBrowserSessionExpiration is the default expiration time for browser sessions.
	DefaultBrowserSessionExpiration = 24 * time.Hour

	// DefaultServiceSessionExpiration is the default expiration time for service sessions.
	DefaultServiceSessionExpiration = 7 * 24 * time.Hour

	// DefaultSessionIdleTimeout is the default idle timeout for sessions.
	// Session expires after this duration of inactivity.
	DefaultSessionIdleTimeout = 2 * time.Hour

	// DefaultSessionCleanupInterval is the default interval for cleaning up expired sessions.
	DefaultSessionCleanupInterval = 1 * time.Hour

	// DefaultCompatibilitySessionExpiration is the default session expiration for API compatibility.
	DefaultCompatibilitySessionExpiration = 15 * time.Minute

	// DefaultShutdownTimeout is the default timeout for graceful shutdown operations.
	DefaultShutdownTimeout = 30 * time.Second
)
