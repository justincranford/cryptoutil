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
	require.Equal(t, cryptoutilSharedMagic.OTLPServiceIdentityAuthz, cfg.OTLPService)
	require.True(t, cfg.DevMode)
	require.Equal(t, defaultIssuer, cfg.Issuer)
	require.Equal(t, defaultTokenLifetime, cfg.TokenLifetime)
	require.Equal(t, defaultRefreshTokenLifetime, cfg.RefreshTokenLifetime)
	require.Equal(t, defaultAuthorizationCodeTTL, cfg.AuthorizationCodeTTL)
	require.True(t, cfg.EnableDiscovery)
	require.False(t, cfg.EnableDynamicRegistration)
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

	cfg := NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, cryptoutilSharedMagic.PKICAServicePort, false)

	require.NotNil(t, cfg)
	require.Equal(t, uint16(cryptoutilSharedMagic.PKICAServicePort), cfg.BindPublicPort)
	require.False(t, cfg.DevMode, "Should not be in dev mode")
}

func TestIdentityAuthzServerSettings_FullConfig(t *testing.T) {
	t.Parallel()

	cfg := NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	// Verify embedded template config is populated.
	require.NotNil(t, cfg.ServiceTemplateServerSettings)
	require.Equal(t, cryptoutilSharedMagic.IPv4Loopback, cfg.BindPublicAddress)

	// Verify authz-specific settings.
	require.Equal(t, defaultIssuer, cfg.Issuer)
	require.Equal(t, defaultTokenLifetime, cfg.TokenLifetime)
	require.Equal(t, defaultRefreshTokenLifetime, cfg.RefreshTokenLifetime)
	require.Equal(t, defaultAuthorizationCodeTTL, cfg.AuthorizationCodeTTL)
	require.True(t, cfg.EnableDiscovery)
	require.False(t, cfg.EnableDynamicRegistration)
}

func TestValidateIdentityAuthzSettings_IssuerFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		issuer    string
		wantError bool
	}{
		{"valid_https", "https://localhost:8200", false},
		{"valid_http", "http://localhost:8200", false},
		{"invalid_no_scheme", "localhost:8200", true},
		{"invalid_ftp_scheme", "ftp://localhost:8200", true},
		{"empty_issuer", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
			cfg.Issuer = tt.issuer

			err := validateIdentityAuthzSettings(cfg)

			if tt.wantError {
				require.Error(t, err, "Expected validation error for issuer: %s", tt.issuer)
			} else {
				require.NoError(t, err, "Unexpected validation error for issuer: %s", tt.issuer)
			}
		})
	}
}

func TestValidateIdentityAuthzSettings_TokenLifetimes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		tokenLifetime        int
		refreshTokenLifetime int
		authCodeTTL          int
		wantError            bool
	}{
		{"valid_defaults", defaultTokenLifetime, defaultRefreshTokenLifetime, defaultAuthorizationCodeTTL, false},
		{"valid_custom", 7200, 172800, 300, false},
		{"invalid_token_zero", 0, defaultRefreshTokenLifetime, defaultAuthorizationCodeTTL, true},
		{"invalid_token_negative", -1, defaultRefreshTokenLifetime, defaultAuthorizationCodeTTL, true},
		{"invalid_refresh_zero", defaultTokenLifetime, 0, defaultAuthorizationCodeTTL, true},
		{"invalid_authcode_zero", defaultTokenLifetime, defaultRefreshTokenLifetime, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
			cfg.TokenLifetime = tt.tokenLifetime
			cfg.RefreshTokenLifetime = tt.refreshTokenLifetime
			cfg.AuthorizationCodeTTL = tt.authCodeTTL

			err := validateIdentityAuthzSettings(cfg)

			if tt.wantError {
				require.Error(t, err, "Expected validation error for token lifetimes")
			} else {
				require.NoError(t, err, "Unexpected validation error for token lifetimes")
			}
		})
	}
}
