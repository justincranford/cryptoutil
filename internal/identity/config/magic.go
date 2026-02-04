// Copyright (c) 2025 Justin Cranford
//
//

package config

// Default ports.
// Port ranges per user specification:
// - identity-authz: 8100-8109 (shares with idp)
// - identity-idp: 8100-8109 (shares with authz)
// - identity-rs: 8110-8119
// All admin ports: 9090
const (
        defaultAuthZPort      = 8100 // Default AuthZ server port.
        defaultAuthZAdminPort = 9090 // Default AuthZ admin port.
        defaultIDPPort        = 8100 // Default IDP server port (shares with authz).
        defaultIDPAdminPort   = 9090 // Default IDP admin port.
        defaultRSPort         = 8110 // Default RS server port.
	defaultRSAdminPort    = 9090 // Default RS admin port.
)

// Default server timeouts (seconds).
const (
	defaultReadTimeoutSeconds  = 30 // Default read timeout (seconds).
	defaultWriteTimeoutSeconds = 30 // Default write timeout (seconds).
	defaultIdleTimeoutSeconds  = 60 // Default idle timeout (seconds).
)

// Default database connection pool settings.
const (
	defaultMaxOpenConns       = 25 // Default max open connections.
	defaultMaxIdleConns       = 5  // Default max idle connections.
	defaultConnMaxLifetimeMin = 5  // Default connection max lifetime (minutes).
	defaultConnMaxIdleTimeMin = 10 // Default connection max idle time (minutes).
)

// Default token lifetimes (seconds).
const (
	defaultAccessTokenLifetimeSeconds  = 3600  // Default access token lifetime (1 hour).
	defaultRefreshTokenLifetimeSeconds = 86400 // Default refresh token lifetime (24 hours).
	defaultIDTokenLifetimeSeconds      = 3600  // Default ID token lifetime (1 hour).
)

// Default session settings (seconds).
const (
	defaultSessionLifetimeSeconds    = 3600 // Default session lifetime (1 hour).
	defaultIdleTimeoutSecondsSession = 900  // Default session idle timeout (15 minutes).
)

// Default rate limiting.
const (
	defaultRateLimitRequests      = 100 // Default rate limit requests.
	defaultRateLimitWindowSeconds = 60  // Default rate limit window (seconds).
)

// Token format constants.
const (
	tokenFormatJWS  = "jws"  // JWS token format.
	tokenFormatJWE  = "jwe"  // JWE token format.
	tokenFormatUUID = "uuid" // UUID token format.
)
