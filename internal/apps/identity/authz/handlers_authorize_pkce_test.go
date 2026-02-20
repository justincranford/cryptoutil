// Copyright (c) 2025 Justin Cranford

package authz_test

import (
	"context"
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
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

// TestHandleAuthorizeGET_PKCE validates PKCE parameter requirements for GET /authorize.
func TestHandleAuthorizeGET_PKCE(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		codeChallenge       string
		codeChallengeMethod string
		wantStatus          int
		wantErrorCode       string
	}{
		{
			name:                "missing code_challenge required",
			codeChallenge:       "",
			codeChallengeMethod: cryptoutilIdentityMagic.PKCEMethodS256,
			wantStatus:          fiber.StatusBadRequest,
			wantErrorCode:       cryptoutilIdentityMagic.ErrorInvalidRequest,
		},
		{
			name:                "valid S256 code_challenge",
			codeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
			codeChallengeMethod: cryptoutilIdentityMagic.PKCEMethodS256,
			wantStatus:          fiber.StatusFound,
			wantErrorCode:       "",
		},
		{
			name:                "default method is S256",
			codeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
			codeChallengeMethod: "",
			wantStatus:          fiber.StatusFound,
			wantErrorCode:       "",
		},
		{
			name:                "plain method rejected",
			codeChallenge:       "test-plain-challenge",
			codeChallengeMethod: cryptoutilIdentityMagic.PKCEMethodPlain,
			wantStatus:          fiber.StatusBadRequest,
			wantErrorCode:       cryptoutilIdentityMagic.ErrorInvalidRequest,
		},
		{
			name:                "invalid method rejected",
			codeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
			codeChallengeMethod: "invalid",
			wantStatus:          fiber.StatusBadRequest,
			wantErrorCode:       cryptoutilIdentityMagic.ErrorInvalidRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := createAuthorizePKCETestConfig(t)
			repoFactory := createAuthorizePKCETestRepoFactory(t)

			testClient := createAuthorizePKCETestClient(t, repoFactory)

			svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
			require.NotNil(t, svc, "Service should not be nil")

			app := fiber.New()
			svc.RegisterRoutes(app)

			query := url.Values{
				cryptoutilIdentityMagic.ParamClientID:     []string{testClient.ClientID},
				cryptoutilIdentityMagic.ParamResponseType: []string{cryptoutilIdentityMagic.ResponseTypeCode},
				cryptoutilIdentityMagic.ParamRedirectURI:  []string{testClient.RedirectURIs[0]},
				cryptoutilIdentityMagic.ParamScope:        []string{"openid profile"},
				cryptoutilIdentityMagic.ParamState:        []string{"test-state"},
			}

			if tc.codeChallenge != "" {
				query.Set(cryptoutilIdentityMagic.ParamCodeChallenge, tc.codeChallenge)
			}

			if tc.codeChallengeMethod != "" {
				query.Set(cryptoutilIdentityMagic.ParamCodeChallengeMethod, tc.codeChallengeMethod)
			}

			req := httptest.NewRequest("GET", "/oauth2/v1/authorize?"+query.Encode(), nil)

			resp, err := app.Test(req, -1)
			require.NoError(t, err, "Request should succeed")

			defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

			require.Equal(t, tc.wantStatus, resp.StatusCode, "Status code should match expected")
		})
	}
}

// TestHandleAuthorizePOST_PKCE validates PKCE parameter requirements for POST /authorize.
func TestHandleAuthorizePOST_PKCE(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		codeChallenge       string
		codeChallengeMethod string
		wantStatus          int
		wantErrorCode       string
	}{
		{
			name:                "missing code_challenge required",
			codeChallenge:       "",
			codeChallengeMethod: cryptoutilIdentityMagic.PKCEMethodS256,
			wantStatus:          fiber.StatusBadRequest,
			wantErrorCode:       cryptoutilIdentityMagic.ErrorInvalidRequest,
		},
		{
			name:                "valid S256 code_challenge",
			codeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
			codeChallengeMethod: cryptoutilIdentityMagic.PKCEMethodS256,
			wantStatus:          fiber.StatusFound,
			wantErrorCode:       "",
		},
		{
			name:                "default method is S256",
			codeChallenge:       "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
			codeChallengeMethod: "",
			wantStatus:          fiber.StatusFound,
			wantErrorCode:       "",
		},
		{
			name:                "plain method rejected",
			codeChallenge:       "test-plain-challenge",
			codeChallengeMethod: cryptoutilIdentityMagic.PKCEMethodPlain,
			wantStatus:          fiber.StatusBadRequest,
			wantErrorCode:       cryptoutilIdentityMagic.ErrorInvalidRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := createAuthorizePKCETestConfig(t)
			repoFactory := createAuthorizePKCETestRepoFactory(t)

			testClient := createAuthorizePKCETestClient(t, repoFactory)

			svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
			require.NotNil(t, svc, "Service should not be nil")

			app := fiber.New()
			svc.RegisterRoutes(app)

			formBody := url.Values{
				cryptoutilIdentityMagic.ParamClientID:     []string{testClient.ClientID},
				cryptoutilIdentityMagic.ParamResponseType: []string{cryptoutilIdentityMagic.ResponseTypeCode},
				cryptoutilIdentityMagic.ParamRedirectURI:  []string{testClient.RedirectURIs[0]},
				cryptoutilIdentityMagic.ParamScope:        []string{"openid profile"},
				cryptoutilIdentityMagic.ParamState:        []string{"test-state"},
			}

			if tc.codeChallenge != "" {
				formBody.Set(cryptoutilIdentityMagic.ParamCodeChallenge, tc.codeChallenge)
			}

			if tc.codeChallengeMethod != "" {
				formBody.Set(cryptoutilIdentityMagic.ParamCodeChallengeMethod, tc.codeChallengeMethod)
			}

			req := httptest.NewRequest("POST", "/oauth2/v1/authorize", strings.NewReader(formBody.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			resp, err := app.Test(req, -1)
			require.NoError(t, err, "Request should succeed")

			defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

			require.Equal(t, tc.wantStatus, resp.StatusCode, "Status code should match expected")
		})
	}
}

