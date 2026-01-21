// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	googleUuid "github.com/google/uuid"

	"github.com/gofiber/fiber/v2/log"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// TLSMode defines the three supported TLS certificate provisioning modes.
type TLSMode string

const (
	// TLSModeStatic uses pre-generated TLS certificates (production).
	// Requires: TLS certificate chain (PEM), private key (PEM).
	// Source: Docker secrets, Kubernetes secrets, CA-signed certificates.
	TLSModeStatic TLSMode = "static"

	// TLSModeMixed uses static CA to sign dynamically generated server certificates (staging/QA).
	// Requires: CA certificate chain (PEM), CA private key (PEM).
	// Auto-generates: Server certificate signed by provided CA on startup.
	TLSModeMixed TLSMode = "mixed"

	// TLSModeAuto fully auto-generates CA hierarchy and server certificates (development/testing).
	// Requires: Configuration parameters only (DNS names, IP addresses, validity).
	// Auto-generates: 3-tier CA hierarchy (Root → Intermediate → Server).
	TLSModeAuto TLSMode = "auto"
)

// TLSMaterial holds the runtime TLS configuration and certificate pools.
type TLSMaterial struct {
	// Config is the tls.Config for HTTPS servers.
	Config *tls.Config

	// RootCAPool is the certificate pool for root CAs (for client certificate validation).
	RootCAPool *x509.CertPool

	// IntermediateCAPool is the certificate pool for intermediate CAs (for chain building).
	IntermediateCAPool *x509.CertPool
}

const (
	defaultLogLevel                    = cryptoutilMagic.DefaultLogLevelInfo                // Balanced verbosity: shows important events without being overwhelming
	defaultBindPublicProtocol          = cryptoutilMagic.DefaultPublicProtocolCryptoutil    // HTTPS by default for security in production environments
	defaultBindPublicAddress           = cryptoutilMagic.DefaultPublicAddressCryptoutil     // IPv4 loopback prevents external access by default, requires explicit configuration for exposure
	defaultBindPublicPort              = cryptoutilMagic.DefaultPublicPortCryptoutil        // Standard HTTP/HTTPS port, well-known and commonly available
	defaultBindPrivateProtocol         = cryptoutilMagic.DefaultPrivateProtocolCryptoutil   // HTTPS for private API security, even in service-to-service communication
	defaultBindPrivateAddress          = cryptoutilMagic.DefaultPrivateAddressCryptoutil    // IPv4 loopback for private API, only accessible from same machine
	defaultBindPrivatePort             = cryptoutilMagic.DefaultPrivatePortCryptoutil       // Non-standard port to avoid conflicts with other services
	defaultPublicBrowserAPIContextPath = cryptoutilMagic.DefaultPublicBrowserAPIContextPath // RESTful API versioning, separates browser from service APIs
	defaultPublicServiceAPIContextPath = cryptoutilMagic.DefaultPublicServiceAPIContextPath // RESTful API versioning, separates service from browser APIs
	defaultAdminServerAPIContextPath   = cryptoutilMagic.DefaultPrivateAdminAPIContextPath  // RESTful API versioning for admin API
	defaultCORSMaxAge                  = cryptoutilMagic.DefaultCORSMaxAge                  // 1 hour cache for CORS preflight requests, balances performance and freshness
	defaultCSRFTokenName               = cryptoutilMagic.DefaultCSRFTokenName               // Standard CSRF token name, widely recognized by frameworks
	defaultCSRFTokenSameSite           = cryptoutilMagic.DefaultCSRFTokenSameSiteStrict     // Strict SameSite prevents CSRF while maintaining usability
	defaultCSRFTokenMaxAge             = cryptoutilMagic.DefaultCSRFTokenMaxAge             // 1 hour expiration balances security and user experience
	defaultCSRFTokenCookieSecure       = cryptoutilMagic.DefaultCSRFTokenCookieSecure       // Secure cookies in production prevent MITM attacks
	defaultCSRFTokenCookieHTTPOnly     = cryptoutilMagic.DefaultCSRFTokenCookieHTTPOnly     // False allows JavaScript access for form submissions (Swagger UI workaround)
	defaultCSRFTokenCookieSessionOnly  = cryptoutilMagic.DefaultCSRFTokenCookieSessionOnly  // Session-only prevents persistent tracking while maintaining security
	defaultCSRFTokenSingleUseToken     = cryptoutilMagic.DefaultCSRFTokenSingleUseToken     // Reusable tokens for better UX, can be changed for high-security needs
	defaultRequestBodyLimit            = cryptoutilMagic.DefaultHTTPRequestBodyLimit        // 2MB limit prevents large payload attacks while allowing reasonable API usage
	defaultBrowserIPRateLimit          = cryptoutilMagic.DefaultPublicBrowserAPIIPRateLimit // More lenient rate limit for browser APIs (user interactions)
	defaultServiceIPRateLimit          = cryptoutilMagic.DefaultPublicServiceAPIIPRateLimit // More restrictive rate limit for service APIs (automated systems)
	defaultDatabaseContainer           = cryptoutilMagic.DefaultDatabaseContainerDisabled   // Disabled by default to avoid unexpected container dependencies
	defaultDatabaseURL                 = cryptoutilMagic.DefaultDatabaseURL                 // pragma: allowlist secret // PostgreSQL default with placeholder credentials, SSL disabled for local dev
	defaultDatabaseInitTotalTimeout    = cryptoutilMagic.DefaultDatabaseInitTotalTimeout    // 5 minutes allows for container startup while preventing indefinite waits
	defaultDatabaseInitRetryWait       = cryptoutilMagic.DefaultDataInitRetryWait           // 1 second retry interval balances responsiveness and resource usage
	defaultServerShutdownTimeout       = cryptoutilMagic.DefaultDataServerShutdownTimeout   // 5 seconds allows graceful shutdown while preventing indefinite waits
	defaultHelp                        = cryptoutilMagic.DefaultHelp
	defaultVerboseMode                 = cryptoutilMagic.DefaultVerboseMode
	defaultDevMode                     = cryptoutilMagic.DefaultDevMode
	defaultDemoMode                    = cryptoutilMagic.DefaultDemoMode
	defaultResetDemoMode               = cryptoutilMagic.DefaultResetDemoMode
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
	defaultTLSPublicMode               = TLSMode(cryptoutilMagic.DefaultTLSPublicMode)
	defaultTLSPrivateMode              = TLSMode(cryptoutilMagic.DefaultTLSPrivateMode)
	defaultBrowserSessionCookie        = cryptoutilMagic.DefaultBrowserSessionCookie
	defaultBrowserSessionAlgorithm     = cryptoutilMagic.DefaultBrowserSessionAlgorithm
	defaultBrowserSessionJWSAlgorithm  = cryptoutilMagic.DefaultBrowserSessionJWSAlgorithm
	defaultBrowserSessionJWEAlgorithm  = cryptoutilMagic.DefaultBrowserSessionJWEAlgorithm
	defaultBrowserSessionExpiration    = cryptoutilMagic.DefaultBrowserSessionExpiration
	defaultServiceSessionAlgorithm     = cryptoutilMagic.DefaultServiceSessionAlgorithm
	defaultServiceSessionJWSAlgorithm  = cryptoutilMagic.DefaultServiceSessionJWSAlgorithm
	defaultServiceSessionJWEAlgorithm  = cryptoutilMagic.DefaultServiceSessionJWEAlgorithm
	defaultServiceSessionExpiration    = cryptoutilMagic.DefaultServiceSessionExpiration
	defaultSessionIdleTimeout          = cryptoutilMagic.DefaultSessionIdleTimeout
	defaultSessionCleanupInterval      = cryptoutilMagic.DefaultSessionCleanupInterval
)

