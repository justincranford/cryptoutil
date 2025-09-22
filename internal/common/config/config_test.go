package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParse_HappyPath_Defaults(t *testing.T) {
	resetFlags()
	commandParameters := []string{"start"}
	s, err := Parse(commandParameters, true) // true => If --help is set, help is printed and the program exits
	require.NoError(t, err)
	require.Equal(t, help.value, s.Help)
	require.Equal(t, configFile.value, s.ConfigFile)
	require.Equal(t, logLevel.value, s.LogLevel)
	require.Equal(t, verboseMode.value, s.VerboseMode)
	require.Equal(t, devMode.value, s.DevMode)
	require.Equal(t, bindPublicProtocol.value, s.BindPublicProtocol)
	require.Equal(t, bindPublicAddress.value, s.BindPublicAddress)
	require.Equal(t, bindPublicPort.value, s.BindPublicPort)
	require.Equal(t, bindPrivateProtocol.value, s.BindPrivateProtocol)
	require.Equal(t, bindPrivateAddress.value, s.BindPrivateAddress)
	require.Equal(t, bindPrivatePort.value, s.BindPrivatePort)
	require.Equal(t, tlsPublicDNSNames.value, s.TLSPublicDNSNames)
	require.Equal(t, tlsPublicIPAddresses.value, s.TLSPublicIPAddresses)
	require.Equal(t, tlsPrivateDNSNames.value, s.TLSPrivateDNSNames)
	require.Equal(t, tlsPrivateIPAddresses.value, s.TLSPrivateIPAddresses)
	require.Equal(t, publicBrowserAPIContextPath.value, s.PublicBrowserAPIContextPath)
	require.Equal(t, publicServiceAPIContextPath.value, s.PublicServiceAPIContextPath)
	require.Equal(t, corsAllowedOrigins.value, s.CORSAllowedOrigins)
	require.Equal(t, corsAllowedMethods.value, s.CORSAllowedMethods)
	require.Equal(t, corsAllowedHeaders.value, s.CORSAllowedHeaders)
	require.Equal(t, corsMaxAge.value, s.CORSMaxAge)
	require.Equal(t, csrfTokenName.value, s.CSRFTokenName)
	require.Equal(t, csrfTokenSameSite.value, s.CSRFTokenSameSite)
	require.Equal(t, csrfTokenMaxAge.value, s.CSRFTokenMaxAge)
	require.Equal(t, csrfTokenCookieSecure.value, s.CSRFTokenCookieSecure)
	require.Equal(t, csrfTokenCookieHTTPOnly.value, s.CSRFTokenCookieHTTPOnly)
	require.Equal(t, csrfTokenCookieSessionOnly.value, s.CSRFTokenCookieSessionOnly)
	require.Equal(t, csrfTokenSingleUseToken.value, s.CSRFTokenSingleUseToken)
	require.Equal(t, ipRateLimit.value, s.IPRateLimit)
	require.Equal(t, allowedIps.value, s.AllowedIPs)
	require.Equal(t, allowedCidrs.value, s.AllowedCIDRs)
	require.Equal(t, databaseContainer.value, s.DatabaseContainer)
	require.Equal(t, databaseURL.value, s.DatabaseURL)
	require.Equal(t, databaseInitTotalTimeout.value, s.DatabaseInitTotalTimeout)
	require.Equal(t, databaseInitRetryWait.value, s.DatabaseInitRetryWait)
	require.Equal(t, otlp.value, s.OTLP)
	require.Equal(t, otlpConsole.value, s.OTLPConsole)
	require.Equal(t, otlpScope.value, s.OTLPScope)
	require.Equal(t, unsealMode.value, s.UnsealMode)
	unsealFilesSlice, ok := unsealFiles.value.([]string)
	require.True(t, ok, "unsealFiles.value should be []string")
	require.Equal(t, unsealFilesSlice, s.UnsealFiles)
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
	require.NoError(t, err)
	require.True(t, s.Help)
	require.Equal(t, "test.yaml", s.ConfigFile)
	require.Equal(t, "debug", s.LogLevel)
	require.True(t, s.VerboseMode)
	require.Equal(t, "http", s.BindPublicProtocol)
	require.Equal(t, "192.168.1.2", s.BindPublicAddress)
	require.Equal(t, uint16(18080), s.BindPublicPort)
	require.Equal(t, "https", s.BindPrivateProtocol)
	require.Equal(t, "192.168.1.3", s.BindPrivateAddress)
	require.Equal(t, uint16(19090), s.BindPrivatePort)
	require.Equal(t, []string{"public1.example.com", "public2.example.com"}, s.TLSPublicDNSNames)
	require.Equal(t, []string{"192.168.1.4", "192.168.1.6"}, s.TLSPublicIPAddresses)
	require.Equal(t, []string{"private1.example.com", "private2.example.com"}, s.TLSPrivateDNSNames)
	require.Equal(t, []string{"192.168.1.5", "192.168.1.7"}, s.TLSPrivateIPAddresses)
	require.Equal(t, "/browser", s.PublicBrowserAPIContextPath)
	require.Equal(t, "/service", s.PublicServiceAPIContextPath)
	require.Equal(t, "https://example.com", s.CORSAllowedOrigins)
	require.Equal(t, "GET,POST", s.CORSAllowedMethods)
	require.Equal(t, "X-Custom-Header", s.CORSAllowedHeaders)
	require.Equal(t, uint16(1800), s.CORSMaxAge)
	require.Equal(t, "custom_csrf", s.CSRFTokenName)
	require.Equal(t, "Lax", s.CSRFTokenSameSite)
	require.Equal(t, 24*time.Hour, s.CSRFTokenMaxAge)
	require.Equal(t, false, s.CSRFTokenCookieSecure)
	require.Equal(t, false, s.CSRFTokenCookieHTTPOnly)
	require.Equal(t, false, s.CSRFTokenCookieSessionOnly)
	require.Equal(t, true, s.CSRFTokenSingleUseToken)
	require.Equal(t, uint16(100), s.IPRateLimit)
	require.Equal(t, []string{"192.168.1.100", "192.168.1.101"}, s.AllowedIPs)
	require.Equal(t, []string{"10.0.0.0/8", "192.168.1.0/24"}, s.AllowedCIDRs)
	require.Equal(t, "required", s.DatabaseContainer)
	require.Equal(t, "postgres://user:pass@db:5432/dbname?sslmode=disable", s.DatabaseURL)
	require.Equal(t, 5*time.Minute, s.DatabaseInitTotalTimeout)
	require.Equal(t, 30*time.Second, s.DatabaseInitRetryWait)
	require.True(t, s.OTLP)
	require.True(t, s.OTLPConsole)
	require.Equal(t, "my-scope", s.OTLPScope)
	require.True(t, s.DevMode)
	require.Equal(t, "2-of-3", s.UnsealMode)
	require.Equal(t, []string{"/docker/secrets/unseal1", "/docker/secrets/unseal2", "/docker/secrets/unseal3"}, s.UnsealFiles)
}

