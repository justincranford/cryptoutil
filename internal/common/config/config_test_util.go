package config

import (
	"strings"
	"time"

	cryptoutilMagic "cryptoutil/internal/common/magic"

	googleUuid "github.com/google/uuid"
)

func RequireNewForTest(applicationName string) *Settings {
	configFileValue, ok := configFile.value.([]string)
	if !ok {
		panic("configFile.value must be []string")
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

	tlsPublicDNSNamesValue, ok := tlsPublicDNSNames.value.([]string)
	if !ok {
		panic("tlsPublicDNSNames.value must be []string")
	}

	tlsPublicIPAddressesValue, ok := tlsPublicIPAddresses.value.([]string)
	if !ok {
		panic("tlsPublicIPAddresses.value must be []string")
	}

	tlsPrivateDNSNamesValue, ok := tlsPrivateDNSNames.value.([]string)
	if !ok {
		panic("tlsPrivateDNSNames.value must be []string")
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

	corsAllowedOriginsValue, ok := corsAllowedOrigins.value.([]string)
	if !ok {
		panic("corsAllowedOrigins.value must be []string")
	}

	corsAllowedMethodsValue, ok := corsAllowedMethods.value.([]string)
	if !ok {
		panic("corsAllowedMethods.value must be []string")
	}

	corsAllowedHeadersValue, ok := corsAllowedHeaders.value.([]string)
	if !ok {
		panic("corsAllowedHeaders.value must be []string")
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

	csrfTokenSingleUseTokenValue, ok := csrfTokenSingleUseToken.value.(bool)
	if !ok {
		panic("csrfTokenSingleUseToken.value must be bool")
	}

	browserIPRateLimitValue, ok := browserIPRateLimit.value.(uint16)
	if !ok {
		panic("browserIPRateLimit.value must be uint16")
	}

	serviceIPRateLimitValue, ok := serviceIPRateLimit.value.(uint16)
	if !ok {
		panic("serviceIPRateLimit.value must be uint16")
	}

	requestBodyLimitValue, ok := requestBodyLimit.value.(int)
	if !ok {
		panic("requestBodyLimit.value must be int")
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

	serverShutdownTimeoutValue, ok := serverShutdownTimeout.value.(time.Duration)
	if !ok {
		panic("serverShutdownTimeout.value must be time.Duration")
	}

	otlpValue, ok := otlp.value.(bool)
	if !ok {
		panic("otlp.value must be bool")
	}

	otlpConsoleValue, ok := otlpConsole.value.(bool)
	if !ok {
		panic("otlpConsole.value must be bool")
	}

	otlpServiceValue, ok := otlpService.value.(string)
	if !ok {
		panic("otlpService.value must be string")
	}

	otlpInstanceValue, ok := otlpInstance.value.(string)
	if !ok {
		panic("otlpInstance.value must be string")
	}

	otlpVersionValue, ok := otlpVersion.value.(string)
	if !ok {
		panic("otlpVersion.value must be string")
	}

	otlpEnvironmentValue, ok := otlpEnvironment.value.(string)
	if !ok {
		panic("otlpEnvironment.value must be string")
	}

	otlpHostnameValue, ok := otlpHostname.value.(string)
	if !ok {
		panic("otlpHostname.value must be string")
	}

	otlpEndpointValue, ok := otlpEndpoint.value.(string)
	if !ok {
		panic("otlpEndpoint.value must be string")
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
		DatabaseContainer:           databaseContainerValue,
		DatabaseURL:                 databaseURLValue,
		DatabaseInitTotalTimeout:    databaseInitTotalTimeoutValue,
		DatabaseInitRetryWait:       databaseInitRetryWaitValue,
		ServerShutdownTimeout:       serverShutdownTimeoutValue,
		OTLP:                        otlpValue,
		OTLPConsole:                 otlpConsoleValue,
		OTLPService:                 otlpServiceValue,
		OTLPInstance:                otlpInstanceValue,
		OTLPVersion:                 otlpVersionValue,
		OTLPEnvironment:             otlpEnvironmentValue,
		OTLPHostname:                otlpHostnameValue,
		OTLPEndpoint:                otlpEndpointValue,
		UnsealMode:                  unsealModeValue,
		UnsealFiles:                 unsealFilesValue,
	}
	// Overrides for testing
	settings.LogLevel = "ALL"
	settings.DevMode = true
	settings.BrowserIPRateLimit = cryptoutilMagic.RateLimitBrowserIP
	settings.ServiceIPRateLimit = cryptoutilMagic.RateLimitServiceIP
	settings.OTLPService = applicationName
	settings.ServerShutdownTimeout = time.Duration(cryptoutilMagic.Timeout1MinuteSeconds) * time.Second // Increase shutdown timeout for tests to allow cleanup of resources
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