var (
	defaultTLSStaticCertPEM  = cryptoutilMagic.DefaultTLSStaticCertPEM
	defaultTLSStaticKeyPEM   = cryptoutilMagic.DefaultTLSStaticKeyPEM
	defaultTLSMixedCACertPEM = cryptoutilMagic.DefaultTLSMixedCACertPEM
	defaultTLSMixedCAKeyPEM  = cryptoutilMagic.DefaultTLSMixedCAKeyPEM
	defaultBrowserRealms     = cryptoutilMagic.DefaultBrowserRealms
	defaultServiceRealms     = cryptoutilMagic.DefaultServiceRealms
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

var defaultSwaggerUIUsername = ""

var defaultSwaggerUIPassword = ""

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

var allServeiceTemplateServerRegisteredSettings []*Setting

// ServiceTemplateServerSettings contains all configuration settings for the service template server.
type ServiceTemplateServerSettings struct {
	SubCommand                  string
	Help                        bool
	ConfigFile                  []string
	LogLevel                    string
	VerboseMode                 bool
	DevMode                     bool
	DemoMode                    bool
	ResetDemoMode               bool
	DryRun                      bool
	Profile                     string // Configuration profile: dev, stg, prod, test
	BindPublicProtocol          string
	BindPublicAddress           string
	BindPublicPort              uint16
	BindPrivateProtocol         string
	BindPrivateAddress          string
	BindPrivatePort             uint16
	TLSPublicMode               TLSMode // Default TLSModeAuto
	TLSPublicDNSNames           []string
	TLSPublicIPAddresses        []string
	TLSPrivateDNSNames          []string
	TLSPrivateMode              TLSMode // Default TLSModeAuto
	TLSPrivateIPAddresses       []string
	TLSStaticCertPEM            []byte // Default nil. PEM-encoded certificate chain (for TLSModeStatic). Should contain: [Server Cert, Intermediate CA(s), Root CA] or [Server Cert, Root CA].
	TLSStaticKeyPEM             []byte // Default nil. PEM-encoded private key (for TLSModeStatic).
	TLSMixedCACertPEM           []byte // Default nil. PEM-encoded CA certificate chain (for TLSModeMixed). Should contain: [Intermediate CA(s), Root CA] or [Root CA].
	TLSMixedCAKeyPEM            []byte // Default nil. PEM-encoded CA private key (for TLSModeMixed).
	PublicBrowserAPIContextPath string
	PublicServiceAPIContextPath string
	PrivateAdminAPIContextPath  string
	CORSAllowedOrigins          []string
	CORSAllowedMethods          []string
	CORSAllowedHeaders          []string
	CORSMaxAge                  uint16
	CSRFTokenName               string
	CSRFTokenSameSite           string
	CSRFTokenMaxAge             time.Duration
	CSRFTokenCookieSecure       bool
	CSRFTokenCookieHTTPOnly     bool
	CSRFTokenCookieSessionOnly  bool
	CSRFTokenSingleUseToken     bool
	RequestBodyLimit            int
	BrowserIPRateLimit          uint16
	ServiceIPRateLimit          uint16
	AllowedIPs                  []string
	AllowedCIDRs                []string
	SwaggerUIUsername           string
	SwaggerUIPassword           string
	DatabaseContainer           string
	DatabaseURL                 string
	DatabaseInitTotalTimeout    time.Duration
	DatabaseInitRetryWait       time.Duration
	ServerShutdownTimeout       time.Duration
	OTLPEnabled                 bool
	OTLPConsole                 bool
	OTLPService                 string
	OTLPInstance                string
	OTLPVersion                 string
	OTLPEnvironment             string
	OTLPHostname                string
	OTLPEndpoint                string
	UnsealMode                  string
	UnsealFiles                 []string
	BrowserRealms               []string      // Paths to browser realm configuration files (session-based auth)
	ServiceRealms               []string      // Paths to service realm configuration files (token-based auth)
	BrowserSessionCookie        string        // Cookie type: jwe (encrypted), jws (signed), opaque (database) - DEPRECATED: use BrowserSessionAlgorithm
	BrowserSessionAlgorithm     string        // Session algorithm: OPAQUE (hashed), JWS (signed JWT), JWE (encrypted JWT)
	BrowserSessionJWSAlgorithm  string        // JWS algorithm for browser sessions (e.g., RS256, ES256, EdDSA)
	BrowserSessionJWEAlgorithm  string        // JWE algorithm for browser sessions (e.g., dir+A256GCM, A256GCMKW+A256GCM)
	BrowserSessionExpiration    time.Duration // Browser session expiration duration
	ServiceSessionAlgorithm     string        // Session algorithm: OPAQUE (hashed), JWS (signed JWT), JWE (encrypted JWT)
	ServiceSessionJWSAlgorithm  string        // JWS algorithm for service sessions (e.g., RS256, ES256, EdDSA)
	ServiceSessionJWEAlgorithm  string        // JWE algorithm for service sessions (e.g., dir+A256GCM, A256GCMKW+A256GCM)
	ServiceSessionExpiration    time.Duration // Service session expiration duration
	SessionIdleTimeout          time.Duration // Session idle timeout duration
	SessionCleanupInterval      time.Duration // Interval for cleaning up expired sessions
}

// PrivateBaseURL returns the private base URL constructed from protocol, address, and port.
func (s *ServiceTemplateServerSettings) PrivateBaseURL() string {
	return fmt.Sprintf("%s://%s:%d", s.BindPrivateProtocol, s.BindPrivateAddress, s.BindPrivatePort)
}

// PublicBaseURL returns the public base URL constructed from protocol, address, and port.
func (s *ServiceTemplateServerSettings) PublicBaseURL() string {
	return fmt.Sprintf("%s://%s:%d", s.BindPublicProtocol, s.BindPublicAddress, s.BindPublicPort)
}

// Setting Input values for pflag.*P(name, shortname, value, usage).
type Setting struct {
	Name        string // unique long name for the flag
	Env         string // unique environment variable name for the flag
	Shorthand   string // unique short name for the flag
	Value       any    // default value for the flag
	Usage       string // description of the flag for help text
	Description string // human-readable description for logging/display
	Redacted    bool   // whether to redact the value in logs (except in dev+verbose mode)
}

type analysisResult struct {
	SettingsByNames      map[string][]*Setting
	SettingsByShorthands map[string][]*Setting
	DuplicateNames       []string
	DuplicateShorthands  []string
}

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
		Value:       cryptoutilMagic.DefaultPrivateAdminAPIContextPath,
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

// getTLSPEMBytes safely retrieves PEM bytes from viper for BytesBase64 flags.
// Returns nil if the value is not set or cannot be converted to []byte.
func getTLSPEMBytes(key string) []byte {
	val := viper.Get(key)
	if val == nil {
		return nil
	}

	// BytesBase64P flags are stored in viper as strings (base64-encoded)
	// We need to decode them manually
	if str, ok := val.(string); ok {
		if str == "" {
			return nil
		}

		bytes, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			return nil
		}

		return bytes
	}

	// Fallback: if already []byte (e.g., from config file), use as-is
	if bytes, ok := val.([]byte); ok {
		return bytes
	}

	return nil
}

