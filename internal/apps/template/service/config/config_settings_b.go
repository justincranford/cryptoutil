// Copyright (c) 2025 Justin Cranford
//
//

package config

var (
	csrfTokenCookieSecure = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "csrf-token-cookie-secure",
		Shorthand:   "R",
		Value:       defaultCSRFTokenCookieSecure,
		Usage:       "CSRF token cookie Secure attribute",
		Description: "CSRF Token Cookie Secure",
	})
	csrfTokenCookieHTTPOnly = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "csrf-token-cookie-http-only",
		Shorthand:   "J",
		Value:       defaultCSRFTokenCookieHTTPOnly, // False needed for Swagger UI submit CSRF workaround
		Usage:       "CSRF token cookie HttpOnly attribute",
		Description: "CSRF Token Cookie HTTPOnly",
	})
	csrfTokenCookieSessionOnly = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "csrf-token-cookie-session-only",
		Shorthand:   "E",
		Value:       defaultCSRFTokenCookieSessionOnly,
		Usage:       "CSRF token cookie SessionOnly attribute",
		Description: "CSRF Token Cookie SessionOnly",
	})
	csrfTokenSingleUseToken = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "csrf-token-single-use-token",
		Shorthand:   "G",
		Value:       defaultCSRFTokenSingleUseToken,
		Usage:       "CSRF token SingleUse attribute",
		Description: "CSRF Token SingleUseToken",
	})
	browserIPRateLimit = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "browser-rate-limit",
		Shorthand:   "e",
		Value:       defaultBrowserIPRateLimit,
		Usage:       "rate limit for browser API requests per second",
		Description: "Browser IP Rate Limit",
	})
	serviceIPRateLimit = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "service-rate-limit",
		Shorthand:   "w",
		Value:       defaultServiceIPRateLimit,
		Usage:       "rate limit for service API requests per second",
		Description: "Service IP Rate Limit",
	})
	allowedIps = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "allowed-ips",
		Shorthand:   "I",
		Value:       defaultAllowedIps,
		Usage:       "comma-separated list of allowed IPs",
		Description: "Allowed IPs",
	})
	allowedCidrs = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "allowed-cidrs",
		Shorthand:   "C",
		Value:       defaultAllowedCIDRs,
		Usage:       "comma-separated list of allowed CIDRs",
		Description: "Allowed CIDRs",
	})
	swaggerUIUsername = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "swagger-ui-username",
		Shorthand:   "1",
		Value:       defaultSwaggerUIUsername,
		Usage:       "username for Swagger UI basic authentication",
		Description: "Swagger UI Username",
	})
	swaggerUIPassword = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "swagger-ui-password",
		Shorthand:   "2",
		Value:       defaultSwaggerUIPassword,
		Usage:       "password for Swagger UI basic authentication",
		Description: "Swagger UI Password",
		Redacted:    true,
	})
	requestBodyLimit = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "request-body-limit",
		Shorthand:   "L",
		Value:       defaultRequestBodyLimit,
		Usage:       "Maximum request body size in bytes",
		Description: "Request Body Limit",
	})
	databaseContainer = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "database-container",
		Shorthand:   "D",
		Value:       defaultDatabaseContainer,
		Usage:       "database container mode; true to use container, false to use local database",
		Description: "Database Container",
	})
	databaseURL = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "database-url",
		Shorthand:   "u",
		Value:       defaultDatabaseURL,
		Usage:       "database URL; start a container with:\ndocker run -d --name postgres -p 5432:5432 -e POSTGRES_USER=USR -e POSTGRES_PASSWORD=PWD -e POSTGRES_DB=DB postgres:18\n", // pragma: allowlist secret
		Description: "Database URL",
		Redacted:    true,
	})
	databaseInitTotalTimeout = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "database-init-total-timeout",
		Shorthand:   "Z",
		Value:       defaultDatabaseInitTotalTimeout,
		Usage:       "database init total timeout",
		Description: "Database Init Total Timeout",
	})
	databaseInitRetryWait = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "database-init-retry-wait",
		Shorthand:   "W",
		Value:       defaultDatabaseInitRetryWait,
		Usage:       "database init retry wait",
		Description: "Database Init Retry Wait",
	})
	serverShutdownTimeout = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "server-shutdown-timeout",
		Shorthand:   "",
		Value:       defaultServerShutdownTimeout,
		Usage:       "server shutdown timeout",
		Description: "Server Shutdown Timeout",
	})
	otlpEnabled = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "otlp",
		Shorthand:   "z",
		Value:       defaultOTLPEnabled,
		Usage:       "enable OTLP export",
		Description: "OTLP Export",
	})
	otlpConsole = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "otlp-console",
		Shorthand:   "q",
		Value:       defaultOTLPConsole,
		Usage:       "enable OTLP logging to console (STDOUT)",
		Description: "OTLP Console",
	})
	otlpService = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "otlp-service",
		Shorthand:   "s",
		Value:       defaultOTLPService,
		Usage:       "OTLP service",
		Description: "OTLP Service",
	})
	otlpVersion = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "otlp-version",
		Shorthand:   "B",
		Value:       defaultOTLPVersion,
		Usage:       "OTLP version",
		Description: "OTLP Version",
	})
	otlpEnvironment = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "otlp-environment",
		Shorthand:   "K",
		Value:       defaultOTLPEnvironment,
		Usage:       "OTLP environment",
		Description: "OTLP Environment",
	})
	otlpHostname = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "otlp-hostname",
		Shorthand:   "O",
		Value:       defaultOTLPHostname,
		Usage:       "OTLP hostname",
		Description: "OTLP Hostname",
	})
	otlpEndpoint = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "otlp-endpoint",
		Shorthand:   "",
		Value:       defaultOTLPEndpoint,
		Usage:       "OTLP endpoint (grpc://host:port or http://host:port)",
		Description: "OTLP Endpoint",
	})
	otlpInstance = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "otlp-instance",
		Shorthand:   "V",
		Value:       defaultOTLPInstance,
		Usage:       "OTLP instance id",
		Description: "OTLP Instance",
	})
	unsealMode = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "unseal-mode",
		Shorthand:   "5",
		Value:       defaultUnsealMode,
		Usage:       "unseal mode: N, M-of-N, sysinfo; N keys, or M-of-N derived keys from shared secrets, or X-of-Y custom sysinfo as shared secrets",
		Description: "Unseal Mode",
	})
	unsealFiles = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:      "unseal-files",
		Shorthand: "F",
		Value:     defaultUnsealFiles,
		Usage: "unseal files; repeat for multiple files; e.g. " +
			"\"--unseal-files=/docker/secrets/unseal_1of3 --unseal-files=/docker/secrets/unseal_2of3\"; " +
			"used for N unseal keys or M-of-N unseal shared secrets",
		Description: "Unseal Files",
	})
	browserRealms = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:      "browser-realms",
		Shorthand: "r",
		Value:     defaultBrowserRealms,
		Usage: "browser realm configuration files; repeat for multiple realms; e.g. " +
			"\"--browser-realms=/config/01-jwe-session-cookie.yml --browser-realms=/config/02-jws-session-cookie.yml\"; " +
			"defines session-based authentication realms for browser clients",
		Description: "Browser Authentication Realms",
	})
	serviceRealms = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:      "service-realms",
		Shorthand: "",
		Value:     defaultServiceRealms,
		Usage: "service realm configuration files; repeat for multiple realms; e.g. " +
			"\"--service-realms=/config/01-bearer-token.yml --service-realms=/config/02-client-cert.yml\"; " +
			"defines token-based authentication realms for service clients",
		Description: "Service Authentication Realms",
	})
	browserSessionCookie = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "browser-session-cookie",
		Shorthand:   "Q",
		Value:       defaultBrowserSessionCookie,
		Usage:       "browser session cookie type: jwe (encrypted), jws (signed), opaque (database); defaults to jws for stateless signed tokens [DEPRECATED: use browser-session-algorithm]",
		Description: "Browser Session Cookie Type",
	})
	browserSessionAlgorithm = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "browser-session-algorithm",
		Shorthand:   "",
		Value:       defaultBrowserSessionAlgorithm,
		Usage:       "browser session algorithm: OPAQUE (hashed UUIDv7), JWS (signed JWT), JWE (encrypted JWT)",
		Description: "Browser Session Algorithm",
	})
	browserSessionJWSAlgorithm = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "browser-session-jws-algorithm",
		Shorthand:   "",
		Value:       defaultBrowserSessionJWSAlgorithm,
		Usage:       "JWS algorithm for browser sessions (e.g., RS256, RS384, RS512, ES256, ES384, ES512, EdDSA)",
		Description: "Browser Session JWS Algorithm",
	})
	browserSessionJWEAlgorithm = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "browser-session-jwe-algorithm",
		Shorthand:   "",
		Value:       defaultBrowserSessionJWEAlgorithm,
		Usage:       "JWE algorithm for browser sessions (e.g., dir+A256GCM, A256GCMKW+A256GCM)",
		Description: "Browser Session JWE Algorithm",
	})
	browserSessionExpiration = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "browser-session-expiration",
		Shorthand:   "",
		Value:       defaultBrowserSessionExpiration,
		Usage:       "browser session expiration duration (e.g., 24h, 48h)",
		Description: "Browser Session Expiration",
	})
	serviceSessionAlgorithm = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "service-session-algorithm",
		Shorthand:   "",
		Value:       defaultServiceSessionAlgorithm,
		Usage:       "service session algorithm: OPAQUE (hashed UUIDv7), JWS (signed JWT), JWE (encrypted JWT)",
		Description: "Service Session Algorithm",
	})
	serviceSessionJWSAlgorithm = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "service-session-jws-algorithm",
		Shorthand:   "",
		Value:       defaultServiceSessionJWSAlgorithm,
		Usage:       "JWS algorithm for service sessions (e.g., RS256, RS384, RS512, ES256, ES384, ES512, EdDSA)",
		Description: "Service Session JWS Algorithm",
	})
	serviceSessionJWEAlgorithm = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "service-session-jwe-algorithm",
		Shorthand:   "",
		Value:       defaultServiceSessionJWEAlgorithm,
		Usage:       "JWE algorithm for service sessions (e.g., dir+A256GCM, A256GCMKW+A256GCM)",
		Description: "Service Session JWE Algorithm",
	})
	serviceSessionExpiration = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "service-session-expiration",
		Shorthand:   "",
		Value:       defaultServiceSessionExpiration,
		Usage:       "service session expiration duration (e.g., 168h for 7 days)",
		Description: "Service Session Expiration",
	})
	sessionIdleTimeout = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "session-idle-timeout",
		Shorthand:   "",
		Value:       defaultSessionIdleTimeout,
		Usage:       "session idle timeout duration (e.g., 2h)",
		Description: "Session Idle Timeout",
	})
	sessionCleanupInterval = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "session-cleanup-interval",
		Shorthand:   "",
		Value:       defaultSessionCleanupInterval,
		Usage:       "interval for cleaning up expired sessions (e.g., 1h)",
		Description: "Session Cleanup Interval",
	})
)
