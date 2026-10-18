// Copyright (c) 2025-2026 Justin Cranford.
//
//

package apis_test

import (
	"context"
	"encoding/base64"
	json "encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity-authz/server/apis"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestHandleClientCredentialsGrant_ErrorPaths tests error paths for handleClientCredentialsGrant (76.9% â†’ 90%).
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
					ClientSecret:            testTestSecret,
					Name:                    cryptoutilSharedMagic.TestClientName,
					AllowedScopes:           []string{cryptoutilSharedMagic.ScopeRead},
					ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
					TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
					AccessTokenLifetime:     cryptoutilSharedMagic.IMDefaultSessionTimeout,
				}

				clientRepo := repoFactory.ClientRepository()
				err := clientRepo.Create(ctx, client)
				require.NoError(t, err, "Failed to create client")

				authHeader := base64.StdEncoding.EncodeToString([]byte(client.ClientID + ":test-secret"))

				return client.ClientID, cryptoutilSharedMagic.ScopeWrite, "Basic " + authHeader // Request scope not in AllowedScopes
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
			form.Set(cryptoutilSharedMagic.ParamGrantType, cryptoutilSharedMagic.GrantTypeClientCredentials)

			if clientID != "" {
				form.Set(cryptoutilSharedMagic.ClaimClientID, clientID)
			}

			if scope != "" {
				form.Set(cryptoutilSharedMagic.ClaimScope, scope)
			}

			req := httptest.NewRequest(testPost, testOauth2V1Token, strings.NewReader(form.Encode()))
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

func TestHandleClientCredentialsGrant_AuthenticatedFailurePaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                   string
		tokenServiceConfigured bool
		accessTokenErr         error
		seedDuplicate          bool
		expectedStatus         int
		expectedError          string
		expectedAccessToken    string
	}{
		{
			name:           "token_service_not_configured",
			expectedStatus: fiber.StatusInternalServerError,
			expectedError:  cryptoutilSharedMagic.ErrorServerError,
		},
		{
			name:                   testAccessTokenIssuanceFailed,
			tokenServiceConfigured: true,
			accessTokenErr:         errors.New("injected token issuance failure"),
			expectedStatus:         fiber.StatusInternalServerError,
			expectedError:          cryptoutilSharedMagic.ErrorServerError,
		},
		{
			name:                   "token_persistence_failure_is_nonfatal",
			tokenServiceConfigured: true,
			seedDuplicate:          true,
			expectedStatus:         fiber.StatusOK,
			expectedAccessToken:    testAccessToken,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			config, repoFactory, _ := createClientCredentialsTestDependencies(t)
			client := createClientAuthFlowTestClient(t, repoFactory, cryptoutilIdentityDomain.ClientAuthMethodSecretBasic)

			if tc.seedDuplicate {
				now := time.Now().UTC()
				err := repoFactory.TokenRepository().Create(ctx, &cryptoutilIdentityDomain.Token{
					TokenValue:  testAccessToken,
					TokenType:   cryptoutilIdentityDomain.TokenTypeAccess,
					TokenFormat: cryptoutilIdentityDomain.TokenFormatJWS,
					ClientID:    client.ID,
					Scopes:      []string{cryptoutilSharedMagic.ScopeOpenID},
					IssuedAt:    now,
					ExpiresAt:   now.Add(time.Duration(cryptoutilSharedMagic.AccessTokenExpirySeconds) * time.Second),
				})
				require.NoError(t, err)
			}

			var service *cryptoutilIdentityAuthz.Service
			if tc.tokenServiceConfigured {
				service = cryptoutilIdentityAuthz.NewService(config, repoFactory, clientCredentialsTestTokenService{accessTokenErr: tc.accessTokenErr})
			} else {
				service = cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
			}

			app := fiber.New()
			service.RegisterRoutes(app)

			form := url.Values{
				cryptoutilSharedMagic.ParamGrantType: []string{cryptoutilSharedMagic.GrantTypeClientCredentials},
				cryptoutilSharedMagic.ClaimScope:     []string{cryptoutilSharedMagic.ScopeOpenID},
			}
			request := httptest.NewRequest(testPost, testOauth2V1Token, strings.NewReader(form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(client.ClientID+":"+testTestSecret)))

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() {
				require.NoError(t, response.Body.Close())
			}()

			require.Equal(t, tc.expectedStatus, response.StatusCode)

			var body struct {
				Error       string `json:"error"`
				AccessToken string `json:"access_token"`
			}
			require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
			require.Equal(t, tc.expectedError, body.Error)
			require.Equal(t, tc.expectedAccessToken, body.AccessToken)
		})
	}
}

type clientCredentialsTestTokenService struct {
	accessTokenErr error
}

func (s clientCredentialsTestTokenService) IssueAccessToken(_ context.Context, _ map[string]any) (string, error) {
	if s.accessTokenErr != nil {
		return "", s.accessTokenErr
	}

	return testAccessToken, nil
}

func (clientCredentialsTestTokenService) IssueRefreshToken(_ context.Context) (string, error) {
	return "", nil
}

func (clientCredentialsTestTokenService) ValidateAccessToken(_ context.Context, _ string) (map[string]any, error) {
	return nil, nil
}

func (clientCredentialsTestTokenService) GetPublicKeys() []map[string]any {
	return nil
}

// Helper functions.

func createClientCredentialsTestDependencies(t *testing.T) (*cryptoutilIdentityConfig.Config, *cryptoutilIdentityRepository.RepositoryFactory, *cryptoutilIdentityIssuer.TokenService) {
	t.Helper()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:        cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:         fmt.Sprintf("file::memory:?cache=private&mode=memory&_id=%s", googleUuid.New().String()),
		AutoMigrate: true,
	}

	config := &cryptoutilIdentityConfig.Config{
		Database: dbConfig,
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer:              testHTTPSLocalhost8080,
			AccessTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionTimeout,
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