// Parse parses command line parameters and returns application settings.
func Parse(commandParameters []string, exitIfHelp bool) (*ServiceTemplateServerSettings, error) {
	if len(commandParameters) == 0 {
		return nil, fmt.Errorf("missing subcommand: use \"start\", \"stop\", \"init\", \"live\", or \"ready\"")
	}

	subCommand := commandParameters[0]
	if _, ok := subcommands[subCommand]; !ok {
		return nil, fmt.Errorf("invalid subcommand: use \"start\", \"stop\", \"init\", \"live\", or \"ready\"")
	}

	subCommandParameters := commandParameters[1:]

	// Enable environment variable support with CRYPTOUTIL_ prefix BEFORE parsing flags
	viper.SetEnvPrefix("CRYPTOUTIL")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// Explicitly bind boolean environment variables (viper.AutomaticEnv may not handle booleans correctly)
	// Note: viper.BindEnv errors are logged but don't prevent startup as they are extremely rare
	for _, setting := range allServeiceTemplateServerRegisteredSettings {
		if _, ok := setting.Value.(bool); ok {
			if err := viper.BindEnv(setting.Name, setting.Env); err != nil {
				fmt.Printf("Warning: failed to bind environment variable %s: %v\n", setting.Env, err)
			}
		}
	}

	// pflag will parse subCommandParameters, and viper will union them with config file contents (if specified)
	pflag.BoolP(help.Name, help.Shorthand, RegisterAsBoolSetting(&help), help.Usage)
	pflag.StringSliceP(configFile.Name, configFile.Shorthand, RegisterAsStringSliceSetting(&configFile), configFile.Usage)
	pflag.StringP(logLevel.Name, logLevel.Shorthand, RegisterAsStringSetting(&logLevel), logLevel.Usage)
	pflag.BoolP(verboseMode.Name, verboseMode.Shorthand, RegisterAsBoolSetting(&verboseMode), verboseMode.Usage)
	pflag.BoolP(devMode.Name, devMode.Shorthand, RegisterAsBoolSetting(&devMode), devMode.Usage)
	pflag.BoolP(demoMode.Name, demoMode.Shorthand, RegisterAsBoolSetting(&demoMode), demoMode.Usage)
	pflag.BoolP(dryRun.Name, dryRun.Shorthand, RegisterAsBoolSetting(&dryRun), dryRun.Usage)
	pflag.StringP(profile.Name, profile.Shorthand, RegisterAsStringSetting(&profile), profile.Usage)
	pflag.StringP(bindPublicProtocol.Name, bindPublicProtocol.Shorthand, RegisterAsStringSetting(&bindPublicProtocol), bindPublicProtocol.Usage)
	pflag.StringP(bindPublicAddress.Name, bindPublicAddress.Shorthand, RegisterAsStringSetting(&bindPublicAddress), bindPublicAddress.Usage)
	pflag.Uint16P(bindPublicPort.Name, bindPublicPort.Shorthand, RegisterAsUint16Setting(&bindPublicPort), bindPublicPort.Usage)
	pflag.StringSliceP(tlsPublicDNSNames.Name, tlsPublicDNSNames.Shorthand, RegisterAsStringSliceSetting(&tlsPublicDNSNames), tlsPublicDNSNames.Usage)
	pflag.StringSliceP(tlsPublicIPAddresses.Name, tlsPublicIPAddresses.Shorthand, RegisterAsStringSliceSetting(&tlsPublicIPAddresses), tlsPublicIPAddresses.Usage)
	pflag.StringSliceP(tlsPrivateDNSNames.Name, tlsPrivateDNSNames.Shorthand, RegisterAsStringSliceSetting(&tlsPrivateDNSNames), tlsPrivateDNSNames.Usage)
	pflag.StringSliceP(tlsPrivateIPAddresses.Name, tlsPrivateIPAddresses.Shorthand, RegisterAsStringSliceSetting(&tlsPrivateIPAddresses), tlsPrivateIPAddresses.Usage)
	pflag.StringP(tlsPublicMode.Name, tlsPublicMode.Shorthand, string(defaultTLSPublicMode), tlsPublicMode.Usage)
	pflag.StringP(tlsPrivateMode.Name, tlsPrivateMode.Shorthand, string(defaultTLSPrivateMode), tlsPrivateMode.Usage)
	pflag.BytesBase64P(tlsStaticCertPEM.Name, tlsStaticCertPEM.Shorthand, []byte(nil), tlsStaticCertPEM.Usage)
	pflag.BytesBase64P(tlsStaticKeyPEM.Name, tlsStaticKeyPEM.Shorthand, []byte(nil), tlsStaticKeyPEM.Usage)
	pflag.BytesBase64P(tlsMixedCACertPEM.Name, tlsMixedCACertPEM.Shorthand, []byte(nil), tlsMixedCACertPEM.Usage)
	pflag.BytesBase64P(tlsMixedCAKeyPEM.Name, tlsMixedCAKeyPEM.Shorthand, []byte(nil), tlsMixedCAKeyPEM.Usage)
	pflag.StringP(bindPrivateProtocol.Name, bindPrivateProtocol.Shorthand, RegisterAsStringSetting(&bindPrivateProtocol), bindPrivateProtocol.Usage)
	pflag.StringP(bindPrivateAddress.Name, bindPrivateAddress.Shorthand, RegisterAsStringSetting(&bindPrivateAddress), bindPrivateAddress.Usage)
	pflag.Uint16P(bindPrivatePort.Name, bindPrivatePort.Shorthand, RegisterAsUint16Setting(&bindPrivatePort), bindPrivatePort.Usage)
	pflag.StringP(publicBrowserAPIContextPath.Name, publicBrowserAPIContextPath.Shorthand, RegisterAsStringSetting(&publicBrowserAPIContextPath), publicBrowserAPIContextPath.Usage)
	pflag.StringP(publicServiceAPIContextPath.Name, publicServiceAPIContextPath.Shorthand, RegisterAsStringSetting(&publicServiceAPIContextPath), publicServiceAPIContextPath.Usage)
	pflag.StringP(privateAdminAPIContextPath.Name, privateAdminAPIContextPath.Shorthand, RegisterAsStringSetting(&privateAdminAPIContextPath), privateAdminAPIContextPath.Usage)
	pflag.StringSliceP(corsAllowedOrigins.Name, corsAllowedOrigins.Shorthand, RegisterAsStringSliceSetting(&corsAllowedOrigins), corsAllowedOrigins.Usage)
	pflag.StringSliceP(corsAllowedMethods.Name, corsAllowedMethods.Shorthand, RegisterAsStringSliceSetting(&corsAllowedMethods), corsAllowedMethods.Usage)
	pflag.StringSliceP(corsAllowedHeaders.Name, corsAllowedHeaders.Shorthand, RegisterAsStringSliceSetting(&corsAllowedHeaders), corsAllowedHeaders.Usage)
	pflag.Uint16P(corsMaxAge.Name, corsMaxAge.Shorthand, RegisterAsUint16Setting(&corsMaxAge), corsMaxAge.Usage)
	pflag.StringP(csrfTokenName.Name, csrfTokenName.Shorthand, RegisterAsStringSetting(&csrfTokenName), csrfTokenName.Usage)
	pflag.StringP(csrfTokenSameSite.Name, csrfTokenSameSite.Shorthand, RegisterAsStringSetting(&csrfTokenSameSite), csrfTokenSameSite.Usage)
	pflag.DurationP(csrfTokenMaxAge.Name, csrfTokenMaxAge.Shorthand, RegisterAsDurationSetting(&csrfTokenMaxAge), csrfTokenMaxAge.Usage)
	pflag.BoolP(csrfTokenCookieSecure.Name, csrfTokenCookieSecure.Shorthand, RegisterAsBoolSetting(&csrfTokenCookieSecure), csrfTokenCookieSecure.Usage)
	pflag.BoolP(csrfTokenCookieHTTPOnly.Name, csrfTokenCookieHTTPOnly.Shorthand, RegisterAsBoolSetting(&csrfTokenCookieHTTPOnly), csrfTokenCookieHTTPOnly.Usage)
	pflag.BoolP(csrfTokenCookieSessionOnly.Name, csrfTokenCookieSessionOnly.Shorthand, RegisterAsBoolSetting(&csrfTokenCookieSessionOnly), csrfTokenCookieSessionOnly.Usage)
	pflag.BoolP(csrfTokenSingleUseToken.Name, csrfTokenSingleUseToken.Shorthand, RegisterAsBoolSetting(&csrfTokenSingleUseToken), csrfTokenSingleUseToken.Usage)
	pflag.Uint16P(browserIPRateLimit.Name, browserIPRateLimit.Shorthand, RegisterAsUint16Setting(&browserIPRateLimit), browserIPRateLimit.Usage)
	pflag.Uint16P(serviceIPRateLimit.Name, serviceIPRateLimit.Shorthand, RegisterAsUint16Setting(&serviceIPRateLimit), serviceIPRateLimit.Usage)
	pflag.StringSliceP(allowedIps.Name, allowedIps.Shorthand, RegisterAsStringSliceSetting(&allowedIps), allowedIps.Usage)
	pflag.StringSliceP(allowedCidrs.Name, allowedCidrs.Shorthand, RegisterAsStringSliceSetting(&allowedCidrs), allowedCidrs.Usage)
	pflag.IntP(requestBodyLimit.Name, requestBodyLimit.Shorthand, RegisterAsIntSetting(&requestBodyLimit), requestBodyLimit.Usage)
	pflag.StringP(databaseContainer.Name, databaseContainer.Shorthand, RegisterAsStringSetting(&databaseContainer), databaseContainer.Usage)
	pflag.StringP(databaseURL.Name, databaseURL.Shorthand, RegisterAsStringSetting(&databaseURL), databaseURL.Usage)
	pflag.DurationP(databaseInitTotalTimeout.Name, databaseInitTotalTimeout.Shorthand, RegisterAsDurationSetting(&databaseInitTotalTimeout), databaseInitTotalTimeout.Usage)
	pflag.DurationP(databaseInitRetryWait.Name, databaseInitRetryWait.Shorthand, RegisterAsDurationSetting(&databaseInitRetryWait), databaseInitRetryWait.Usage)
	pflag.DurationP(serverShutdownTimeout.Name, serverShutdownTimeout.Shorthand, RegisterAsDurationSetting(&serverShutdownTimeout), serverShutdownTimeout.Usage)
	pflag.BoolP(otlpEnabled.Name, otlpEnabled.Shorthand, RegisterAsBoolSetting(&otlpEnabled), otlpEnabled.Usage)
	pflag.BoolP(otlpConsole.Name, otlpConsole.Shorthand, RegisterAsBoolSetting(&otlpConsole), otlpConsole.Usage)
	pflag.StringP(otlpService.Name, otlpService.Shorthand, RegisterAsStringSetting(&otlpService), otlpService.Usage)
	pflag.StringP(otlpVersion.Name, otlpVersion.Shorthand, RegisterAsStringSetting(&otlpVersion), otlpVersion.Usage)
	pflag.StringP(otlpEnvironment.Name, otlpEnvironment.Shorthand, RegisterAsStringSetting(&otlpEnvironment), otlpEnvironment.Usage)
	pflag.StringP(otlpHostname.Name, otlpHostname.Shorthand, RegisterAsStringSetting(&otlpHostname), otlpHostname.Usage)
	pflag.StringP(otlpEndpoint.Name, otlpEndpoint.Shorthand, RegisterAsStringSetting(&otlpEndpoint), otlpEndpoint.Usage)
	pflag.StringP(otlpInstance.Name, otlpInstance.Shorthand, RegisterAsStringSetting(&otlpInstance), otlpInstance.Usage)
	pflag.StringP(unsealMode.Name, unsealMode.Shorthand, RegisterAsStringSetting(&unsealMode), unsealMode.Usage)
	pflag.StringArrayP(unsealFiles.Name, unsealFiles.Shorthand, RegisterAsStringArraySetting(&unsealFiles), unsealFiles.Usage)
	pflag.StringSliceP(browserRealms.Name, browserRealms.Shorthand, RegisterAsStringSliceSetting(&browserRealms), browserRealms.Usage)
	pflag.StringSliceP(serviceRealms.Name, serviceRealms.Shorthand, RegisterAsStringSliceSetting(&serviceRealms), serviceRealms.Usage)

	err := pflag.CommandLine.Parse(subCommandParameters)
	if err != nil {
		return nil, fmt.Errorf("error parsing flags: %w", err)
	}

	err = viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return nil, fmt.Errorf("failed to bind flags: %w", err)
	}

	// Enable environment variable support for all configuration settings.
	// Environment variables use CRYPTOUTIL_ prefix with underscores instead of hyphens.
	// Example: CRYPTOUTIL_DATABASE_URL overrides --database-url flag.
	// Precedence: flags > env vars > config files > defaults
	viper.AutomaticEnv()
	viper.SetEnvPrefix("CRYPTOUTIL")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	configFiles := viper.GetStringSlice(configFile.Name)
	if len(configFiles) > 0 {
		// Set the first config file
		if info, err := os.Stat(configFiles[0]); err == nil && !info.IsDir() {
			viper.SetConfigFile(configFiles[0])

			if err := viper.ReadInConfig(); err != nil {
				return nil, fmt.Errorf("error reading config file %s: %w", configFiles[0], err)
			}
		}
		// Merge additional config files
		for i := 1; i < len(configFiles); i++ {
			if info, err := os.Stat(configFiles[i]); err == nil && !info.IsDir() {
				viper.SetConfigFile(configFiles[i])

				if err := viper.MergeInConfig(); err != nil {
					return nil, fmt.Errorf("error merging config file %s: %w", configFiles[i], err)
				}
			}
		}
	}

	// Apply configuration profile if specified
	profileName := viper.GetString(profile.Name)
	if profileName != "" {
		if profileConfig, exists := profiles[profileName]; exists {
			// Apply profile settings (these can be overridden by config files or command line flags)
			for key, value := range profileConfig {
				if !viper.IsSet(key) {
					viper.Set(key, value)
				}
			}
		} else {
			return nil, fmt.Errorf("unknown configuration profile: %s (available: dev, stg, prod, test)", profileName)
		}
	}

	// Parse TLS mode and PEM fields
	tlsPublicModeStr := viper.GetString(tlsPublicMode.Name)
	if tlsPublicModeStr == "" {
		tlsPublicModeStr = string(defaultTLSPublicMode)
	}

	tlsPrivateModeStr := viper.GetString(tlsPrivateMode.Name)
	if tlsPrivateModeStr == "" {
		tlsPrivateModeStr = string(defaultTLSPrivateMode)
	}

	s := &ServiceTemplateServerSettings{
		TLSPublicMode:               TLSMode(tlsPublicModeStr),
		TLSPrivateMode:              TLSMode(tlsPrivateModeStr),
		TLSStaticCertPEM:            getTLSPEMBytes(tlsStaticCertPEM.Name),
		TLSStaticKeyPEM:             getTLSPEMBytes(tlsStaticKeyPEM.Name),
		TLSMixedCACertPEM:           getTLSPEMBytes(tlsMixedCACertPEM.Name),
		TLSMixedCAKeyPEM:            getTLSPEMBytes(tlsMixedCAKeyPEM.Name),
		SubCommand:                  subCommand,
		Help:                        viper.GetBool(help.Name),
		ConfigFile:                  viper.GetStringSlice(configFile.Name),
		LogLevel:                    viper.GetString(logLevel.Name),
		VerboseMode:                 viper.GetBool(verboseMode.Name),
		DevMode:                     viper.GetBool(devMode.Name),
		DemoMode:                    viper.GetBool(demoMode.Name),
		ResetDemoMode:               viper.GetBool(resetDemoMode.Name),
		DryRun:                      viper.GetBool(dryRun.Name),
		Profile:                     viper.GetString(profile.Name),
		BindPublicProtocol:          viper.GetString(bindPublicProtocol.Name),
		BindPublicAddress:           viper.GetString(bindPublicAddress.Name),
		BindPublicPort:              viper.GetUint16(bindPublicPort.Name),
		TLSPublicDNSNames:           viper.GetStringSlice(tlsPublicDNSNames.Name),
		TLSPublicIPAddresses:        viper.GetStringSlice(tlsPublicIPAddresses.Name),
		TLSPrivateDNSNames:          viper.GetStringSlice(tlsPrivateDNSNames.Name),
		TLSPrivateIPAddresses:       viper.GetStringSlice(tlsPrivateIPAddresses.Name),
		BindPrivateProtocol:         viper.GetString(bindPrivateProtocol.Name),
		BindPrivateAddress:          viper.GetString(bindPrivateAddress.Name),
		BindPrivatePort:             viper.GetUint16(bindPrivatePort.Name),
		PublicBrowserAPIContextPath: viper.GetString(publicBrowserAPIContextPath.Name),
		PublicServiceAPIContextPath: viper.GetString(publicServiceAPIContextPath.Name),
		PrivateAdminAPIContextPath:  viper.GetString(privateAdminAPIContextPath.Name),
		CORSAllowedOrigins:          viper.GetStringSlice(corsAllowedOrigins.Name),
		CORSAllowedMethods:          viper.GetStringSlice(corsAllowedMethods.Name),
		CORSAllowedHeaders:          viper.GetStringSlice(corsAllowedHeaders.Name),
		CORSMaxAge:                  viper.GetUint16(corsMaxAge.Name),
		RequestBodyLimit:            viper.GetInt(requestBodyLimit.Name),
		CSRFTokenName:               viper.GetString(csrfTokenName.Name),
		CSRFTokenSameSite:           viper.GetString(csrfTokenSameSite.Name),
		CSRFTokenMaxAge:             viper.GetDuration(csrfTokenMaxAge.Name),
		CSRFTokenCookieSecure:       viper.GetBool(csrfTokenCookieSecure.Name),
		CSRFTokenCookieHTTPOnly:     viper.GetBool(csrfTokenCookieHTTPOnly.Name),
		CSRFTokenCookieSessionOnly:  viper.GetBool(csrfTokenCookieSessionOnly.Name),
		CSRFTokenSingleUseToken:     viper.GetBool(csrfTokenSingleUseToken.Name),
		BrowserIPRateLimit:          viper.GetUint16(browserIPRateLimit.Name),
		ServiceIPRateLimit:          viper.GetUint16(serviceIPRateLimit.Name),
		AllowedIPs:                  viper.GetStringSlice(allowedIps.Name),
		AllowedCIDRs:                viper.GetStringSlice(allowedCidrs.Name),
		DatabaseContainer:           viper.GetString(databaseContainer.Name),
		DatabaseURL:                 viper.GetString(databaseURL.Name),
		DatabaseInitTotalTimeout:    viper.GetDuration(databaseInitTotalTimeout.Name),
		DatabaseInitRetryWait:       viper.GetDuration(databaseInitRetryWait.Name),
		ServerShutdownTimeout:       viper.GetDuration(serverShutdownTimeout.Name),
		OTLPEnabled:                 viper.GetBool(otlpEnabled.Name),
		OTLPConsole:                 viper.GetBool(otlpConsole.Name),
		OTLPService:                 viper.GetString(otlpService.Name),
		OTLPInstance:                viper.GetString(otlpInstance.Name),
		OTLPVersion:                 viper.GetString(otlpVersion.Name),
		OTLPEnvironment:             viper.GetString(otlpEnvironment.Name),
		OTLPHostname:                viper.GetString(otlpHostname.Name),
		OTLPEndpoint:                viper.GetString(otlpEndpoint.Name),
		UnsealMode:                  viper.GetString(unsealMode.Name),
		UnsealFiles:                 viper.GetStringSlice(unsealFiles.Name),
		BrowserRealms:               viper.GetStringSlice(browserRealms.Name),
		ServiceRealms:               viper.GetStringSlice(serviceRealms.Name),
		BrowserSessionCookie:        viper.GetString(browserSessionCookie.Name),
		BrowserSessionAlgorithm:     viper.GetString(browserSessionAlgorithm.Name),
		BrowserSessionJWSAlgorithm:  viper.GetString(browserSessionJWSAlgorithm.Name),
		BrowserSessionJWEAlgorithm:  viper.GetString(browserSessionJWEAlgorithm.Name),
		BrowserSessionExpiration:    viper.GetDuration(browserSessionExpiration.Name),
		ServiceSessionAlgorithm:     viper.GetString(serviceSessionAlgorithm.Name),
		ServiceSessionJWSAlgorithm:  viper.GetString(serviceSessionJWSAlgorithm.Name),
		ServiceSessionJWEAlgorithm:  viper.GetString(serviceSessionJWEAlgorithm.Name),
		ServiceSessionExpiration:    viper.GetDuration(serviceSessionExpiration.Name),
		SessionIdleTimeout:          viper.GetDuration(sessionIdleTimeout.Name),
		SessionCleanupInterval:      viper.GetDuration(sessionCleanupInterval.Name),
	}

	// Resolve file:// URLs for sensitive settings from Docker secrets or Kubernetes secrets.
	// This allows configuration to reference secret files rather than embedding sensitive values directly.
	s.DatabaseURL = resolveFileURL(s.DatabaseURL)

	logSettings(s)

	if s.Help {
		pflag.CommandLine.SetOutput(os.Stdout)
		fmt.Println("cryptoutil - Cryptographic utility server")
		fmt.Println()
		fmt.Println("USAGE:")
		fmt.Println("  cryptoutil [subcommand] [flags]")
		fmt.Println()
		fmt.Println("SUBCOMMANDS:")
		fmt.Println("  start    Start the server")
		fmt.Println("  stop     Stop the server")
		fmt.Println("  init     Initialize the server")
		fmt.Println("  live     Check server liveness")
		fmt.Println("  ready    Check server readiness")
		fmt.Println()
		fmt.Println("CONFIGURATION SETTINGS:")
		fmt.Println("  -d, --dev                           run in development mode; enables in-memory SQLite")
		fmt.Println("  -h, --help                          print help")
		fmt.Println("  -y, --config strings                path to config file (can be specified multiple times)")
		fmt.Println("  -Y, --dry-run                       validate configuration and exit without starting server")
		fmt.Println("  -P, --profile strings                configuration profile: dev, stg, prod, test")
		fmt.Println()
		fmt.Println("DATABASE SETTINGS:")
		fmt.Println("  -u, --database-url string           database URL (default " + formatDefault(defaultDatabaseURL) + ")")
		fmt.Println("  -D, --database-container string     database container mode (default " + formatDefault(defaultDatabaseContainer) + ")")
		fmt.Println("  -Z, --database-init-total-timeout duration database init total timeout (default " + formatDefault(defaultDatabaseInitTotalTimeout) + ")")
		fmt.Println("  -W, --database-init-retry-wait duration database init retry wait (default " + formatDefault(defaultDatabaseInitRetryWait) + ")")
		fmt.Println()
		fmt.Println("SERVER SETTINGS:")
		fmt.Println("  -a, --bind-public-address string    bind public address (default " + formatDefault(defaultBindPublicAddress) + ")")
		fmt.Println("  -p, --bind-public-port uint16       bind public port (default " + formatDefault(defaultBindPublicPort) + ")")
		fmt.Println("  -t, --bind-public-protocol string   bind public protocol (http or https) (default " + formatDefault(defaultBindPublicProtocol) + ")")
		fmt.Println("  -A, --bind-private-address string   bind private address (default " + formatDefault(defaultBindPrivateAddress) + ")")
		fmt.Println("  -P, --bind-private-port uint16      bind private port (default " + formatDefault(defaultBindPrivatePort) + ")")
		fmt.Println("  -T, --bind-private-protocol string  bind private protocol (http or https) (default " + formatDefault(defaultBindPrivateProtocol) + ")")
		fmt.Println("  -c, --browser-api-context-path string  context path for Public Browser API (default " + formatDefault(defaultPublicBrowserAPIContextPath) + ")")
		fmt.Println("  -b, --service-api-context-path string  context path for Public Service API (default " + formatDefault(defaultPublicServiceAPIContextPath) + ")")
		fmt.Println()
		fmt.Println("NETWORK SECURITY SETTINGS:")
		fmt.Println("  -I, --allowed-ips strings           comma-separated list of allowed IPs (default " + formatDefault(defaultAllowedIps) + ")")
		fmt.Println("  -C, --allowed-cidrs strings         comma-separated list of allowed CIDRs (default " + formatDefault(defaultAllowedCIDRs) + ")")
		fmt.Println("  -e, --browser-rate-limit uint16     rate limit for browser API requests per second (default " + formatDefault(defaultBrowserIPRateLimit) + ")")
		fmt.Println("  -w, --service-rate-limit uint16     rate limit for service API requests per second (default " + formatDefault(defaultServiceIPRateLimit) + ")")
		fmt.Println("  -L, --request-body-limit int        Maximum request body size in bytes (default " + formatDefault(defaultRequestBodyLimit) + ")")
		fmt.Println()
		fmt.Println("SWAGGER UI SETTINGS:")
		fmt.Println("      --swagger-ui-username string    username for Swagger UI basic authentication")
		fmt.Println("      --swagger-ui-password string    password for Swagger UI basic authentication")
		fmt.Println()
		fmt.Println("BROWSER CORS SECURITY SETTINGS:")
		fmt.Println("  -o, --cors-origins strings          CORS allowed origins")
		fmt.Println("  -m, --cors-methods strings          CORS allowed methods (default " + formatDefault(defaultCORSAllowedMethods) + ")")
		fmt.Println("  -H, --cors-headers strings          CORS allowed headers (default " + formatDefault(defaultCORSAllowedHeaders) + ")")
		fmt.Println("  -x, --cors-max-age uint16           CORS max age in seconds (default " + formatDefault(defaultCORSMaxAge) + ")")
		fmt.Println()
		fmt.Println("BROWSER CSRF SECURITY SETTINGS:")
		fmt.Println("  -N, --csrf-token-name string        CSRF token name (default " + formatDefault(defaultCSRFTokenName) + ")")
		fmt.Println("  -S, --csrf-token-same-site string   CSRF token SameSite attribute (default " + formatDefault(defaultCSRFTokenSameSite) + ")")
		fmt.Println("  -M, --csrf-token-max-age duration   CSRF token max age (expiration) (default " + formatDefault(defaultCSRFTokenMaxAge) + ")")
		fmt.Println("  -R, --csrf-token-cookie-secure      CSRF token cookie Secure attribute (default " + formatDefault(defaultCSRFTokenCookieSecure) + ")")
		fmt.Println("  -J, --csrf-token-cookie-http-only   CSRF token cookie HttpOnly attribute (default " + formatDefault(defaultCSRFTokenCookieHTTPOnly) + ")")
		fmt.Println("  -E, --csrf-token-cookie-session-only CSRF token cookie SessionOnly attribute (default " + formatDefault(defaultCSRFTokenCookieSessionOnly) + ")")
		fmt.Println("  -G, --csrf-token-single-use-token   CSRF token SingleUse attribute (default " + formatDefault(defaultCSRFTokenSingleUseToken) + ")")
		fmt.Println()
		fmt.Println("TLS SECURITY SETTINGS:")
		fmt.Println("  -n, --tls-public-dns-names strings  TLS public DNS names (default " + formatDefault(defaultTLSPublicDNSNames) + ")")
		fmt.Println("  -i, --tls-public-ip-addresses strings TLS public IP addresses (default " + formatDefault(defaultTLSPublicIPAddresses) + ")")
		fmt.Println("  -j, --tls-private-dns-names strings TLS private DNS names (default " + formatDefault(defaultTLSPrivateDNSNames) + ")")
		fmt.Println("  -k, --tls-private-ip-addresses strings TLS private IP addresses (default " + formatDefault(defaultTLSPrivateIPAddresses) + ")")
		fmt.Println()
		fmt.Println("BARRIER ENCRYPTION SECURITY SETTINGS:")
		fmt.Println("  -U, --unseal-mode string            unseal mode: N, M-of-N, sysinfo (default " + formatDefault(defaultUnsealMode) + ")")
		fmt.Println("  -F, --unseal-files strings          unseal files")
		fmt.Println()
		fmt.Println("OBSERVABILITY SETTINGS:")
		fmt.Println("  -l, --log-level string              log level: ALL, TRACE, DEBUG, CONFIG, INFO, NOTICE, WARN, ERROR, FATAL, OFF (default " + formatDefault(defaultLogLevel) + ")")
		fmt.Println("  -v, --verbose                       verbose modifier for log level")
		fmt.Println("  -z, --otlp                          enable OTLP export")
		fmt.Println("  -q, --otlp-console                  enable OTLP logging to console (STDOUT)")
		fmt.Println("  -s, --otlp-service string           OTLP service (default " + formatDefault(defaultOTLPService) + ")")
		fmt.Println("  -B, --otlp-version string           OTLP version (default " + formatDefault(defaultOTLPVersion) + ")")
		fmt.Println("  -I, --otlp-instance string          OTLP instance id (default " + formatDefault(defaultOTLPInstance) + ")")
		fmt.Println("  -K, --otlp-environment string       OTLP environment (default " + formatDefault(defaultOTLPEnvironment) + ")")
		fmt.Println("  -O, --otlp-hostname string          OTLP hostname (default " + formatDefault(defaultOTLPHostname) + ")")
		fmt.Println("  -Q, --otlp-endpoint string          OTLP endpoint (default " + formatDefault(defaultOTLPEndpoint) + ")")
		fmt.Println()
		fmt.Println("ENVIRONMENT VARIABLES:")
		fmt.Println("  All flags can be set via environment variables using the CRYPTOUTIL_ prefix.")
		fmt.Println("  Examples: CRYPTOUTIL_LOG_LEVEL=DEBUG, CRYPTOUTIL_DATABASE_URL=...")
		fmt.Println()
		fmt.Println("Quickstart Examples:")
		fmt.Println("  kms cryptoutil server start --d                              Start server with in-memory SQLite")
		fmt.Println("  kms cryptoutil server stop  --d                              Stop server")
		fmt.Println("  kms cryptoutil server start --D required                     Start server with PostgreSQL container")
		fmt.Println("  kms cryptoutil server start --y global.yml --y preprod.yml   Start server with settings in YAML config files")
		fmt.Println("  kms cryptoutil server start --Y --y config.yml               Validate configuration without starting")
		fmt.Println("  kms cryptoutil server stop                                   Stop server")

		if exitIfHelp {
			os.Exit(0)
		}
	}

	// Validate configuration before returning
	if err := validateConfiguration(s); err != nil {
		return nil, err
	}

	return s, nil
}

