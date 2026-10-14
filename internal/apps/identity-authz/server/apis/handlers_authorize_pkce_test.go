// Copyright (c) 2025-2026 Justin Cranford.
package apis_test

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

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity-authz/server/apis"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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
			name:                testMissingCodeChallengeRequired,
			codeChallenge:       "",
			codeChallengeMethod: cryptoutilSharedMagic.PKCEMethodS256,
			wantStatus:          fiber.StatusBadRequest,
			wantErrorCode:       cryptoutilSharedMagic.ErrorInvalidRequest,
		},
		{
			name:                testValidS256CodeChallenge,
			codeChallenge:       testE9Melhoa2OwvFrEmtjguChaoeK1t8UrwbuGjsstwCM,
			codeChallengeMethod: cryptoutilSharedMagic.PKCEMethodS256,
			wantStatus:          fiber.StatusFound,
			wantErrorCode:       "",
		},
		{
			name:                testDefaultMethodIsS256,
			codeChallenge:       testE9Melhoa2OwvFrEmtjguChaoeK1t8UrwbuGjsstwCM,
			codeChallengeMethod: "",
			wantStatus:          fiber.StatusFound,
			wantErrorCode:       "",
		},
		{
			name:                testPlainMethodRejected,
			codeChallenge:       testTestPlainChallenge,
			codeChallengeMethod: cryptoutilSharedMagic.PKCEMethodPlain,
			wantStatus:          fiber.StatusBadRequest,
			wantErrorCode:       cryptoutilSharedMagic.ErrorInvalidRequest,
		},
		{
			name:                "invalid method rejected",
			codeChallenge:       testE9Melhoa2OwvFrEmtjguChaoeK1t8UrwbuGjsstwCM,
			codeChallengeMethod: cryptoutilSharedMagic.DevSetupTestInvalidName,
			wantStatus:          fiber.StatusBadRequest,
			wantErrorCode:       cryptoutilSharedMagic.ErrorInvalidRequest,
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
				cryptoutilSharedMagic.ParamClientID:     []string{testClient.ClientID},
				cryptoutilSharedMagic.ParamResponseType: []string{cryptoutilSharedMagic.ResponseTypeCode},
				cryptoutilSharedMagic.ParamRedirectURI:  []string{testClient.RedirectURIs[0]},
				cryptoutilSharedMagic.ParamScope:        []string{testOpenidProfile},
				cryptoutilSharedMagic.ParamState:        []string{testTestState},
			}

			if tc.codeChallenge != "" {
				query.Set(cryptoutilSharedMagic.ParamCodeChallenge, tc.codeChallenge)
			}

			if tc.codeChallengeMethod != "" {
				query.Set(cryptoutilSharedMagic.ParamCodeChallengeMethod, tc.codeChallengeMethod)
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
			name:                testMissingCodeChallengeRequired,
			codeChallenge:       "",
			codeChallengeMethod: cryptoutilSharedMagic.PKCEMethodS256,
			wantStatus:          fiber.StatusBadRequest,
			wantErrorCode:       cryptoutilSharedMagic.ErrorInvalidRequest,
		},
		{
			name:                testValidS256CodeChallenge,
			codeChallenge:       testE9Melhoa2OwvFrEmtjguChaoeK1t8UrwbuGjsstwCM,
			codeChallengeMethod: cryptoutilSharedMagic.PKCEMethodS256,
			wantStatus:          fiber.StatusFound,
			wantErrorCode:       "",
		},
		{
			name:                testDefaultMethodIsS256,
			codeChallenge:       testE9Melhoa2OwvFrEmtjguChaoeK1t8UrwbuGjsstwCM,
			codeChallengeMethod: "",
			wantStatus:          fiber.StatusFound,
			wantErrorCode:       "",
		},
		{
			name:                testPlainMethodRejected,
			codeChallenge:       testTestPlainChallenge,
			codeChallengeMethod: cryptoutilSharedMagic.PKCEMethodPlain,
			wantStatus:          fiber.StatusBadRequest,
			wantErrorCode:       cryptoutilSharedMagic.ErrorInvalidRequest,
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
				cryptoutilSharedMagic.ParamClientID:     []string{testClient.ClientID},
				cryptoutilSharedMagic.ParamResponseType: []string{cryptoutilSharedMagic.ResponseTypeCode},
				cryptoutilSharedMagic.ParamRedirectURI:  []string{testClient.RedirectURIs[0]},
				cryptoutilSharedMagic.ParamScope:        []string{testOpenidProfile},
				cryptoutilSharedMagic.ParamState:        []string{testTestState},
			}

			if tc.codeChallenge != "" {
				formBody.Set(cryptoutilSharedMagic.ParamCodeChallenge, tc.codeChallenge)
			}

			if tc.codeChallengeMethod != "" {
				formBody.Set(cryptoutilSharedMagic.ParamCodeChallengeMethod, tc.codeChallengeMethod)
			}

			req := httptest.NewRequest(testPost, testOauth2V1Authorize, strings.NewReader(formBody.Encode()))
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
		cryptoutilSharedMagic.ParamClientID:            []string{testClient.ClientID},
		cryptoutilSharedMagic.ParamResponseType:        []string{cryptoutilSharedMagic.ResponseTypeCode},
		cryptoutilSharedMagic.ParamRedirectURI:         []string{testClient.RedirectURIs[0]},
		cryptoutilSharedMagic.ParamScope:               []string{testOpenidProfile},
		cryptoutilSharedMagic.ParamState:               []string{testTestState},
		cryptoutilSharedMagic.ParamCodeChallenge:       []string{testE9Melhoa2OwvFrEmtjguChaoeK1t8UrwbuGjsstwCM},
		cryptoutilSharedMagic.ParamCodeChallengeMethod: []string{cryptoutilSharedMagic.PKCEMethodS256},
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
			Type: cryptoutilSharedMagic.TestDatabaseSQLite,
			DSN:  fmt.Sprintf("file:test_%s.db?mode=memory&cache=shared", testID),
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: testHTTPSLocalhost8080,
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
		Name:                    cryptoutilSharedMagic.TestClientName,
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		AllowedGrantTypes:       []string{cryptoutilSharedMagic.GrantTypeAuthorizationCode},
		AllowedScopes:           []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile, cryptoutilSharedMagic.ClaimEmail},
		RedirectURIs:            []string{cryptoutilSharedMagic.TestRedirectURI},
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
	}

	err = clientRepo.Create(ctx, testClient)
	require.NoError(t, err, "Failed to create test client")

	return testClient
}