func TestAnalyzeSettings_RealSettings(t *testing.T) {
	result := analyzeSettings(allRegisteredSettings)

	totalMappedByName := 0
	for _, settings := range result.SettingsByNames {
		totalMappedByName += len(settings)
	}
	require.Equal(t, len(allRegisteredSettings), totalMappedByName, "All settings should be accounted for by name")

	totalMappedByShorthand := 0
	for _, settings := range result.SettingsByShorthands {
		totalMappedByShorthand += len(settings)
	}
	require.Equal(t, len(allRegisteredSettings), totalMappedByShorthand, "All settings should be accounted for by shorthand")

	require.Empty(t, result.DuplicateNames, "Production settings should have no duplicate names")
	require.Empty(t, result.DuplicateShorthands, "Production settings should have no duplicate shorthands")

	for name, settings := range result.SettingsByNames {
		require.Len(t, settings, 1, "Setting name '%s' should be unique", name)
	}

	for shorthand, settings := range result.SettingsByShorthands {
		require.Len(t, settings, 1, "Setting shorthand '%s' should be unique", shorthand)
	}
}

func TestAnalyzeSettings_NoDuplicates(t *testing.T) {
	result := analyzeSettings([]*Setting{
		{name: "unique1", shorthand: "a", value: "value1", usage: "usage1"},
		{name: "unique2", shorthand: "b", value: "value2", usage: "usage2"},
		{name: "unique3", shorthand: "c", value: "value3", usage: "usage3"},
	})

	require.Len(t, result.SettingsByNames, 3)
	require.Len(t, result.SettingsByNames["unique1"], 1)
	require.Len(t, result.SettingsByNames["unique2"], 1)
	require.Len(t, result.SettingsByNames["unique3"], 1)
	require.Equal(t, "unique1", result.SettingsByNames["unique1"][0].name)
	require.Equal(t, "unique2", result.SettingsByNames["unique2"][0].name)
	require.Equal(t, "unique3", result.SettingsByNames["unique3"][0].name)

	require.Len(t, result.SettingsByShorthands, 3)
	require.Len(t, result.SettingsByShorthands["a"], 1)
	require.Len(t, result.SettingsByShorthands["b"], 1)
	require.Len(t, result.SettingsByShorthands["c"], 1)
	require.Equal(t, "a", result.SettingsByShorthands["a"][0].shorthand)
	require.Equal(t, "b", result.SettingsByShorthands["b"][0].shorthand)
	require.Equal(t, "c", result.SettingsByShorthands["c"][0].shorthand)

	require.Empty(t, result.DuplicateNames)

	require.Empty(t, result.DuplicateShorthands)
}