func logSettings(s *ServiceTemplateServerSettings) {
	if s.VerboseMode {
		log.Info("Sub Command: ", s.SubCommand)

		// Create a map to get values by setting name
		valueMap := map[string]any{
			help.Name:                        s.Help,
			configFile.Name:                  s.ConfigFile,
			logLevel.Name:                    s.LogLevel,
			verboseMode.Name:                 s.VerboseMode,
			devMode.Name:                     s.DevMode,
			dryRun.Name:                      s.DryRun,
			profile.Name:                     s.Profile,
			bindPublicProtocol.Name:          s.BindPublicProtocol,
			bindPublicAddress.Name:           s.BindPublicAddress,
			bindPublicPort.Name:              s.BindPublicPort,
			tlsPublicDNSNames.Name:           s.TLSPublicDNSNames,
			tlsPublicIPAddresses.Name:        s.TLSPublicIPAddresses,
			tlsPrivateDNSNames.Name:          s.TLSPrivateDNSNames,
			tlsPrivateIPAddresses.Name:       s.TLSPrivateIPAddresses,
			bindPrivateProtocol.Name:         s.BindPrivateProtocol,
			bindPrivateAddress.Name:          s.BindPrivateAddress,
			bindPrivatePort.Name:             s.BindPrivatePort,
			publicBrowserAPIContextPath.Name: s.PublicBrowserAPIContextPath,
			publicServiceAPIContextPath.Name: s.PublicServiceAPIContextPath,
			privateAdminAPIContextPath.Name:  s.PrivateAdminAPIContextPath,
			corsAllowedOrigins.Name:          s.CORSAllowedOrigins,
			corsAllowedMethods.Name:          s.CORSAllowedMethods,
			corsAllowedHeaders.Name:          s.CORSAllowedHeaders,
			corsMaxAge.Name:                  s.CORSMaxAge,
			requestBodyLimit.Name:            s.RequestBodyLimit,
			csrfTokenName.Name:               s.CSRFTokenName,
			csrfTokenSameSite.Name:           s.CSRFTokenSameSite,
			csrfTokenMaxAge.Name:             s.CSRFTokenMaxAge,
			csrfTokenCookieSecure.Name:       s.CSRFTokenCookieSecure,
			csrfTokenCookieHTTPOnly.Name:     s.CSRFTokenCookieHTTPOnly,
			csrfTokenCookieSessionOnly.Name:  s.CSRFTokenCookieSessionOnly,
			csrfTokenSingleUseToken.Name:     s.CSRFTokenSingleUseToken,
			browserIPRateLimit.Name:          s.BrowserIPRateLimit,
			serviceIPRateLimit.Name:          s.ServiceIPRateLimit,
			allowedIps.Name:                  s.AllowedIPs,
			allowedCidrs.Name:                s.AllowedCIDRs,
			databaseContainer.Name:           s.DatabaseContainer,
			databaseURL.Name:                 s.DatabaseURL,
			databaseInitTotalTimeout.Name:    s.DatabaseInitTotalTimeout,
			databaseInitRetryWait.Name:       s.DatabaseInitRetryWait,
			otlpEnabled.Name:                 s.OTLPEnabled,
			otlpConsole.Name:                 s.OTLPConsole,
			otlpService.Name:                 s.OTLPService,
			otlpVersion.Name:                 s.OTLPVersion,
			otlpEnvironment.Name:             s.OTLPEnvironment,
			otlpHostname.Name:                s.OTLPHostname,
			otlpEndpoint.Name:                s.OTLPEndpoint,
			unsealMode.Name:                  s.UnsealMode,
			unsealFiles.Name:                 s.UnsealFiles,
			browserRealms.Name:               s.BrowserRealms,
			serviceRealms.Name:               s.ServiceRealms,
			browserSessionCookie.Name:        s.BrowserSessionCookie,
		}

		// Iterate through all registered settings and log them
		for _, setting := range allServeiceTemplateServerRegisteredSettings {
			value := valueMap[setting.Name]
			if setting.Redacted && (!s.DevMode || !s.VerboseMode) {
				value = "REDACTED"
			}

			log.Info(setting.Description+" (-"+setting.Shorthand+"): ", value)
		}

		analysis := analyzeSettings(allServeiceTemplateServerRegisteredSettings)

		var usedShorthands []string

		var unusedShorthands []string

		// Check all letters (lowercase and uppercase) and digits
		allPossibleShorthands := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		for _, r := range allPossibleShorthands {
			possibleShorthand := string(r)
			if _, ok := analysis.SettingsByShorthands[possibleShorthand]; ok {
				usedShorthands = append(usedShorthands, possibleShorthand)
			} else {
				unusedShorthands = append(unusedShorthands, possibleShorthand)
			}
		}

		sort.Strings(usedShorthands)
		sort.Strings(unusedShorthands)
		log.Info("Shorthands, Used:   ", len(usedShorthands), ", Values: ", usedShorthands)
		log.Info("Shorthands, Unused: ", len(unusedShorthands), ", Values: ", unusedShorthands)
	}
}

