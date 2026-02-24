// Copyright (c) 2025 Justin Cranford

package authz_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestHandleRevoke_AdditionalErrorPaths tests remaining uncovered error paths in handleRevoke (90.0% coverage after initial tests).
func TestHandleRevoke_AdditionalErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) (string, string)
		expectedStatus int
	}{
		{
			name: "token_type_hint_mismatch_access_token",
			setupFunc: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) (string, string) {
				t.Helper()

				ctx := context.Background()

				// Create client first (foreign key constraint).
				client := &cryptoutilIdentityDomain.Client{
					ClientID:                "test-client-" + googleUuid.NewString(),
					ClientSecret:            "test-secret",
					Name:                    "Test Client",
					AllowedScopes:           []string{cryptoutilSharedMagic.ScopeRead, cryptoutilSharedMagic.ScopeWrite},
					ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
					TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
				}

				clientRepo := repoFactory.ClientRepository()
				err := clientRepo.Create(ctx, client)
				require.NoError(t, err, "Failed to create client")

				// Create refresh token but provide access_token hint.
				tokenRepo := repoFactory.TokenRepository()

				refreshToken := &cryptoutilIdentityDomain.Token{
					ID:          googleUuid.Must(googleUuid.NewV7()),
					TokenValue:  "refresh-token-" + googleUuid.NewString(),
					TokenType:   cryptoutilIdentityDomain.TokenTypeRefresh,
					TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
					ClientID:    client.ID,
					Scopes:      []string{cryptoutilSharedMagic.ScopeRead},
					ExpiresAt:   time.Now().UTC().Add(time.Hour),
					IssuedAt:    time.Now().UTC(),
					NotBefore:   time.Now().UTC(),
				}

				err = tokenRepo.Create(ctx, refreshToken)
				require.NoError(t, err, "Failed to create refresh token")

				return refreshToken.TokenValue, cryptoutilSharedMagic.TokenTypeAccessToken
			},
			expectedStatus: fiber.StatusBadRequest, // 400 - hint mismatch
		},
		{
			name: "token_type_hint_mismatch_refresh_token",
			setupFunc: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) (string, string) {
				t.Helper()

				ctx := context.Background()

				// Create client first (foreign key constraint).
				client := &cryptoutilIdentityDomain.Client{
					ClientID:                "test-client-" + googleUuid.NewString(),
					ClientSecret:            "test-secret",
					Name:                    "Test Client",
					AllowedScopes:           []string{cryptoutilSharedMagic.ScopeRead, cryptoutilSharedMagic.ScopeWrite},
					ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
					TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
				}

				clientRepo := repoFactory.ClientRepository()
				err := clientRepo.Create(ctx, client)
				require.NoError(t, err, "Failed to create client")

				// Create access token but provide refresh_token hint.
				tokenRepo := repoFactory.TokenRepository()

				accessToken := &cryptoutilIdentityDomain.Token{
					ID:          googleUuid.Must(googleUuid.NewV7()),
					TokenValue:  "access-token-" + googleUuid.NewString(),
					TokenType:   cryptoutilIdentityDomain.TokenTypeAccess,
					TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
					ClientID:    client.ID,
					Scopes:      []string{cryptoutilSharedMagic.ScopeRead},
					ExpiresAt:   time.Now().UTC().Add(time.Hour),
					IssuedAt:    time.Now().UTC(),
					NotBefore:   time.Now().UTC(),
				}

				err = tokenRepo.Create(ctx, accessToken)
				require.NoError(t, err, "Failed to create access token")

				return accessToken.TokenValue, cryptoutilSharedMagic.TokenTypeRefreshToken
			},
			expectedStatus: fiber.StatusBadRequest, // 400 - hint mismatch
		},
	}

	for _, tc := range tests {
		// Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup dependencies.
			config, repoFactory, tokenSvc := createRevokeErrorPathTestDependencies(t)

			// Create service and register routes.
			service := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)
			app := fiber.New()
			service.RegisterRoutes(app)

			// Run test setup.
			tokenValue, tokenTypeHint := tc.setupFunc(t, repoFactory)

			// Build revocation request form.
			form := url.Values{}
			form.Set(cryptoutilSharedMagic.ParamToken, tokenValue)
			form.Set(cryptoutilSharedMagic.ParamTokenTypeHint, tokenTypeHint)

			req := httptest.NewRequest("POST", "/oauth2/v1/revoke", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { //nolint:errcheck // Test cleanup - error intentionally ignored
				_ = resp.Body.Close()
			}()

			require.Equal(t, tc.expectedStatus, resp.StatusCode, "Should return expected status code for token_type_hint mismatch")
		})
	}
}

// Helper functions.

func createRevokeErrorPathTestDependencies(t *testing.T) (*cryptoutilIdentityConfig.Config, *cryptoutilIdentityRepository.RepositoryFactory, *cryptoutilIdentityIssuer.TokenService) {
	t.Helper()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:        "sqlite",
		DSN:         fmt.Sprintf("file::memory:?cache=private&mode=memory&_id=%s", googleUuid.New().String()),
		AutoMigrate: true,
	}

	config := &cryptoutilIdentityConfig.Config{
		Database: dbConfig,
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer:              "https://localhost:8080",
			AccessTokenLifetime: 3600,
		},
	}

	ctx := context.Background()

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, config.Database)
	require.NoError(t, err, "Failed to create repository factory")
	require.NotNil(t, repoFactory, "Repository factory should not be nil")

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run auto migrations")

	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(nil, nil, nil, config.Tokens)
	require.NotNil(t, tokenSvc, "Token service should not be nil")

	return config, repoFactory, tokenSvc
}
