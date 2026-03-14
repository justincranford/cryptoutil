// Copyright (c) 2025 Justin Cranford

// Package config provides configuration management for cryptoutil services.
package config

import (
	"fmt"
	"os"

	googleUuid "github.com/google/uuid"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// NewTestConfig creates a ServiceTemplateServerSettings instance for testing without calling Parse().
// This bypasses pflag's global FlagSet to allow multiple config creations in tests.
//
// Use this in tests instead of NewForJOSEServer/NewForCAServer/etc to avoid
// "flag redefined" panics when creating multiple server instances.
//
// Parameters:
//   - bindAddr: public bind address (typically cryptoutilMagic.IPv4Loopback)
//   - bindPort: public bind port (use 0 for dynamic allocation)
//   - devMode: enable development mode (in-memory SQLite, relaxed security)
//
// Returns directly populated ServiceTemplateServerSettings matching NewForJOSEServer/NewForCAServer behavior.
func NewTestConfig(bindAddr string, bindPort uint16, devMode bool) *ServiceTemplateServerSettings {
	// Generate unique OTLP instance ID for test isolation.
	instanceID := googleUuid.New().String()

	// Determine database URL based on dev mode.
	dbURL := cryptoutilSharedMagic.DefaultDatabaseURL
	if devMode {
		dbURL = cryptoutilSharedMagic.SQLiteInMemoryDSN // In-memory SQLite for dev/test mode.
	}

	s := &ServiceTemplateServerSettings{
		TLSPublicMode:               TLSModeAuto,
		TLSPrivateMode:              TLSModeAuto,
		ConfigFile:                  []string{},
		LogLevel:                    cryptoutilSharedMagic.DefaultLogLevelInfo,
		VerboseMode:                 cryptoutilSharedMagic.DefaultVerboseMode,
		DevMode:                     devMode,
		DemoMode:                    cryptoutilSharedMagic.DefaultDemoMode,
		DryRun:                      cryptoutilSharedMagic.DefaultDryRun,
		Profile:                     cryptoutilSharedMagic.DefaultProfile,
		BindPublicProtocol:          cryptoutilSharedMagic.DefaultPublicProtocolCryptoutil,
		BindPublicAddress:           bindAddr,
		BindPublicPort:              bindPort,
		TLSPublicDNSNames:           defaultTLSPublicDNSNames,
		TLSPublicIPAddresses:        defaultTLSPublicIPAddresses,
		TLSPrivateDNSNames:          defaultTLSPrivateDNSNames,
		TLSPrivateIPAddresses:       defaultTLSPrivateIPAddresses,
		BindPrivateProtocol:         cryptoutilSharedMagic.DefaultPrivateProtocolCryptoutil,
		BindPrivateAddress:          bindAddr,
		BindPrivatePort:             0, // Dynamic port allocation for tests (avoids port conflicts in parallel testing)
		PublicBrowserAPIContextPath: cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath,
		PublicServiceAPIContextPath: cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath,
		PrivateAdminAPIContextPath:  cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath,
		CORSAllowedOrigins:          defaultCORSAllowedOrigins,
		CORSAllowedMethods:          defaultCORSAllowedMethods,
		CORSAllowedHeaders:          defaultCORSAllowedHeaders,
		CORSMaxAge:                  cryptoutilSharedMagic.DefaultCORSMaxAge,
		CSRFTokenName:               cryptoutilSharedMagic.DefaultCSRFTokenName,
		CSRFTokenSameSite:           cryptoutilSharedMagic.DefaultCSRFTokenSameSiteStrict,
		CSRFTokenMaxAge:             cryptoutilSharedMagic.DefaultCSRFTokenMaxAge,
		CSRFTokenCookieSecure:       cryptoutilSharedMagic.DefaultCSRFTokenCookieSecure,
		CSRFTokenCookieHTTPOnly:     cryptoutilSharedMagic.DefaultCSRFTokenCookieHTTPOnly,
		CSRFTokenCookieSessionOnly:  cryptoutilSharedMagic.DefaultCSRFTokenCookieSessionOnly,
		CSRFTokenSingleUseToken:     cryptoutilSharedMagic.DefaultCSRFTokenSingleUseToken,
		BrowserIPRateLimit:          cryptoutilSharedMagic.DefaultPublicBrowserAPIIPRateLimit,
		ServiceIPRateLimit:          cryptoutilSharedMagic.DefaultPublicServiceAPIIPRateLimit,
		AllowedIPs:                  []string{},
		AllowedCIDRs:                []string{},
		RequestBodyLimit:            cryptoutilSharedMagic.DefaultHTTPRequestBodyLimit,
		DatabaseContainer:           cryptoutilSharedMagic.DefaultDatabaseContainerDisabled,
		DatabaseURL:                 dbURL,
		DatabaseInitTotalTimeout:    cryptoutilSharedMagic.DefaultDatabaseInitTotalTimeout,
		DatabaseInitRetryWait:       cryptoutilSharedMagic.DefaultDataInitRetryWait,
		ServerShutdownTimeout:       cryptoutilSharedMagic.DefaultDataServerShutdownTimeout,
		OTLPEnabled:                 cryptoutilSharedMagic.DefaultOTLPEnabled,
		OTLPConsole:                 cryptoutilSharedMagic.DefaultOTLPConsole,
		OTLPService:                 cryptoutilSharedMagic.DefaultOTLPServiceDefault,
		OTLPVersion:                 cryptoutilSharedMagic.DefaultOTLPVersionDefault,
		OTLPEnvironment:             cryptoutilSharedMagic.DefaultOTLPEnvironmentDefault,
		OTLPHostname:                cryptoutilSharedMagic.DefaultOTLPHostnameDefault,
		OTLPEndpoint:                cryptoutilSharedMagic.DefaultOTLPEndpointDefault,
		OTLPInstance:                instanceID,
		UnsealMode:                  cryptoutilSharedMagic.DefaultUnsealModeSysInfo,
		UnsealFiles:                 []string{},
		BrowserRealms:               []string{},
		ServiceRealms:               []string{},
		BrowserSessionCookie:        cryptoutilSharedMagic.DefaultBrowserSessionCookie,
		BrowserSessionAlgorithm:     cryptoutilSharedMagic.DefaultBrowserSessionAlgorithm,
		BrowserSessionJWSAlgorithm:  cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		BrowserSessionJWEAlgorithm:  cryptoutilSharedMagic.DefaultBrowserSessionJWEAlgorithm,
		BrowserSessionExpiration:    cryptoutilSharedMagic.DefaultBrowserSessionExpiration,
		ServiceSessionAlgorithm:     cryptoutilSharedMagic.DefaultServiceSessionAlgorithm,
		ServiceSessionJWSAlgorithm:  cryptoutilSharedMagic.DefaultServiceSessionJWSAlgorithm,
		ServiceSessionJWEAlgorithm:  cryptoutilSharedMagic.DefaultServiceSessionJWEAlgorithm,
		ServiceSessionExpiration:    cryptoutilSharedMagic.DefaultServiceSessionExpiration,
		SessionIdleTimeout:          cryptoutilSharedMagic.DefaultSessionIdleTimeout,
		SessionCleanupInterval:      cryptoutilSharedMagic.DefaultSessionCleanupInterval,
	}

	// Validate configuration before returning.
	if err := validateConfiguration(s); err != nil {
		fmt.Fprintf(os.Stderr, "NewTestConfig validation error: %v\n", err)
		panic(fmt.Sprintf("NewTestConfig failed validation: %v", err))
	}

	return s
}
