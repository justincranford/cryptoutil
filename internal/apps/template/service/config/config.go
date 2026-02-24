// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"sync"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// viperMutex protects concurrent access to the global viper instance used in ParseWithFlagSet.
// This prevents "concurrent map writes" panics when tests run in parallel with ParseWithFlagSet().
// Note: viper uses global maps for environment variable bindings and other state, so we must serialize access.
var viperMutex sync.Mutex

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
	defaultLogLevel                    = cryptoutilSharedMagic.DefaultLogLevelInfo                // Balanced verbosity: shows important events without being overwhelming
	defaultBindPublicProtocol          = cryptoutilSharedMagic.DefaultPublicProtocolCryptoutil    // HTTPS by default for security in production environments
	defaultBindPublicAddress           = cryptoutilSharedMagic.DefaultPublicAddressCryptoutil     // IPv4 loopback prevents external access by default, requires explicit configuration for exposure
	defaultBindPublicPort              = cryptoutilSharedMagic.DefaultPublicPortCryptoutil        // Standard HTTP/HTTPS port, well-known and commonly available
	defaultBindPrivateProtocol         = cryptoutilSharedMagic.DefaultPrivateProtocolCryptoutil   // HTTPS for private API security, even in service-to-service communication
	defaultBindPrivateAddress          = cryptoutilSharedMagic.DefaultPrivateAddressCryptoutil    // IPv4 loopback for private API, only accessible from same machine
	defaultBindPrivatePort             = cryptoutilSharedMagic.DefaultPrivatePortCryptoutil       // Non-standard port to avoid conflicts with other services
	defaultPublicBrowserAPIContextPath = cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath // RESTful API versioning, separates browser from service APIs
	defaultPublicServiceAPIContextPath = cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath // RESTful API versioning, separates service from browser APIs
	defaultAdminServerAPIContextPath   = cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath  // RESTful API versioning for admin API
	defaultCORSMaxAge                  = cryptoutilSharedMagic.DefaultCORSMaxAge                  // 1 hour cache for CORS preflight requests, balances performance and freshness
	defaultCSRFTokenName               = cryptoutilSharedMagic.DefaultCSRFTokenName               // Standard CSRF token name, widely recognized by frameworks
	defaultCSRFTokenSameSite           = cryptoutilSharedMagic.DefaultCSRFTokenSameSiteStrict     // Strict SameSite prevents CSRF while maintaining usability
	defaultCSRFTokenMaxAge             = cryptoutilSharedMagic.DefaultCSRFTokenMaxAge             // 1 hour expiration balances security and user experience
	defaultCSRFTokenCookieSecure       = cryptoutilSharedMagic.DefaultCSRFTokenCookieSecure       // Secure cookies in production prevent MITM attacks
	defaultCSRFTokenCookieHTTPOnly     = cryptoutilSharedMagic.DefaultCSRFTokenCookieHTTPOnly     // False allows JavaScript access for form submissions (Swagger UI workaround)
	defaultCSRFTokenCookieSessionOnly  = cryptoutilSharedMagic.DefaultCSRFTokenCookieSessionOnly  // Session-only prevents persistent tracking while maintaining security
	defaultCSRFTokenSingleUseToken     = cryptoutilSharedMagic.DefaultCSRFTokenSingleUseToken     // Reusable tokens for better UX, can be changed for high-security needs
	defaultRequestBodyLimit            = cryptoutilSharedMagic.DefaultHTTPRequestBodyLimit        // 2MB limit prevents large payload attacks while allowing reasonable API usage
	defaultBrowserIPRateLimit          = cryptoutilSharedMagic.DefaultPublicBrowserAPIIPRateLimit // More lenient rate limit for browser APIs (user interactions)
	defaultServiceIPRateLimit          = cryptoutilSharedMagic.DefaultPublicServiceAPIIPRateLimit // More restrictive rate limit for service APIs (automated systems)
	defaultDatabaseContainer           = cryptoutilSharedMagic.DefaultDatabaseContainerDisabled   // Disabled by default to avoid unexpected container dependencies
	defaultDatabaseURL                 = cryptoutilSharedMagic.DefaultDatabaseURL                 // pragma: allowlist secret // PostgreSQL default with placeholder credentials, SSL disabled for local dev
	defaultDatabaseInitTotalTimeout    = cryptoutilSharedMagic.DefaultDatabaseInitTotalTimeout    // 5 minutes allows for container startup while preventing indefinite waits
	defaultDatabaseInitRetryWait       = cryptoutilSharedMagic.DefaultDataInitRetryWait           // 1 second retry interval balances responsiveness and resource usage
	defaultServerShutdownTimeout       = cryptoutilSharedMagic.DefaultDataServerShutdownTimeout   // 5 seconds allows graceful shutdown while preventing indefinite waits
	defaultHelp                        = cryptoutilSharedMagic.DefaultHelp
	defaultVerboseMode                 = cryptoutilSharedMagic.DefaultVerboseMode
	defaultDevMode                     = cryptoutilSharedMagic.DefaultDevMode
	defaultDemoMode                    = cryptoutilSharedMagic.DefaultDemoMode
	defaultResetDemoMode               = cryptoutilSharedMagic.DefaultResetDemoMode
	defaultDryRun                      = cryptoutilSharedMagic.DefaultDryRun
	defaultProfile                     = cryptoutilSharedMagic.DefaultProfile
	defaultOTLPEnabled                 = cryptoutilSharedMagic.DefaultOTLPEnabled
	defaultOTLPConsole                 = cryptoutilSharedMagic.DefaultOTLPConsole
	defaultOTLPService                 = cryptoutilSharedMagic.DefaultOTLPServiceDefault
	defaultOTLPVersion                 = cryptoutilSharedMagic.DefaultOTLPVersionDefault
	defaultOTLPEnvironment             = cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault
	defaultOTLPHostname                = cryptoutilSharedMagic.DefaultOTLPHostnameDefault
	defaultOTLPEndpoint                = cryptoutilSharedMagic.DefaultOTLPEndpointDefault
	defaultUnsealMode                  = cryptoutilSharedMagic.DefaultUnsealModeSysInfo
	defaultTLSPublicMode               = TLSMode(cryptoutilSharedMagic.DefaultTLSPublicMode)
	defaultTLSPrivateMode              = TLSMode(cryptoutilSharedMagic.DefaultTLSPrivateMode)
	defaultBrowserSessionCookie        = cryptoutilSharedMagic.DefaultBrowserSessionCookie
	defaultBrowserSessionAlgorithm     = cryptoutilSharedMagic.DefaultBrowserSessionAlgorithm
	defaultBrowserSessionJWSAlgorithm  = cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm
	defaultBrowserSessionJWEAlgorithm  = cryptoutilSharedMagic.DefaultBrowserSessionJWEAlgorithm
	defaultBrowserSessionExpiration    = cryptoutilSharedMagic.DefaultBrowserSessionExpiration
	defaultServiceSessionAlgorithm     = cryptoutilSharedMagic.DefaultServiceSessionAlgorithm
	defaultServiceSessionJWSAlgorithm  = cryptoutilSharedMagic.DefaultServiceSessionJWSAlgorithm
	defaultServiceSessionJWEAlgorithm  = cryptoutilSharedMagic.DefaultServiceSessionJWEAlgorithm
	defaultServiceSessionExpiration    = cryptoutilSharedMagic.DefaultServiceSessionExpiration
	defaultSessionIdleTimeout          = cryptoutilSharedMagic.DefaultSessionIdleTimeout
	defaultSessionCleanupInterval      = cryptoutilSharedMagic.DefaultSessionCleanupInterval
)

