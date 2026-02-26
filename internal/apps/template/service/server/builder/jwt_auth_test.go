// Copyright (c) 2025 Justin Cranford
//
//

package builder

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewDefaultJWTAuthConfig(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultJWTAuthConfig()

	require.Equal(t, JWTAuthModeSession, cfg.Mode)
	require.Equal(t, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Minute, cfg.CacheTTL)
	require.Equal(t, "minimal", cfg.ErrorDetailLevel)
	require.False(t, cfg.IsEnabled())
	require.False(t, cfg.IsRequired())
}

func TestNewKMSJWTAuthConfig(t *testing.T) {
	t.Parallel()

	cfg := NewKMSJWTAuthConfig("https://example.com/.well-known/jwks.json", "https://auth.example.com", "kms-service")

	require.Equal(t, JWTAuthModeRequired, cfg.Mode)
	require.Equal(t, "https://example.com/.well-known/jwks.json", cfg.JWKSURL)
	require.Equal(t, "https://auth.example.com", cfg.RequiredIssuer)
	require.Equal(t, "kms-service", cfg.RequiredAudience)
	require.True(t, cfg.IsEnabled())
	require.True(t, cfg.IsRequired())
}

func TestJWTAuthConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *JWTAuthConfig
		wantErr bool
	}{
		{
			name:    "disabled mode - no validation needed",
			config:  NewDefaultJWTAuthConfig(),
			wantErr: false,
		},
		{
			name: "required mode with valid config",
			config: &JWTAuthConfig{
				Mode:              JWTAuthModeRequired,
				JWKSURL:           "https://example.com/.well-known/jwks.json",
				AllowedAlgorithms: []string{cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm},
			},
			wantErr: false,
		},
		{
			name: "required mode - missing JWKS URL",
			config: &JWTAuthConfig{
				Mode:              JWTAuthModeRequired,
				AllowedAlgorithms: []string{cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm},
			},
			wantErr: true,
		},
		{
			name: "required mode - missing algorithms",
			config: &JWTAuthConfig{
				Mode:              JWTAuthModeRequired,
				JWKSURL:           "https://example.com/.well-known/jwks.json",
				AllowedAlgorithms: []string{},
			},
			wantErr: true,
		},
		{
			name: "optional mode with valid config",
			config: &JWTAuthConfig{
				Mode:              JWTAuthModeOptional,
				JWKSURL:           "https://example.com/.well-known/jwks.json",
				AllowedAlgorithms: []string{cryptoutilSharedMagic.JoseAlgES256},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestJWTClaims_Scopes(t *testing.T) {
	t.Parallel()

	claims := &JWTClaims{
		Subject: "user123",
		Scopes:  []string{cryptoutilSharedMagic.ScopeRead, cryptoutilSharedMagic.ScopeWrite, "admin"},
	}

	t.Run("HasScope", func(t *testing.T) {
		t.Parallel()

		require.True(t, claims.HasScope(cryptoutilSharedMagic.ScopeRead))
		require.True(t, claims.HasScope(cryptoutilSharedMagic.ScopeWrite))
		require.True(t, claims.HasScope("admin"))
		require.False(t, claims.HasScope("delete"))
	})

	t.Run("HasAnyScope", func(t *testing.T) {
		t.Parallel()

		require.True(t, claims.HasAnyScope(cryptoutilSharedMagic.ScopeRead, "unknown"))
		require.True(t, claims.HasAnyScope("unknown", cryptoutilSharedMagic.ScopeWrite))
		require.False(t, claims.HasAnyScope("unknown1", "unknown2"))
	})

	t.Run("HasAllScopes", func(t *testing.T) {
		t.Parallel()

		require.True(t, claims.HasAllScopes(cryptoutilSharedMagic.ScopeRead, cryptoutilSharedMagic.ScopeWrite))
		require.True(t, claims.HasAllScopes(cryptoutilSharedMagic.ScopeRead))
		require.False(t, claims.HasAllScopes(cryptoutilSharedMagic.ScopeRead, "delete"))
	})
}

func TestGetJWTClaims(t *testing.T) {
	t.Parallel()

	t.Run("claims present", func(t *testing.T) {
		t.Parallel()

		claims := &JWTClaims{Subject: "user123"}
		ctx := context.WithValue(context.Background(), JWTContextKey{}, claims)

		result := GetJWTClaims(ctx)
		require.NotNil(t, result)
		require.Equal(t, "user123", result.Subject)
	})

	t.Run("claims absent", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		result := GetJWTClaims(ctx)
		require.Nil(t, result)
	})
}

func TestJWTAuthConfig_ShouldSkipPath(t *testing.T) {
	t.Parallel()

	cfg := &JWTAuthConfig{
		SkipPaths: []string{"/health", "/metrics", "/public"},
	}

	require.True(t, cfg.ShouldSkipPath("/health"))
	require.True(t, cfg.ShouldSkipPath("/metrics"))
	require.True(t, cfg.ShouldSkipPath("/public"))
	require.False(t, cfg.ShouldSkipPath("/api/protected"))
}

func TestWithJWTAuth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *JWTAuthConfig
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "invalid config - required mode missing JWKS URL",
			config: &JWTAuthConfig{
				Mode: JWTAuthModeRequired,
			},
			wantErr: true,
		},
		{
			name:    "valid disabled config",
			config:  NewDefaultJWTAuthConfig(),
			wantErr: false,
		},
		{
			name:    "valid required config",
			config:  NewKMSJWTAuthConfig("https://example.com/.well-known/jwks.json", "https://auth.example.com", cryptoutilSharedMagic.KMSServiceName),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a minimal builder to test WithJWTAuth
			// We can't use NewServerBuilder since it requires full config, so we'll test error accumulation.
			builder := &ServerBuilder{
				ctx: context.Background(),
			}

			result := builder.WithJWTAuth(tt.config)

			if tt.wantErr {
				require.Error(t, result.err)
			} else {
				require.NoError(t, result.err)
				require.NotNil(t, result.jwtAuthConfig)
			}
		})
	}
}

func TestWithJWTAuth_ErrorAccumulation(t *testing.T) {
	t.Parallel()

	// Builder with existing error should preserve it.
	builder := &ServerBuilder{
		err: fmt.Errorf("previous error"),
	}

	result := builder.WithJWTAuth(NewDefaultJWTAuthConfig())

	require.Error(t, result.err)
	require.Contains(t, result.err.Error(), "previous error")
}
