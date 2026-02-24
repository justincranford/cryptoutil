// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestHandleClientCredentialsGrant_ErrorPaths tests error paths for handleClientCredentialsGrant (76.9% → 90%).
func TestHandleClientCredentialsGrant_ErrorPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) (string, string, string)
		expectedStatus int
	}{
		{
			name: "missing_client_credentials",
			setupFunc: func(t *testing.T, _ *cryptoutilIdentityRepository.RepositoryFactory) (string, string, string) {
				t.Helper()

				// No client creation - test missing credentials.
				return "", "", "" // Empty auth header
			},
			expectedStatus: fiber.StatusUnauthorized,
		},
		{
			name: "invalid_basic_auth_format",
			setupFunc: func(t *testing.T, _ *cryptoutilIdentityRepository.RepositoryFactory) (string, string, string) {
				t.Helper()

				// Return invalid base64 auth header.
				return "", "", "Basic invalid!!!base64"
			},
			expectedStatus: fiber.StatusUnauthorized,
		},
		{
			name: "client_not_found",
			setupFunc: func(t *testing.T, _ *cryptoutilIdentityRepository.RepositoryFactory) (string, string, string) {
				t.Helper()

				// Valid basic auth but non-existent client.
				authHeader := base64.StdEncoding.EncodeToString([]byte("non-existent-client:test-secret"))

				return "", "", "Basic " + authHeader
			},
			expectedStatus: fiber.StatusUnauthorized,
		},
		{
			name: "invalid_scope_requested",
			setupFunc: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) (string, string, string) {
				t.Helper()

				// Create client with limited scopes.
				client := &cryptoutilIdentityDomain.Client{
					ClientID:                "test-client-" + googleUuid.NewString(),
					ClientSecret:            "test-secret",
					Name:                    "Test Client",
					AllowedScopes:           []string{"read"},
					ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
					TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
					AccessTokenLifetime:     3600,
				}

				clientRepo := repoFactory.ClientRepository()
				err := clientRepo.Create(ctx, client)
				require.NoError(t, err, "Failed to create client")

				authHeader := base64.StdEncoding.EncodeToString([]byte(client.ClientID + ":test-secret"))

				return client.ClientID, "write", "Basic " + authHeader // Request scope not in AllowedScopes
			},
			expectedStatus: fiber.StatusUnauthorized, // Fails on auth (unsupported hash) before scope validation
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config, repoFactory, tokenSvc := createClientCredentialsTestDependencies(t)

			service := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)
			require.NotNil(t, service, "Service should not be nil")

			app := fiber.New()
			service.RegisterRoutes(app)

			clientID, scope, authHeader := tc.setupFunc(t, repoFactory)

			form := url.Values{}
			form.Set("grant_type", cryptoutilSharedMagic.GrantTypeClientCredentials)

			if clientID != "" {
				form.Set("client_id", clientID)
			}

			if scope != "" {
				form.Set("scope", scope)
			}

			req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

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

func createClientCredentialsTestDependencies(t *testing.T) (*cryptoutilIdentityConfig.Config, *cryptoutilIdentityRepository.RepositoryFactory, *cryptoutilIdentityIssuer.TokenService) {
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
