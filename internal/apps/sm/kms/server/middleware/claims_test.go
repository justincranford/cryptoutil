// Copyright (c) 2025 Justin Cranford
//
//

package middleware

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewClaimsExtractor(t *testing.T) {
	t.Parallel()

	extractor := NewClaimsExtractor()

	require.NotNil(t, extractor)
	require.NotEmpty(t, extractor.KnownClaims)
	require.Contains(t, extractor.KnownClaims, cryptoutilSharedMagic.ClaimSub)
	require.Contains(t, extractor.KnownClaims, "tenant_id")
}

func TestClaimsExtractor_ExtractFromMap(t *testing.T) {
	t.Parallel()

	extractor := NewClaimsExtractor()

	tests := []struct {
		name      string
		rawClaims map[string]any
		wantErr   bool
		validate  func(*testing.T, *OIDCClaims)
	}{
		{
			name: "standard claims",
			rawClaims: map[string]any{
				cryptoutilSharedMagic.ClaimSub:   "user-123",
				cryptoutilSharedMagic.ClaimIss:   "https://issuer.example.com",
				cryptoutilSharedMagic.ClaimAud:   []any{"api-1", "api-2"},
				cryptoutilSharedMagic.ClaimEmail: "user@example.com",
				cryptoutilSharedMagic.ClaimName:  "Test User",
			},
			wantErr: false,
			validate: func(t *testing.T, claims *OIDCClaims) {
				t.Helper()
				require.Equal(t, "user-123", claims.Subject)
				require.Equal(t, "https://issuer.example.com", claims.Issuer)
				require.Equal(t, "user@example.com", claims.Email)
				require.Equal(t, "Test User", claims.Name)
			},
		},
		{
			name: "with scopes",
			rawClaims: map[string]any{
				cryptoutilSharedMagic.ClaimSub:   "service-123",
				cryptoutilSharedMagic.ClaimScope: "kms:read kms:write",
			},
			wantErr: false,
			validate: func(t *testing.T, claims *OIDCClaims) {
				t.Helper()
				require.Equal(t, "kms:read kms:write", claims.Scope)
				require.Len(t, claims.Scopes, 2)
				require.Contains(t, claims.Scopes, "kms:read")
				require.Contains(t, claims.Scopes, "kms:write")
			},
		},
		{
			name: "with tenant claims",
			rawClaims: map[string]any{
				cryptoutilSharedMagic.ClaimSub: "user-123",
				"tenant_id":                    "tenant-abc",
				"tenant_name":                  "ACME Corp",
			},
			wantErr: false,
			validate: func(t *testing.T, claims *OIDCClaims) {
				t.Helper()
				require.Equal(t, "tenant-abc", claims.TenantID)
				require.Equal(t, "ACME Corp", claims.TenantName)
			},
		},
		{
			name: "with groups and roles",
			rawClaims: map[string]any{
				cryptoutilSharedMagic.ClaimSub: "user-123",
				"groups":                       []any{"admins", "developers"},
				"roles":                        []any{"admin", "user"},
				"permissions":                  []any{cryptoutilSharedMagic.ScopeRead, cryptoutilSharedMagic.ScopeWrite},
			},
			wantErr: false,
			validate: func(t *testing.T, claims *OIDCClaims) {
				t.Helper()
				require.Len(t, claims.Groups, 2)
				require.Contains(t, claims.Groups, "admins")
				require.Len(t, claims.Roles, 2)
				require.Contains(t, claims.Roles, "admin")
				require.Len(t, claims.Permissions, 2)
			},
		},
		{
			name: "with custom claims",
			rawClaims: map[string]any{
				cryptoutilSharedMagic.ClaimSub: "user-123",
				"custom_field":                 "custom_value",
				"urn:example:custom_claim":     "namespaced_value",
			},
			wantErr: false,
			validate: func(t *testing.T, claims *OIDCClaims) {
				t.Helper()
				require.NotEmpty(t, claims.Custom)
				require.Equal(t, "custom_value", claims.Custom["custom_field"])
				require.Equal(t, "namespaced_value", claims.Custom["urn:example:custom_claim"])
			},
		},
		{
			name:      "nil claims",
			rawClaims: nil,
			wantErr:   true,
		},
		{
			name:      "marshal error - unsupported type",
			rawClaims: map[string]any{cryptoutilSharedMagic.ClaimSub: "user-123", "bad": make(chan int)},
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			claims, err := extractor.ExtractFromMap(tc.rawClaims)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, claims)
			} else {
				require.NoError(t, err)
				require.NotNil(t, claims)
				tc.validate(t, claims)
			}
		})
	}
}