func resetFlags() {
	pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)

	viper.Reset()
}

// SetEnvAndRegisterSetting sets the environment variable name and registers the setting.
func SetEnvAndRegisterSetting(_ []*Setting, setting *Setting) *Setting {
	setting.Env = "CRYPTOUTIL_" + strings.ToUpper(strings.ReplaceAll(setting.Name, "-", "_"))

	allServeiceTemplateServerRegisteredSettings = append(allServeiceTemplateServerRegisteredSettings, setting)

	return setting
}

// RegisterAsBoolSetting extracts a bool value from a Setting with type assertion.
func RegisterAsBoolSetting(s *Setting) bool {
	if v, ok := s.Value.(bool); ok {
		return v
	}

	panic(fmt.Sprintf("setting %s value is not bool", s.Name))
}

// RegisterAsStringSetting extracts a string value from a Setting with type assertion.
func RegisterAsStringSetting(s *Setting) string {
	if v, ok := s.Value.(string); ok {
		return v
	}

	panic(fmt.Sprintf("setting %s value is not string", s.Name))
}

// RegisterAsUint16Setting extracts a uint16 value from a Setting with type assertion.
func RegisterAsUint16Setting(s *Setting) uint16 {
	if v, ok := s.Value.(uint16); ok {
		return v
	}

	panic(fmt.Sprintf("setting %s value is not uint16", s.Name))
}

