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

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestHandleAuthorizeGET_HappyPath tests successful authorization request flow.
func TestHandleAuthorizeGET_HappyPath(t *testing.T) {
	t.Parallel()

	config := createAuthorizeComprehensiveTestConfig(t)
	repoFactory := createAuthorizeComprehensiveTestRepoFactory(t)

	// Create test client.
	ctx := context.Background()
	testClient := createTestClient(ctx, t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	query := url.Values{
		cryptoutilSharedMagic.ParamClientID:            []string{testClient.ClientID},
		cryptoutilSharedMagic.ParamResponseType:        []string{cryptoutilSharedMagic.ResponseTypeCode},
		cryptoutilSharedMagic.ParamRedirectURI:         []string{testClient.RedirectURIs[0]},
		cryptoutilSharedMagic.ParamScope:               []string{"openid profile"},
		cryptoutilSharedMagic.ParamState:               []string{"test-state"},
		cryptoutilSharedMagic.ParamCodeChallenge:       []string{"test-challenge"},
		cryptoutilSharedMagic.ParamCodeChallengeMethod: []string{cryptoutilSharedMagic.PKCEMethodS256},
	}

	req := httptest.NewRequest("GET", "/oauth2/v1/authorize?"+query.Encode(), nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusFound, resp.StatusCode, "Should return 302 redirect to login")
	require.Contains(t, resp.Header.Get("Location"), "/oidc/v1/login", "Should redirect to login with request_id")
}

// TestHandleAuthorizeGET_InvalidClientID tests authorization with non-existent client.
func TestHandleAuthorizeGET_InvalidClientID(t *testing.T) {
	t.Parallel()

	config := createAuthorizeComprehensiveTestConfig(t)
	repoFactory := createAuthorizeComprehensiveTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	query := url.Values{
		cryptoutilSharedMagic.ParamClientID:            []string{"non-existent-client"},
		cryptoutilSharedMagic.ParamResponseType:        []string{cryptoutilSharedMagic.ResponseTypeCode},
		cryptoutilSharedMagic.ParamRedirectURI:         []string{cryptoutilSharedMagic.DemoRedirectURI},
		cryptoutilSharedMagic.ParamScope:               []string{"openid profile"},
		cryptoutilSharedMagic.ParamState:               []string{"test-state"},
		cryptoutilSharedMagic.ParamCodeChallenge:       []string{"test-challenge"},
		cryptoutilSharedMagic.ParamCodeChallengeMethod: []string{cryptoutilSharedMagic.PKCEMethodS256},
	}

	req := httptest.NewRequest("GET", "/oauth2/v1/authorize?"+query.Encode(), nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusNotFound, resp.StatusCode, "Should return 404 Not Found for invalid client")
}

// TestHandleAuthorizeGET_InvalidRedirectURI tests authorization with unregistered redirect URI.
func TestHandleAuthorizeGET_InvalidRedirectURI(t *testing.T) {
	t.Parallel()

	config := createAuthorizeComprehensiveTestConfig(t)
	repoFactory := createAuthorizeComprehensiveTestRepoFactory(t)

	// Create test client.
	ctx := context.Background()
	testClient := createTestClient(ctx, t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	query := url.Values{
		cryptoutilSharedMagic.ParamClientID:            []string{testClient.ClientID},
		cryptoutilSharedMagic.ParamResponseType:        []string{cryptoutilSharedMagic.ResponseTypeCode},
		cryptoutilSharedMagic.ParamRedirectURI:         []string{"https://malicious.com/callback"},
		cryptoutilSharedMagic.ParamScope:               []string{"openid profile"},
		cryptoutilSharedMagic.ParamState:               []string{"test-state"},
		cryptoutilSharedMagic.ParamCodeChallenge:       []string{"test-challenge"},
		cryptoutilSharedMagic.ParamCodeChallengeMethod: []string{cryptoutilSharedMagic.PKCEMethodS256},
	}

	req := httptest.NewRequest("GET", "/oauth2/v1/authorize?"+query.Encode(), nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() {
		_ = resp.Body.Close() //nolint:errcheck // Test cleanup
	}()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request for invalid client")
}

// TestHandleAuthorizeGET_UnsupportedResponseType tests authorization with non-code response type.
func TestHandleAuthorizeGET_UnsupportedResponseType(t *testing.T) {
	t.Parallel()

	config := createAuthorizeComprehensiveTestConfig(t)
	repoFactory := createAuthorizeComprehensiveTestRepoFactory(t)

	// Create test client.
	ctx := context.Background()
	testClient := createTestClient(ctx, t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	query := url.Values{
		cryptoutilSharedMagic.ParamClientID:            []string{testClient.ClientID},
		cryptoutilSharedMagic.ParamResponseType:        []string{cryptoutilSharedMagic.ParamToken},
		cryptoutilSharedMagic.ParamRedirectURI:         []string{testClient.RedirectURIs[0]},
		cryptoutilSharedMagic.ParamScope:               []string{"openid profile"},
		cryptoutilSharedMagic.ParamState:               []string{"test-state"},
		cryptoutilSharedMagic.ParamCodeChallenge:       []string{"test-challenge"},
		cryptoutilSharedMagic.ParamCodeChallengeMethod: []string{cryptoutilSharedMagic.PKCEMethodS256},
	}

	req := httptest.NewRequest("GET", "/oauth2/v1/authorize?"+query.Encode(), nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() {
		_ = resp.Body.Close() //nolint:errcheck // Test cleanup
	}()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should reject implicit flow (OAuth 2.1)")
}

// TestHandleAuthorizeGET_MissingCodeChallenge tests authorization without PKCE.
func TestHandleAuthorizeGET_MissingCodeChallenge(t *testing.T) {
	t.Parallel()

	config := createAuthorizeComprehensiveTestConfig(t)
	repoFactory := createAuthorizeComprehensiveTestRepoFactory(t)

	// Create test client.
	ctx := context.Background()
	testClient := createTestClient(ctx, t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	query := url.Values{
		cryptoutilSharedMagic.ParamClientID:     []string{testClient.ClientID},
		cryptoutilSharedMagic.ParamResponseType: []string{cryptoutilSharedMagic.ResponseTypeCode},
		cryptoutilSharedMagic.ParamRedirectURI:  []string{testClient.RedirectURIs[0]},
		cryptoutilSharedMagic.ParamScope:        []string{"openid profile"},
		cryptoutilSharedMagic.ParamState:        []string{"test-state"},
	}

	req := httptest.NewRequest("GET", "/oauth2/v1/authorize?"+query.Encode(), nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() {
		_ = resp.Body.Close() //nolint:errcheck // Test cleanup
	}()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should only accept S256 PKCE method")
}

// TestHandleAuthorizeGET_InvalidCodeChallengeMethod tests authorization with unsupported PKCE method.
func TestHandleAuthorizeGET_InvalidCodeChallengeMethod(t *testing.T) {
	t.Parallel()

	config := createAuthorizeComprehensiveTestConfig(t)
	repoFactory := createAuthorizeComprehensiveTestRepoFactory(t)

	// Create test client.
	ctx := context.Background()
	testClient := createTestClient(ctx, t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	query := url.Values{
		cryptoutilSharedMagic.ParamClientID:            []string{testClient.ClientID},
		cryptoutilSharedMagic.ParamResponseType:        []string{cryptoutilSharedMagic.ResponseTypeCode},
		cryptoutilSharedMagic.ParamRedirectURI:         []string{testClient.RedirectURIs[0]},
		cryptoutilSharedMagic.ParamScope:               []string{"openid profile"},
		cryptoutilSharedMagic.ParamState:               []string{"test-state"},
		cryptoutilSharedMagic.ParamCodeChallenge:       []string{"test-challenge"},
		cryptoutilSharedMagic.ParamCodeChallengeMethod: []string{cryptoutilSharedMagic.PKCEMethodPlain},
	}

	req := httptest.NewRequest("GET", "/oauth2/v1/authorize?"+query.Encode(), nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() {
		_ = resp.Body.Close() //nolint:errcheck // Test cleanup
	}()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request for invalid redirect URI")
}

// TestHandleAuthorizePOST_HappyPath tests successful POST authorization request.
func TestHandleAuthorizePOST_HappyPath(t *testing.T) {
	t.Parallel()

	config := createAuthorizeComprehensiveTestConfig(t)
	repoFactory := createAuthorizeComprehensiveTestRepoFactory(t)

	// Create test client.
	ctx := context.Background()
	testClient := createTestClient(ctx, t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formData := url.Values{
		cryptoutilSharedMagic.ParamClientID:            []string{testClient.ClientID},
		cryptoutilSharedMagic.ParamResponseType:        []string{cryptoutilSharedMagic.ResponseTypeCode},
		cryptoutilSharedMagic.ParamRedirectURI:         []string{testClient.RedirectURIs[0]},
		cryptoutilSharedMagic.ParamScope:               []string{"openid profile"},
		cryptoutilSharedMagic.ParamState:               []string{"test-state"},
		cryptoutilSharedMagic.ParamCodeChallenge:       []string{"test-challenge"},
		cryptoutilSharedMagic.ParamCodeChallengeMethod: []string{cryptoutilSharedMagic.PKCEMethodS256},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/authorize", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusFound, resp.StatusCode, "Should return 302 redirect to login")
	require.Contains(t, resp.Header.Get("Location"), "/oidc/v1/login", "Should redirect to login with request_id")
}

// TestHandleAuthorizePOST_MissingPKCE tests POST authorization without PKCE.
func TestHandleAuthorizePOST_MissingPKCE(t *testing.T) {
	t.Parallel()

	config := createAuthorizeComprehensiveTestConfig(t)
	repoFactory := createAuthorizeComprehensiveTestRepoFactory(t)

	// Create test client.
	ctx := context.Background()
	testClient := createTestClient(ctx, t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formData := url.Values{
		cryptoutilSharedMagic.ParamClientID:     []string{testClient.ClientID},
		cryptoutilSharedMagic.ParamResponseType: []string{cryptoutilSharedMagic.ResponseTypeCode},
		cryptoutilSharedMagic.ParamRedirectURI:  []string{testClient.RedirectURIs[0]},
		cryptoutilSharedMagic.ParamScope:        []string{"openid profile"},
		cryptoutilSharedMagic.ParamState:        []string{"test-state"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/authorize", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() {
		_ = resp.Body.Close() //nolint:errcheck // Test cleanup
	}()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should require PKCE in POST requests")
}

// Helper functions.

func createAuthorizeComprehensiveTestConfig(t *testing.T) *cryptoutilIdentityConfig.Config {
	t.Helper()

	return &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: cryptoutilSharedMagic.TestDatabaseSQLite,
			DSN:  cryptoutilSharedMagic.SQLiteInMemoryDSN,
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: "https://localhost:8080",
		},
	}
}

func createAuthorizeComprehensiveTestRepoFactory(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	cfg := createAuthorizeComprehensiveTestConfig(t)
	ctx := context.Background()

	// Clear migration state to ensure fresh database for this test.
	cryptoutilIdentityRepository.ResetMigrationStateForTesting()

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, cfg.Database)
	require.NoError(t, err, "Failed to create repository factory")
	require.NotNil(t, repoFactory, "Repository factory should not be nil")

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run auto migrations")

	return repoFactory
}

func createTestClient(ctx context.Context, t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) *cryptoutilIdentityDomain.Client {
	t.Helper()

	clientID := fmt.Sprintf("test-client-%s", googleUuid.New().String())
	client := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.New(),
		ClientID:                clientID,
		ClientSecret:            "test-secret",
		Name:                    "Test Client",
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		RedirectURIs:            []string{cryptoutilSharedMagic.DemoRedirectURI},
		AllowedScopes:           []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile, cryptoutilSharedMagic.ClaimEmail},
		AllowedGrantTypes:       []string{cryptoutilSharedMagic.GrantTypeAuthorizationCode, cryptoutilSharedMagic.GrantTypeRefreshToken},
		AllowedResponseTypes:    []string{cryptoutilSharedMagic.ResponseTypeCode},
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
		AccessTokenLifetime:     cryptoutilSharedMagic.IMDefaultSessionTimeout,
		RefreshTokenLifetime:    cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax,
		IDTokenLifetime:         cryptoutilSharedMagic.IMDefaultSessionTimeout,
		RequirePKCE:             boolPtr(true),
		PKCEChallengeMethod:     cryptoutilSharedMagic.PKCEMethodS256,
		Enabled:                 boolPtr(true),
		CreatedAt:               time.Now().UTC(),
		UpdatedAt:               time.Now().UTC(),
	}

	clientRepo := repoFactory.ClientRepository()
	err := clientRepo.Create(ctx, client)
	require.NoError(t, err, "Failed to create test client")

	return client
}
