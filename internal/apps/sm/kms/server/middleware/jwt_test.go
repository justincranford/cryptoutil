// Copyright (c) 2025 Justin Cranford
//
//

package middleware

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewJWTValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  JWTValidatorConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: JWTValidatorConfig{
				JWKSURL: "https://example.com/.well-known/jwks.json",
			},
			wantErr: false,
		},
		{
			name:    "missing JWKS URL",
			config:  JWTValidatorConfig{},
			wantErr: true,
			errMsg:  "JWKS URL is required",
		},
		{
			name: "with all options",
			config: JWTValidatorConfig{
				JWKSURL:                   "https://example.com/.well-known/jwks.json",
				CacheTTL:                  defaultJWKSCacheTTL * 2,
				RequiredIssuer:            "https://issuer.example.com",
				RequiredAudience:          "my-api",
				AllowedAlgorithms:         []string{cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilSharedMagic.JoseAlgES256},
				RevocationCheckEnabled:    true,
				IntrospectionURL:          "https://example.com/oauth2/introspect",
				IntrospectionClientID:     "client-id",
				IntrospectionClientSecret: "client-secret",
				ErrorDetailLevel:          errorDetailLevelDebug,
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			validator, err := NewJWTValidator(tc.config)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
				require.Nil(t, validator)
			} else {
				require.NoError(t, err)
				require.NotNil(t, validator)
			}
		})
	}
}