func TestConvertFromJWTClaims(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		jwtClaims *JWTClaims
		expected  *OIDCClaims
	}{
		{
			name: "full conversion",
			jwtClaims: &JWTClaims{
				Subject:           "user-123",
				Issuer:            "https://issuer.example.com",
				Audience:          []string{"api-1"},
				ExpiresAt:         time.Date(2025, cryptoutilSharedMagic.HashPrefixLength, 31, 0, 0, 0, 0, time.UTC),
				Name:              "Test User",
				PreferredUsername: "testuser",
				Email:             "test@example.com",
				EmailVerified:     true,
				Scope:             "kms:read",
				Scopes:            []string{"kms:read"},
				Custom:            map[string]any{"custom": "value"},
			},
			expected: &OIDCClaims{
				Subject:           "user-123",
				Issuer:            "https://issuer.example.com",
				Audience:          []string{"api-1"},
				ExpiresAt:         time.Date(2025, cryptoutilSharedMagic.HashPrefixLength, 31, 0, 0, 0, 0, time.UTC),
				Name:              "Test User",
				PreferredUsername: "testuser",
				Email:             "test@example.com",
				EmailVerified:     true,
				Scope:             "kms:read",
				Scopes:            []string{"kms:read"},
				Custom:            map[string]any{"custom": "value"},
			},
		},
		{
			name:      "nil input",
			jwtClaims: nil,
			expected:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := ConvertFromJWTClaims(tc.jwtClaims)

			if tc.expected == nil {
				require.Nil(t, result)
			} else {
				require.NotNil(t, result)
				require.Equal(t, tc.expected.Subject, result.Subject)
				require.Equal(t, tc.expected.Issuer, result.Issuer)
				require.Equal(t, tc.expected.Email, result.Email)
				require.Equal(t, tc.expected.Scopes, result.Scopes)
			}
		})
	}
}

func TestGetOIDCClaims(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func() context.Context
		expected string // Subject value for comparison.
	}{
		{
			name: "from OIDC context",
			setup: func() context.Context {
				claims := &OIDCClaims{Subject: "oidc-user"}

				return context.WithValue(context.Background(), OIDCClaimsContextKey{}, claims)
			},
			expected: "oidc-user",
		},
		{
			name: "from JWT context",
			setup: func() context.Context {
				claims := &JWTClaims{Subject: "jwt-user"}

				return context.WithValue(context.Background(), JWTContextKey{}, claims)
			},
			expected: "jwt-user",
		},
		{
			name: "no claims",
			setup: func() context.Context {
				return context.Background()
			},
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := tc.setup()
			claims := GetOIDCClaims(ctx)

			if tc.expected == "" {
				require.Nil(t, claims)
			} else {
				require.NotNil(t, claims)
				require.Equal(t, tc.expected, claims.Subject)
			}
		})
	}
}

func TestOIDCClaims_HasScope(t *testing.T) {
	t.Parallel()

	claims := &OIDCClaims{
		Scopes: []string{"kms:read", "kms:write"},
	}

	require.True(t, claims.HasScope("kms:read"))
	require.True(t, claims.HasScope("kms:write"))
	require.False(t, claims.HasScope("kms:admin"))
}

func TestOIDCClaims_HasAnyScope(t *testing.T) {
	t.Parallel()

	claims := &OIDCClaims{
		Scopes: []string{"kms:read"},
	}

	require.True(t, claims.HasAnyScope("kms:read", "kms:write"))
	require.False(t, claims.HasAnyScope("kms:admin", "kms:delete"))
}

