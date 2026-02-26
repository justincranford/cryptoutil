// Copyright (c) 2025 Justin Cranford
//
//

package issuer_test

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
)

func TestValidateAccessToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		tokenFormat string
		setupToken  func(*testing.T, *cryptoutilIdentityIssuer.TokenService) string
		wantErr     bool
	}{
		{
			name:        "valid_jws_token",
			tokenFormat: cryptoutilSharedMagic.DefaultBrowserSessionCookie,
			setupToken: func(t *testing.T, service *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				ctx := context.Background()
				claims := map[string]any{
					cryptoutilSharedMagic.ClaimSub:   googleUuid.Must(googleUuid.NewV7()).String(),
					cryptoutilSharedMagic.ClaimScope: "openid profile",
					cryptoutilSharedMagic.ClaimIat:   time.Now().UTC().Unix(),
					cryptoutilSharedMagic.ClaimExp:   time.Now().UTC().Add(1 * time.Hour).Unix(),
				}
				token, err := service.IssueAccessToken(ctx, claims)
				require.NoError(t, err)

				return token
			},
			wantErr: false,
		},
		{
			name:        "valid_jwe_token",
			tokenFormat: cryptoutilSharedMagic.IdentityTokenFormatJWE,
			setupToken: func(t *testing.T, service *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				ctx := context.Background()
				claims := map[string]any{
					cryptoutilSharedMagic.ClaimSub:   googleUuid.Must(googleUuid.NewV7()).String(),
					cryptoutilSharedMagic.ClaimScope: "openid profile",
					cryptoutilSharedMagic.ClaimIat:   time.Now().UTC().Unix(),
					cryptoutilSharedMagic.ClaimExp:   time.Now().UTC().Add(1 * time.Hour).Unix(),
				}
				token, err := service.IssueAccessToken(ctx, claims)
				require.NoError(t, err)

				return token
			},
			wantErr: false,
		},
		{
			name:        "valid_uuid_token",
			tokenFormat: cryptoutilSharedMagic.IdentityTokenFormatUUID,
			setupToken: func(t *testing.T, service *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				ctx := context.Background()
				claims := map[string]any{}
				token, err := service.IssueAccessToken(ctx, claims)
				require.NoError(t, err)

				return token
			},
			wantErr: false,
		},
		{
			name:        "invalid_jws_token",
			tokenFormat: cryptoutilSharedMagic.DefaultBrowserSessionCookie,
			setupToken: func(t *testing.T, _ *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				return "invalid.jws.token"
			},
			wantErr: true,
		},
		{
			name:        "invalid_jwe_token",
			tokenFormat: cryptoutilSharedMagic.IdentityTokenFormatJWE,
			setupToken: func(t *testing.T, _ *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				return "invalid_jwe_token"
			},
			wantErr: true,
		},
		{
			name:        "unsupported_format",
			tokenFormat: "jwt",
			setupToken: func(t *testing.T, _ *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				return "any-token"
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			service := setupTestService(t, tc.tokenFormat)

			token := tc.setupToken(t, service)

			claims, err := service.ValidateAccessToken(ctx, token)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, claims)
			}
		})
	}
}

// TestValidateIDToken validates ID token validation.
func TestValidateIDToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupToken func(*testing.T, *cryptoutilIdentityIssuer.TokenService) string
		wantErr    bool
	}{
		{
			name: "valid_id_token",
			setupToken: func(t *testing.T, service *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				ctx := context.Background()
				claims := map[string]any{
					cryptoutilSharedMagic.ClaimSub:   googleUuid.Must(googleUuid.NewV7()).String(),
					cryptoutilSharedMagic.ClaimAud:   googleUuid.Must(googleUuid.NewV7()).String(),
					cryptoutilSharedMagic.ClaimNonce: "test-nonce",
					cryptoutilSharedMagic.ClaimIat:   time.Now().UTC().Unix(),
					cryptoutilSharedMagic.ClaimExp:   time.Now().UTC().Add(1 * time.Hour).Unix(),
				}
				token, err := service.IssueIDToken(ctx, claims)
				require.NoError(t, err)

				return token
			},
			wantErr: false,
		},
		{
			name: "invalid_id_token",
			setupToken: func(t *testing.T, _ *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				return "invalid.id.token"
			},
			wantErr: true,
		},
		{
			name: "expired_id_token",
			setupToken: func(t *testing.T, service *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				ctx := context.Background()
				claims := map[string]any{
					cryptoutilSharedMagic.ClaimSub:   googleUuid.Must(googleUuid.NewV7()).String(),
					cryptoutilSharedMagic.ClaimAud:   googleUuid.Must(googleUuid.NewV7()).String(),
					cryptoutilSharedMagic.ClaimNonce: "test-nonce",
					cryptoutilSharedMagic.ClaimIat:   time.Now().UTC().Add(-2 * time.Hour).Unix(),
					cryptoutilSharedMagic.ClaimExp:   time.Now().UTC().Add(-1 * time.Hour).Unix(),
				}
				token, err := service.IssueIDToken(ctx, claims)
				require.NoError(t, err)

				return token
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			service := setupTestService(t, cryptoutilSharedMagic.DefaultBrowserSessionCookie)

			token := tc.setupToken(t, service)

			claims, err := service.ValidateIDToken(ctx, token)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, claims)
				require.Contains(t, claims, cryptoutilSharedMagic.ClaimSub)
			}
		})
	}
}

