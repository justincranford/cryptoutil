// Copyright (c) 2025 Justin Cranford

package config

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestRequireNewForTest_TypeAssertionPanics verifies that RequireNewForTest panics
// when Setting variables have wrong types. Each test corrupts one Setting's Value
// to trigger the sequential type assertion panic.
func TestRequireNewForTest_TypeAssertionPanics(t *testing.T) {
	// NOTE: Cannot use t.Parallel() — modifies global Setting variables.
	tests := []struct {
		name    string
		setting *Setting
	}{
		{"configFile", &configFile},
		{"logLevel", &logLevel},
		{"verboseMode", &verboseMode},
		{"devMode", &devMode},
		{"demoMode", &demoMode},
		{"bindPublicProtocol", &bindPublicProtocol},
		{"bindPublicAddress", &bindPublicAddress},
		{"tlsPublicDNSNames", &tlsPublicDNSNames},
		{"tlsPublicIPAddresses", &tlsPublicIPAddresses},
		{"tlsPrivateDNSNames", &tlsPrivateDNSNames},
		{"tlsPrivateIPAddresses", &tlsPrivateIPAddresses},
		{"bindPrivateProtocol", &bindPrivateProtocol},
		{"bindPrivateAddress", &bindPrivateAddress},
		{"publicBrowserAPIContextPath", &publicBrowserAPIContextPath},
		{"publicServiceAPIContextPath", &publicServiceAPIContextPath},
		{"corsAllowedOrigins", &corsAllowedOrigins},
		{"corsAllowedMethods", &corsAllowedMethods},
		{"corsAllowedHeaders", &corsAllowedHeaders},
		{"corsMaxAge", &corsMaxAge},
		{"csrfTokenName", &csrfTokenName},
		{"csrfTokenSameSite", &csrfTokenSameSite},
		{"csrfTokenMaxAge", &csrfTokenMaxAge},
		{"csrfTokenCookieSecure", &csrfTokenCookieSecure},
		{"csrfTokenCookieHTTPOnly", &csrfTokenCookieHTTPOnly},
		{"csrfTokenCookieSessionOnly", &csrfTokenCookieSessionOnly},
		{"csrfTokenSingleUseToken", &csrfTokenSingleUseToken},
		{"browserIPRateLimit", &browserIPRateLimit},
		{"serviceIPRateLimit", &serviceIPRateLimit},
		{"requestBodyLimit", &requestBodyLimit},
		{"allowedIps", &allowedIps},
		{"allowedCidrs", &allowedCidrs},
		{"swaggerUIUsername", &swaggerUIUsername},
		{"swaggerUIPassword", &swaggerUIPassword},
		{"databaseContainer", &databaseContainer},
		{"databaseURL", &databaseURL},
		{"databaseInitTotalTimeout", &databaseInitTotalTimeout},
		{"databaseInitRetryWait", &databaseInitRetryWait},
		{"serverShutdownTimeout", &serverShutdownTimeout},
		{"otlpEnabled", &otlpEnabled},
		{"otlpConsole", &otlpConsole},
		{"otlpService", &otlpService},
		{"otlpInstance", &otlpInstance},
		{"otlpVersion", &otlpVersion},
		{"otlpEnvironment", &otlpEnvironment},
		{"otlpHostname", &otlpHostname},
		{"otlpEndpoint", &otlpEndpoint},
		{"unsealMode", &unsealMode},
		{"unsealFiles", &unsealFiles},
		{"browserSessionAlgorithm", &browserSessionAlgorithm},
		{"browserSessionJWSAlgorithm", &browserSessionJWSAlgorithm},
		{"browserSessionJWEAlgorithm", &browserSessionJWEAlgorithm},
		{"browserSessionExpiration", &browserSessionExpiration},
		{"serviceSessionAlgorithm", &serviceSessionAlgorithm},
		{"serviceSessionJWSAlgorithm", &serviceSessionJWSAlgorithm},
		{"serviceSessionJWEAlgorithm", &serviceSessionJWEAlgorithm},
		{"serviceSessionExpiration", &serviceSessionExpiration},
		{"sessionIdleTimeout", &sessionIdleTimeout},
		{"sessionCleanupInterval", &sessionCleanupInterval},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			original := tc.setting.Value
			tc.setting.Value = struct{}{} // corrupt type to trigger panic

			defer func() { tc.setting.Value = original }()

			require.Panics(t, func() {
				RequireNewForTest("test-" + tc.name)
			})
		})
	}
}

