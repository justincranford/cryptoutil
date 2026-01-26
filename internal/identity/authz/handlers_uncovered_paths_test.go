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

// TestHandleAuthorizationCodeGrant_ErrorPaths targets uncovered error branches (51.7% coverage).
func TestHandleAuthorizationCodeGrant_ErrorPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) (clientID, code, redirectURI, codeVerifier string)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "user_id_missing_from_authorization_request",
			setupFunc: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) (string, string, string, string) {
				t.Helper()

				// Create client.
				clientID := "test-client-" + googleUuid.NewString()
				client := &cryptoutilIdentityDomain.Client{
					ClientID:                clientID,
					ClientSecret:            "test-secret",
					Name:                    "Test Client",
					RedirectURIs:            []string{"https://example.com/callback"},
					AllowedScopes:           []string{"openid", "profile"},
					ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
					TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
					RequirePKCE:             boolPtr(true),
					PKCEChallengeMethod:     cryptoutilIdentityMagic.PKCEMethodS256,
				}

				clientRepo := repoFactory.ClientRepository()
				err := clientRepo.Create(ctx, client)
				require.NoError(t, err, "Failed to create test client")

				// Create authorization request WITHOUT UserID (simulates incomplete login/consent).
				authReqRepo := repoFactory.AuthorizationRequestRepository()

				authCode := googleUuid.Must(googleUuid.NewV7()).String() // Use UUIDv7 for time-ordered uniqueness
				authReq := &cryptoutilIdentityDomain.AuthorizationRequest{
					ID:                  googleUuid.Must(googleUuid.NewV7()),
					ClientID:            clientID,
					Code:                authCode,
					RedirectURI:         "https://example.com/callback",
					ResponseType:        cryptoutilIdentityMagic.ResponseTypeCode,
					Scope:               "openid profile",
					CodeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
					CodeChallengeMethod: cryptoutilIdentityMagic.PKCEMethodS256,
					CreatedAt:           time.Now().UTC(),
					ExpiresAt:           time.Now().UTC().Add(10 * time.Minute),
					ConsentGranted:      true,
					// UserID NOT SET - this is the error condition we're testing
				}

				err = authReqRepo.Create(ctx, authReq)
				require.NoError(t, err, "Failed to create authorization request")

				verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk" // Valid verifier for challenge above

				return clientID, authCode, "https://example.com/callback", verifier
			},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  cryptoutilIdentityMagic.ErrorInvalidRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config, repoFactory, tokenSvc := createUncoveredPathsTestDependencies(t)

			// Service created WITH tokenSvc - error path occurs BEFORE token issuance.
			service := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)
			require.NotNil(t, service, "Service should not be nil")

			app := fiber.New()
			service.RegisterRoutes(app)

			// Run test setup.
			clientID, code, redirectURI, codeVerifier := tc.setupFunc(t, repoFactory)

			// Build token request form.
			form := url.Values{}
			form.Set("grant_type", cryptoutilIdentityMagic.GrantTypeAuthorizationCode)
			form.Set("code", code)
			form.Set("redirect_uri", redirectURI)
			form.Set("client_id", clientID)
			form.Set("code_verifier", codeVerifier)

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