// TestIsTokenActive validates token expiration and not-before checks.
func TestIsTokenActive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		claims     map[string]any
		wantActive bool
	}{
		{
			name: "valid_active_token",
			claims: map[string]any{
				cryptoutilSharedMagic.ClaimExp: float64(time.Now().UTC().Add(1 * time.Hour).Unix()),
				cryptoutilSharedMagic.ClaimNbf: float64(time.Now().UTC().Add(-1 * time.Minute).Unix()),
			},
			wantActive: true,
		},
		{
			name: cryptoutilSharedMagic.ErrorExpiredToken,
			claims: map[string]any{
				cryptoutilSharedMagic.ClaimExp: float64(time.Now().UTC().Add(-1 * time.Hour).Unix()),
				cryptoutilSharedMagic.ClaimNbf: float64(time.Now().UTC().Add(-2 * time.Hour).Unix()),
			},
			wantActive: false,
		},
		{
			name: "not_yet_valid_token",
			claims: map[string]any{
				cryptoutilSharedMagic.ClaimExp: float64(time.Now().UTC().Add(2 * time.Hour).Unix()),
				cryptoutilSharedMagic.ClaimNbf: float64(time.Now().UTC().Add(1 * time.Hour).Unix()),
			},
			wantActive: false,
		},
		{
			name:       "no_expiration_or_nbf",
			claims:     map[string]any{},
			wantActive: true,
		},
		{
			name: "only_expiration_valid",
			claims: map[string]any{
				cryptoutilSharedMagic.ClaimExp: float64(time.Now().UTC().Add(1 * time.Hour).Unix()),
			},
			wantActive: true,
		},
		{
			name: "only_nbf_valid",
			claims: map[string]any{
				cryptoutilSharedMagic.ClaimNbf: float64(time.Now().UTC().Add(-1 * time.Minute).Unix()),
			},
			wantActive: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service := setupTestService(t, cryptoutilSharedMagic.DefaultBrowserSessionCookie)

			active := service.IsTokenActive(tc.claims)

			require.Equal(t, tc.wantActive, active)
		})
	}
}

// TestIntrospectToken validates token introspection.
func TestIntrospectToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupToken  func(*testing.T, *cryptoutilIdentityIssuer.TokenService) string
		wantActive  bool
		checkExpiry bool
	}{
		{
			name: "valid_active_token",
			setupToken: func(t *testing.T, service *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				ctx := context.Background()
				claims := map[string]any{
					cryptoutilSharedMagic.ClaimSub:   googleUuid.Must(googleUuid.NewV7()).String(),
					cryptoutilSharedMagic.ClaimScope: "openid profile",
					cryptoutilSharedMagic.ClaimIat:   time.Now().UTC().Unix(),
					cryptoutilSharedMagic.ClaimExp:   time.Now().UTC().Add(1 * time.Hour).Unix(),
				}
				token, err := service.IssueAccessToken(ctx, claims)
				require.NoError(t, err)

				return token
			},
			wantActive:  true,
			checkExpiry: true,
		},
		{
			name: cryptoutilSharedMagic.ErrorExpiredToken,
			setupToken: func(t *testing.T, service *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				ctx := context.Background()
				claims := map[string]any{
					cryptoutilSharedMagic.ClaimSub:   googleUuid.Must(googleUuid.NewV7()).String(),
					cryptoutilSharedMagic.ClaimScope: "openid profile",
					cryptoutilSharedMagic.ClaimIat:   time.Now().UTC().Add(-2 * time.Hour).Unix(),
					cryptoutilSharedMagic.ClaimExp:   time.Now().UTC().Add(-1 * time.Hour).Unix(),
				}
				token, err := service.IssueAccessToken(ctx, claims)
				require.NoError(t, err)

				return token
			},
			wantActive:  true,
			checkExpiry: true,
		},
		{
			name: cryptoutilSharedMagic.ErrorInvalidToken,
			setupToken: func(t *testing.T, _ *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				return "invalid.token.here"
			},
			wantActive:  false,
			checkExpiry: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			service := setupTestService(t, cryptoutilSharedMagic.DefaultBrowserSessionCookie)

			token := tc.setupToken(t, service)

			metadata, err := service.IntrospectToken(ctx, token)

			require.NoError(t, err)
			require.NotNil(t, metadata)
			require.Equal(t, tc.wantActive, metadata.Active)

			if tc.checkExpiry {
				require.NotNil(t, metadata.ExpiresAt)
				require.NotNil(t, metadata.Claims)
			}
		})
	}
}

// TestIssueUserInfoJWT validates UserInfo JWT issuance.
func TestIssueUserInfoJWT(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		clientID string
		claims   map[string]any
		wantErr  bool
	}{
		{
			name:     "valid_userinfo_jwt",
			clientID: googleUuid.Must(googleUuid.NewV7()).String(),
			claims: map[string]any{
				cryptoutilSharedMagic.ClaimSub:   googleUuid.Must(googleUuid.NewV7()).String(),
				cryptoutilSharedMagic.ClaimEmail: "user@example.com",
				cryptoutilSharedMagic.ClaimName:  "Test User",
			},
			wantErr: false,
		},
		{
			name:     "missing_sub_claim",
			clientID: googleUuid.Must(googleUuid.NewV7()).String(),
			claims: map[string]any{
				cryptoutilSharedMagic.ClaimEmail: "user@example.com",
				cryptoutilSharedMagic.ClaimName:  "Test User",
			},
			wantErr: true,
		},
		{
			name:     "empty_client_id",
			clientID: "",
			claims: map[string]any{
				cryptoutilSharedMagic.ClaimSub:   googleUuid.Must(googleUuid.NewV7()).String(),
				cryptoutilSharedMagic.ClaimEmail: "user@example.com",
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			service := setupTestService(t, cryptoutilSharedMagic.DefaultBrowserSessionCookie)

			jwt, err := service.IssueUserInfoJWT(ctx, tc.clientID, tc.claims)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, jwt)
			}
		})
	}
}