// RegisterAsStringSliceSetting extracts a string slice value from a Setting with type assertion.
func RegisterAsStringSliceSetting(s *Setting) []string {
	if v, ok := s.Value.([]string); ok {
		return v
	}

	panic(fmt.Sprintf("setting %s value is not []string", s.Name))
}

// RegisterAsStringArraySetting extracts a string array value from a Setting with type assertion.
func RegisterAsStringArraySetting(s *Setting) []string {
	if v, ok := s.Value.([]string); ok {
		return v
	}

	panic(fmt.Sprintf("setting %s value is not []string for array", s.Name))
}

// RegisterAsDurationSetting extracts a time.Duration value from a Setting with type assertion.
func RegisterAsDurationSetting(s *Setting) time.Duration {
	if v, ok := s.Value.(time.Duration); ok {
		return v
	}

	panic(fmt.Sprintf("setting %s value is not time.Duration", s.Name))
}

// RegisterAsIntSetting extracts an int value from a Setting with type assertion.
func RegisterAsIntSetting(s *Setting) int {
	if v, ok := s.Value.(int); ok {
		return v
	}

	panic(fmt.Sprintf("setting %s value is not int", s.Name))
}

func formatDefault(value any) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", v)
	case []string:
		if len(v) == 0 {
			return "[]"
		}

		return fmt.Sprintf("[%s]", strings.Join(v, ","))
	case bool:
		return fmt.Sprintf("%t", v)
	case uint16:
		return fmt.Sprintf("%d", v)
	case time.Duration:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func analyzeSettings(settings []*Setting) analysisResult {
	result := analysisResult{
		SettingsByNames:      make(map[string][]*Setting),
		SettingsByShorthands: make(map[string][]*Setting),
	}
	for _, setting := range settings {
		result.SettingsByNames[setting.Name] = append(result.SettingsByNames[setting.Name], setting)
		// Only track non-empty shorthands
		if setting.Shorthand != "" {
			result.SettingsByShorthands[setting.Shorthand] = append(result.SettingsByShorthands[setting.Shorthand], setting)
		}
	}

	for _, setting := range settings {
		if len(result.SettingsByNames[setting.Name]) > 1 {
			result.DuplicateNames = append(result.DuplicateNames, setting.Name)
		}

		if setting.Shorthand != "" && len(result.SettingsByShorthands[setting.Shorthand]) > 1 {
			result.DuplicateShorthands = append(result.DuplicateShorthands, setting.Shorthand)
		}
	}

	return result
}

