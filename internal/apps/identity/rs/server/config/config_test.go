// Copyright (c) 2025 Justin Cranford

package config

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestNewTestConfig(t *testing.T) {
	t.Parallel()

	cfg := NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	require.NotNil(t, cfg)
	require.Equal(t, cryptoutilSharedMagic.OTLPServiceIdentityRS, cfg.OTLPService)
	require.True(t, cfg.DevMode)
	require.Equal(t, defaultRSAuthzServerURL, cfg.AuthzServerURL)
	require.Equal(t, defaultJWKSEndpoint, cfg.JWKSEndpoint)
	require.Equal(t, defaultIntrospectionURL, cfg.IntrospectionURL)
	require.Equal(t, defaultAllowBearerToken, cfg.AllowBearerToken)
	require.Equal(t, defaultAllowClientCert, cfg.AllowClientCert)
	require.Equal(t, defaultJWKSCacheTTL, cfg.JWKSCacheTTL)
	require.Equal(t, defaultTokenCacheTTL, cfg.TokenCacheTTL)
	require.Equal(t, defaultEnableTokenCaching, cfg.EnableTokenCaching)
}

func TestDefaultTestConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultTestConfig()

	require.NotNil(t, cfg)
	require.Equal(t, cryptoutilSharedMagic.IPv4Loopback, cfg.BindPublicAddress)
	require.Equal(t, uint16(0), cfg.BindPublicPort, "Should use dynamic port allocation")
	require.True(t, cfg.DevMode, "Should be in dev mode")
}

func TestNewTestConfig_ProductionMode(t *testing.T) {
	t.Parallel()

	cfg := NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 8110, false)

	require.NotNil(t, cfg)
	require.Equal(t, uint16(8110), cfg.BindPublicPort)
	require.False(t, cfg.DevMode, "Should not be in dev mode")
}

func TestIdentityRSServerSettings_FullConfig(t *testing.T) {
	t.Parallel()

	cfg := NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	// Verify embedded template config is populated.
	require.NotNil(t, cfg.ServiceTemplateServerSettings)
	require.Equal(t, cryptoutilSharedMagic.IPv4Loopback, cfg.BindPublicAddress)

	// Verify rs-specific settings.
	require.Equal(t, defaultRSAuthzServerURL, cfg.AuthzServerURL)
	require.Equal(t, defaultJWKSEndpoint, cfg.JWKSEndpoint)
	require.Equal(t, defaultIntrospectionURL, cfg.IntrospectionURL)
	require.Equal(t, defaultRequiredScopes, cfg.RequiredScopes)
	require.Equal(t, defaultRequiredAudiences, cfg.RequiredAudiences)
	require.Equal(t, defaultAllowBearerToken, cfg.AllowBearerToken)
	require.Equal(t, defaultAllowClientCert, cfg.AllowClientCert)
	require.Equal(t, defaultJWKSCacheTTL, cfg.JWKSCacheTTL)
	require.Equal(t, defaultTokenCacheTTL, cfg.TokenCacheTTL)
	require.Equal(t, defaultEnableTokenCaching, cfg.EnableTokenCaching)
}

func TestValidateIdentityRSSettings_AuthzURLFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		authzURL  string
		devMode   bool
		wantError bool
	}{
		{"valid_https", "https://localhost:8100", false, false},
		{"valid_http", "http://localhost:8100", false, false},
		{"empty_prod_mode", "", false, true},
		{"empty_dev_mode", "", true, false}, // Empty allowed in dev mode.
		{"invalid_no_scheme", "localhost:8100", true, true},
		{"invalid_ftp_scheme", "ftp://localhost:8100", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, tt.devMode)
			cfg.AuthzServerURL = tt.authzURL

			err := validateIdentityRSSettings(cfg)

			if tt.wantError {
				require.Error(t, err, "Expected validation error for authz URL: %s", tt.authzURL)
			} else {
				require.NoError(t, err, "Unexpected validation error for authz URL: %s", tt.authzURL)
			}
		})
	}
}

func TestValidateIdentityRSSettings_AuthMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		allowBearerToken bool
		allowClientCert  bool
		wantError        bool
	}{
		{"both_enabled", true, true, false},
		{"bearer_only", true, false, false},
		{"client_cert_only", false, true, false},
		{"neither_enabled", false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
			cfg.AllowBearerToken = tt.allowBearerToken
			cfg.AllowClientCert = tt.allowClientCert

			err := validateIdentityRSSettings(cfg)

			if tt.wantError {
				require.Error(t, err, "Expected validation error for auth method config")
			} else {
				require.NoError(t, err, "Unexpected validation error for auth method config")
			}
		})
	}
}

func TestValidateIdentityRSSettings_CacheTTLs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		jwksCacheTTL  int
		tokenCacheTTL int
		wantError     bool
	}{
		{"valid_defaults", defaultJWKSCacheTTL, defaultTokenCacheTTL, false},
		{"valid_custom", 7200, 600, false},
		{"valid_zero", 0, 0, false},
		{"invalid_jwks_negative", -1, defaultTokenCacheTTL, true},
		{"invalid_token_negative", defaultJWKSCacheTTL, -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
			cfg.JWKSCacheTTL = tt.jwksCacheTTL
			cfg.TokenCacheTTL = tt.tokenCacheTTL

			err := validateIdentityRSSettings(cfg)

			if tt.wantError {
				require.Error(t, err, "Expected validation error for cache TTLs")
			} else {
				require.NoError(t, err, "Unexpected validation error for cache TTLs")
			}
		})
	}
}
