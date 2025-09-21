package config

import (
	"math"
	"sync/atomic"
	"time"
)

var (
	currentBindPublicPort  atomic.Uint32
	currentBindPrivatePort atomic.Uint32
)

func init() {
	currentBindPublicPort.Store(uint32(bindPublicPort.value.(uint16)))
	currentBindPrivatePort.Store(uint32(bindPrivatePort.value.(uint16)))
}

func RequireNewForTest(applicationName string) *Settings {
	// Add bounds checking for port conversion to prevent integer overflow
	nextPublicPort := currentBindPublicPort.Add(1)
	if nextPublicPort > math.MaxUint16 {
		nextPublicPort = 10000 // Reset to safe starting value
	}
	nextPrivatePort := currentBindPrivatePort.Add(1)
	if nextPrivatePort > math.MaxUint16 {
		nextPrivatePort = 20000 // Reset to safe starting value
	}

	settings := &Settings{
		ConfigFile:                  configFile.value.(string),
		LogLevel:                    logLevel.value.(string),
		VerboseMode:                 verboseMode.value.(bool),
		DevMode:                     devMode.value.(bool),
		BindPublicProtocol:          bindPublicProtocol.value.(string),
		BindPublicAddress:           bindPublicAddress.value.(string),
		BindPublicPort:              uint16(nextPublicPort),
		TLSPublicDNSNames:           tlsPublicDnsNames.value.([]string),
		TLSPublicIPAddresses:        tlsPublicIPAddresses.value.([]string),
		TLSPrivateDNSNames:          tlsPrivateDnsNames.value.([]string),
		TLSPrivateIPAddresses:       tlsPrivateIPAddresses.value.([]string),
		BindPrivateProtocol:         bindPrivateProtocol.value.(string),
		BindPrivateAddress:          bindPrivateAddress.value.(string),
		BindPrivatePort:             uint16(nextPrivatePort),
		PublicBrowserAPIContextPath: publicBrowserAPIContextPath.value.(string),
		PublicServiceAPIContextPath: publicServiceAPIContextPath.value.(string),
		CORSAllowedOrigins:          corsAllowedOrigins.value.(string),
		CORSAllowedMethods:          corsAllowedMethods.value.(string),
		CORSAllowedHeaders:          corsAllowedHeaders.value.(string),
		CORSMaxAge:                  corsMaxAge.value.(uint16),
		CSRFTokenName:               csrfTokenName.value.(string),
		CSRFTokenSameSite:           csrfTokenSameSite.value.(string),
		CSRFTokenMaxAge:             csrfTokenMaxAge.value.(time.Duration),
		CSRFTokenCookieSecure:       csrfTokenCookieSecure.value.(bool),
		CSRFTokenCookieHTTPOnly:     csrfTokenCookieHTTPOnly.value.(bool),
		CSRFTokenCookieSessionOnly:  csrfTokenCookieSessionOnly.value.(bool),
		IPRateLimit:                 ipRateLimit.value.(uint16),
		AllowedIPs:                  allowedIps.value.([]string),
		AllowedCIDRs:                allowedCidrs.value.([]string),
		DatabaseContainer:           databaseContainer.value.(string),
		DatabaseURL:                 databaseURL.value.(string),
		DatabaseInitTotalTimeout:    databaseInitTotalTimeout.value.(time.Duration),
		DatabaseInitRetryWait:       databaseInitRetryWait.value.(time.Duration),
		OTLP:                        otlp.value.(bool),
		OTLPConsole:                 otlpConsole.value.(bool),
		OTLPScope:                   otlpScope.value.(string),
		UnsealMode:                  unsealMode.value.(string),
		UnsealFiles:                 unsealFiles.value.([]string),
	}
	// Overrides for testing
	settings.LogLevel = "ALL"
	settings.DevMode = true
	settings.IPRateLimit = 1000
	settings.OTLPScope = applicationName
	return settings
}
