// Copyright (c) 2025 Justin Cranford
//
//

package config

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTokenConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *TokenConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_jws_tokens",
			config: &TokenConfig{
				AccessTokenLifetime:  cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				RefreshTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax * time.Second,
				IDTokenLifetime:      cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				AccessTokenFormat:    cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				RefreshTokenFormat:   cryptoutilSharedMagic.IdentityTokenFormatUUID,
				IDTokenFormat:        cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				Issuer:               "https://example.com",
				SigningAlgorithm:     cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			},
			expectError: false,
		},
		{
			name: "valid_jwe_access_token",
			config: &TokenConfig{
				AccessTokenLifetime:  cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				RefreshTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax * time.Second,
				IDTokenLifetime:      cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				AccessTokenFormat:    cryptoutilSharedMagic.IdentityTokenFormatJWE,
				RefreshTokenFormat:   cryptoutilSharedMagic.IdentityTokenFormatUUID,
				IDTokenFormat:        cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				Issuer:               "https://example.com",
				SigningAlgorithm:     cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			},
			expectError: false,
		},
		{
			name: "valid_uuid_access_token",
			config: &TokenConfig{
				AccessTokenLifetime:  cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				RefreshTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax * time.Second,
				IDTokenLifetime:      cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				AccessTokenFormat:    cryptoutilSharedMagic.IdentityTokenFormatUUID,
				RefreshTokenFormat:   cryptoutilSharedMagic.IdentityTokenFormatUUID,
				IDTokenFormat:        cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				Issuer:               "https://example.com",
				SigningAlgorithm:     cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			},
			expectError: false,
		},
		{
			name: "invalid_access_token_lifetime_zero",
			config: &TokenConfig{
				AccessTokenLifetime:  0,
				RefreshTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax * time.Second,
				IDTokenLifetime:      cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				AccessTokenFormat:    cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				RefreshTokenFormat:   cryptoutilSharedMagic.IdentityTokenFormatUUID,
				IDTokenFormat:        cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				Issuer:               "https://example.com",
				SigningAlgorithm:     cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			},
			expectError: true,
			errorMsg:    "access token lifetime must be positive",
		},
		{
			name: "invalid_access_token_lifetime_negative",
			config: &TokenConfig{
				AccessTokenLifetime:  -1 * time.Second,
				RefreshTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax * time.Second,
				IDTokenLifetime:      cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				AccessTokenFormat:    cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				RefreshTokenFormat:   cryptoutilSharedMagic.IdentityTokenFormatUUID,
				IDTokenFormat:        cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				Issuer:               "https://example.com",
				SigningAlgorithm:     cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			},
			expectError: true,
			errorMsg:    "access token lifetime must be positive",
		},
		{
			name: "invalid_refresh_token_lifetime_zero",
			config: &TokenConfig{
				AccessTokenLifetime:  cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				RefreshTokenLifetime: 0,
				IDTokenLifetime:      cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				AccessTokenFormat:    cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				RefreshTokenFormat:   cryptoutilSharedMagic.IdentityTokenFormatUUID,
				IDTokenFormat:        cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				Issuer:               "https://example.com",
				SigningAlgorithm:     cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			},
			expectError: true,
			errorMsg:    "refresh token lifetime must be positive",
		},
		{
			name: "invalid_id_token_lifetime_zero",
			config: &TokenConfig{
				AccessTokenLifetime:  cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				RefreshTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax * time.Second,
				IDTokenLifetime:      0,
				AccessTokenFormat:    cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				RefreshTokenFormat:   cryptoutilSharedMagic.IdentityTokenFormatUUID,
				IDTokenFormat:        cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				Issuer:               "https://example.com",
				SigningAlgorithm:     cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			},
			expectError: true,
			errorMsg:    "ID token lifetime must be positive",
		},
		{
			name: "invalid_access_token_format",
			config: &TokenConfig{
				AccessTokenLifetime:  cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				RefreshTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax * time.Second,
				IDTokenLifetime:      cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				AccessTokenFormat:    "invalid",
				RefreshTokenFormat:   cryptoutilSharedMagic.IdentityTokenFormatUUID,
				IDTokenFormat:        cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				Issuer:               "https://example.com",
				SigningAlgorithm:     cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			},
			expectError: true,
			errorMsg:    "access token format must be 'jws', 'jwe', or 'uuid'",
		},
		{
			name: "invalid_refresh_token_format",
			config: &TokenConfig{
				AccessTokenLifetime:  cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				RefreshTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax * time.Second,
				IDTokenLifetime:      cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				AccessTokenFormat:    cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				RefreshTokenFormat:   cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				IDTokenFormat:        cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				Issuer:               "https://example.com",
				SigningAlgorithm:     cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			},
			expectError: true,
			errorMsg:    "refresh token format must be 'uuid'",
		},
		{
			name: "invalid_id_token_format",
			config: &TokenConfig{
				AccessTokenLifetime:  cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				RefreshTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax * time.Second,
				IDTokenLifetime:      cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				AccessTokenFormat:    cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				RefreshTokenFormat:   cryptoutilSharedMagic.IdentityTokenFormatUUID,
				IDTokenFormat:        cryptoutilSharedMagic.IdentityTokenFormatUUID,
				Issuer:               "https://example.com",
				SigningAlgorithm:     cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			},
			expectError: true,
			errorMsg:    "ID token format must be 'jws'",
		},
		{
			name: "missing_issuer",
			config: &TokenConfig{
				AccessTokenLifetime:  cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				RefreshTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax * time.Second,
				IDTokenLifetime:      cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				AccessTokenFormat:    cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				RefreshTokenFormat:   cryptoutilSharedMagic.IdentityTokenFormatUUID,
				IDTokenFormat:        cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				Issuer:               "",
				SigningAlgorithm:     cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			},
			expectError: true,
			errorMsg:    "token issuer is required",
		},
		{
			name: "missing_signing_algorithm",
			config: &TokenConfig{
				AccessTokenLifetime:  cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				RefreshTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax * time.Second,
				IDTokenLifetime:      cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				AccessTokenFormat:    cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				RefreshTokenFormat:   cryptoutilSharedMagic.IdentityTokenFormatUUID,
				IDTokenFormat:        cryptoutilSharedMagic.DefaultBrowserSessionCookie,
				Issuer:               "https://example.com",
				SigningAlgorithm:     "",
			},
			expectError: true,
			errorMsg:    "signing algorithm is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.config.Validate()
			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSessionConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *SessionConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_strict_samesite",
			config: &SessionConfig{
				SessionLifetime: cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				IdleTimeout:     cryptoutilSharedMagic.IMEnterpriseSessionTimeout * time.Second,
				CookieName:      "session_id",
				CookieSameSite:  cryptoutilSharedMagic.DefaultCSRFTokenSameSiteStrict,
				CookieSecure:    true,
				CookieHTTPOnly:  true,
			},
			expectError: false,
		},
		{
			name: "valid_lax_samesite",
			config: &SessionConfig{
				SessionLifetime: cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				IdleTimeout:     cryptoutilSharedMagic.IMEnterpriseSessionTimeout * time.Second,
				CookieName:      "session_id",
				CookieSameSite:  "Lax",
				CookieSecure:    true,
				CookieHTTPOnly:  true,
			},
			expectError: false,
		},
		{
			name: "valid_none_samesite",
			config: &SessionConfig{
				SessionLifetime: cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				IdleTimeout:     cryptoutilSharedMagic.IMEnterpriseSessionTimeout * time.Second,
				CookieName:      "session_id",
				CookieSameSite:  "None",
				CookieSecure:    true,
				CookieHTTPOnly:  true,
			},
			expectError: false,
		},
		{
			name: "invalid_session_lifetime_zero",
			config: &SessionConfig{
				SessionLifetime: 0,
				IdleTimeout:     cryptoutilSharedMagic.IMEnterpriseSessionTimeout * time.Second,
				CookieName:      "session_id",
				CookieSameSite:  cryptoutilSharedMagic.DefaultCSRFTokenSameSiteStrict,
			},
			expectError: true,
			errorMsg:    "session lifetime must be positive",
		},
		{
			name: "invalid_session_lifetime_negative",
			config: &SessionConfig{
				SessionLifetime: -1 * time.Second,
				IdleTimeout:     cryptoutilSharedMagic.IMEnterpriseSessionTimeout * time.Second,
				CookieName:      "session_id",
				CookieSameSite:  cryptoutilSharedMagic.DefaultCSRFTokenSameSiteStrict,
			},
			expectError: true,
			errorMsg:    "session lifetime must be positive",
		},
		{
			name: "invalid_idle_timeout_zero",
			config: &SessionConfig{
				SessionLifetime: cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				IdleTimeout:     0,
				CookieName:      "session_id",
				CookieSameSite:  cryptoutilSharedMagic.DefaultCSRFTokenSameSiteStrict,
			},
			expectError: true,
			errorMsg:    "idle timeout must be positive",
		},
		{
			name: "invalid_idle_timeout_negative",
			config: &SessionConfig{
				SessionLifetime: cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				IdleTimeout:     -1 * time.Second,
				CookieName:      "session_id",
				CookieSameSite:  cryptoutilSharedMagic.DefaultCSRFTokenSameSiteStrict,
			},
			expectError: true,
			errorMsg:    "idle timeout must be positive",
		},
		{
			name: "missing_cookie_name",
			config: &SessionConfig{
				SessionLifetime: cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				IdleTimeout:     cryptoutilSharedMagic.IMEnterpriseSessionTimeout * time.Second,
				CookieName:      "",
				CookieSameSite:  cryptoutilSharedMagic.DefaultCSRFTokenSameSiteStrict,
			},
			expectError: true,
			errorMsg:    "cookie name is required",
		},
		{
			name: "invalid_cookie_samesite",
			config: &SessionConfig{
				SessionLifetime: cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
				IdleTimeout:     cryptoutilSharedMagic.IMEnterpriseSessionTimeout * time.Second,
				CookieName:      "session_id",
				CookieSameSite:  "Invalid",
			},
			expectError: true,
			errorMsg:    "cookie SameSite must be 'Strict', 'Lax', or 'None'",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.config.Validate()
			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
