package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParse_HappyPath_Defaults(t *testing.T) {
	resetFlags()
	commandParameters := []string{"start"}
	s, err := Parse(commandParameters, true) // true => If --help is set, help is printed and the program exits
	assert.NoError(t, err)
	assert.Equal(t, help.value, s.Help)
	assert.Equal(t, configFile.value, s.ConfigFile)
	assert.Equal(t, logLevel.value, s.LogLevel)
	assert.Equal(t, verboseMode.value, s.VerboseMode)
	assert.Equal(t, devMode.value, s.DevMode)
	assert.Equal(t, bindPublicProtocol.value, s.BindPublicProtocol)
	assert.Equal(t, bindPublicAddress.value, s.BindPublicAddress)
	assert.Equal(t, bindPublicPort.value, s.BindPublicPort)
	assert.Equal(t, bindPrivateProtocol.value, s.BindPrivateProtocol)
	assert.Equal(t, bindPrivateAddress.value, s.BindPrivateAddress)
	assert.Equal(t, bindPrivatePort.value, s.BindPrivatePort)
	assert.Equal(t, tlsPublicDnsNames.value, s.TLSPublicDNSNames)
	assert.Equal(t, tlsPublicIPAddresses.value, s.TLSPublicIPAddresses)
	assert.Equal(t, tlsPrivateDnsNames.value, s.TLSPrivateDNSNames)
	assert.Equal(t, tlsPrivateIPAddresses.value, s.TLSPrivateIPAddresses)
	assert.Equal(t, publicBrowserAPIContextPath.value, s.PublicBrowserAPIContextPath)
	assert.Equal(t, publicServiceAPIContextPath.value, s.PublicServiceAPIContextPath)
	assert.Equal(t, corsAllowedOrigins.value, s.CORSAllowedOrigins)
	assert.Equal(t, corsAllowedMethods.value, s.CORSAllowedMethods)
	assert.Equal(t, corsAllowedHeaders.value, s.CORSAllowedHeaders)
	assert.Equal(t, corsMaxAge.value, s.CORSMaxAge)
	assert.Equal(t, csrfTokenName.value, s.CSRFTokenName)
	assert.Equal(t, csrfTokenSameSite.value, s.CSRFTokenSameSite)
	assert.Equal(t, csrfTokenMaxAge.value, s.CSRFTokenMaxAge)
	assert.Equal(t, csrfTokenCookieSecure.value, s.CSRFTokenCookieSecure)
	assert.Equal(t, csrfTokenCookieHTTPOnly.value, s.CSRFTokenCookieHTTPOnly)
	assert.Equal(t, csrfTokenCookieSessionOnly.value, s.CSRFTokenCookieSessionOnly)
	assert.Equal(t, csrfTokenSingleUseToken.value, s.CSRFTokenSingleUseToken)
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
	commandParameters := []string{
		"start",
		"--help",
		"--config=test.yaml",
		"--log-level=debug",
		"--verbose",
		"--dev",
		"--bind-public-protocol=http",
		"--bind-public-address=192.168.1.2",
		"--bind-public-port=18080",
		"--bind-private-protocol=https",
		"--bind-private-address=192.168.1.3",
		"--bind-private-port=19090",
		"--tls-public-dns-names=public1.example.com,public2.example.com",
		"--tls-public-ip-addresses=192.168.1.4,192.168.1.6",
		"--tls-private-dns-names=private1.example.com,private2.example.com",
		"--tls-private-ip-addresses=192.168.1.5,192.168.1.7",
		"--browser-api-context-path=/browser",
		"--service-api-context-path=/service",
		"--cors-origins=https://example.com",
		"--cors-methods=GET,POST",
		"--cors-headers=X-Custom-Header",
		"--cors-max-age=1800",
		"--csrf-token-name=custom_csrf",
		"--csrf-token-same-site=Lax",
		"--csrf-token-max-age=24h",
		"--csrf-token-cookie-secure=false",
		"--csrf-token-cookie-http-only=false",
		"--csrf-token-cookie-session-only=false",
		"--csrf-token-single-use-token=true",
		"--rate-limit=100",
		"--allowed-ips=192.168.1.100,192.168.1.101",
		"--allowed-cidrs=10.0.0.0/8,192.168.1.0/24",
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

	s, err := Parse(commandParameters, false) // false => If --help is set, help is printed but the program doesn't exit
	assert.NoError(t, err)
	assert.True(t, s.Help)
	assert.Equal(t, "test.yaml", s.ConfigFile)
	assert.Equal(t, "debug", s.LogLevel)
	assert.True(t, s.VerboseMode)
	assert.Equal(t, "http", s.BindPublicProtocol)
	assert.Equal(t, "192.168.1.2", s.BindPublicAddress)
	assert.Equal(t, uint16(18080), s.BindPublicPort)
	assert.Equal(t, "https", s.BindPrivateProtocol)
	assert.Equal(t, "192.168.1.3", s.BindPrivateAddress)
	assert.Equal(t, uint16(19090), s.BindPrivatePort)
	assert.Equal(t, []string{"public1.example.com", "public2.example.com"}, s.TLSPublicDNSNames)
	assert.Equal(t, []string{"192.168.1.4", "192.168.1.6"}, s.TLSPublicIPAddresses)
	assert.Equal(t, []string{"private1.example.com", "private2.example.com"}, s.TLSPrivateDNSNames)
	assert.Equal(t, []string{"192.168.1.5", "192.168.1.7"}, s.TLSPrivateIPAddresses)
	assert.Equal(t, "/browser", s.PublicBrowserAPIContextPath)
	assert.Equal(t, "/service", s.PublicServiceAPIContextPath)
	assert.Equal(t, "https://example.com", s.CORSAllowedOrigins)
	assert.Equal(t, "GET,POST", s.CORSAllowedMethods)
	assert.Equal(t, "X-Custom-Header", s.CORSAllowedHeaders)
	assert.Equal(t, uint16(1800), s.CORSMaxAge)
	assert.Equal(t, "custom_csrf", s.CSRFTokenName)
	assert.Equal(t, "Lax", s.CSRFTokenSameSite)
	assert.Equal(t, 24*time.Hour, s.CSRFTokenMaxAge)
	assert.Equal(t, false, s.CSRFTokenCookieSecure)
	assert.Equal(t, false, s.CSRFTokenCookieHTTPOnly)
	assert.Equal(t, false, s.CSRFTokenCookieSessionOnly)
	assert.Equal(t, true, s.CSRFTokenSingleUseToken)
	assert.Equal(t, uint16(100), s.IPRateLimit)
	assert.Equal(t, []string{"192.168.1.100", "192.168.1.101"}, s.AllowedIPs)
	assert.Equal(t, []string{"10.0.0.0/8", "192.168.1.0/24"}, s.AllowedCIDRs)
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
