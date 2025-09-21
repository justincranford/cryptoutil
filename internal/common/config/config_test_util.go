package config

import (
	"fmt"
	"math"
	"sync/atomic"
	"time"
)

var (
	currentBindPublicPort  atomic.Uint32
	currentBindPrivatePort atomic.Uint32
)

func init() {
	publicPort, ok := bindPublicPort.value.(uint16)
	if !ok {
		panic("bindPublicPort.value must be uint16")
	}
	privatePort, ok := bindPrivatePort.value.(uint16)
	if !ok {
		panic("bindPrivatePort.value must be uint16")
	}
	currentBindPublicPort.Store(uint32(publicPort))
	currentBindPrivatePort.Store(uint32(privatePort))
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

	// Validate port ranges before conversion
	if nextPublicPort > math.MaxUint16 {
		panic(fmt.Sprintf("public port %d exceeds uint16 maximum %d", nextPublicPort, math.MaxUint16))
	}
	if nextPrivatePort > math.MaxUint16 {
		panic(fmt.Sprintf("private port %d exceeds uint16 maximum %d", nextPrivatePort, math.MaxUint16))
	}

	configFileValue, ok := configFile.value.(string)
	if !ok {
		panic("configFile.value must be string")
	}
	logLevelValue, ok := logLevel.value.(string)
	if !ok {
		panic("logLevel.value must be string")
	}
	verboseModeValue, ok := verboseMode.value.(bool)
	if !ok {
		panic("verboseMode.value must be bool")
	}
	devModeValue, ok := devMode.value.(bool)
	if !ok {
		panic("devMode.value must be bool")
	}
	bindPublicProtocolValue, ok := bindPublicProtocol.value.(string)
	if !ok {
		panic("bindPublicProtocol.value must be string")
	}
	bindPublicAddressValue, ok := bindPublicAddress.value.(string)
	if !ok {
		panic("bindPublicAddress.value must be string")
	}
	tlsPublicDnsNamesValue, ok := tlsPublicDnsNames.value.([]string)
	if !ok {
		panic("tlsPublicDnsNames.value must be []string")
	}
	tlsPublicIPAddressesValue, ok := tlsPublicIPAddresses.value.([]string)
	if !ok {
		panic("tlsPublicIPAddresses.value must be []string")
	}
	tlsPrivateDnsNamesValue, ok := tlsPrivateDnsNames.value.([]string)
	if !ok {
		panic("tlsPrivateDnsNames.value must be []string")
	}
	tlsPrivateIPAddressesValue, ok := tlsPrivateIPAddresses.value.([]string)
	if !ok {
		panic("tlsPrivateIPAddresses.value must be []string")
	}
	bindPrivateProtocolValue, ok := bindPrivateProtocol.value.(string)
	if !ok {
		panic("bindPrivateProtocol.value must be string")
	}
	bindPrivateAddressValue, ok := bindPrivateAddress.value.(string)
	if !ok {
		panic("bindPrivateAddress.value must be string")
	}

	publicBrowserAPIContextPathValue, ok := publicBrowserAPIContextPath.value.(string)
	if !ok {
		panic("publicBrowserAPIContextPath.value must be string")
	}
	publicServiceAPIContextPathValue, ok := publicServiceAPIContextPath.value.(string)
	if !ok {
		panic("publicServiceAPIContextPath.value must be string")
	}
	corsAllowedOriginsValue, ok := corsAllowedOrigins.value.(string)
	if !ok {
		panic("corsAllowedOrigins.value must be string")
	}
	corsAllowedMethodsValue, ok := corsAllowedMethods.value.(string)
	if !ok {
		panic("corsAllowedMethods.value must be string")
	}
	corsAllowedHeadersValue, ok := corsAllowedHeaders.value.(string)
	if !ok {
		panic("corsAllowedHeaders.value must be string")
	}
	corsMaxAgeValue, ok := corsMaxAge.value.(uint16)
	if !ok {
		panic("corsMaxAge.value must be uint16")
	}
	csrfTokenNameValue, ok := csrfTokenName.value.(string)
	if !ok {
		panic("csrfTokenName.value must be string")
	}
	csrfTokenSameSiteValue, ok := csrfTokenSameSite.value.(string)
	if !ok {
		panic("csrfTokenSameSite.value must be string")
	}
	csrfTokenMaxAgeValue, ok := csrfTokenMaxAge.value.(time.Duration)
	if !ok {
		panic("csrfTokenMaxAge.value must be time.Duration")
	}
	csrfTokenCookieSecureValue, ok := csrfTokenCookieSecure.value.(bool)
	if !ok {
		panic("csrfTokenCookieSecure.value must be bool")
	}
	csrfTokenCookieHTTPOnlyValue, ok := csrfTokenCookieHTTPOnly.value.(bool)
	if !ok {
		panic("csrfTokenCookieHTTPOnly.value must be bool")
	}
	csrfTokenCookieSessionOnlyValue, ok := csrfTokenCookieSessionOnly.value.(bool)
	if !ok {
		panic("csrfTokenCookieSessionOnly.value must be bool")
	}
	ipRateLimitValue, ok := ipRateLimit.value.(uint16)
	if !ok {
		panic("ipRateLimit.value must be uint16")
	}
	allowedIPsValue, ok := allowedIps.value.([]string)
	if !ok {
		panic("allowedIps.value must be []string")
	}
	allowedCIDRsValue, ok := allowedCidrs.value.([]string)
	if !ok {
		panic("allowedCidrs.value must be []string")
	}

	databaseContainerValue, ok := databaseContainer.value.(string)
	if !ok {
		panic("databaseContainer.value must be string")
	}
	databaseURLValue, ok := databaseURL.value.(string)
	if !ok {
		panic("databaseURL.value must be string")
	}
	databaseInitTotalTimeoutValue, ok := databaseInitTotalTimeout.value.(time.Duration)
	if !ok {
		panic("databaseInitTotalTimeout.value must be time.Duration")
	}
	databaseInitRetryWaitValue, ok := databaseInitRetryWait.value.(time.Duration)
	if !ok {
		panic("databaseInitRetryWait.value must be time.Duration")
	}
	otlpValue, ok := otlp.value.(bool)
	if !ok {
		panic("otlp.value must be bool")
	}
	otlpConsoleValue, ok := otlpConsole.value.(bool)
	if !ok {
		panic("otlpConsole.value must be bool")
	}
	otlpScopeValue, ok := otlpScope.value.(string)
	if !ok {
		panic("otlpScope.value must be string")
	}
	unsealModeValue, ok := unsealMode.value.(string)
	if !ok {
		panic("unsealMode.value must be string")
	}
	unsealFilesValue, ok := unsealFiles.value.([]string)
	if !ok {
		panic("unsealFiles.value must be []string")
	}

	settings := &Settings{
		ConfigFile:                  configFileValue,
		LogLevel:                    logLevelValue,
		VerboseMode:                 verboseModeValue,
		DevMode:                     devModeValue,
		BindPublicProtocol:          bindPublicProtocolValue,
		BindPublicAddress:           bindPublicAddressValue,
		BindPublicPort:              uint16(nextPublicPort),
		TLSPublicDNSNames:           tlsPublicDnsNamesValue,
		TLSPublicIPAddresses:        tlsPublicIPAddressesValue,
		TLSPrivateDNSNames:          tlsPrivateDnsNamesValue,
		TLSPrivateIPAddresses:       tlsPrivateIPAddressesValue,
		BindPrivateProtocol:         bindPrivateProtocolValue,
		BindPrivateAddress:          bindPrivateAddressValue,
		BindPrivatePort:             uint16(nextPrivatePort),
		PublicBrowserAPIContextPath: publicBrowserAPIContextPathValue,
		PublicServiceAPIContextPath: publicServiceAPIContextPathValue,
		CORSAllowedOrigins:          corsAllowedOriginsValue,
		CORSAllowedMethods:          corsAllowedMethodsValue,
		CORSAllowedHeaders:          corsAllowedHeadersValue,
		CORSMaxAge:                  corsMaxAgeValue,
		CSRFTokenName:               csrfTokenNameValue,
		CSRFTokenSameSite:           csrfTokenSameSiteValue,
		CSRFTokenMaxAge:             csrfTokenMaxAgeValue,
		CSRFTokenCookieSecure:       csrfTokenCookieSecureValue,
		CSRFTokenCookieHTTPOnly:     csrfTokenCookieHTTPOnlyValue,
		CSRFTokenCookieSessionOnly:  csrfTokenCookieSessionOnlyValue,
		IPRateLimit:                 ipRateLimitValue,
		AllowedIPs:                  allowedIPsValue,
		AllowedCIDRs:                allowedCIDRsValue,
		DatabaseContainer:           databaseContainerValue,
		DatabaseURL:                 databaseURLValue,
		DatabaseInitTotalTimeout:    databaseInitTotalTimeoutValue,
		DatabaseInitRetryWait:       databaseInitRetryWaitValue,
		OTLP:                        otlpValue,
		OTLPConsole:                 otlpConsoleValue,
		OTLPScope:                   otlpScopeValue,
		UnsealMode:                  unsealModeValue,
		UnsealFiles:                 unsealFilesValue,
	}
	// Overrides for testing
	settings.LogLevel = "ALL"
	settings.DevMode = true
	settings.IPRateLimit = 1000
	settings.OTLPScope = applicationName
	return settings
}
