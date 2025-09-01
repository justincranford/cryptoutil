package config

import (
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
	settings := &Settings{
		ConfigFile:               configFile.value.(string),
		VerboseMode:              verboseMode.value.(bool),
		LogLevel:                 logLevel.value.(string),
		DevMode:                  devMode.value.(bool),
		BindPublicAddress:        bindPublicAddress.value.(string),
		BindPublicPort:           uint16(currentBindPublicPort.Add(1)),
		BindPrivateAddress:       bindPrivateAddress.value.(string),
		BindPrivatePort:          uint16(currentBindPrivatePort.Add(1)),
		ContextPath:              contextPath.value.(string),
		CORSAllowedOrigins:       corsAllowedOrigins.value.(string),
		CORSAllowedMethods:       corsAllowedMethods.value.(string),
		CORSAllowedHeaders:       corsAllowedHeaders.value.(string),
		CORSMaxAge:               corsMaxAge.value.(uint16),
		CSRFTokenName:            csrfTokenName.value.(string),
		CSRFTokenSameSite:        csrfTokenSameSite.value.(string),
		CSRFTokenMaxAge:          csrfTokenMaxAge.value.(time.Duration),
		IPRateLimit:              ipRateLimit.value.(uint16),
		AllowedIPs:               allowedIps.value.(string),
		AllowedCIDRs:             allowedCidrs.value.(string),
		DatabaseContainer:        databaseContainer.value.(string),
		DatabaseURL:              databaseURL.value.(string),
		DatabaseInitTotalTimeout: databaseInitTotalTimeout.value.(time.Duration),
		DatabaseInitRetryWait:    databaseInitRetryWait.value.(time.Duration),
		OTLP:                     otlp.value.(bool),
		OTLPConsole:              otlpConsole.value.(bool),
		OTLPScope:                otlpScope.value.(string),
		UnsealMode:               unsealMode.value.(string),
		UnsealFiles:              unsealFiles.value.([]string),
	}
	// Overrides for testing
	settings.LogLevel = "ALL"
	settings.DevMode = true
	settings.IPRateLimit = 1000
	settings.OTLPScope = applicationName
	return settings
}
