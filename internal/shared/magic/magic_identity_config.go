// Copyright (c) 2025 Justin Cranford
//
//

package magic

// Identity service default ports.
// Port ranges per service catalog (architecture.md):
// - identity-authz: 8200-8299
// - identity-idp: 8300-8399
// - identity-rs: 8400-8499
// All admin ports: 9090.
const (
IdentityDefaultAuthZPort      = 8200 // Default AuthZ server port.
IdentityDefaultAuthZAdminPort = 9090 // Default AuthZ admin port.
IdentityDefaultIDPPort        = 8300 // Default IDP server port.
IdentityDefaultIDPAdminPort   = 9090 // Default IDP admin port.
IdentityDefaultRSPort         = 8400 // Default RS server port.
IdentityDefaultRSAdminPort    = 9090 // Default RS admin port.
)

// Identity default server timeouts (seconds).
const (
IdentityDefaultReadTimeoutSeconds  = 30 // Default read timeout (seconds).
IdentityDefaultWriteTimeoutSeconds = 30 // Default write timeout (seconds).
IdentityDefaultIdleTimeoutSeconds  = 60 // Default idle timeout (seconds).
)

// Identity default database connection pool settings.
const (
IdentityDefaultMaxOpenConns       = 25 // Default max open connections.
IdentityDefaultMaxIdleConns       = 5  // Default max idle connections.
IdentityDefaultConnMaxLifetimeMin = 5  // Default connection max lifetime (minutes).
IdentityDefaultConnMaxIdleTimeMin = 10 // Default connection max idle time (minutes).
)

// Identity default token lifetimes (seconds).
const (
IdentityDefaultAccessTokenLifetimeSeconds  = 3600  // Default access token lifetime (1 hour).
IdentityDefaultRefreshTokenLifetimeSeconds = 86400 // Default refresh token lifetime (24 hours).
IdentityDefaultIDTokenLifetimeSeconds      = 3600  // Default ID token lifetime (1 hour).
)

// Identity default session settings (seconds).
const (
IdentityDefaultSessionLifetimeSeconds    = 3600 // Default session lifetime (1 hour).
IdentityDefaultIdleTimeoutSecondsSession = 900  // Default session idle timeout (15 minutes).
)

// Identity default rate limiting.
const (
IdentityDefaultRateLimitRequests      = 100 // Default rate limit requests.
IdentityDefaultRateLimitWindowSeconds = 60  // Default rate limit window (seconds).
)

// Identity token format constants.
const (
IdentityTokenFormatJWS  = "jws"  // JWS token format.
IdentityTokenFormatJWE  = "jwe"  // JWE token format.
IdentityTokenFormatUUID = "uuid" // UUID token format.
)
