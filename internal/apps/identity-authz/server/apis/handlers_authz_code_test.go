// Copyright (c) 2025-2026 Justin Cranford.
package apis_test

import (
	"context"
	json "encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthzPkce "cryptoutil/internal/apps/identity-authz/pkce"
	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity-authz/server/apis"
	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestHandleAuthorizationCodeGrant_AdditionalErrorPaths tests remaining error paths
// to improve handleAuthorizationCodeGrant coverage from 68.3% to target.
func TestHandleAuthorizationCodeGrant_AdditionalErrorPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name             string
		setupFunc        func(*testing.T, *cryptoutilIdentityRepository.RepositoryFactory) (code, redirectURI, clientID, codeVerifier string)
		tokenSvc         *authzCodeTokenServiceStub
		expectedStatus   int
		wantError        string
		wantDescription  string
		wantAccessCalls  int
		wantRefreshCalls int
	}{
		{
			name: "redirect_uri_mismatch",
			setupFunc: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) (string, string, string, string) {
				t.Helper()

				// Create client.
				clientRepo := repoFactory.ClientRepository()
				clientUUID, err := googleUuid.NewV7()
				require.NoError(t, err)

				client := &cryptoutilIdentityDomain.Client{
					ClientID:             "test-client-" + clientUUID.String(),
					Name:                 cryptoutilSharedMagic.TestClientName,
					RedirectURIs:         []string{cryptoutilSharedMagic.TestRedirectURI},
					AllowedScopes:        []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile},
					AccessTokenLifetime:  cryptoutilSharedMagic.IMDefaultSessionTimeout,
					RefreshTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax,
				}
				require.NoError(t, clientRepo.Create(ctx, client))

				// Create authorization request with different redirect URI.
				authzReqRepo := repoFactory.AuthorizationRequestRepository()
				codeVerifier, err := cryptoutilIdentityAuthzPkce.GenerateCodeVerifier()
				require.NoError(t, err)
				codeID, err := googleUuid.NewV7()
				require.NoError(t, err)
				userID, err := googleUuid.NewV7()
				require.NoError(t, err)

				codeChallenge := cryptoutilIdentityAuthzPkce.GenerateCodeChallenge(codeVerifier, cryptoutilSharedMagic.PKCEMethodS256)
				authRequest := &cryptoutilIdentityDomain.AuthorizationRequest{
					ClientID:            client.ClientID,
					RedirectURI:         cryptoutilSharedMagic.TestRedirectURI,
					Code:                codeID.String(),
					Scope:               testOpenidProfile,
					State:               testTestState,
					CodeChallenge:       codeChallenge,
					CodeChallengeMethod: cryptoutilSharedMagic.PKCEMethodS256,
					ExpiresAt:           time.Now().UTC().Add(cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Minute),
					UserID:              cryptoutilIdentityDomain.NullableUUID{UUID: userID, Valid: true},
				}
				require.NoError(t, authzReqRepo.Create(ctx, authRequest))

				// Return mismatched redirect URI.
				return authRequest.Code, "https://different.com/callback", client.ClientID, codeVerifier
			},
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			name: "client_not_found_after_code_validation",
			setupFunc: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) (string, string, string, string) {
				t.Helper()

				// Create authorization request WITHOUT creating client in database.
				authzReqRepo := repoFactory.AuthorizationRequestRepository()
				codeVerifier, err := cryptoutilIdentityAuthzPkce.GenerateCodeVerifier()
				require.NoError(t, err)
				codeID, err := googleUuid.NewV7()
				require.NoError(t, err)
				userID, err := googleUuid.NewV7()
				require.NoError(t, err)
				nonExistentClientUUID, err := googleUuid.NewV7()
				require.NoError(t, err)

				codeChallenge := cryptoutilIdentityAuthzPkce.GenerateCodeChallenge(codeVerifier, cryptoutilSharedMagic.PKCEMethodS256)
				nonExistentClientID := "non-existent-client-" + nonExistentClientUUID.String()
				authRequest := &cryptoutilIdentityDomain.AuthorizationRequest{
					ClientID:            nonExistentClientID,
					RedirectURI:         cryptoutilSharedMagic.TestRedirectURI,
					Code:                codeID.String(),
					Scope:               testOpenidProfile,
					State:               testTestState,
					CodeChallenge:       codeChallenge,
					CodeChallengeMethod: cryptoutilSharedMagic.PKCEMethodS256,
					ExpiresAt:           time.Now().UTC().Add(cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Minute),
					UserID:              cryptoutilIdentityDomain.NullableUUID{UUID: userID, Valid: true},
				}
				require.NoError(t, authzReqRepo.Create(ctx, authRequest))

				// Client lookup will fail after authorization code validation.
				return authRequest.Code, authRequest.RedirectURI, nonExistentClientID, codeVerifier
			},
			expectedStatus: fiber.StatusNotFound,
		},
		{
			name: testAccessTokenIssuanceFailed,
			setupFunc: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) (string, string, string, string) {
				return createValidAuthorizationCodeGrant(t, repoFactory)
			},
			tokenSvc:         &authzCodeTokenServiceStub{accessTokenErr: fmt.Errorf("injected access token error")},
			expectedStatus:   fiber.StatusInternalServerError,
			wantError:        cryptoutilSharedMagic.ErrorServerError,
			wantDescription:  cryptoutilIdentityAppErr.ErrTokenIssuanceFailed.Message,
			wantAccessCalls:  1,
			wantRefreshCalls: 0,
		},
		{
			name: "refresh_token_issuance_failed",
			setupFunc: func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) (string, string, string, string) {
				return createValidAuthorizationCodeGrant(t, repoFactory)
			},
			tokenSvc:         &authzCodeTokenServiceStub{refreshTokenErr: fmt.Errorf("injected refresh token error")},
			expectedStatus:   fiber.StatusInternalServerError,
			wantError:        cryptoutilSharedMagic.ErrorServerError,
			wantDescription:  cryptoutilIdentityAppErr.ErrTokenIssuanceFailed.Message,
			wantAccessCalls:  1,
			wantRefreshCalls: 1,
		},
	}

	for _, tc := range tests {
		// Capture range variable.
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup test dependencies.
			config, repoFactory, tokenSvc := createAuthzCodeAdditionalTestDependencies(t)

			var serviceTokenSvc cryptoutilIdentityAuthz.TokenService = tokenSvc

			// Setup test scenario.
			code, redirectURI, clientID, codeVerifier := tc.setupFunc(t, repoFactory)
			if tc.tokenSvc != nil {
				serviceTokenSvc = tc.tokenSvc
			}

			// Create service.
			svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, serviceTokenSvc)
			require.NotNil(t, svc)

			// Create Fiber app and mount routes.
			app := fiber.New()
			svc.RegisterRoutes(app)

			// Make POST request to /oauth2/v1/token.
			form := url.Values{}
			form.Set(cryptoutilSharedMagic.ParamGrantType, cryptoutilSharedMagic.GrantTypeAuthorizationCode)
			form.Set(cryptoutilSharedMagic.ParamCode, code)
			form.Set(cryptoutilSharedMagic.ParamRedirectURI, redirectURI)
			form.Set(cryptoutilSharedMagic.ClaimClientID, clientID)
			form.Set(cryptoutilSharedMagic.ParamCodeVerifier, codeVerifier)

			req := httptest.NewRequest(testPost, testOauth2V1Token, strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			responseBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			defer func() { //nolint:errcheck // Test cleanup - error intentionally ignored
				_ = resp.Body.Close()
			}()

			require.Equal(t, tc.expectedStatus, resp.StatusCode, "response body: %s", responseBody)

			if tc.wantError != "" {
				var errorResponse struct {
					Error            string `json:"error"`
					ErrorDescription string `json:"error_description"`
				}
				require.NoError(t, json.Unmarshal(responseBody, &errorResponse))
				require.Equal(t, tc.wantError, errorResponse.Error)
				require.Equal(t, tc.wantDescription, errorResponse.ErrorDescription)
			}

			if tc.tokenSvc != nil {
				require.Equal(t, tc.wantAccessCalls, tc.tokenSvc.accessTokenCalls)
				require.Equal(t, tc.wantRefreshCalls, tc.tokenSvc.refreshTokenCalls)
			}
		})
	}
}