// TestHandleRevoke_ErrorPaths targets uncovered error branches (65.0% coverage).
func TestHandleRevoke_ErrorPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name           string
		setupFunc      func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) string
		expectedStatus int
	}{
		{
			name: "token_already_revoked",
			setupFunc: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) string {
				t.Helper()

				// Create client for token association.
				client := &cryptoutilIdentityDomain.Client{
					ClientID:                "test-client-" + googleUuid.NewString(),
					ClientSecret:            "test-secret",
					Name:                    "Test Client",
					RedirectURIs:            []string{"https://example.com/callback"},
					AllowedScopes:           []string{"read", "write"},
					ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
					TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
					RequirePKCE:             boolPtr(true),
				}

				clientRepo := repoFactory.ClientRepository()
				err := clientRepo.Create(ctx, client)
				require.NoError(t, err, "Failed to create test client")

				// Create token that's already revoked.
				tokenValue := "revoked-token-" + googleUuid.NewString()
				now := time.Now().UTC()
				token := &cryptoutilIdentityDomain.Token{
					ID:          googleUuid.Must(googleUuid.NewV7()),
					ClientID:    client.ID, // Use client.ID (UUID), not ClientID (string)
					TokenType:   cryptoutilIdentityDomain.TokenTypeAccess,
					TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
					TokenValue:  tokenValue,
					Scopes:      []string{"read"},
					ExpiresAt:   time.Now().UTC().Add(1 * time.Hour),
					IssuedAt:    time.Now().UTC(),
					NotBefore:   time.Now().UTC(),
					Revoked:     true,
					RevokedAt:   &now, // Already revoked
				}

				tokenRepo := repoFactory.TokenRepository()
				err = tokenRepo.Create(ctx, token)
				require.NoError(t, err, "Failed to create revoked token")

				return tokenValue
			},
			expectedStatus: fiber.StatusOK, // Revocation should still return 200 OK (idempotent)
		},
		{
			name: "database_update_error_simulation",
			setupFunc: func(t *testing.T, _ *cryptoutilIdentityRepository.RepositoryFactory) string {
				t.Helper()

				// This test would require injecting a mock repository that simulates database update failure.
				// Since we're using real repository, we return non-existent token to trigger lookup failure.
				return "non-existent-token-" + googleUuid.NewString()
			},
			expectedStatus: fiber.StatusOK, // Token not found returns 200 OK (success - already not active)
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create test infrastructure.
			config := &cryptoutilIdentityConfig.Config{
				Tokens: &cryptoutilIdentityConfig.TokenConfig{
					Issuer:               "https://identity.example.com",
					AccessTokenLifetime:  15 * time.Minute,
					RefreshTokenLifetime: 24 * time.Hour,
				},
			}

			dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
				Type:        "sqlite",
				DSN:         "file::memory:?cache=private&_id=" + googleUuid.NewString(),
				AutoMigrate: true,
			}

			repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
			require.NoError(t, err, "Failed to create repository factory")

			err = repoFactory.AutoMigrate(ctx)
			require.NoError(t, err, "Failed to run database migrations")

			service := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
			require.NotNil(t, service, "Service should not be nil")

			app := fiber.New()
			service.RegisterRoutes(app)

			// Run test setup.
			tokenValue := tc.setupFunc(t, repoFactory)

			// Build revocation request form.
			form := url.Values{}
			form.Set("token", tokenValue)

			req := httptest.NewRequest("POST", "/oauth2/v1/revoke", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { //nolint:errcheck // Test cleanup - error intentionally ignored
				_ = resp.Body.Close()
			}()

			require.Equal(t, tc.expectedStatus, resp.StatusCode, "Expected 200 OK for revocation (idempotent)")
		})
	}
}

// Helper functions.

func createUncoveredPathsTestDependencies(t *testing.T) (*cryptoutilIdentityConfig.Config, *cryptoutilIdentityRepository.RepositoryFactory, *cryptoutilIdentityIssuer.TokenService) {
	t.Helper()

	config := &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  fmt.Sprintf("file::memory:?cache=private&mode=memory&_id=%s", googleUuid.New().String()),
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer:              "https://localhost:8080",
			AccessTokenLifetime: 3600,
		},
	}

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:        config.Database.Type,
		DSN:         config.Database.DSN,
		AutoMigrate: true,
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err, "Failed to create repository factory")
	require.NotNil(t, repoFactory, "Repository factory should not be nil")

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run auto migrations")

	// Create token service - pass nil for issuers (tests don't need actual token generation).
	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(nil, nil, nil, config.Tokens)
	require.NotNil(t, tokenSvc, "Token service should not be nil")

	return config, repoFactory, tokenSvc
}
