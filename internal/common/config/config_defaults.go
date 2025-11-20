// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	googleUuid "github.com/google/uuid"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

const (
	defaultLogLevel                    = cryptoutilMagic.DefaultLogLevelInfo                // Balanced verbosity: shows important events without being overwhelming.
	defaultBindPublicProtocol          = cryptoutilMagic.DefaultPublicProtocolCryptoutil    // HTTPS by default for security in production environments.
	defaultBindPublicAddress           = cryptoutilMagic.DefaultPublicAddressCryptoutil     // IPv4 loopback prevents external access by default, requires explicit configuration for exposure.
	defaultBindPublicPort              = cryptoutilMagic.DefaultPublicPortCryptoutil        // Standard HTTP/HTTPS port, well-known and commonly available.
	defaultBindPrivateProtocol         = cryptoutilMagic.DefaultPrivateProtocolCryptoutil   // HTTPS for private API security, even in service-to-service communication.
	defaultBindPrivateAddress          = cryptoutilMagic.DefaultPrivateAddressCryptoutil    // IPv4 loopback for private API, only accessible from same machine.
	defaultBindPrivatePort             = cryptoutilMagic.DefaultPrivatePortCryptoutil       // Non-standard port to avoid conflicts with other services.
	defaultPublicBrowserAPIContextPath = cryptoutilMagic.DefaultPublicBrowserAPIContextPath // RESTful API versioning, separates browser from service APIs.
	defaultPublicServiceAPIContextPath = cryptoutilMagic.DefaultPublicServiceAPIContextPath // RESTful API versioning, separates service from browser APIs.
	defaultCORSMaxAge                  = cryptoutilMagic.DefaultCORSMaxAge                  // 1 hour cache for CORS preflight requests, balances performance and freshness.
	defaultCSRFTokenName               = cryptoutilMagic.DefaultCSRFTokenName               // Standard CSRF token name, widely recognized by frameworks.
	defaultCSRFTokenSameSite           = cryptoutilMagic.DefaultCSRFTokenSameSiteStrict     // Strict SameSite prevents CSRF while maintaining usability.
	defaultCSRFTokenMaxAge             = cryptoutilMagic.DefaultCSRFTokenMaxAge             // 1 hour expiration balances security and user experience.
	defaultCSRFTokenCookieSecure       = cryptoutilMagic.DefaultCSRFTokenCookieSecure       // Secure cookies in production prevent MITM attacks.
	defaultCSRFTokenCookieHTTPOnly     = cryptoutilMagic.DefaultCSRFTokenCookieHTTPOnly     // False allows JavaScript access for form submissions (Swagger UI workaround).
	defaultCSRFTokenCookieSessionOnly  = cryptoutilMagic.DefaultCSRFTokenCookieSessionOnly  // Session-only prevents persistent tracking while maintaining security.
	defaultCSRFTokenSingleUseToken     = cryptoutilMagic.DefaultCSRFTokenSingleUseToken     // Reusable tokens for better UX, can be changed for high-security needs.
	defaultRequestBodyLimit            = cryptoutilMagic.DefaultHTTPRequestBodyLimit        // 2MB limit prevents large payload attacks while allowing reasonable API usage.
	defaultBrowserIPRateLimit          = cryptoutilMagic.DefaultPublicBrowserAPIIPRateLimit // More lenient rate limit for browser APIs (user interactions).
	defaultServiceIPRateLimit          = cryptoutilMagic.DefaultPublicServiceAPIIPRateLimit // More restrictive rate limit for service APIs (automated systems).
	defaultDatabaseContainer           = cryptoutilMagic.DefaultDatabaseContainerDisabled   // Disabled by default to avoid unexpected container dependencies.
	defaultDatabaseURL                 = cryptoutilMagic.DefaultDatabaseURL                 // pragma: allowlist secret // PostgreSQL default with placeholder credentials, SSL disabled for local dev.
	defaultDatabaseInitTotalTimeout    = cryptoutilMagic.DefaultDatabaseInitTotalTimeout    // 5 minutes allows for container startup while preventing indefinite waits.
	defaultDatabaseInitRetryWait       = cryptoutilMagic.DefaultDataInitRetryWait           // 1 second retry interval balances responsiveness and resource usage.
	defaultServerShutdownTimeout       = cryptoutilMagic.DefaultDataServerShutdownTimeout   // 5 seconds allows graceful shutdown while preventing indefinite waits.
	defaultHelp                        = cryptoutilMagic.DefaultHelp
	defaultVerboseMode                 = cryptoutilMagic.DefaultVerboseMode
	defaultDevMode                     = cryptoutilMagic.DefaultDevMode
	defaultDryRun                      = cryptoutilMagic.DefaultDryRun
	defaultProfile                     = cryptoutilMagic.DefaultProfile
	defaultOTLPEnabled                 = cryptoutilMagic.DefaultOTLPEnabled
	defaultOTLPConsole                 = cryptoutilMagic.DefaultOTLPConsole
	defaultOTLPService                 = cryptoutilMagic.DefaultOTLPServiceDefault
	defaultOTLPVersion                 = cryptoutilMagic.DefaultOTLPVersionDefault
	defaultOTLPEnvironment             = cryptoutilMagic.DefaultOTLPEnvironmentDefault
	defaultOTLPHostname                = cryptoutilMagic.DefaultOTLPHostnameDefault
	defaultOTLPEndpoint                = cryptoutilMagic.DefaultOTLPEndpointDefault
	defaultUnsealMode                  = cryptoutilMagic.DefaultUnsealModeSysInfo
)

