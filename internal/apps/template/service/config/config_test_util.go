// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	"strings"
	"time"

	cryptoutilMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
)

// RequireNewForTest creates a new ServiceTemplateServerSettings with test defaults.
func RequireNewForTest(applicationName string) *ServiceTemplateServerSettings {
	configFileValue, ok := configFile.Value.([]string)
	if !ok {
		panic("configFile.value must be []string")
	}

	logLevelValue, ok := logLevel.Value.(string)
	if !ok {
		panic("logLevel.value must be string")
	}

	verboseModeValue, ok := verboseMode.Value.(bool)
	if !ok {
		panic("verboseMode.value must be bool")
	}

	devModeValue, ok := devMode.Value.(bool)
	if !ok {
		panic("devMode.value must be bool")
	}

	demoModeValue, ok := demoMode.Value.(bool)
	if !ok {
		panic("demoMode.value must be bool")
	}

	bindPublicProtocolValue, ok := bindPublicProtocol.Value.(string)
	if !ok {
		panic("bindPublicProtocol.value must be string")
	}

	bindPublicAddressValue, ok := bindPublicAddress.Value.(string)
	if !ok {
		panic("bindPublicAddress.value must be string")
	}

	tlsPublicDNSNamesValue, ok := tlsPublicDNSNames.Value.([]string)
	if !ok {
		panic("tlsPublicDNSNames.value must be []string")
	}

	tlsPublicIPAddressesValue, ok := tlsPublicIPAddresses.Value.([]string)
	if !ok {
		panic("tlsPublicIPAddresses.value must be []string")
	}

	tlsPrivateDNSNamesValue, ok := tlsPrivateDNSNames.Value.([]string)
	if !ok {
		panic("tlsPrivateDNSNames.value must be []string")
	}

	tlsPrivateIPAddressesValue, ok := tlsPrivateIPAddresses.Value.([]string)
	if !ok {
		panic("tlsPrivateIPAddresses.value must be []string")
	}

	bindPrivateProtocolValue, ok := bindPrivateProtocol.Value.(string)
	if !ok {
		panic("bindPrivateProtocol.value must be string")
	}

	bindPrivateAddressValue, ok := bindPrivateAddress.Value.(string)
	if !ok {
		panic("bindPrivateAddress.value must be string")
	}

	publicBrowserAPIContextPathValue, ok := publicBrowserAPIContextPath.Value.(string)
	if !ok {
		panic("publicBrowserAPIContextPath.value must be string")
	}

	publicServiceAPIContextPathValue, ok := publicServiceAPIContextPath.Value.(string)
	if !ok {
		panic("publicServiceAPIContextPath.value must be string")
	}

	corsAllowedOriginsValue, ok := corsAllowedOrigins.Value.([]string)
	if !ok {
		panic("corsAllowedOrigins.value must be []string")
	}

	corsAllowedMethodsValue, ok := corsAllowedMethods.Value.([]string)
	if !ok {
		panic("corsAllowedMethods.value must be []string")
	}

	corsAllowedHeadersValue, ok := corsAllowedHeaders.Value.([]string)
	if !ok {
		panic("corsAllowedHeaders.value must be []string")
	}

	corsMaxAgeValue, ok := corsMaxAge.Value.(uint16)
	if !ok {
		panic("corsMaxAge.value must be uint16")
	}

	csrfTokenNameValue, ok := csrfTokenName.Value.(string)
	if !ok {
		panic("csrfTokenName.value must be string")
	}

	csrfTokenSameSiteValue, ok := csrfTokenSameSite.Value.(string)
	if !ok {
		panic("csrfTokenSameSite.value must be string")
	}

	csrfTokenMaxAgeValue, ok := csrfTokenMaxAge.Value.(time.Duration)
	if !ok {
		panic("csrfTokenMaxAge.value must be time.Duration")
	}

	csrfTokenCookieSecureValue, ok := csrfTokenCookieSecure.Value.(bool)
	if !ok {
		panic("csrfTokenCookieSecure.value must be bool")
	}

	csrfTokenCookieHTTPOnlyValue, ok := csrfTokenCookieHTTPOnly.Value.(bool)
	if !ok {
		panic("csrfTokenCookieHTTPOnly.value must be bool")
	}

	csrfTokenCookieSessionOnlyValue, ok := csrfTokenCookieSessionOnly.Value.(bool)
	if !ok {
		panic("csrfTokenCookieSessionOnly.value must be bool")
	}

	csrfTokenSingleUseTokenValue, ok := csrfTokenSingleUseToken.Value.(bool)
	if !ok {
		panic("csrfTokenSingleUseToken.value must be bool")
	}

	browserIPRateLimitValue, ok := browserIPRateLimit.Value.(uint16)
	if !ok {
		panic("browserIPRateLimit.value must be uint16")
	}

	serviceIPRateLimitValue, ok := serviceIPRateLimit.Value.(uint16)
	if !ok {
		panic("serviceIPRateLimit.value must be uint16")
	}

	requestBodyLimitValue, ok := requestBodyLimit.Value.(int)
	if !ok {
		panic("requestBodyLimit.value must be int")
	}

	allowedIPsValue, ok := allowedIps.Value.([]string)
	if !ok {
		panic("allowedIps.value must be []string")
	}

	allowedCIDRsValue, ok := allowedCidrs.Value.([]string)
	if !ok {
		panic("allowedCidrs.value must be []string")
	}

	swaggerUIUsernameValue, ok := swaggerUIUsername.Value.(string)
	if !ok {
		panic("swaggerUIUsername.value must be string")
	}

	swaggerUIPasswordValue, ok := swaggerUIPassword.Value.(string)
	if !ok {
		panic("swaggerUIPassword.value must be string")
	}

	databaseContainerValue, ok := databaseContainer.Value.(string)
	if !ok {
		panic("databaseContainer.value must be string")
	}

	databaseURLValue, ok := databaseURL.Value.(string)
	if !ok {
		panic("databaseURL.value must be string")
	}

	databaseInitTotalTimeoutValue, ok := databaseInitTotalTimeout.Value.(time.Duration)
	if !ok {
		panic("databaseInitTotalTimeout.value must be time.Duration")
	}

	databaseInitRetryWaitValue, ok := databaseInitRetryWait.Value.(time.Duration)
	if !ok {
		panic("databaseInitRetryWait.value must be time.Duration")
	}

	serverShutdownTimeoutValue, ok := serverShutdownTimeout.Value.(time.Duration)
	if !ok {
		panic("serverShutdownTimeout.value must be time.Duration")
	}

	otlpValue, ok := otlpEnabled.Value.(bool)
	if !ok {
		panic("otlp.value must be bool")
	}

	otlpConsoleValue, ok := otlpConsole.Value.(bool)
	if !ok {
		panic("otlpConsole.value must be bool")
	}

	otlpServiceValue, ok := otlpService.Value.(string)
	if !ok {
		panic("otlpService.value must be string")
	}

	otlpInstanceValue, ok := otlpInstance.Value.(string)
	if !ok {
		panic("otlpInstance.value must be string")
	}

	otlpVersionValue, ok := otlpVersion.Value.(string)
	if !ok {
		panic("otlpVersion.value must be string")
	}

	otlpEnvironmentValue, ok := otlpEnvironment.Value.(string)
	if !ok {
		panic("otlpEnvironment.value must be string")
	}

	otlpHostnameValue, ok := otlpHostname.Value.(string)
	if !ok {
		panic("otlpHostname.value must be string")
	}

	otlpEndpointValue, ok := otlpEndpoint.Value.(string)
	if !ok {
		panic("otlpEndpoint.value must be string")
	}

	unsealModeValue, ok := unsealMode.Value.(string)
	if !ok {
		panic("unsealMode.value must be string")
	}

	unsealFilesValue, ok := unsealFiles.Value.([]string)
	if !ok {
		panic("unsealFiles.value must be []string")
	}

	browserSessionAlgorithmValue, ok := browserSessionAlgorithm.Value.(string)
	if !ok {
		panic("browserSessionAlgorithm.value must be string")
	}

	browserSessionJWSAlgorithmValue, ok := browserSessionJWSAlgorithm.Value.(string)
	if !ok {
		panic("browserSessionJWSAlgorithm.value must be string")
	}

	browserSessionJWEAlgorithmValue, ok := browserSessionJWEAlgorithm.Value.(string)
	if !ok {
		panic("browserSessionJWEAlgorithm.value must be string")
	}

	browserSessionExpirationValue, ok := browserSessionExpiration.Value.(time.Duration)
	if !ok {
		panic("browserSessionExpiration.value must be time.Duration")
	}

	serviceSessionAlgorithmValue, ok := serviceSessionAlgorithm.Value.(string)
	if !ok {
		panic("serviceSessionAlgorithm.value must be string")
	}

	serviceSessionJWSAlgorithmValue, ok := serviceSessionJWSAlgorithm.Value.(string)
	if !ok {
		panic("serviceSessionJWSAlgorithm.value must be string")
	}

	serviceSessionJWEAlgorithmValue, ok := serviceSessionJWEAlgorithm.Value.(string)
	if !ok {
		panic("serviceSessionJWEAlgorithm.value must be string")
	}

	serviceSessionExpirationValue, ok := serviceSessionExpiration.Value.(time.Duration)
	if !ok {
		panic("serviceSessionExpiration.value must be time.Duration")
	}

	sessionIdleTimeoutValue, ok := sessionIdleTimeout.Value.(time.Duration)
	if !ok {
		panic("sessionIdleTimeout.value must be time.Duration")
	}

	sessionCleanupIntervalValue, ok := sessionCleanupInterval.Value.(time.Duration)
	if !ok {
		panic("sessionCleanupInterval.value must be time.Duration")
	}

	settings := &ServiceTemplateServerSettings{
		TLSPublicMode:               TLSModeAuto,
		TLSPrivateMode:              TLSModeAuto,
		ConfigFile:                  configFileValue,
		LogLevel:                    logLevelValue,
		VerboseMode:                 verboseModeValue,
		DevMode:                     devModeValue,
		DemoMode:                    demoModeValue,
		ResetDemoMode:               false, // Default to false for tests
		BindPublicProtocol:          bindPublicProtocolValue,
		BindPublicAddress:           bindPublicAddressValue,
		BindPublicPort:              uint16(0), // Let OS assign port to avoid conflict during parallel testing
		TLSPublicDNSNames:           tlsPublicDNSNamesValue,
		TLSPublicIPAddresses:        tlsPublicIPAddressesValue,
		TLSPrivateDNSNames:          tlsPrivateDNSNamesValue,
		TLSPrivateIPAddresses:       tlsPrivateIPAddressesValue,
		BindPrivateProtocol:         bindPrivateProtocolValue,
		BindPrivateAddress:          bindPrivateAddressValue,
		BindPrivatePort:             uint16(0), // Let OS assign port to avoid conflict during parallel testing
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
		CSRFTokenSingleUseToken:     csrfTokenSingleUseTokenValue,
		RequestBodyLimit:            requestBodyLimitValue,
		BrowserIPRateLimit:          browserIPRateLimitValue,
		ServiceIPRateLimit:          serviceIPRateLimitValue,
		AllowedIPs:                  allowedIPsValue,
		AllowedCIDRs:                allowedCIDRsValue,
		SwaggerUIUsername:           swaggerUIUsernameValue,
		SwaggerUIPassword:           swaggerUIPasswordValue,
		DatabaseContainer:           databaseContainerValue,
		DatabaseURL:                 databaseURLValue,
		DatabaseInitTotalTimeout:    databaseInitTotalTimeoutValue,
		DatabaseInitRetryWait:       databaseInitRetryWaitValue,
		ServerShutdownTimeout:       serverShutdownTimeoutValue,
		OTLPEnabled:                 otlpValue,
		OTLPConsole:                 otlpConsoleValue,
		OTLPService:                 otlpServiceValue,
		OTLPInstance:                otlpInstanceValue,
		OTLPVersion:                 otlpVersionValue,
		OTLPEnvironment:             otlpEnvironmentValue,
		OTLPHostname:                otlpHostnameValue,
		OTLPEndpoint:                otlpEndpointValue,
		UnsealMode:                  unsealModeValue,
		UnsealFiles:                 unsealFilesValue,
		// Session Manager settings
		BrowserSessionAlgorithm:    browserSessionAlgorithmValue,
		BrowserSessionJWSAlgorithm: browserSessionJWSAlgorithmValue,
		BrowserSessionJWEAlgorithm: browserSessionJWEAlgorithmValue,
		BrowserSessionExpiration:   browserSessionExpirationValue,
		ServiceSessionAlgorithm:    serviceSessionAlgorithmValue,
		ServiceSessionJWSAlgorithm: serviceSessionJWSAlgorithmValue,
		ServiceSessionJWEAlgorithm: serviceSessionJWEAlgorithmValue,
		ServiceSessionExpiration:   serviceSessionExpirationValue,
		SessionIdleTimeout:         sessionIdleTimeoutValue,
		SessionCleanupInterval:     sessionCleanupIntervalValue,
	}
	// Overrides for testing
	settings.LogLevel = cryptoutilMagic.TestDefaultLogLevelAll
	settings.DevMode = cryptoutilMagic.TestDefaultDevMode
	settings.BrowserIPRateLimit = cryptoutilMagic.TestDefaultRateLimitBrowserIP
	settings.ServiceIPRateLimit = cryptoutilMagic.TestDefaultRateLimitServiceIP
	settings.OTLPService = applicationName
	settings.ServerShutdownTimeout = cryptoutilMagic.TestDefaultServerShutdownTimeout // Increase shutdown timeout for tests to allow cleanup of resources
	uniqueSuffix := strings.ReplaceAll(googleUuid.Must(googleUuid.NewV7()).String(), "-", "")

	if strings.Contains(settings.DatabaseURL, "/DB?") {
		// SQLite: Randomize the in-memory database name, so it is unique during concurrent testing
		settings.DatabaseURL = strings.Replace(settings.DatabaseURL, "/DB?", "/DB_"+applicationName+"_"+uniqueSuffix+"?", 1)
	} else if strings.Contains(settings.DatabaseURL, ":memory:") {
		// SQLite in-memory: Randomize the database name to avoid sharing between concurrent tests
		settings.DatabaseURL = strings.Replace(settings.DatabaseURL, ":memory:", ":memory:"+applicationName+"_"+uniqueSuffix, 1)
	} else if strings.Contains(settings.DatabaseURL, "postgres://") {
		// PostgreSQL: Use a unique schema within the shared database for concurrent testing
		schemaName := applicationName + "_" + uniqueSuffix
		if strings.Contains(settings.DatabaseURL, "?") {
			// URL already has query parameters, add search_path
			settings.DatabaseURL = strings.Replace(settings.DatabaseURL, "?", "?search_path="+schemaName+"&", 1)
		} else {
			// No query parameters, add search_path
			settings.DatabaseURL = settings.DatabaseURL + "?search_path=" + schemaName
		}
	} else {
		panic("unsupported database type in DATABASE_URL for RequireNewForTest()")
	}

	return settings
}

// NewFromFile loads *ServiceTemplateServerSettings from a YAML configuration file.
func NewFromFile(filePath string) (*ServiceTemplateServerSettings, error) {
	return Parse([]string{"--config-file", filePath}, false)
}