var (
	defaultTLSStaticCertPEM  = cryptoutilSharedMagic.DefaultTLSStaticCertPEM
	defaultTLSStaticKeyPEM   = cryptoutilSharedMagic.DefaultTLSStaticKeyPEM
	defaultTLSMixedCACertPEM = cryptoutilSharedMagic.DefaultTLSMixedCACertPEM
	defaultTLSMixedCAKeyPEM  = cryptoutilSharedMagic.DefaultTLSMixedCAKeyPEM
	defaultBrowserRealms     = cryptoutilSharedMagic.DefaultBrowserRealms
	defaultServiceRealms     = cryptoutilSharedMagic.DefaultServiceRealms
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

var defaultCORSAllowedOrigins = cryptoutilSharedMagic.DefaultCORSAllowedOrigins

var defaultAllowedIps = cryptoutilSharedMagic.DefaultIPFilterAllowedIPs

var defaultTLSPublicDNSNames = cryptoutilSharedMagic.DefaultTLSPublicDNSNames

var defaultTLSPublicIPAddresses = cryptoutilSharedMagic.DefaultTLSPublicIPAddresses

var defaultTLSPrivateDNSNames = cryptoutilSharedMagic.DefaultTLSPrivateDNSNames

var defaultTLSPrivateIPAddresses = cryptoutilSharedMagic.DefaultTLSPrivateIPAddresses

var defaultAllowedCIDRs = cryptoutilSharedMagic.DefaultIPFilterAllowedCIDRs

var defaultSwaggerUIUsername = ""

var defaultSwaggerUIPassword = ""

var defaultCORSAllowedMethods = cryptoutilSharedMagic.DefaultCORSAllowedMethods

var defaultCORSAllowedHeaders = cryptoutilSharedMagic.DefaultCORSAllowedHeaders

var defaultOTLPInstance = func() string {
	return googleUuid.Must(googleUuid.NewV7()).String()
}()
var defaultUnsealFiles = cryptoutilSharedMagic.DefaultUnsealFiles

var defaultConfigFiles = cryptoutilSharedMagic.DefaultConfigFiles

// set of valid subcommands.
var subcommands = map[string]struct{}{
	"start": {},
	"stop":  {},
	"init":  {},
	"live":  {},
	"ready": {},
}

var allServiceTemplateServerRegisteredSettings []*Setting

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

// ToTelemetrySettings converts the server settings to TelemetrySettings for the shared telemetry package.
func (s *ServiceTemplateServerSettings) ToTelemetrySettings() *cryptoutilSharedTelemetry.TelemetrySettings {
	return &cryptoutilSharedTelemetry.TelemetrySettings{
		LogLevel:        s.LogLevel,
		VerboseMode:     s.VerboseMode,
		OTLPEnabled:     s.OTLPEnabled,
		OTLPConsole:     s.OTLPConsole,
		OTLPService:     s.OTLPService,
		OTLPInstance:    s.OTLPInstance,
		OTLPVersion:     s.OTLPVersion,
		OTLPEnvironment: s.OTLPEnvironment,
		OTLPHostname:    s.OTLPHostname,
		OTLPEndpoint:    s.OTLPEndpoint,
	}
}
