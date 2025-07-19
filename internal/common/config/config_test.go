package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse_HappyPath_Defaults(t *testing.T) {
	resetFlags()
	os.Args = []string{"cmd"} // No flags
	s, err := Parse()
	assert.NoError(t, err)
	assert.Equal(t, configFile.value, s.ConfigFile)
	assert.Equal(t, verboseMode.value, s.VerboseMode)
	assert.Equal(t, devMode.value, s.DevMode)
	assert.Equal(t, bindAddress.value, s.BindAddress)
	assert.Equal(t, bindPort.value, s.BindPort)
	assert.Equal(t, contextPath.value, s.ContextPath)
	assert.Equal(t, databaseURL.value, s.DatabaseURL)
	assert.Equal(t, corsOrigins.value, s.CorsOrigins)
	assert.Equal(t, corsMethods.value, s.CorsMethods)
	assert.Equal(t, corsHeaders.value, s.CorsHeaders)
	assert.Equal(t, corsMaxAge.value, s.CorsMaxAge)
	assert.Equal(t, rateLimit.value, s.RateLimit)
	assert.Equal(t, allowedIps.value, s.AllowedIPs)
	assert.Equal(t, allowedCidrs.value, s.AllowedCIDRs)
}

func TestParse_HappyPath_Overrides(t *testing.T) {
	resetFlags()
	os.Args = []string{
		"cmd",
		"--config=test.yaml",
		"--verbose",
		"--dev",
		"--bind-address=192.168.1.1",
		"--bind-port=8080",
		"--context-path=/custom",
		"--database-url=postgres://user:pass@db:5432/dbname?sslmode=disable",
		"--cors-origins=https://example.com",
		"--cors-methods=GET,POST",
		"--cors-headers=X-Custom-Header",
		"--cors-max-age=1800",
		"--rate-limit=100",
		"--allowed-ips=192.168.1.100",
		"--allowed-cidrs=10.0.0.0/8",
	}

	s, err := Parse()
	assert.NoError(t, err)
	assert.Equal(t, "test.yaml", s.ConfigFile)
	assert.True(t, s.VerboseMode)
	assert.True(t, s.DevMode)
	assert.Equal(t, "192.168.1.1", s.BindAddress)
	assert.Equal(t, uint16(8080), s.BindPort)
	assert.Equal(t, "/custom", s.ContextPath)
	assert.Equal(t, "postgres://user:pass@db:5432/dbname?sslmode=disable", s.DatabaseURL)
	assert.Equal(t, "https://example.com", s.CorsOrigins)
	assert.Equal(t, "GET,POST", s.CorsMethods)
	assert.Equal(t, "X-Custom-Header", s.CorsHeaders)
	assert.Equal(t, uint16(1800), s.CorsMaxAge)
	assert.Equal(t, uint16(100), s.RateLimit)
	assert.Equal(t, "192.168.1.100", s.AllowedIPs)
	assert.Equal(t, "10.0.0.0/8", s.AllowedCIDRs)
}
