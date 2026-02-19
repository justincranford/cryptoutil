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
// Port ranges per service catalog (architecture.md):
// - identity-authz: 8200-8299
// - identity-idp: 8300-8399
// - identity-rs: 8400-8499
// - identity-rp: 8500-8599
// - identity-spa: 8600-8699.
const (
	// OTLPServiceIdentityAuthz is the OTLP service name for identity-authz.
	OTLPServiceIdentityAuthz = "identity-authz"

	// IdentityAuthzServicePort is the default public port for identity-authz service.
	IdentityAuthzServicePort = uint16(8200)

	// OTLPServiceIdentityIDP is the OTLP service name for identity-idp.
	OTLPServiceIdentityIDP = "identity-idp"

	// IdentityIDPServicePort is the default public port for identity-idp service.
	IdentityIDPServicePort = uint16(8300)

	// OTLPServiceIdentityRS is the OTLP service name for identity-rs.
	OTLPServiceIdentityRS = "identity-rs"

	// IdentityRSServicePort is the default public port for identity-rs service.
	IdentityRSServicePort = uint16(8400)

	// OTLPServiceIdentityRP is the OTLP service name for identity-rp.
	OTLPServiceIdentityRP = "identity-rp"

	// IdentityRPServicePort is the default public port for identity-rp service.
	IdentityRPServicePort = uint16(8500)

	// OTLPServiceIdentitySPA is the OTLP service name for identity-spa.
	OTLPServiceIdentitySPA = "identity-spa"

	// IdentitySPAServicePort is the default public port for identity-spa service.
	IdentitySPAServicePort = uint16(8600)
)

// E2E Test Configuration for identity services.
// Uses PRODUCT-level compose (deployments/identity/compose.yml) with PRODUCT ports (18XXX).
const (
	// IdentityE2EComposeFile is the path to the identity PRODUCT docker compose file (relative from e2e test directory).
	// Path: internal/apps/identity/e2e → ../../../../deployments/identity/compose.yml.
	IdentityE2EComposeFile = "../../../../deployments/identity/compose.yml"

	// IdentityE2EAuthzContainer is the identity-authz service name in PRODUCT compose.
	IdentityE2EAuthzContainer = "identity-authz-app-sqlite-1"

	// IdentityE2EIDPContainer is the identity-idp service name in PRODUCT compose.
	IdentityE2EIDPContainer = "identity-idp-app-sqlite-1"

	// IdentityE2ERSContainer is the identity-rs service name in PRODUCT compose.
	IdentityE2ERSContainer = "identity-rs-app-sqlite-1"

	// IdentityE2ERPContainer is the identity-rp service name in PRODUCT compose.
	IdentityE2ERPContainer = "identity-rp-app-sqlite-1"

	// IdentityE2ESPAContainer is the identity-spa service name in PRODUCT compose.
	IdentityE2ESPAContainer = "identity-spa-app-sqlite-1"

	// IdentityE2EHealthTimeout is the timeout for health checks during E2E tests.
	// Must account for cascade dependencies: authz (30s) → idp (30s) → rs (30s) → rp (30s) → spa (30s) = 150s worst case.
	// Increased to 240s to handle slower CI/CD environments and Windows systems.
	IdentityE2EHealthTimeout = 240 * time.Second

	// IdentityE2EHealthPollInterval is the interval between health check attempts.
	IdentityE2EHealthPollInterval = 2 * time.Second

	// IdentityE2EAuthzPublicPort is the identity-authz PRODUCT-level public HTTPS port.
	IdentityE2EAuthzPublicPort = 18200

	// IdentityE2EIDPPublicPort is the identity-idp PRODUCT-level public HTTPS port.
	IdentityE2EIDPPublicPort = 18300

	// IdentityE2ERSPublicPort is the identity-rs PRODUCT-level public HTTPS port.
	IdentityE2ERSPublicPort = 18400

	// IdentityE2ERPPublicPort is the identity-rp PRODUCT-level public HTTPS port.
	IdentityE2ERPPublicPort = 18500

	// IdentityE2ESPAPublicPort is the identity-spa PRODUCT-level public HTTPS port.
	IdentityE2ESPAPublicPort = 18600

	// IdentityE2EHealthEndpoint is the public health check endpoint.
	// Uses /service/api/v1/health for headless client health checks (per 02-03.https-ports.instructions.md).
	IdentityE2EHealthEndpoint = "/service/api/v1/health"
)
