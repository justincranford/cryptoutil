// Copyright (c) 2025-2026 Justin Cranford.
//
//

package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilSharedCryptoTls "cryptoutil/internal/shared/crypto/tls"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// TLSClientPolicy defines the supported runtime client-certificate policies.
type TLSClientPolicy = cryptoutilSharedCryptoTls.TLSClientPolicy

const (
	// TLSClientPolicyNone does not request client certificates.
	TLSClientPolicyNone = cryptoutilSharedCryptoTls.TLSClientPolicyNone
	// TLSClientPolicyRequest requests a client certificate but does not require or verify it.
	TLSClientPolicyRequest = cryptoutilSharedCryptoTls.TLSClientPolicyRequest
	// TLSClientPolicyRequireAny requires a client certificate without CA verification.
	TLSClientPolicyRequireAny = cryptoutilSharedCryptoTls.TLSClientPolicyRequireAny
	// TLSClientPolicyVerifyIfGiven verifies client certificates when presented.
	TLSClientPolicyVerifyIfGiven = cryptoutilSharedCryptoTls.TLSClientPolicyVerifyIfGiven
	// TLSClientPolicyRequireAndVerify requires and verifies client certificates.
	TLSClientPolicyRequireAndVerify = cryptoutilSharedCryptoTls.TLSClientPolicyRequireAndVerify
)

// TLSProvisionMode defines the three supported TLS certificate provisioning modes.
type TLSProvisionMode string

