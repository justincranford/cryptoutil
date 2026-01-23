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

// DefaultBrowserRealms - Default browser realm configuration files (6 non-federated session-based auth methods).
// See 02-10.authn.instructions.md for complete browser authentication methods (28 total: 6 non-federated + 22 federated).
var DefaultBrowserRealms = []string{
	"jwe-session-cookie",      // Browser Realm #1: JWE Session Cookie (encrypted session tokens)
	"jws-session-cookie",      // Browser Realm #2: JWS Session Cookie (signed session tokens)
	"opaque-session-cookie",   // Browser Realm #3: Opaque Session Cookie (database-backed session tokens)
	"basic-username-password", // Browser Realm #4: Basic (Username/Password) - HTTP Basic auth with user credentials
	"bearer-api-token",        // Browser Realm #5: Bearer (API Token) - Bearer token authentication from browser
	"https-client-cert",       // Browser Realm #6: HTTPS Client Certificate - mTLS client certificate from browser
}

// DefaultServiceRealms - Default service realm configuration files (6 non-federated token-based auth methods).
// See 02-10.authn.instructions.md for complete headless authentication methods (13 total: 6 non-federated + 4 federated + 3 additional).
var DefaultServiceRealms = []string{
	"jwe-session-token",      // Service Realm #1: JWE Session Token (encrypted session tokens for headless clients)
	"jws-session-token",      // Service Realm #2: JWS Session Token (signed session tokens for headless clients)
	"opaque-session-token",   // Service Realm #3: Opaque Session Token (non-JWT session tokens)
	"basic-client-id-secret", // Service Realm #4: Basic (Client ID/Secret) - HTTP Basic with client credentials
	"bearer-api-token",       // Service Realm #5: Bearer (API Token) - Long-lived service credentials
	"https-client-cert",      // Service Realm #6: HTTPS Client Certificate - mTLS for high-security service-to-service
}

// Realm configuration validation constants.
const (
	// RealmMinTokenLengthBytes - Minimum token length in bytes for random tokens (16 bytes = 128 bits).
	RealmMinTokenLengthBytes = 16

	// RealmMinBearerTokenLengthBytes - Minimum bearer token length in bytes (32 bytes = 256 bits).
	RealmMinBearerTokenLengthBytes = 32

	// RealmStorageTypeDatabase - Database storage type for session/token persistence.
	RealmStorageTypeDatabase = "database"

	// RealmStorageTypeRedis - Redis storage type for session/token persistence.
	RealmStorageTypeRedis = "redis"
)

// Identity service port and OTLP constants.
const (
	// OTLPServiceIdentityRP is the OTLP service name for identity-rp.
	OTLPServiceIdentityRP = "identity-rp"

	// IdentityRPServicePort is the default public port for identity-rp service.
	IdentityRPServicePort = uint16(18300)

	// OTLPServiceIdentitySPA is the OTLP service name for identity-spa.
	OTLPServiceIdentitySPA = "identity-spa"

	// IdentitySPAServicePort is the default public port for identity-spa service.
	IdentitySPAServicePort = uint16(18400)
)
