// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// TestHandleRefreshTokenGrant_ErrorPaths tests error paths for handleRefreshTokenGrant (80.0% → 90%).
func TestHandleRefreshTokenGrant_ErrorPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) (string, string, string)
		expectedStatus int
	}{
		{
			name: "missing_refresh_token",
			setupFunc: func(t *testing.T, _ *cryptoutilIdentityRepository.RepositoryFactory) (string, string, string) {
				t.Helper()

				// No token provided.
				return "", cryptoutilIdentityMagic.TestClientID, cryptoutilIdentityMagic.ScopeRead
			},
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			name: "missing_client_id",
			setupFunc: func(t *testing.T, _ *cryptoutilIdentityRepository.RepositoryFactory) (string, string, string) {
				t.Helper()

				// Token but no client ID.
				return "test-refresh-token", "", "read"
			},
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			name: "refresh_token_not_found",
			setupFunc: func(t *testing.T, _ *cryptoutilIdentityRepository.RepositoryFactory) (string, string, string) {
				t.Helper()

				// Token does not exist.
				return "non-existent-refresh-token-" + googleUuid.NewString(), cryptoutilIdentityMagic.TestClientID, cryptoutilIdentityMagic.ScopeRead
			},
			expectedStatus: fiber.StatusNotFound, // Token lookup returns 404
		},
		{
			name: "token_type_not_refresh",
			setupFunc: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) (string, string, string) {
				t.Helper()

				// Create client.
				client := &cryptoutilIdentityDomain.Client{
					ClientID:                "test-client-" + googleUuid.NewString(),
					ClientSecret:            "test-secret",
					Name:                    "Test Client",
					AllowedScopes:           []string{"read", "write"},
					ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
					TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
				}

				clientRepo := repoFactory.ClientRepository()
				err := clientRepo.Create(ctx, client)
				require.NoError(t, err, "Failed to create client")

				// Create ACCESS token (not refresh).
				tokenValue := "access-token-" + googleUuid.NewString()
				token := &cryptoutilIdentityDomain.Token{
					ID:          googleUuid.Must(googleUuid.NewV7()),
					ClientID:    client.ID,
					TokenType:   cryptoutilIdentityDomain.TokenTypeAccess, // WRONG type
					TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
					TokenValue:  tokenValue,
					Scopes:      []string{cryptoutilIdentityMagic.ScopeRead},
					ExpiresAt:   time.Now().Add(1 * time.Hour),
					IssuedAt:    time.Now(),
					NotBefore:   time.Now(),
				}

				tokenRepo := repoFactory.TokenRepository()
				err = tokenRepo.Create(ctx, token)
				require.NoError(t, err, "Failed to create access token")

				return tokenValue, client.ClientID, cryptoutilIdentityMagic.ScopeRead
			},
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			name: "refresh_token_revoked",
			setupFunc: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) (string, string, string) {
				t.Helper()

				// Create client.
				client := &cryptoutilIdentityDomain.Client{
					ClientID:                "test-client-" + googleUuid.NewString(),
					ClientSecret:            "test-secret",
					Name:                    "Test Client",
					AllowedScopes:           []string{"read", "write"},
					ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
					TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
				}

				clientRepo := repoFactory.ClientRepository()
				err := clientRepo.Create(ctx, client)
				require.NoError(t, err, "Failed to create client")

				// Create REVOKED refresh token.
				tokenValue := "refresh-token-" + googleUuid.NewString()
				now := time.Now()
				token := &cryptoutilIdentityDomain.Token{
					ID:          googleUuid.Must(googleUuid.NewV7()),
					ClientID:    client.ID,
					TokenType:   cryptoutilIdentityDomain.TokenTypeRefresh,
					TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
					TokenValue:  tokenValue,
					Scopes:      []string{cryptoutilIdentityMagic.ScopeRead},
					ExpiresAt:   time.Now().Add(24 * time.Hour),
					IssuedAt:    time.Now(),
					NotBefore:   time.Now(),
					Revoked:     true,
					RevokedAt:   &now,
				}

				tokenRepo := repoFactory.TokenRepository()
				err = tokenRepo.Create(ctx, token)
				require.NoError(t, err, "Failed to create revoked refresh token")

				return tokenValue, client.ClientID, cryptoutilIdentityMagic.ScopeRead
			},
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			name: "refresh_token_expired",
			setupFunc: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) (string, string, string) {
				t.Helper()

				// Create client.
				client := &cryptoutilIdentityDomain.Client{
					ClientID:                "test-client-" + googleUuid.NewString(),
					ClientSecret:            "test-secret",
					Name:                    "Test Client",
					AllowedScopes:           []string{"read", "write"},
					ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
					TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
				}

				clientRepo := repoFactory.ClientRepository()
				err := clientRepo.Create(ctx, client)
				require.NoError(t, err, "Failed to create client")

				// Create EXPIRED refresh token.
				tokenValue := "refresh-token-" + googleUuid.NewString()
				token := &cryptoutilIdentityDomain.Token{
					ID:          googleUuid.Must(googleUuid.NewV7()),
					ClientID:    client.ID,
					TokenType:   cryptoutilIdentityDomain.TokenTypeRefresh,
					TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
					TokenValue:  tokenValue,
					Scopes:      []string{cryptoutilIdentityMagic.ScopeRead},
					ExpiresAt:   time.Now().Add(-1 * time.Hour), // Already expired
					IssuedAt:    time.Now().Add(-25 * time.Hour),
					NotBefore:   time.Now().Add(-25 * time.Hour),
				}

				tokenRepo := repoFactory.TokenRepository()
				err = tokenRepo.Create(ctx, token)
				require.NoError(t, err, "Failed to create expired refresh token")

				return tokenValue, client.ClientID, cryptoutilIdentityMagic.ScopeRead
			},
			expectedStatus: fiber.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config, repoFactory, tokenSvc := createRefreshTokenTestDependencies(t)

			service := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)
			require.NotNil(t, service, "Service should not be nil")

			app := fiber.New()
			service.RegisterRoutes(app)

			refreshToken, clientID, scope := tc.setupFunc(t, repoFactory)

			form := url.Values{}
			form.Set("grant_type", cryptoutilIdentityMagic.GrantTypeRefreshToken)

			if refreshToken != "" {
				form.Set("refresh_token", refreshToken)
			}

			if clientID != "" {
				form.Set("client_id", clientID)
			}

			if scope != "" {
				form.Set("scope", scope)
			}

			req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { //nolint:errcheck // Test cleanup - error intentionally ignored
				_ = resp.Body.Close()
			}()

			require.Equal(t, tc.expectedStatus, resp.StatusCode, "Expected specific HTTP status for error path")
		})
	}
}

// Helper functions.

func createRefreshTokenTestDependencies(t *testing.T) (*cryptoutilIdentityConfig.Config, *cryptoutilIdentityRepository.RepositoryFactory, *cryptoutilIdentityIssuer.TokenService) {
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

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err, "Failed to create repository factory")
	require.NotNil(t, repoFactory, "Repository factory should not be nil")

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run auto migrations")

	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(nil, nil, nil, config.Tokens)
	require.NotNil(t, tokenSvc, "Token service should not be nil")

	return config, repoFactory, tokenSvc
}