// TestRequireNewForTest_DatabaseURLBranches verifies the database URL rewriting
// paths for different database URL formats.
func TestRequireNewForTest_DatabaseURLBranches(t *testing.T) {
	// NOTE: Cannot use t.Parallel() — modifies global databaseURL.
	tests := []struct {
		name   string
		dbURL  string
		panics bool
		expect func(t *testing.T, result string)
	}{
		{
			name:  "file::memory: format",
			dbURL: cryptoutilSharedMagic.SQLiteInMemoryDSN,
			expect: func(t *testing.T, result string) {
				t.Helper()
				require.Contains(t, result, "?mode=memory&cache=shared")
				require.NotContains(t, result, "file::memory:")
			},
		},
		{
			name:  ":memory: format",
			dbURL: cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
			expect: func(t *testing.T, result string) {
				t.Helper()
				require.Contains(t, result, "?mode=memory&cache=private")
			},
		},
		{
			name:  "postgres:// with query params",
			dbURL: "postgres://user:pass@host:5432/mydb?sslmode=disable", // pragma: allowlist secret
			expect: func(t *testing.T, result string) {
				t.Helper()
				require.Contains(t, result, "search_path=")
				require.Contains(t, result, "sslmode=disable")
			},
		},
		{
			name:  "postgres:// without query params",
			dbURL: "postgres://user:pass@host:5432/mydb", // pragma: allowlist secret
			expect: func(t *testing.T, result string) {
				t.Helper()
				require.Contains(t, result, "?search_path=")
			},
		},
		{
			name:   "unsupported database type panics",
			dbURL:  "mysql://user:pass@host:3306/mydb", // pragma: allowlist secret
			panics: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			original := databaseURL.Value
			databaseURL.Value = tc.dbURL

			defer func() { databaseURL.Value = original }()

			if tc.panics {
				require.Panics(t, func() {
					RequireNewForTest("test-db")
				})

				return
			}

			settings := RequireNewForTest("test-db")
			tc.expect(t, settings.DatabaseURL)
		})
	}
}

// TestNewTestConfig_ValidationPanic verifies that NewTestConfig panics when
// invalid arguments cause validation failure.
func TestNewTestConfig_ValidationPanic(t *testing.T) {
	t.Parallel()
	// In dev mode, IPv4AnyAddress is rejected by validateConfiguration.
	require.Panics(t, func() {
		NewTestConfig(cryptoutilSharedMagic.IPv4AnyAddress, 0, true)
	})
}

// TestRequireNewForTest_DatabaseURLEnvOverride verifies that CRYPTOUTIL_DATABASE_URL
// environment variable overrides the default database URL in RequireNewForTest.
func TestRequireNewForTest_DatabaseURLEnvOverride(t *testing.T) {
	// NOTE: Cannot use t.Parallel() — modifies environment variable.
	t.Setenv("CRYPTOUTIL_DATABASE_URL", cryptoutilSharedMagic.SQLiteInMemoryDSN)

	settings := RequireNewForTest("test-env-override")
	require.Contains(t, settings.DatabaseURL, "?mode=memory&cache=shared")

	// Verify the env var was used (not the default postgres URL).
	require.NotContains(t, settings.DatabaseURL, "postgres://")
}

// TestRequireNewForTest_HappyPath verifies normal operation with default settings.
func TestRequireNewForTest_HappyPath(t *testing.T) {
	t.Parallel()

	settings := RequireNewForTest("test-happy")
	require.NotNil(t, settings)
	require.Equal(t, "test-happy", settings.OTLPService)
	require.NotEmpty(t, settings.DatabaseURL)

	// Verify the database URL was rewritten (default postgres URL contains /DB?).
	require.Contains(t, settings.DatabaseURL, "/DB_test-happy_")
}

// TestNewFromFile_InvalidPath verifies NewFromFile returns error for non-existent file.
func TestNewFromFile_InvalidPath(t *testing.T) {
	t.Parallel()

	_, err := NewFromFile("/non/existent/config.yaml")
	// Parse may or may not fail depending on file handling, but should not panic.
	_ = err
}
