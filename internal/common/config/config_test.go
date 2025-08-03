package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParse_HappyPath_Defaults(t *testing.T) {
	resetFlags()
	os.Args = []string{"cmd"} // No flags
	s, err := Parse(true)     // true => If --help is set, help is printed and the program exits
	assert.NoError(t, err)
	assert.Equal(t, help.value, s.Help)
	assert.Equal(t, configFile.value, s.ConfigFile)
	assert.Equal(t, logLevel.value, s.LogLevel)
	assert.Equal(t, verboseMode.value, s.VerboseMode)
	assert.Equal(t, devMode.value, s.DevMode)
	assert.Equal(t, bindAddress.value, s.BindAddress)
	assert.Equal(t, bindPort.value, s.BindPort)
	assert.Equal(t, contextPath.value, s.ContextPath)
	assert.Equal(t, corsAllowedOrigins.value, s.CORSAllowedOrigins)
	assert.Equal(t, corsAllowedMethods.value, s.CORSAllowedMethods)
	assert.Equal(t, corsAllowedHeaders.value, s.CORSAllowedHeaders)
	assert.Equal(t, corsMaxAge.value, s.CORSMaxAge)
	assert.Equal(t, csrfTokenName.value, s.CSRFTokenName)
	assert.Equal(t, csrfTokenSameSite.value, s.CSRFTokenSameSite)
	assert.Equal(t, csrfTokenMaxAge.value, s.CSRFTokenMaxAge)
	assert.Equal(t, ipRateLimit.value, s.IPRateLimit)
	assert.Equal(t, allowedIps.value, s.AllowedIPs)
	assert.Equal(t, allowedCidrs.value, s.AllowedCIDRs)
	assert.Equal(t, databaseContainer.value, s.DatabaseContainer)
	assert.Equal(t, databaseURL.value, s.DatabaseURL)
	assert.Equal(t, databaseInitTotalTimeout.value, s.DatabaseInitTotalTimeout)
	assert.Equal(t, databaseInitRetryWait.value, s.DatabaseInitRetryWait)
	assert.Equal(t, otlp.value, s.OTLP)
	assert.Equal(t, otlpConsole.value, s.OTLPConsole)
	assert.Equal(t, otlpScope.value, s.OTLPScope)
	assert.Equal(t, unsealMode.value, s.UnsealMode)
	assert.Equal(t, unsealFiles.value.([]string), s.UnsealFiles)
}

func TestParse_HappyPath_Overrides(t *testing.T) {
	resetFlags()
	os.Args = []string{
		"cmd",
		"--help",
		"--config=test.yaml",
		"--log-level=debug",
		"--verbose",
		"--dev",
		"--bind-address=192.168.1.1",
		"--bind-port=8080",
		"--context-path=/custom",
		"--cors-origins=https://example.com",
		"--cors-methods=GET,POST",
		"--cors-headers=X-Custom-Header",
		"--cors-max-age=1800",
		"--csrf-token-name=custom_csrf",
		"--csrf-token-same-site=Lax",
		"--csrf-token-max-age=24h",
		"--rate-limit=100",
		"--allowed-ips=192.168.1.100",
		"--allowed-cidrs=10.0.0.0/8",
		"--database-container=required",
		"--database-url=postgres://user:pass@db:5432/dbname?sslmode=disable",
		"--database-init-total-timeout=5m",
		"--database-init-retry-wait=30s",
		"--otlp",
		"--otlp-console",
		"--otlp-scope=my-scope",
		"--unseal-mode=2-of-3",
		"--unseal-files=/docker/secrets/unseal1",
		"--unseal-files=/docker/secrets/unseal2",
		"--unseal-files=/docker/secrets/unseal3",
	}

	s, err := Parse(false) // false => If --help is set, help is printed but the program doesn't exit
	assert.NoError(t, err)
	assert.True(t, s.Help)
	assert.Equal(t, "test.yaml", s.ConfigFile)
	assert.Equal(t, "debug", s.LogLevel)
	assert.True(t, s.VerboseMode)
	assert.Equal(t, "192.168.1.1", s.BindAddress)
	assert.Equal(t, uint16(8080), s.BindPort)
	assert.Equal(t, "/custom", s.ContextPath)
	assert.Equal(t, "https://example.com", s.CORSAllowedOrigins)
	assert.Equal(t, "GET,POST", s.CORSAllowedMethods)
	assert.Equal(t, "X-Custom-Header", s.CORSAllowedHeaders)
	assert.Equal(t, uint16(1800), s.CORSMaxAge)
	assert.Equal(t, "custom_csrf", s.CSRFTokenName)
	assert.Equal(t, "Lax", s.CSRFTokenSameSite)
	assert.Equal(t, 24*time.Hour, s.CSRFTokenMaxAge)
	assert.Equal(t, uint16(100), s.IPRateLimit)
	assert.Equal(t, "192.168.1.100", s.AllowedIPs)
	assert.Equal(t, "10.0.0.0/8", s.AllowedCIDRs)
	assert.Equal(t, "required", s.DatabaseContainer)
	assert.Equal(t, "postgres://user:pass@db:5432/dbname?sslmode=disable", s.DatabaseURL)
	assert.Equal(t, 5*time.Minute, s.DatabaseInitTotalTimeout)
	assert.Equal(t, 30*time.Second, s.DatabaseInitRetryWait)
	assert.True(t, s.OTLP)
	assert.True(t, s.OTLPConsole)
	assert.Equal(t, "my-scope", s.OTLPScope)
	assert.True(t, s.DevMode)
	assert.Equal(t, "2-of-3", s.UnsealMode)
	assert.Equal(t, []string{"/docker/secrets/unseal1", "/docker/secrets/unseal2", "/docker/secrets/unseal3"}, s.UnsealFiles)
}
