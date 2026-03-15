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
	require.Equal(t, cryptoutilSharedMagic.OTLPServiceIdentityIDP, cfg.OTLPService)
	require.True(t, cfg.DevMode)
	require.Equal(t, defaultIDPAuthzServerURL, cfg.AuthzServerURL)
	require.Equal(t, defaultLoginPagePath, cfg.LoginPagePath)
	require.Equal(t, defaultConsentPagePath, cfg.ConsentPagePath)
	require.Equal(t, defaultEnableMFAEnrollment, cfg.EnableMFAEnrollment)
	require.Equal(t, defaultRequireMFA, cfg.RequireMFA)
	require.Equal(t, defaultLoginSessionTimeout, cfg.LoginSessionTimeout)
	require.Equal(t, defaultConsentSessionTimeout, cfg.ConsentSessionTimeout)
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

func TestIdentityIDPServerSettings_FullConfig(t *testing.T) {
	t.Parallel()

	cfg := NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	// Verify embedded template config is populated.
	require.NotNil(t, cfg.ServiceTemplateServerSettings)
	require.Equal(t, cryptoutilSharedMagic.IPv4Loopback, cfg.BindPublicAddress)

	// Verify idp-specific settings.
	require.Equal(t, defaultIDPAuthzServerURL, cfg.AuthzServerURL)
	require.Equal(t, defaultLoginPagePath, cfg.LoginPagePath)
	require.Equal(t, defaultConsentPagePath, cfg.ConsentPagePath)
	require.Equal(t, defaultEnableMFAEnrollment, cfg.EnableMFAEnrollment)
	require.Equal(t, defaultRequireMFA, cfg.RequireMFA)
	require.Equal(t, defaultMFAMethods, cfg.MFAMethods)
	require.Equal(t, defaultLoginSessionTimeout, cfg.LoginSessionTimeout)
	require.Equal(t, defaultConsentSessionTimeout, cfg.ConsentSessionTimeout)
}

func TestValidateIdentityIDPSettings_AuthzURLFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		authzURL  string
		devMode   bool
		wantError bool
	}{
		{"valid_https", "https://localhost:8200", false, false},
		{"valid_http", "http://localhost:8200", false, false},
		{"empty_prod_mode", "", false, true},
		{"empty_dev_mode", "", true, false}, // Empty allowed in dev mode.
		{"invalid_no_scheme", "localhost:8200", true, true},
		{"invalid_ftp_scheme", "ftp://localhost:8200", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, tt.devMode)
			cfg.AuthzServerURL = tt.authzURL

			err := validateIdentityIDPSettings(cfg)

			if tt.wantError {
				require.Error(t, err, "Expected validation error for authz URL: %s", tt.authzURL)
			} else {
				require.NoError(t, err, "Unexpected validation error for authz URL: %s", tt.authzURL)
			}
		})
	}
}

func TestValidateIdentityIDPSettings_SessionTimeouts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		loginTimeout   int
		consentTimeout int
		wantError      bool
	}{
		{"valid_defaults", defaultLoginSessionTimeout, defaultConsentSessionTimeout, false},
		{"valid_custom", 600, 600, false},
		{"invalid_login_zero", 0, defaultConsentSessionTimeout, true},
		{"invalid_login_negative", -1, defaultConsentSessionTimeout, true},
		{"invalid_consent_zero", defaultLoginSessionTimeout, 0, true},
		{"invalid_consent_negative", defaultLoginSessionTimeout, -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
			cfg.LoginSessionTimeout = tt.loginTimeout
			cfg.ConsentSessionTimeout = tt.consentTimeout

			err := validateIdentityIDPSettings(cfg)

			if tt.wantError {
				require.Error(t, err, "Expected validation error for session timeouts")
			} else {
				require.NoError(t, err, "Unexpected validation error for session timeouts")
			}
		})
	}
}