// Configuration profiles for common deployment scenarios.
var profiles = map[string]map[string]any{
	"test": {
		"log-level":                "ERROR",
		"dev":                      defaultDevMode,
		"bind-public-protocol":     defaultBindPublicProtocol,
		"bind-public-address":      defaultBindPublicAddress,
		"bind-public-port":         defaultBindPublicPort,
		"bind-private-protocol":    defaultBindPrivateProtocol,
		"bind-private-address":     defaultBindPrivateAddress,
		"bind-private-port":        defaultBindPrivatePort,
		"database-container":       defaultDatabaseContainer,
		"database-url":             "sqlite://file::memory:?cache=shared",
		"csrf-token-cookie-secure": false,
		"otlp":                     false,
		"otlp-console":             false,
		"otlp-environment":         "test",
	},
	"dev": {
		"log-level":                "DEBUG",
		"dev":                      defaultDevMode,
		"bind-public-protocol":     defaultBindPrivateProtocol,
		"bind-public-address":      defaultBindPublicAddress,
		"bind-public-port":         defaultBindPublicPort,
		"bind-private-protocol":    defaultBindPrivateProtocol,
		"bind-private-address":     defaultBindPrivateAddress,
		"bind-private-port":        defaultBindPrivatePort,
		"database-container":       defaultDatabaseContainer,
		"database-url":             "sqlite://file::memory:?cache=shared",
		"csrf-token-cookie-secure": false,
		"otlp":                     false,
		"otlp-console":             true,
		"otlp-environment":         "dev",
	},
	"stg": {
		"log-level":                "INFO",
		"dev":                      false,
		"bind-public-protocol":     defaultBindPrivateProtocol,
		"bind-public-address":      "0.0.0.0",
		"bind-public-port":         defaultBindPublicPort,
		"bind-private-protocol":    defaultBindPrivateProtocol,
		"bind-private-address":     defaultBindPrivateAddress,
		"bind-private-port":        defaultBindPrivatePort,
		"database-container":       defaultDatabaseContainer,
		"csrf-token-cookie-secure": true,
		"otlp":                     true,
		"otlp-console":             false,
		"otlp-environment":         "stg",
	},
	"prod": {
		"log-level":                "WARN",
		"dev":                      false,
		"bind-public-protocol":     defaultBindPublicProtocol,
		"bind-public-address":      "0.0.0.0",
		"bind-public-port":         defaultBindPublicPort,
		"bind-private-protocol":    defaultBindPrivateProtocol,
		"bind-private-address":     defaultBindPrivateAddress,
		"bind-private-port":        defaultBindPrivatePort,
		"database-container":       defaultDatabaseContainer,
		"rate-limit":               defaultBrowserIPRateLimit,
		"csrf-token-cookie-secure": true,
		"otlp":                     true,
		"otlp-console":             false,
		"otlp-environment":         "prod",
	},
}

var defaultCORSAllowedOrigins = cryptoutilMagic.DefaultCORSAllowedOrigins

var defaultAllowedIps = cryptoutilMagic.DefaultIPFilterAllowedIPs

var defaultTLSPublicDNSNames = cryptoutilMagic.DefaultTLSPublicDNSNames

var defaultTLSPublicIPAddresses = cryptoutilMagic.DefaultTLSPublicIPAddresses

var defaultTLSPrivateDNSNames = cryptoutilMagic.DefaultTLSPrivateDNSNames

var defaultTLSPrivateIPAddresses = cryptoutilMagic.DefaultTLSPrivateIPAddresses

var defaultAllowedCIDRs = cryptoutilMagic.DefaultIPFilterAllowedCIDRs

var defaultCORSAllowedMethods = cryptoutilMagic.DefaultCORSAllowedMethods

var defaultCORSAllowedHeaders = cryptoutilMagic.DefaultCORSAllowedHeaders

var defaultOTLPInstance = func() string {
	return googleUuid.Must(googleUuid.NewV7()).String()
}()

var defaultUnsealFiles = cryptoutilMagic.DefaultUnsealFiles

var defaultConfigFiles = cryptoutilMagic.DefaultConfigFiles

// set of valid subcommands.
var subcommands = map[string]struct{}{
	"start": {},
	"stop":  {},
	"init":  {},
	"live":  {},
	"ready": {},
}