func TestAnalyzeSettings_DuplicateNamesOnly(t *testing.T) {
	result := analyzeSettings([]*Setting{
		{name: "duplicate", shorthand: "a", value: "value1", usage: "usage1"},
		{name: "duplicate", shorthand: "b", value: "value2", usage: "usage2"},
		{name: "unique", shorthand: "c", value: "value3", usage: "usage3"},
	})

	require.Len(t, result.SettingsByNames, 2)
	require.Len(t, result.SettingsByNames["duplicate"], 2)
	require.Len(t, result.SettingsByNames["unique"], 1)

	require.Contains(t, result.DuplicateNames, "duplicate")
	require.NotContains(t, result.DuplicateNames, "unique")

	require.Empty(t, result.DuplicateShorthands)
}

func TestAnalyzeSettings_DuplicateShorthandsOnly(t *testing.T) {
	result := analyzeSettings([]*Setting{
		{name: "unique1", shorthand: "d", value: "value1", usage: "usage1"},
		{name: "unique2", shorthand: "d", value: "value2", usage: "usage2"},
		{name: "unique3", shorthand: "u", value: "value3", usage: "usage3"},
	})

	require.Len(t, result.SettingsByShorthands, 2)
	require.Len(t, result.SettingsByShorthands["d"], 2)
	require.Len(t, result.SettingsByShorthands["u"], 1)

	require.Empty(t, result.DuplicateNames)

	require.Contains(t, result.DuplicateShorthands, "d")
	require.NotContains(t, result.DuplicateShorthands, "u")
}

func TestAnalyzeSettings_DuplicateNames_And_DuplicateShorthands(t *testing.T) {
	result := analyzeSettings([]*Setting{
		{name: "duplicate", shorthand: "d", value: "value1", usage: "usage1"},
		{name: "duplicate", shorthand: "d", value: "value2", usage: "usage2"},
		{name: "unique1", shorthand: "u", value: "value3", usage: "usage3"},
		{name: "unique2", shorthand: "U", value: "value4", usage: "usage4"},
	})

	require.Len(t, result.SettingsByNames, 3)
	require.Len(t, result.SettingsByNames["duplicate"], 2)
	require.Len(t, result.SettingsByNames["unique1"], 1)
	require.Len(t, result.SettingsByNames["unique2"], 1)

	require.Len(t, result.SettingsByShorthands, 3)
	require.Len(t, result.SettingsByShorthands["d"], 2)
	require.Len(t, result.SettingsByShorthands["u"], 1)
	require.Len(t, result.SettingsByShorthands["U"], 1)

	require.Contains(t, result.DuplicateNames, "duplicate")
	require.NotContains(t, result.DuplicateNames, "unique1")
	require.NotContains(t, result.DuplicateNames, "unique2")

	require.Contains(t, result.DuplicateShorthands, "d")
	require.NotContains(t, result.DuplicateShorthands, "u")
	require.NotContains(t, result.DuplicateShorthands, "U")
}
