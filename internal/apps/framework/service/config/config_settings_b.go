// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	csrfTokenCookieSecure = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "csrf-token-cookie-secure",
		Shorthand:   "R",
		Value:       cryptoutilSharedMagic.DefaultCSRFTokenCookieSecure,
		Usage:       "CSRF token cookie Secure attribute",
		Description: "CSRF Token Cookie Secure",
	})
	csrfTokenCookieHTTPOnly = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "csrf-token-cookie-http-only",
		Shorthand:   "J",
		Value:       cryptoutilSharedMagic.DefaultCSRFTokenCookieHTTPOnly, // False needed for Swagger UI submit CSRF workaround
		Usage:       "CSRF token cookie HttpOnly attribute",
		Description: "CSRF Token Cookie HTTPOnly",
	})
	csrfTokenCookieSessionOnly = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "csrf-token-cookie-session-only",
		Shorthand:   "E",
		Value:       cryptoutilSharedMagic.DefaultCSRFTokenCookieSessionOnly,
		Usage:       "CSRF token cookie SessionOnly attribute",
		Description: "CSRF Token Cookie SessionOnly",
	})
	csrfTokenSingleUseToken = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "csrf-token-single-use-token",
		Shorthand:   "G",
		Value:       cryptoutilSharedMagic.DefaultCSRFTokenSingleUseToken,
		Usage:       "CSRF token SingleUse attribute",
		Description: "CSRF Token SingleUseToken",
	})
	browserIPRateLimit = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "browser-rate-limit",
		Shorthand:   "e",
		Value:       cryptoutilSharedMagic.DefaultPublicBrowserAPIIPRateLimit,
		Usage:       "rate limit for browser API requests per second",
		Description: "Browser IP Rate Limit",
	})
	serviceIPRateLimit = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "service-rate-limit",
		Shorthand:   "w",
		Value:       cryptoutilSharedMagic.DefaultPublicServiceAPIIPRateLimit,
		Usage:       "rate limit for service API requests per second",
		Description: "Service IP Rate Limit",
	})
	allowedIps = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "allowed-ips",
		Shorthand:   "I",
		Value:       defaultAllowedIps,
		Usage:       "comma-separated list of allowed IPs",
		Description: "Allowed IPs",
	})
	allowedCidrs = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "allowed-cidrs",
		Shorthand:   "C",
		Value:       defaultAllowedCIDRs,
		Usage:       "comma-separated list of allowed CIDRs",
		Description: "Allowed CIDRs",
	})
	swaggerUIUsername = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "swagger-ui-username",
		Shorthand:   "1",
		Value:       defaultSwaggerUIUsername,
		Usage:       "username for Swagger UI basic authentication",
		Description: "Swagger UI Username",
	})
	swaggerUIPassword = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "swagger-ui-password",
		Shorthand:   "2",
		Value:       defaultSwaggerUIPassword,
		Usage:       "password for Swagger UI basic authentication",
		Description: "Swagger UI Password",
		Redacted:    true,
	})
	requestBodyLimit = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "request-body-limit",
		Shorthand:   "L",
		Value:       cryptoutilSharedMagic.DefaultHTTPRequestBodyLimit,
		Usage:       "Maximum request body size in bytes",
		Description: "Request Body Limit",
	})
	databaseContainer = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "database-container",
		Shorthand:   "D",
		Value:       cryptoutilSharedMagic.DefaultDatabaseContainerDisabled,
		Usage:       "database container mode; true to use container, false to use local database",
		Description: "Database Container",
	})
	databaseURL = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "database-url",
		Shorthand:   "u",
		Value:       cryptoutilSharedMagic.DefaultDatabaseURL,
		Usage:       "database URL; start a container with:\ndocker run -d --name postgres -p 5432:5432 -e POSTGRES_USER=USR -e POSTGRES_PASSWORD=PWD -e POSTGRES_DB=DB postgres:latest\n", // pragma: allowlist secret
		Description: "Database URL",
		Redacted:    true,
	})
	databaseSSLMode = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "database-sslmode",
		Shorthand:   "",
		Value:       "",
		Usage:       "PostgreSQL SSL mode: disable, require, verify-ca, verify-full (empty = use DSN default)",
		Description: "Database SSL Mode",
	})
	databaseSSLCert = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "database-sslcert",
		Shorthand:   "",
		Value:       "",
		Usage:       "path to client TLS certificate file for PostgreSQL mTLS (Cat 14)",
		Description: "Database SSL Client Cert",
	})
	databaseSSLKey = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "database-sslkey",
		Shorthand:   "",
		Value:       "",
		Usage:       "path to client TLS private key file for PostgreSQL mTLS (Cat 14)",
		Description: "Database SSL Client Key",
		Redacted:    true,
	})
	databaseSSLRootCert = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "database-sslrootcert",
		Shorthand:   "",
		Value:       "",
		Usage:       "path to CA truststore for verifying PostgreSQL server cert (Cat 10)",
		Description: "Database SSL Root Cert",
	})
	adminTLSCertFile = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "server-admin-tls-cert-file",
		Shorthand:   "",
		Value:       "",
		Usage:       "path to admin TLS server certificate file for private admin mTLS (Cat 7)",
		Description: "Admin TLS Cert File",
	})
	adminTLSKeyFile = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "server-admin-tls-key-file",
		Shorthand:   "",
		Value:       "",
		Usage:       "path to admin TLS server private key file for private admin mTLS (Cat 7)",
		Description: "Admin TLS Key File",
		Redacted:    true,
	})
	adminTLSCAFile = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "server-admin-tls-ca-file",
		Shorthand:   "",
		Value:       "",
		Usage:       "path to CA truststore for verifying admin client certs (Cat 6)",
		Description: "Admin TLS CA File",
	})
	databaseInitTotalTimeout = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "database-init-total-timeout",
		Shorthand:   "Z",
		Value:       cryptoutilSharedMagic.DefaultDatabaseInitTotalTimeout,
		Usage:       "database init total timeout",
		Description: "Database Init Total Timeout",
	})
	databaseInitRetryWait = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "database-init-retry-wait",
		Shorthand:   "W",
		Value:       cryptoutilSharedMagic.DefaultDataInitRetryWait,
		Usage:       "database init retry wait",
		Description: "Database Init Retry Wait",
	})
	serverShutdownTimeout = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "server-shutdown-timeout",
		Shorthand:   "",
		Value:       cryptoutilSharedMagic.DefaultDataServerShutdownTimeout,
		Usage:       "server shutdown timeout",
		Description: "Server Shutdown Timeout",
	})
	otlpEnabled = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "otlp",
		Shorthand:   "z",
		Value:       cryptoutilSharedMagic.DefaultOTLPEnabled,
		Usage:       "enable OTLP export",
		Description: "OTLP Export",
	})
	otlpConsole = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "otlp-console",
		Shorthand:   "q",
		Value:       cryptoutilSharedMagic.DefaultOTLPConsole,
		Usage:       "enable OTLP logging to console (STDOUT)",
		Description: "OTLP Console",
	})
	otlpService = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "otlp-service",
		Shorthand:   "s",
		Value:       cryptoutilSharedMagic.DefaultOTLPServiceDefault,
		Usage:       "OTLP service",
		Description: "OTLP Service",
	})
	otlpVersion = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "otlp-version",
		Shorthand:   "B",
		Value:       cryptoutilSharedMagic.DefaultOTLPVersionDefault,
		Usage:       "OTLP version",
		Description: "OTLP Version",
	})
	otlpEnvironment = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "otlp-environment",
		Shorthand:   "K",
		Value:       cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault,
		Usage:       "OTLP environment",
		Description: "OTLP Environment",
	})
	otlpHostname = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "otlp-hostname",
		Shorthand:   "O",
		Value:       cryptoutilSharedMagic.DefaultOTLPHostnameDefault,
		Usage:       "OTLP hostname",
		Description: "OTLP Hostname",
	})
	otlpEndpoint = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "otlp-endpoint",
		Shorthand:   "",
		Value:       cryptoutilSharedMagic.DefaultOTLPEndpointDefault,
		Usage:       "OTLP endpoint (grpc://host:port or http://host:port)",
		Description: "OTLP Endpoint",
	})
	otlpInstance = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "otlp-instance",
		Shorthand:   "V",
		Value:       defaultOTLPInstance,
		Usage:       "OTLP instance id",
		Description: "OTLP Instance",
	})
	unsealMode = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "unseal-mode",
		Shorthand:   "5",
		Value:       cryptoutilSharedMagic.DefaultUnsealModeSysInfo,
		Usage:       "unseal mode: N, M-of-N, sysinfo; N keys, or M-of-N derived keys from shared secrets, or X-of-Y custom sysinfo as shared secrets",
		Description: "Unseal Mode",
	})
	unsealFiles = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:      "unseal-files",
		Shorthand: "F",
		Value:     defaultUnsealFiles,
		Usage: "unseal files; repeat for multiple files; e.g. " +
			"\"--unseal-files=/docker/secrets/unseal_1of3 --unseal-files=/docker/secrets/unseal_2of3\"; " +
			"used for N unseal keys or M-of-N unseal shared secrets",
		Description: "Unseal Files",
	})
	browserRealms = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:      "browser-realms",
		Shorthand: "r",
		Value:     defaultBrowserRealms,
		Usage: "browser realm configuration files; repeat for multiple realms; e.g. " +
			"\"--browser-realms=/config/01-jwe-session-cookie.yml --browser-realms=/config/02-jws-session-cookie.yml\"; " +
			"defines session-based authentication realms for browser clients",
		Description: "Browser Authentication Realms",
	})
	serviceRealms = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:      "service-realms",
		Shorthand: "",
		Value:     defaultServiceRealms,
		Usage: "service realm configuration files; repeat for multiple realms; e.g. " +
			"\"--service-realms=/config/01-bearer-token.yml --service-realms=/config/02-client-cert.yml\"; " +
			"defines token-based authentication realms for service clients",
		Description: "Service Authentication Realms",
	})
	browserSessionCookie = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "browser-session-cookie",
		Shorthand:   "Q",
		Value:       cryptoutilSharedMagic.DefaultBrowserSessionCookie,
		Usage:       "browser session cookie type: jwe (encrypted), jws (signed), opaque (database); defaults to jws for stateless signed tokens [DEPRECATED: use browser-session-algorithm]",
		Description: "Browser Session Cookie Type",
	})
	browserSessionAlgorithm = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "browser-session-algorithm",
		Shorthand:   "",
		Value:       cryptoutilSharedMagic.DefaultBrowserSessionAlgorithm,
		Usage:       "browser session algorithm: OPAQUE (hashed UUIDv7), JWS (signed JWT), JWE (encrypted JWT)",
		Description: "Browser Session Algorithm",
	})
	browserSessionJWSAlgorithm = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "browser-session-jws-algorithm",
		Shorthand:   "",
		Value:       cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		Usage:       "JWS algorithm for browser sessions (e.g., RS256, RS384, RS512, ES256, ES384, ES512, EdDSA)",
		Description: "Browser Session JWS Algorithm",
	})
	browserSessionJWEAlgorithm = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "browser-session-jwe-algorithm",
		Shorthand:   "",
		Value:       cryptoutilSharedMagic.DefaultBrowserSessionJWEAlgorithm,
		Usage:       "JWE algorithm for browser sessions (e.g., dir+A256GCM, A256GCMKW+A256GCM)",
		Description: "Browser Session JWE Algorithm",
	})
	browserSessionExpiration = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "browser-session-expiration",
		Shorthand:   "",
		Value:       cryptoutilSharedMagic.DefaultBrowserSessionExpiration,
		Usage:       "browser session expiration duration (e.g., 24h, 48h)",
		Description: "Browser Session Expiration",
	})
	serviceSessionAlgorithm = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "service-session-algorithm",
		Shorthand:   "",
		Value:       cryptoutilSharedMagic.DefaultServiceSessionAlgorithm,
		Usage:       "service session algorithm: OPAQUE (hashed UUIDv7), JWS (signed JWT), JWE (encrypted JWT)",
		Description: "Service Session Algorithm",
	})
	serviceSessionJWSAlgorithm = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "service-session-jws-algorithm",
		Shorthand:   "",
		Value:       cryptoutilSharedMagic.DefaultServiceSessionJWSAlgorithm,
		Usage:       "JWS algorithm for service sessions (e.g., RS256, RS384, RS512, ES256, ES384, ES512, EdDSA)",
		Description: "Service Session JWS Algorithm",
	})
	serviceSessionJWEAlgorithm = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "service-session-jwe-algorithm",
		Shorthand:   "",
		Value:       cryptoutilSharedMagic.DefaultServiceSessionJWEAlgorithm,
		Usage:       "JWE algorithm for service sessions (e.g., dir+A256GCM, A256GCMKW+A256GCM)",
		Description: "Service Session JWE Algorithm",
	})
	serviceSessionExpiration = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "service-session-expiration",
		Shorthand:   "",
		Value:       cryptoutilSharedMagic.DefaultServiceSessionExpiration,
		Usage:       "service session expiration duration (e.g., 168h for 7 days)",
		Description: "Service Session Expiration",
	})
	sessionIdleTimeout = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "session-idle-timeout",
		Shorthand:   "",
		Value:       cryptoutilSharedMagic.DefaultSessionIdleTimeout,
		Usage:       "session idle timeout duration (e.g., 2h)",
		Description: "Session Idle Timeout",
	})
	sessionCleanupInterval = *SetEnvAndRegisterSetting(allServiceFrameworkServerRegisteredSettings, &Setting{
		Name:        "session-cleanup-interval",
		Shorthand:   "",
		Value:       cryptoutilSharedMagic.DefaultSessionCleanupInterval,
		Usage:       "interval for cleaning up expired sessions (e.g., 1h)",
		Description: "Session Cleanup Interval",
	})
)
