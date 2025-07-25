package config

import (
	"sync/atomic"
	"time"
)

var (
	currentBindPort atomic.Uint32
)

func init() {
	currentBindPort.Store(uint32(bindPort.value.(uint16)))
}

func RequireNewForTest(applicationName string) *Settings {
	settings := &Settings{
		ConfigFile:               configFile.value.(string),
		VerboseMode:              verboseMode.value.(bool),
		LogLevel:                 logLevel.value.(string),
		DevMode:                  devMode.value.(bool),
		OTLP:                     otlp.value.(bool),
		OTLPConsole:              otlpConsole.value.(bool),
		OTLPScope:                otlpScope.value.(string),
		BindAddress:              bindAddress.value.(string),
		BindPort:                 uint16(currentBindPort.Add(1)),
		ContextPath:              contextPath.value.(string),
		CorsOrigins:              corsOrigins.value.(string),
		CorsMethods:              corsMethods.value.(string),
		CorsHeaders:              corsHeaders.value.(string),
		CorsMaxAge:               corsMaxAge.value.(uint16),
		RateLimit:                rateLimit.value.(uint16),
		AllowedIPs:               allowedIps.value.(string),
		AllowedCIDRs:             allowedCidrs.value.(string),
		DatabaseContainer:        databaseContainer.value.(string),
		DatabaseURL:              databaseURL.value.(string),
		DatabaseInitTotalTimeout: databaseInitTotalTimeout.value.(time.Duration),
		DatabaseInitRetryWait:    databaseInitRetryWait.value.(time.Duration),
		Migrations:               migrations.value.(bool),
	}
	// Overrides for testing
	settings.LogLevel = "ALL"
	settings.DevMode = true
	settings.Migrations = true
	settings.OTLPScope = applicationName
	return settings
}