func TestOIDCClaims_HasAllScopes(t *testing.T) {
	t.Parallel()

	claims := &OIDCClaims{
		Scopes: []string{"kms:read", "kms:write"},
	}

	require.True(t, claims.HasAllScopes("kms:read"))
	require.True(t, claims.HasAllScopes("kms:read", "kms:write"))
	require.False(t, claims.HasAllScopes("kms:read", "kms:admin"))
}

func TestOIDCClaims_HasGroup(t *testing.T) {
	t.Parallel()

	claims := &OIDCClaims{
		Groups: []string{"admins", "developers"},
	}

	require.True(t, claims.HasGroup("admins"))
	require.False(t, claims.HasGroup("managers"))
}

func TestOIDCClaims_HasRole(t *testing.T) {
	t.Parallel()

	claims := &OIDCClaims{
		Roles: []string{"admin", "user"},
	}

	require.True(t, claims.HasRole("admin"))
	require.False(t, claims.HasRole("superuser"))
}

func TestOIDCClaims_HasPermission(t *testing.T) {
	t.Parallel()

	claims := &OIDCClaims{
		Permissions: []string{cryptoutilSharedMagic.ScopeRead, cryptoutilSharedMagic.ScopeWrite},
	}

	require.True(t, claims.HasPermission(cryptoutilSharedMagic.ScopeRead))
	require.False(t, claims.HasPermission("delete"))
}

func TestOIDCClaims_GetCustomClaim(t *testing.T) {
	t.Parallel()

	claims := &OIDCClaims{
		Custom: map[string]any{
			"key1": "value1",
			"key2": 123,
		},
	}

	val, exists := claims.GetCustomClaim("key1")
	require.True(t, exists)
	require.Equal(t, "value1", val)

	val, exists = claims.GetCustomClaim("key2")
	require.True(t, exists)
	require.Equal(t, 123, val)

	val, exists = claims.GetCustomClaim("missing")
	require.False(t, exists)
	require.Nil(t, val)
}

func TestOIDCClaims_GetCustomString(t *testing.T) {
	t.Parallel()

	claims := &OIDCClaims{
		Custom: map[string]any{
			"str":    "string_value",
			"nonstr": 123,
		},
	}

	require.Equal(t, "string_value", claims.GetCustomString("str"))
	require.Equal(t, "", claims.GetCustomString("nonstr"))
	require.Equal(t, "", claims.GetCustomString("missing"))
}

func TestOIDCClaims_IsExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		exp      time.Time
		expected bool
	}{
		{
			name:     "not expired",
			exp:      time.Now().UTC().Add(time.Hour),
			expected: false,
		},
		{
			name:     "expired",
			exp:      time.Now().UTC().Add(-time.Hour),
			expected: true,
		},
		{
			name:     "zero time (no expiry)",
			exp:      time.Time{},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			claims := &OIDCClaims{ExpiresAt: tc.exp}
			require.Equal(t, tc.expected, claims.IsExpired())
		})
	}
}

func TestOIDCClaims_TimeUntilExpiry(t *testing.T) {
	t.Parallel()

	// Future expiry.
	claims := &OIDCClaims{ExpiresAt: time.Now().UTC().Add(time.Hour)}
	dur := claims.TimeUntilExpiry()
	require.Greater(t, dur, time.Duration(0))
	require.LessOrEqual(t, dur, time.Hour)

	// Zero time.
	zeroClaims := &OIDCClaims{}
	require.Equal(t, time.Duration(0), zeroClaims.TimeUntilExpiry())
}

func TestOIDCClaims_NilCustomMap(t *testing.T) {
	t.Parallel()

	claims := &OIDCClaims{
		Custom: nil,
	}

	val, exists := claims.GetCustomClaim("any")
	require.False(t, exists)
	require.Nil(t, val)

	require.Equal(t, "", claims.GetCustomString("any"))
}