type authzCodeTokenServiceStub struct {
	accessTokenErr    error
	refreshTokenErr   error
	accessTokenCalls  int
	refreshTokenCalls int
}

func (s *authzCodeTokenServiceStub) IssueAccessToken(_ context.Context, _ map[string]any) (string, error) {
	s.accessTokenCalls++

	if s.accessTokenErr != nil {
		return "", s.accessTokenErr
	}

	return testAccessToken, nil
}

func (s *authzCodeTokenServiceStub) IssueRefreshToken(_ context.Context) (string, error) {
	s.refreshTokenCalls++

	if s.refreshTokenErr != nil {
		return "", s.refreshTokenErr
	}

	return testRefreshToken, nil
}

func (*authzCodeTokenServiceStub) ValidateAccessToken(_ context.Context, _ string) (map[string]any, error) {
	return nil, nil
}

func (*authzCodeTokenServiceStub) GetPublicKeys() []map[string]any {
	return nil
}

func createValidAuthorizationCodeGrant(
	t *testing.T,
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory,
) (code, redirectURI, clientID, codeVerifier string) {
	t.Helper()

	ctx := context.Background()
	clientUUID, err := googleUuid.NewV7()
	require.NoError(t, err)

	clientID = "test-client-" + clientUUID.String()
	client := &cryptoutilIdentityDomain.Client{
		ClientID:             clientID,
		Name:                 cryptoutilSharedMagic.TestClientName,
		RedirectURIs:         []string{cryptoutilSharedMagic.TestRedirectURI},
		AllowedScopes:        []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile},
		AccessTokenLifetime:  cryptoutilSharedMagic.IMDefaultSessionTimeout,
		RefreshTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax,
	}
	require.NoError(t, repoFactory.ClientRepository().Create(ctx, client))

	codeVerifier, err = cryptoutilIdentityAuthzPkce.GenerateCodeVerifier()
	require.NoError(t, err)
	codeID, err := googleUuid.NewV7()
	require.NoError(t, err)
	userID, err := googleUuid.NewV7()
	require.NoError(t, err)

	authRequest := &cryptoutilIdentityDomain.AuthorizationRequest{
		ClientID:            clientID,
		RedirectURI:         cryptoutilSharedMagic.TestRedirectURI,
		Code:                codeID.String(),
		Scope:               testOpenidProfile,
		State:               testTestState,
		CodeChallenge:       cryptoutilIdentityAuthzPkce.GenerateCodeChallenge(codeVerifier, cryptoutilSharedMagic.PKCEMethodS256),
		CodeChallengeMethod: cryptoutilSharedMagic.PKCEMethodS256,
		ExpiresAt:           time.Now().UTC().Add(cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Minute),
		UserID:              cryptoutilIdentityDomain.NullableUUID{UUID: userID, Valid: true},
	}
	require.NoError(t, repoFactory.AuthorizationRequestRepository().Create(ctx, authRequest))

	return authRequest.Code, authRequest.RedirectURI, clientID, codeVerifier
}

func createAuthzCodeAdditionalTestDependencies(t *testing.T) (*cryptoutilIdentityConfig.Config, *cryptoutilIdentityRepository.RepositoryFactory, *cryptoutilIdentityIssuer.TokenService) {
	t.Helper()

	ctx := context.Background()
	databaseUUID, err := googleUuid.NewV7()
	require.NoError(t, err)

	// Create unique in-memory database.
	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:        cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:         fmt.Sprintf("file::memory:?cache=private&mode=memory&_id=%s", databaseUUID.String()),
		AutoMigrate: true,
	}

	config := &cryptoutilIdentityConfig.Config{
		Database: dbConfig,
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer:               testHTTPSLocalhost8080,
			AccessTokenLifetime:  cryptoutilSharedMagic.IMDefaultSessionTimeout,
			RefreshTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax,
		},
	}

	// Create repository factory and run migrations.
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)
	require.NoError(t, repoFactory.AutoMigrate(ctx))

	// Create real tokenSvc (nil issuers OK for error path tests that don't reach token issuance).
	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(nil, nil, nil, config.Tokens)

	return config, repoFactory, tokenSvc
}
