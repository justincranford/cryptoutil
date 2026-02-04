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
// Port ranges per user specification:
// - identity-authz: 8100-8109 (shares with idp)
// - identity-idp: 8100-8109 (shares with authz)
// - identity-rs: 8110-8119
// - identity-rp: 8120-8129
// - identity-spa: 8130-8139
const (
	// OTLPServiceIdentityAuthz is the OTLP service name for identity-authz.
	OTLPServiceIdentityAuthz = "identity-authz"

	// IdentityAuthzServicePort is the default public port for identity-authz service.
	IdentityAuthzServicePort = uint16(8100)

	// OTLPServiceIdentityIDP is the OTLP service name for identity-idp.
	OTLPServiceIdentityIDP = "identity-idp"

	// IdentityIDPServicePort is the default public port for identity-idp service.
	// Same as authz (8100) per specification - both share 8100-8109 range.
	IdentityIDPServicePort = uint16(8100)

	// OTLPServiceIdentityRS is the OTLP service name for identity-rs.
	OTLPServiceIdentityRS = "identity-rs"

	// IdentityRSServicePort is the default public port for identity-rs service.
	IdentityRSServicePort = uint16(8110)

	// OTLPServiceIdentityRP is the OTLP service name for identity-rp.
	OTLPServiceIdentityRP = "identity-rp"

	// IdentityRPServicePort is the default public port for identity-rp service.
	IdentityRPServicePort = uint16(8120)

	// OTLPServiceIdentitySPA is the OTLP service name for identity-spa.
	OTLPServiceIdentitySPA = "identity-spa"

	// IdentitySPAServicePort is the default public port for identity-spa service.
	IdentitySPAServicePort = uint16(8130)
)

// E2E Test Configuration for identity services.
const (
	// IdentityE2EComposeFile is the path to the identity docker compose file (relative from e2e test directory).
	// Path: internal/apps/identity/e2e → ../../../../deployments/identity/compose.e2e.yml
	IdentityE2EComposeFile = "../../../../deployments/identity/compose.e2e.yml"

	// IdentityE2EAuthzContainer is the identity-authz container name.
	IdentityE2EAuthzContainer = "identity-authz-e2e"

	// IdentityE2EIDPContainer is the identity-idp container name.
	IdentityE2EIDPContainer = "identity-idp-e2e"

	// IdentityE2ERSContainer is the identity-rs container name.
	IdentityE2ERSContainer = "identity-rs-e2e"

	// IdentityE2ERPContainer is the identity-rp container name.
	IdentityE2ERPContainer = "identity-rp-e2e"

	// IdentityE2ESPAContainer is the identity-spa container name.
	IdentityE2ESPAContainer = "identity-spa-e2e"

	// IdentityE2EHealthTimeout is the timeout for health checks during E2E tests.
	// Must account for cascade dependencies: authz (30s) → idp (30s) → rs (30s) → rp (30s) → spa (30s) = 150s worst case.
	// Increased to 240s to handle slower CI/CD environments and Windows systems.
	IdentityE2EHealthTimeout = 240 * time.Second

	// IdentityE2EHealthPollInterval is the interval between health check attempts.
	IdentityE2EHealthPollInterval = 2 * time.Second

	// IdentityE2EAuthzPublicPort is the identity-authz E2E public HTTPS port.
	IdentityE2EAuthzPublicPort = 8100

	// IdentityE2EIDPPublicPort is the identity-idp E2E public HTTPS port.
	IdentityE2EIDPPublicPort = 8100

	// IdentityE2ERSPublicPort is the identity-rs E2E public HTTPS port.
	IdentityE2ERSPublicPort = 8110

	// IdentityE2ERPPublicPort is the identity-rp E2E public HTTPS port.
	IdentityE2ERPPublicPort = 8120

	// IdentityE2ESPAPublicPort is the identity-spa E2E public HTTPS port.
	IdentityE2ESPAPublicPort = 8130
	// IdentityE2EHealthEndpoint is the public health check endpoint.
	// Uses /health for liveness checks (matches cipher-im pattern).
	IdentityE2EHealthEndpoint = "/health"
)