const (
	// TLSProvisionModeStatic uses pre-generated TLS certificates (production).
	// Requires: TLS certificate chain (PEM), private key (PEM).
	// Source: Docker secrets, Kubernetes secrets, CA-signed certificates.
	TLSProvisionModeStatic TLSProvisionMode = cryptoutilSharedMagic.TLSProvisionModeStatic

	// TLSProvisionModeMixed uses static CA to sign dynamically generated server certificates (staging/QA).
	// Requires: CA certificate chain (PEM), CA private key (PEM).
	// Auto-generates: Server certificate signed by provided CA on startup.
	TLSProvisionModeMixed TLSProvisionMode = cryptoutilSharedMagic.TLSProvisionModeMixed

	// TLSProvisionModeAuto fully auto-generates CA hierarchy and server certificates (development/testing).
	// Requires: Configuration parameters only (DNS names, IP addresses, validity).
	// Auto-generates: 3-tier CA hierarchy (Root → Intermediate → Server).
	TLSProvisionModeAuto TLSProvisionMode = cryptoutilSharedMagic.TLSProvisionModeAuto
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
	defaultTLSPublicProvisionMode  = TLSProvisionMode(cryptoutilSharedMagic.DefaultTLSPublicMode)
	defaultTLSPrivateProvisionMode = TLSProvisionMode(cryptoutilSharedMagic.DefaultTLSPrivateMode)
	defaultTLSPublicClientPolicy   = TLSClientPolicy(cryptoutilSharedMagic.DefaultTLSPublicClientPolicy)
	defaultTLSPrivateClientPolicy  = TLSClientPolicy(cryptoutilSharedMagic.DefaultTLSPrivateClientPolicy)
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
		"log-level": "ERROR",
		cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault: cryptoutilSharedMagic.DefaultDevMode,
		"bind-public-protocol":                              cryptoutilSharedMagic.DefaultPublicProtocolCryptoutil,
		"bind-public-address":                               cryptoutilSharedMagic.DefaultPublicAddressCryptoutil,
		"bind-public-port":                                  cryptoutilSharedMagic.DefaultPublicPortCryptoutil,
		"bind-private-protocol":                             cryptoutilSharedMagic.DefaultPrivateProtocolCryptoutil,
		"bind-private-address":                              cryptoutilSharedMagic.DefaultPrivateAddressCryptoutil,
		"bind-private-port":                                 cryptoutilSharedMagic.DefaultPrivatePortCryptoutil,
		"database-container":                                cryptoutilSharedMagic.DefaultDatabaseContainerDisabled,
		"database-url":                                      cryptoutilSharedMagic.SQLiteInMemoryURL,
		"csrf-token-cookie-secure":                          false,
		"otlp":                                              false,
		"otlp-console":                                      false,
		"otlp-environment":                                  "test",
	},
	cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault: {
		"log-level": "DEBUG",
		cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault: true,
		"bind-public-protocol":                              cryptoutilSharedMagic.DefaultPrivateProtocolCryptoutil,
		"bind-public-address":                               cryptoutilSharedMagic.DefaultPublicAddressCryptoutil,
		"bind-public-port":                                  cryptoutilSharedMagic.DefaultPublicPortCryptoutil,
		"bind-private-protocol":                             cryptoutilSharedMagic.DefaultPrivateProtocolCryptoutil,
		"bind-private-address":                              cryptoutilSharedMagic.DefaultPrivateAddressCryptoutil,
		"bind-private-port":                                 cryptoutilSharedMagic.DefaultPrivatePortCryptoutil,
		"database-container":                                cryptoutilSharedMagic.DefaultDatabaseContainerDisabled,
		"database-url":                                      cryptoutilSharedMagic.SQLiteInMemoryURL,
		"csrf-token-cookie-secure":                          false,
		"otlp":                                              false,
		"otlp-console":                                      true,
		"otlp-environment":                                  cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault,
	},
	"stg": {
		"log-level": cryptoutilSharedMagic.DefaultLogLevelInfo,
		cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault: false,
		"bind-public-protocol":                              cryptoutilSharedMagic.DefaultPrivateProtocolCryptoutil,
		"bind-public-address":                               cryptoutilSharedMagic.IPv4AnyAddress,
		"bind-public-port":                                  cryptoutilSharedMagic.DefaultPublicPortCryptoutil,
		"bind-private-protocol":                             cryptoutilSharedMagic.DefaultPrivateProtocolCryptoutil,
		"bind-private-address":                              cryptoutilSharedMagic.DefaultPrivateAddressCryptoutil,
		"bind-private-port":                                 cryptoutilSharedMagic.DefaultPrivatePortCryptoutil,
		"database-container":                                cryptoutilSharedMagic.DefaultDatabaseContainerDisabled,
		"csrf-token-cookie-secure":                          true,
		"otlp":                                              true,
		"otlp-console":                                      false,
		"otlp-environment":                                  "stg",
	},
	"prod": {
		"log-level": "WARN",
		cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault: false,
		"bind-public-protocol":                              cryptoutilSharedMagic.DefaultPublicProtocolCryptoutil,
		"bind-public-address":                               cryptoutilSharedMagic.IPv4AnyAddress,
		"bind-public-port":                                  cryptoutilSharedMagic.DefaultPublicPortCryptoutil,
		"bind-private-protocol":                             cryptoutilSharedMagic.DefaultPrivateProtocolCryptoutil,
		"bind-private-address":                              cryptoutilSharedMagic.DefaultPrivateAddressCryptoutil,
		"bind-private-port":                                 cryptoutilSharedMagic.DefaultPrivatePortCryptoutil,
		"database-container":                                cryptoutilSharedMagic.DefaultDatabaseContainerDisabled,
		"rate-limit":                                        cryptoutilSharedMagic.DefaultPublicBrowserAPIIPRateLimit,
		"csrf-token-cookie-secure":                          true,
		"otlp":                                              true,
		"otlp-console":                                      false,
		"otlp-environment":                                  "prod",
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

var allServiceFrameworkServerRegisteredSettings []*Setting

// ServiceFrameworkServerSettings contains all configuration settings for the service template server.
type ServiceFrameworkServerSettings struct {
	SubCommand                  string
	Help                        bool
	ConfigFile                  []string
	LogLevel                    string
	VerboseMode                 bool
	DevMode                     bool
	DryRun                      bool
	Profile                     string // Configuration profile: dev, stg, prod, test
	BindPublicProtocol          string
	BindPublicAddress           string
	BindPublicPort              uint16
	BindPrivateProtocol         string
	BindPrivateAddress          string
	BindPrivatePort             uint16
	TLSPublicProvisionMode      TLSProvisionMode // Default TLSProvisionModeAuto
	TLSPublicDNSNames           []string
	TLSPublicIPAddresses        []string
	TLSPrivateDNSNames          []string
	TLSPrivateProvisionMode     TLSProvisionMode // Default TLSProvisionModeAuto
	TLSPrivateIPAddresses       []string
	TLSStaticCertPEM            []byte // Default nil. PEM-encoded certificate chain (for TLSProvisionModeStatic). Should contain: [Server Cert, Intermediate CA(s), Root CA] or [Server Cert, Root CA].
	TLSStaticKeyPEM             []byte // Default nil. PEM-encoded private key (for TLSProvisionModeStatic).
	TLSMixedCACertPEM           []byte // Default nil. PEM-encoded CA certificate chain (for TLSProvisionModeMixed). Should contain: [Intermediate CA(s), Root CA] or [Root CA].
	TLSMixedCAKeyPEM            []byte // Default nil. PEM-encoded CA private key (for TLSProvisionModeMixed).
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
	BrowserRealms               []string        // Paths to browser realm configuration files (session-based auth)
	ServiceRealms               []string        // Paths to service realm configuration files (token-based auth)
	BrowserSessionCookie        string          // Cookie type: jwe (encrypted), jws (signed), opaque (database) - DEPRECATED: use BrowserSessionAlgorithm
	BrowserSessionAlgorithm     string          // Session algorithm: OPAQUE (hashed), JWS (signed JWT), JWE (encrypted JWT)
	BrowserSessionJWSAlgorithm  string          // JWS algorithm for browser sessions (e.g., RS256, ES256, EdDSA)
	BrowserSessionJWEAlgorithm  string          // JWE algorithm for browser sessions (e.g., dir+A256GCM, A256GCMKW+A256GCM)
	BrowserSessionExpiration    time.Duration   // Browser session expiration duration
	ServiceSessionAlgorithm     string          // Session algorithm: OPAQUE (hashed), JWS (signed JWT), JWE (encrypted JWT)
	ServiceSessionJWSAlgorithm  string          // JWS algorithm for service sessions (e.g., RS256, ES256, EdDSA)
	ServiceSessionJWEAlgorithm  string          // JWE algorithm for service sessions (e.g., dir+A256GCM, A256GCMKW+A256GCM)
	ServiceSessionExpiration    time.Duration   // Service session expiration duration
	SessionIdleTimeout          time.Duration   // Session idle timeout duration
	SessionCleanupInterval      time.Duration   // Interval for cleaning up expired sessions
	DatabaseSSLMode             string          // PostgreSQL SSL mode: disable, require, verify-ca, verify-full (empty = use DSN default)
	DatabaseSSLCert             string          // Path to client TLS certificate file for PostgreSQL mTLS (Cat 14)
	DatabaseSSLKey              string          // Path to client TLS private key file for PostgreSQL mTLS (Cat 14)
	DatabaseSSLRootCert         string          // Path to CA truststore for verifying PostgreSQL server cert (Cat 10)
	AdminTLSCertFile            string          // Path to admin TLS server certificate file for private admin mTLS (Cat 7)
	AdminTLSKeyFile             string          // Path to admin TLS server private key file for private admin mTLS (Cat 7)
	AdminTLSCAFile              string          // Path to CA truststore for verifying admin client certs (Cat 6)
	AdminTLSClientPolicy        TLSClientPolicy // Default TLSClientPolicyNone. Selects the admin listener client-certificate policy.
	PublicTLSCertFile           string          // Path to public TLS server certificate file (Cat 3); absent = auto-TLS
	PublicTLSKeyFile            string          // Path to public TLS server private key file (Cat 3); absent = auto-TLS
	PublicTLSCAFile             string          // Path to CA truststore for verifying public client certs (Cat 4)
	PublicTLSClientPolicy       TLSClientPolicy // Default TLSClientPolicyNone. Selects the public listener client-certificate policy.
	OTLPTLSCertFile             string          // Path to client TLS certificate file for OTLP mTLS (Cat 9)
	OTLPTLSKeyFile              string          // Path to client TLS private key file for OTLP mTLS (Cat 9)
	OTLPTLSCAFile               string          // Path to CA truststore for verifying the OTLP server cert (Cat 1)
}

// PrivateBaseURL returns the private base URL constructed from protocol, address, and port.
func (s *ServiceFrameworkServerSettings) PrivateBaseURL() string {
	return fmt.Sprintf("%s://%s:%d", s.BindPrivateProtocol, s.BindPrivateAddress, s.BindPrivatePort)
}

// PublicBaseURL returns the public base URL constructed from protocol, address, and port.
func (s *ServiceFrameworkServerSettings) PublicBaseURL() string {
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
func (s *ServiceFrameworkServerSettings) ToTelemetrySettings() *cryptoutilSharedTelemetry.TelemetrySettings {
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
		OTLPTLSCertFile: s.OTLPTLSCertFile,
		OTLPTLSKeyFile:  s.OTLPTLSKeyFile,
		OTLPTLSCAFile:   s.OTLPTLSCAFile,
	}
}