// validateConfiguration performs comprehensive validation of the configuration
// and returns detailed error messages with suggestions for fixes.
func validateConfiguration(s *ServiceTemplateServerSettings) error {
	var errors []string

	// Validate bind addresses (CRITICAL: blank address produces ":port" which binds to 0.0.0.0 triggering Windows Firewall).
	if s.BindPublicAddress == "" {
		errors = append(errors, "bind public address cannot be blank (would bind to 0.0.0.0 triggering Windows Firewall): use '127.0.0.1' for localhost or explicit IP")
	}

	if s.BindPrivateAddress == "" {
		errors = append(errors, "bind private address cannot be blank (would bind to 0.0.0.0 triggering Windows Firewall): use '127.0.0.1' for localhost or explicit IP")
	}

	// CRITICAL: In test/dev environments, reject 0.0.0.0 to prevent Windows Firewall prompts.
	// Production containers may use 0.0.0.0 for external access (isolated network namespace).
	if s.DevMode && s.BindPublicAddress == cryptoutilMagic.IPv4AnyAddress {
		errors = append(errors, "CRITICAL: bind public address cannot be 0.0.0.0 in test/dev mode (triggers Windows Firewall prompts): use '127.0.0.1' for localhost")
	}

	if s.DevMode && s.BindPrivateAddress == cryptoutilMagic.IPv4AnyAddress {
		errors = append(errors, "CRITICAL: bind private address cannot be 0.0.0.0 in test/dev mode (triggers Windows Firewall prompts): use '127.0.0.1' for localhost")
	}

	// Validate port ranges (port 0 is valid - OS assigns dynamic port).
	if s.BindPublicPort > cryptoutilMagic.MaxPortNumber {
		errors = append(errors, fmt.Sprintf("invalid public port %d: must be between 0 and 65535", s.BindPublicPort))
	}

	if s.BindPrivatePort > cryptoutilMagic.MaxPortNumber {
		errors = append(errors, fmt.Sprintf("invalid private port %d: must be between 0 and 65535", s.BindPrivatePort))
	}

	// Ports cannot be the same unless both are 0 (OS assigns different dynamic ports).
	if s.BindPublicPort == s.BindPrivatePort && s.BindPublicPort != 0 {
		errors = append(errors, fmt.Sprintf("public port (%d) and private port (%d) cannot be the same", s.BindPublicPort, s.BindPrivatePort))
	}

	// Validate protocols
	if s.BindPublicProtocol != cryptoutilMagic.ProtocolHTTP && s.BindPublicProtocol != cryptoutilMagic.ProtocolHTTPS {
		errors = append(errors, fmt.Sprintf("invalid public protocol '%s': must be '%s' or '%s'", s.BindPublicProtocol, cryptoutilMagic.ProtocolHTTP, cryptoutilMagic.ProtocolHTTPS))
	}

	if s.BindPrivateProtocol != cryptoutilMagic.ProtocolHTTP && s.BindPrivateProtocol != cryptoutilMagic.ProtocolHTTPS {
		errors = append(errors, fmt.Sprintf("invalid private protocol '%s': must be '%s' or '%s'", s.BindPrivateProtocol, cryptoutilMagic.ProtocolHTTP, cryptoutilMagic.ProtocolHTTPS))
	}

	// Validate HTTPS requirements
	if s.BindPublicProtocol == cryptoutilMagic.ProtocolHTTPS && len(s.TLSPublicDNSNames) == 0 && len(s.TLSPublicIPAddresses) == 0 {
		errors = append(errors, "HTTPS public protocol requires TLS DNS names or IP addresses to be configured")
	}

	if s.BindPrivateProtocol == "https" && len(s.TLSPrivateDNSNames) == 0 && len(s.TLSPrivateIPAddresses) == 0 {
		errors = append(errors, "HTTPS private protocol requires TLS DNS names or IP addresses to be configured")
	}

	// Validate database URL format
	// Allow special SQLite formats: ":memory:", "file::memory:?cache=shared"
	// Standard formats must contain "://" (e.g., "postgres://...", "file://...")
	if s.DatabaseURL != "" &&
		s.DatabaseURL != ":memory:" &&
		s.DatabaseURL != "file::memory:?cache=shared" &&
		!strings.Contains(s.DatabaseURL, "://") {
		errors = append(errors, fmt.Sprintf("invalid database URL format '%s': must contain '://' (e.g., 'postgres://user:pass@host:port/db') or use SQLite special formats (':memory:', 'file::memory:?cache=shared')", s.DatabaseURL))
	}

	// Validate CORS origins format
	for _, origin := range s.CORSAllowedOrigins {
		if !strings.Contains(origin, "://") {
			errors = append(errors, fmt.Sprintf("invalid CORS origin format '%s': must contain '://' (e.g., 'https://example.com')", origin))
		}
	}

	// Validate log level
	validLogLevels := []string{"ALL", "TRACE", "DEBUG", "CONFIG", "INFO", "NOTICE", "WARN", "WARNING", "ERROR", "FATAL", "OFF"}
	logLevelValid := false

	for _, level := range validLogLevels {
		if strings.EqualFold(s.LogLevel, level) {
			logLevelValid = true

			break
		}
	}

	if !logLevelValid {
		errors = append(errors, fmt.Sprintf("invalid log level '%s': must be one of %v", s.LogLevel, validLogLevels))
	}

	// Validate rate limits
	if s.BrowserIPRateLimit == 0 {
		errors = append(errors, "browser rate limit cannot be 0 (would block all browser requests)")
	} else if s.BrowserIPRateLimit > cryptoutilMagic.MaxIPRateLimit {
		errors = append(errors, fmt.Sprintf("browser rate limit %d is very high (>%d), may impact performance", s.BrowserIPRateLimit, cryptoutilMagic.MaxIPRateLimit))
	}

	if s.ServiceIPRateLimit == 0 {
		errors = append(errors, "service rate limit cannot be 0 (would block all service requests)")
	} else if s.ServiceIPRateLimit > cryptoutilMagic.MaxIPRateLimit {
		errors = append(errors, fmt.Sprintf("service rate limit %d is very high (>%d), may impact performance", s.ServiceIPRateLimit, cryptoutilMagic.MaxIPRateLimit))
	}

	// Validate OTLP endpoint format
	if s.OTLPEnabled && s.OTLPEndpoint != "" {
		if !strings.HasPrefix(s.OTLPEndpoint, "grpc://") && !strings.HasPrefix(s.OTLPEndpoint, "http://") && !strings.HasPrefix(s.OTLPEndpoint, "https://") {
			errors = append(errors, fmt.Sprintf("invalid OTLP endpoint format '%s': must start with 'grpc://', 'http://', or 'https://'", s.OTLPEndpoint))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n%s\n\nSuggestions:\n- Use --dry-run to validate configuration without starting\n- Check configuration file syntax\n- Use --profile flag for common deployment scenarios\n- See --help for detailed option descriptions", strings.Join(errors, "\n"))
	}

	return nil
}

// resolveFileURL reads the content of a file if the value starts with "file://".
// This pattern is used for Docker secrets and Kubernetes secrets mounted as files.
// Example: "file:///run/secrets/database_url" reads the secret file content.
func resolveFileURL(value string) string {
	if !strings.HasPrefix(value, "file://") {
		return value
	}

	filePath := strings.TrimPrefix(value, "file://")

	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Warnf("Failed to read file URL %s: %v (using value as-is)", value, err)

		return value
	}

	return strings.TrimSpace(string(content))
}