func TestJWTClaims_HasScope(t *testing.T) {
	t.Parallel()

	claims := &JWTClaims{
		Scopes: []string{cryptoutilSharedMagic.ScopeRead, cryptoutilSharedMagic.ScopeWrite, "admin"},
	}

	tests := []struct {
		name     string
		scope    string
		expected bool
	}{
		{name: "has read scope", scope: cryptoutilSharedMagic.ScopeRead, expected: true},
		{name: "has write scope", scope: cryptoutilSharedMagic.ScopeWrite, expected: true},
		{name: "has admin scope", scope: "admin", expected: true},
		{name: "missing delete scope", scope: "delete", expected: false},
		{name: "empty scope", scope: "", expected: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := claims.HasScope(tc.scope)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestJWTClaims_HasAnyScope(t *testing.T) {
	t.Parallel()

	claims := &JWTClaims{
		Scopes: []string{cryptoutilSharedMagic.ScopeRead, cryptoutilSharedMagic.ScopeWrite},
	}

	tests := []struct {
		name     string
		scopes   []string
		expected bool
	}{
		{name: "has one of the scopes", scopes: []string{cryptoutilSharedMagic.ScopeRead, "admin"}, expected: true},
		{name: "has exact match", scopes: []string{cryptoutilSharedMagic.ScopeRead}, expected: true},
		{name: "has both scopes", scopes: []string{cryptoutilSharedMagic.ScopeRead, cryptoutilSharedMagic.ScopeWrite}, expected: true},
		{name: "has none of the scopes", scopes: []string{"admin", "delete"}, expected: false},
		{name: "empty scopes", scopes: []string{}, expected: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := claims.HasAnyScope(tc.scopes...)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestJWTClaims_HasAllScopes(t *testing.T) {
	t.Parallel()

	claims := &JWTClaims{
		Scopes: []string{cryptoutilSharedMagic.ScopeRead, cryptoutilSharedMagic.ScopeWrite, "admin"},
	}

	tests := []struct {
		name     string
		scopes   []string
		expected bool
	}{
		{name: "has all scopes", scopes: []string{cryptoutilSharedMagic.ScopeRead, cryptoutilSharedMagic.ScopeWrite}, expected: true},
		{name: "has single scope", scopes: []string{cryptoutilSharedMagic.ScopeRead}, expected: true},
		{name: "missing one scope", scopes: []string{cryptoutilSharedMagic.ScopeRead, "delete"}, expected: false},
		{name: "missing all scopes", scopes: []string{"delete", "execute"}, expected: false},
		{name: "empty scopes (vacuously true)", scopes: []string{}, expected: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := claims.HasAllScopes(tc.scopes...)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestDefaultAllowedAlgorithms(t *testing.T) {
	t.Parallel()

	algorithms := DefaultAllowedAlgorithms()

	// Should contain FIPS-approved algorithms.
	require.Contains(t, algorithms, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm)
	require.Contains(t, algorithms, cryptoutilSharedMagic.JoseAlgRS384)
	require.Contains(t, algorithms, cryptoutilSharedMagic.JoseAlgRS512)
	require.Contains(t, algorithms, cryptoutilSharedMagic.JoseAlgES256)
	require.Contains(t, algorithms, cryptoutilSharedMagic.JoseAlgES384)
	require.Contains(t, algorithms, cryptoutilSharedMagic.JoseAlgES512)
	require.Contains(t, algorithms, cryptoutilSharedMagic.JoseAlgPS256)
	require.Contains(t, algorithms, cryptoutilSharedMagic.JoseAlgPS384)
	require.Contains(t, algorithms, cryptoutilSharedMagic.JoseAlgPS512)
	require.Contains(t, algorithms, cryptoutilSharedMagic.JoseAlgEdDSA)

	// Should not contain non-FIPS algorithms.
	require.NotContains(t, algorithms, cryptoutilSharedMagic.JoseAlgHS256)
	require.NotContains(t, algorithms, cryptoutilSharedMagic.JoseAlgHS384)
	require.NotContains(t, algorithms, cryptoutilSharedMagic.JoseAlgHS512)
	require.NotContains(t, algorithms, cryptoutilSharedMagic.PromptNone)
}

func TestIsAlgorithmAllowed(t *testing.T) {
	t.Parallel()

	validator, err := NewJWTValidator(JWTValidatorConfig{
		JWKSURL:           "https://example.com/.well-known/jwks.json",
		AllowedAlgorithms: []string{cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilSharedMagic.JoseAlgES256},
	})
	require.NoError(t, err)

	tests := []struct {
		name     string
		alg      string
		expected bool
	}{
		{name: "RS256 allowed", alg: cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, expected: true},
		{name: "ES256 allowed", alg: cryptoutilSharedMagic.JoseAlgES256, expected: true},
		{name: "RS384 not allowed", alg: cryptoutilSharedMagic.JoseAlgRS384, expected: false},
		{name: "HS256 not allowed", alg: cryptoutilSharedMagic.JoseAlgHS256, expected: false},
		{name: "empty not allowed", alg: "", expected: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := validator.isAlgorithmAllowed(tc.alg)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestDefaultsApplied(t *testing.T) {
	t.Parallel()

	validator, err := NewJWTValidator(JWTValidatorConfig{
		JWKSURL: "https://example.com/.well-known/jwks.json",
	})
	require.NoError(t, err)

	// Verify defaults are applied.
	require.Equal(t, defaultJWKSCacheTTL, validator.config.CacheTTL)
	require.Equal(t, errorDetailLevelMin, validator.config.ErrorDetailLevel)
	require.NotNil(t, validator.httpClient)
	require.NotNil(t, validator.cache)
}

func TestShouldPerformRevocationCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   JWTValidatorConfig
		claims   *JWTClaims
		expected bool
	}{
		{
			name: "disabled mode",
			config: JWTValidatorConfig{
				JWKSURL:             "https://example.com/.well-known/jwks.json",
				IntrospectionURL:    "https://example.com/introspect",
				RevocationCheckMode: RevocationCheckDisabled,
			},
			claims:   &JWTClaims{Scopes: []string{cryptoutilSharedMagic.ScopeRead}},
			expected: false,
		},
		{
			name: "every-request mode",
			config: JWTValidatorConfig{
				JWKSURL:             "https://example.com/.well-known/jwks.json",
				IntrospectionURL:    "https://example.com/introspect",
				RevocationCheckMode: RevocationCheckEveryRequest,
			},
			claims:   &JWTClaims{Scopes: []string{cryptoutilSharedMagic.ScopeRead}},
			expected: true,
		},
		{
			name: "sensitive-only with sensitive scope",
			config: JWTValidatorConfig{
				JWKSURL:             "https://example.com/.well-known/jwks.json",
				IntrospectionURL:    "https://example.com/introspect",
				RevocationCheckMode: RevocationCheckSensitiveOnly,
			},
			claims:   &JWTClaims{Scopes: []string{cryptoutilSharedMagic.ScopeRead, cryptoutilSharedMagic.ScopeWrite}},
			expected: true,
		},
		{
			name: "sensitive-only without sensitive scope",
			config: JWTValidatorConfig{
				JWKSURL:             "https://example.com/.well-known/jwks.json",
				IntrospectionURL:    "https://example.com/introspect",
				RevocationCheckMode: RevocationCheckSensitiveOnly,
			},
			claims:   &JWTClaims{Scopes: []string{cryptoutilSharedMagic.ScopeRead, cryptoutilSharedMagic.ClaimProfile}},
			expected: false,
		},
		{
			name: "sensitive-only with custom sensitive scopes",
			config: JWTValidatorConfig{
				JWKSURL:             "https://example.com/.well-known/jwks.json",
				IntrospectionURL:    "https://example.com/introspect",
				RevocationCheckMode: RevocationCheckSensitiveOnly,
				SensitiveScopes:     []string{"special:admin"},
			},
			claims:   &JWTClaims{Scopes: []string{cryptoutilSharedMagic.ScopeRead, "special:admin"}},
			expected: true,
		},
		{
			name: "no introspection URL",
			config: JWTValidatorConfig{
				JWKSURL:             "https://example.com/.well-known/jwks.json",
				RevocationCheckMode: RevocationCheckEveryRequest,
			},
			claims:   &JWTClaims{Scopes: []string{cryptoutilSharedMagic.ScopeRead}},
			expected: false,
		},
		{
			name: "backwards compatibility - old boolean flag",
			config: JWTValidatorConfig{
				JWKSURL:                "https://example.com/.well-known/jwks.json",
				IntrospectionURL:       "https://example.com/introspect",
				RevocationCheckEnabled: true,
			},
			claims:   &JWTClaims{Scopes: []string{cryptoutilSharedMagic.ScopeRead}},
			expected: true,
		},
		{
			name: "interval mode",
			config: JWTValidatorConfig{
				JWKSURL:             "https://example.com/.well-known/jwks.json",
				IntrospectionURL:    "https://example.com/introspect",
				RevocationCheckMode: RevocationCheckInterval,
			},
			claims:   &JWTClaims{Scopes: []string{cryptoutilSharedMagic.ScopeRead}},
			expected: true,
		},
		{
			name: "unknown mode falls through to default",
			config: JWTValidatorConfig{
				JWKSURL:             "https://example.com/.well-known/jwks.json",
				IntrospectionURL:    "https://example.com/introspect",
				RevocationCheckMode: "unknown-mode",
			},
			claims:   &JWTClaims{Scopes: []string{cryptoutilSharedMagic.ScopeRead}},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			validator, err := NewJWTValidator(tc.config)
			require.NoError(t, err)

			result := validator.shouldPerformRevocationCheck(tc.claims)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestHasSensitiveScope(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		sensitiveScopes []string
		claims          *JWTClaims
		expected        bool
	}{
		{
			name:            "default sensitive scopes - has admin",
			sensitiveScopes: nil,
			claims:          &JWTClaims{Scopes: []string{cryptoutilSharedMagic.ScopeRead, "admin"}},
			expected:        true,
		},
		{
			name:            "default sensitive scopes - has write",
			sensitiveScopes: nil,
			claims:          &JWTClaims{Scopes: []string{cryptoutilSharedMagic.ScopeRead, cryptoutilSharedMagic.ScopeWrite}},
			expected:        true,
		},
		{
			name:            "default sensitive scopes - has kms:admin",
			sensitiveScopes: nil,
			claims:          &JWTClaims{Scopes: []string{cryptoutilSharedMagic.ScopeRead, "kms:admin"}},
			expected:        true,
		},
		{
			name:            "default sensitive scopes - read only",
			sensitiveScopes: nil,
			claims:          &JWTClaims{Scopes: []string{cryptoutilSharedMagic.ScopeRead, cryptoutilSharedMagic.ClaimProfile}},
			expected:        false,
		},
		{
			name:            "custom sensitive scopes - match",
			sensitiveScopes: []string{"custom:write"},
			claims:          &JWTClaims{Scopes: []string{cryptoutilSharedMagic.ScopeRead, "custom:write"}},
			expected:        true,
		},
		{
			name:            "custom sensitive scopes - no match",
			sensitiveScopes: []string{"custom:write"},
			claims:          &JWTClaims{Scopes: []string{cryptoutilSharedMagic.ScopeRead, cryptoutilSharedMagic.ScopeWrite}},
			expected:        false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			validator, err := NewJWTValidator(JWTValidatorConfig{
				JWKSURL:         "https://example.com/.well-known/jwks.json",
				SensitiveScopes: tc.sensitiveScopes,
			})
			require.NoError(t, err)

			result := validator.hasSensitiveScope(tc.claims)
			require.Equal(t, tc.expected, result)
		})
	}
}