// TestHandleAuthorizeGET_ValidRequestCreatesAuthorizationRequest validates successful authorization request creation.
func TestHandleAuthorizeGET_ValidRequestCreatesAuthorizationRequest(t *testing.T) {
	t.Parallel()

	config := createAuthorizePKCETestConfig(t)
	repoFactory := createAuthorizePKCETestRepoFactory(t)

	testClient := createAuthorizePKCETestClient(t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	query := url.Values{
		cryptoutilIdentityMagic.ParamClientID:            []string{testClient.ClientID},
		cryptoutilIdentityMagic.ParamResponseType:        []string{cryptoutilIdentityMagic.ResponseTypeCode},
		cryptoutilIdentityMagic.ParamRedirectURI:         []string{testClient.RedirectURIs[0]},
		cryptoutilIdentityMagic.ParamScope:               []string{"openid profile"},
		cryptoutilIdentityMagic.ParamState:               []string{"test-state"},
		cryptoutilIdentityMagic.ParamCodeChallenge:       []string{"E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"},
		cryptoutilIdentityMagic.ParamCodeChallengeMethod: []string{cryptoutilIdentityMagic.PKCEMethodS256},
	}

	req := httptest.NewRequest("GET", "/oauth2/v1/authorize?"+query.Encode(), nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() {
		_ = resp.Body.Close() //nolint:errcheck // Test cleanup
	}()

	require.Equal(t, fiber.StatusFound, resp.StatusCode, "Should redirect to login")

	locationHeader := resp.Header.Get("Location")
	require.Contains(t, locationHeader, "/oidc/v1/login?request_id=", "Should redirect to login with request_id")
}

func createAuthorizePKCETestConfig(t *testing.T) *cryptoutilIdentityConfig.Config {
	t.Helper()

	testID := googleUuid.Must(googleUuid.NewV7()).String()

	return &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  fmt.Sprintf("file:test_%s.db?mode=memory&cache=shared", testID),
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: "https://localhost:8080",
		},
	}
}

func createAuthorizePKCETestRepoFactory(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	cfg := createAuthorizePKCETestConfig(t)
	ctx := context.Background()

	// Clear migration state to ensure fresh database for this test.
	cryptoutilIdentityRepository.ResetMigrationStateForTesting()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:        cfg.Database.Type,
		DSN:         cfg.Database.DSN,
		AutoMigrate: true,
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err, "Failed to create repository factory")
	require.NotNil(t, repoFactory, "Repository factory should not be nil")

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run auto migrations")

	return repoFactory
}

func createAuthorizePKCETestClient(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) *cryptoutilIdentityDomain.Client {
	t.Helper()

	ctx := context.Background()
	clientRepo := repoFactory.ClientRepository()

	clientUUID, err := googleUuid.NewV7()
	require.NoError(t, err, "Failed to generate client UUID")

	testClient := &cryptoutilIdentityDomain.Client{
		ID:                      clientUUID,
		ClientID:                fmt.Sprintf("test-client-%s", clientUUID.String()),
		Name:                    "Test Client",
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		AllowedGrantTypes:       []string{cryptoutilIdentityMagic.GrantTypeAuthorizationCode},
		AllowedScopes:           []string{"openid", "profile", "email"},
		RedirectURIs:            []string{"https://example.com/callback"},
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
	}

	err = clientRepo.Create(ctx, testClient)
	require.NoError(t, err, "Failed to create test client")

	return testClient
}