// NewForJOSEServer creates settings suitable for the JOSE Authority Server.
func NewForJOSEServer(bindAddr string, bindPort uint16, devMode bool) *ServiceTemplateServerSettings {
	// Build args for Parse()
	args := []string{
		"start", // Subcommand required
		"--bind-public-address", bindAddr,
		"--bind-public-port", fmt.Sprintf("%d", bindPort),
		"--otlp-service", "jose-server",
	}

	if devMode {
		args = append(args, "--dev")
	}

	settings, err := Parse(args, false)
	if err != nil {
		// Should not fail with valid default args
		panic(fmt.Sprintf("NewForJOSEServer failed to parse args: %v", err))
	}

	return settings
}

// NewForCAServer creates settings suitable for the CA Server.
func NewForCAServer(bindAddr string, bindPort uint16, devMode bool) *ServiceTemplateServerSettings {
	// Build args for Parse()
	args := []string{
		"start", // Subcommand required
		"--bind-public-address", bindAddr,
		"--bind-public-port", fmt.Sprintf("%d", bindPort),
		"--otlp-service", "ca-server",
	}

	if devMode {
		args = append(args, "--dev")
	}

	settings, err := Parse(args, false)
	if err != nil {
		// Should not fail with valid default args
		panic(fmt.Sprintf("NewForCAServer failed to parse args: %v", err))
	}

	return settings
}
