// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	help = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:      "help",
		Shorthand: "h",
		Value:     defaultHelp,
		Usage: "print help; you can run the server with parameters like this:\n" +
			"cmd -l=INFO -v -M -u=postgres://USR:PWD@localhost:5432/DB?sslmode=disable\n", // pragma: allowlist secret
		Description: "Help",
	})
	configFile = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "config",
		Shorthand:   "y",
		Value:       defaultConfigFiles,
		Usage:       "path to config file (can be specified multiple times)",
		Description: "Config files",
	})
	logLevel = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "log-level",
		Shorthand:   "l",
		Value:       defaultLogLevel,
		Usage:       "log level: ALL, TRACE, DEBUG, CONFIG, INFO, NOTICE, WARN, ERROR, FATAL, OFF",
		Description: "Log Level",
	})
	verboseMode = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "verbose",
		Shorthand:   "v",
		Value:       defaultVerboseMode,
		Usage:       "verbose modifier for log level",
		Description: "Verbose mode",
	})
	devMode = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "dev",
		Shorthand:   "d",
		Value:       defaultDevMode,
		Usage:       "run in development mode; enables in-memory SQLite",
		Description: "Dev mode",
	})
	demoMode = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "demo",
		Shorthand:   "X",
		Value:       defaultDemoMode,
		Usage:       "run in demo mode; auto-seeds demo data on startup",
		Description: "Demo mode",
	})
	resetDemoMode = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "reset-demo",
		Shorthand:   "g",
		Value:       defaultResetDemoMode,
		Usage:       "reset demo mode; clears and re-seeds demo data on startup",
		Description: "Reset demo mode",
	})
	dryRun = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "dry-run",
		Shorthand:   "Y",
		Value:       defaultDryRun,
		Usage:       "validate configuration and exit without starting server",
		Description: "Dry run",
	})
	profile = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "profile",
		Shorthand:   "f",
		Value:       defaultProfile,
		Usage:       "configuration profile: dev, stg, prod, test",
		Description: "Configuration profile",
	})
	bindPublicProtocol = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "bind-public-protocol",
		Shorthand:   "t",
		Value:       defaultBindPublicProtocol,
		Usage:       "bind public protocol (http or https)",
		Description: "Bind Public Protocol",
	})
	bindPublicAddress = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "bind-public-address",
		Shorthand:   "a",
		Value:       defaultBindPublicAddress,
		Usage:       "bind public address",
		Description: "Bind Public Address",
	})
	bindPublicPort = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "bind-public-port",
		Shorthand:   "p",
		Value:       defaultBindPublicPort,
		Usage:       "bind public port",
		Description: "Bind Public Port",
	})
	bindPrivateProtocol = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "bind-private-protocol",
		Shorthand:   "T",
		Value:       defaultBindPrivateProtocol,
		Usage:       "bind private protocol (http or https)",
		Description: "Bind Private Protocol",
	})
	bindPrivateAddress = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "bind-private-address",
		Shorthand:   "A",
		Value:       defaultBindPrivateAddress,
		Usage:       "bind private address",
		Description: "Bind Private Address",
	})
	bindPrivatePort = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "bind-private-port",
		Shorthand:   "P",
		Value:       defaultBindPrivatePort,
		Usage:       "bind private port",
		Description: "Bind Private Port",
	})
	tlsPublicDNSNames = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "tls-public-dns-names",
		Shorthand:   "n",
		Value:       defaultTLSPublicDNSNames,
		Usage:       "TLS public DNS names",
		Description: "TLS Public DNS Names",
	})
	tlsPrivateDNSNames = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "tls-private-dns-names",
		Shorthand:   "j",
		Value:       defaultTLSPrivateDNSNames,
		Usage:       "TLS private DNS names",
		Description: "TLS Private DNS Names",
	})
	tlsPublicIPAddresses = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "tls-public-ip-addresses",
		Shorthand:   "i",
		Value:       defaultTLSPublicIPAddresses,
		Usage:       "TLS public IP addresses",
		Description: "TLS Public IP Addresses",
	})
	tlsPrivateIPAddresses = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "tls-private-ip-addresses",
		Shorthand:   "k",
		Value:       defaultTLSPrivateIPAddresses,
		Usage:       "TLS private IP addresses",
		Description: "TLS Private IP Addresses",
	})
	tlsPublicMode = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "tls-public-mode",
		Shorthand:   "",
		Value:       defaultTLSPublicMode,
		Usage:       "TLS public mode (static, mixed, auto)",
		Description: "TLS Public Mode",
	})
	tlsPrivateMode = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "tls-private-mode",
		Shorthand:   "",
		Value:       defaultTLSPrivateMode,
		Usage:       "TLS private mode (static, mixed, auto)",
		Description: "TLS Private Mode",
	})
	tlsStaticCertPEM = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "tls-static-cert-pem",
		Shorthand:   "",
		Value:       defaultTLSStaticCertPEM,
		Usage:       "TLS static cert PEM (for static mode)",
		Description: "TLS Static Cert PEM",
		Redacted:    true,
	})
	tlsStaticKeyPEM = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "tls-static-key-pem",
		Shorthand:   "",
		Value:       defaultTLSStaticKeyPEM,
		Usage:       "TLS static key PEM (for static mode)",
		Description: "TLS Static Key PEM",
		Redacted:    true,
	})
	tlsMixedCACertPEM = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "tls-mixed-ca-cert-pem",
		Shorthand:   "",
		Value:       defaultTLSMixedCACertPEM,
		Usage:       "TLS mixed CA cert PEM (for mixed mode)",
		Description: "TLS Mixed CA Cert PEM",
		Redacted:    true,
	})
	tlsMixedCAKeyPEM = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "tls-mixed-ca-key-pem",
		Shorthand:   "",
		Value:       defaultTLSMixedCAKeyPEM,
		Usage:       "TLS mixed CA key PEM (for mixed mode)",
		Description: "TLS Mixed CA Key PEM",
		Redacted:    true,
	})
	publicBrowserAPIContextPath = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "browser-api-context-path",
		Shorthand:   "c",
		Value:       defaultPublicBrowserAPIContextPath,
		Usage:       "context path for Public Browser API",
		Description: "Public Browser API Context Path",
	})
	publicServiceAPIContextPath = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "service-api-context-path",
		Shorthand:   "b",
		Value:       defaultPublicServiceAPIContextPath,
		Usage:       "context path for Public Server API",
		Description: "Public Service API Context Path",
	})
	privateAdminAPIContextPath = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "admin-api-context-path",
		Shorthand:   "",
		Value:       cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath,
		Usage:       "context path for Private Admin API",
		Description: "Private Admin API Context Path",
	})
	corsAllowedOrigins = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "cors-origins",
		Shorthand:   "o",
		Value:       defaultCORSAllowedOrigins,
		Usage:       "CORS allowed origins",
		Description: "CORS Allowed Origins",
	})
	corsAllowedMethods = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "cors-methods",
		Shorthand:   "m",
		Value:       defaultCORSAllowedMethods,
		Usage:       "CORS allowed methods",
		Description: "CORS Allowed Methods",
	})
	corsAllowedHeaders = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "cors-headers",
		Shorthand:   "H",
		Value:       defaultCORSAllowedHeaders,
		Usage:       "CORS allowed headers",
		Description: "CORS Allowed Headers",
	})
	corsMaxAge = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "cors-max-age",
		Shorthand:   "x",
		Value:       defaultCORSMaxAge,
		Usage:       "CORS max age in seconds",
		Description: "CORS Max Age",
	})
	csrfTokenName = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "csrf-token-name",
		Shorthand:   "N",
		Value:       defaultCSRFTokenName,
		Usage:       "CSRF token name",
		Description: "CSRF Token Name",
	})
	csrfTokenSameSite = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "csrf-token-same-site",
		Shorthand:   "S",
		Value:       defaultCSRFTokenSameSite,
		Usage:       "CSRF token SameSite attribute",
		Description: "CSRF Token SameSite",
	})
	csrfTokenMaxAge = *SetEnvAndRegisterSetting(allServeiceTemplateServerRegisteredSettings, &Setting{
		Name:        "csrf-token-max-age",
		Shorthand:   "M",
		Value:       defaultCSRFTokenMaxAge,
		Usage:       "CSRF token max age (expiration)",
		Description: "CSRF Token Max Age",
	})
)
