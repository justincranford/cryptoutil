// Copyright (c) 2025 Justin Cranford

package magic

import "time"

// Identity service scaling constants.
const (
	// IdentityScaling1x - Single instance scaling (demo, ci).
	IdentityScaling1x = 1
	// IdentityScaling2x - High availability scaling (development).
	IdentityScaling2x = 2
	// IdentityScaling3x - Production-like scaling (production).
	IdentityScaling3x = 3
)

// Secret rotation configuration constants.
const (
	// SecretRotationExpirationThreshold - Rotate secrets expiring within 7 days.
	SecretRotationExpirationThreshold = 7 * 24 * time.Hour

	// SecretRotationCheckInterval - Check for expiring secrets every hour.
	SecretRotationCheckInterval = 1 * time.Hour

	// SystemInitiatorName - System initiator name for automated operations.
	SystemInitiatorName = "system"
)

// Authentication realm configuration constants.
const (
	// DefaultBrowserSessionCookie - Default browser session cookie type (JWS signed stateless tokens).
	// DEPRECATED: Use DefaultBrowserSessionAlgorithm instead.
	DefaultBrowserSessionCookie = "jws"

	// DefaultBrowserSessionAlgorithm - Default browser session algorithm (OPAQUE hashed tokens).
	DefaultBrowserSessionAlgorithm = "OPAQUE"

	// DefaultServiceSessionAlgorithm - Default service session algorithm (JWS signed tokens).
	DefaultServiceSessionAlgorithm = "JWS"

	// DefaultBrowserSessionJWSAlgorithm - Default JWS algorithm for browser sessions.
	DefaultBrowserSessionJWSAlgorithm = "RS256"

	// DefaultServiceSessionJWSAlgorithm - Default JWS algorithm for service sessions.
	DefaultServiceSessionJWSAlgorithm = "RS256"

	// DefaultBrowserSessionJWEAlgorithm - Default JWE algorithm for browser sessions.
	DefaultBrowserSessionJWEAlgorithm = "dir+A256GCM"

	// DefaultServiceSessionJWEAlgorithm - Default JWE algorithm for service sessions.
	DefaultServiceSessionJWEAlgorithm = "dir+A256GCM"
)

// DefaultBrowserRealms - Default browser realm configuration files (session-based auth).
var DefaultBrowserRealms = []string{
	"jwe-session-cookie",
	"jws-session-cookie",
	"opaque-session-cookie",
}

// DefaultServiceRealms - Default service realm configuration files (token-based auth).
var DefaultServiceRealms = []string{
	"bearer-api-token",
	"https-client-cert",
	"basic-username-password",
}
