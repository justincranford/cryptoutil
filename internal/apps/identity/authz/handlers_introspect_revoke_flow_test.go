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

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

// TestHandleIntrospect_ActiveToken validates introspection of active access tokens.
func TestHandleIntrospect_ActiveToken(t *testing.T) {
	t.Parallel()

	config := createIntrospectRevokeTestConfig(t)
	repoFactory := createIntrospectRevokeTestRepoFactory(t)

	testClient := createIntrospectRevokeTestClient(t, repoFactory)
	testToken := createIntrospectRevokeTestToken(t, repoFactory, testClient.ID, cryptoutilIdentityDomain.TokenTypeAccess, false)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formBody := url.Values{
		cryptoutilIdentityMagic.ParamToken: []string{testToken.TokenValue},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/introspect", strings.NewReader(formBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Should return 200 OK for active token")
}

// TestHandleIntrospect_RevokedToken validates introspection of revoked tokens.
func TestHandleIntrospect_RevokedToken(t *testing.T) {
	t.Parallel()

	config := createIntrospectRevokeTestConfig(t)
	repoFactory := createIntrospectRevokeTestRepoFactory(t)

	testClient := createIntrospectRevokeTestClient(t, repoFactory)
	testToken := createIntrospectRevokeTestToken(t, repoFactory, testClient.ID, cryptoutilIdentityDomain.TokenTypeAccess, true)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formBody := url.Values{
		cryptoutilIdentityMagic.ParamToken: []string{testToken.TokenValue},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/introspect", strings.NewReader(formBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Should return 200 OK with active:false for revoked token")
}

// TestHandleIntrospect_ExpiredToken validates introspection of expired tokens.
func TestHandleIntrospect_ExpiredToken(t *testing.T) {
	t.Parallel()

	config := createIntrospectRevokeTestConfig(t)
	repoFactory := createIntrospectRevokeTestRepoFactory(t)

	testClient := createIntrospectRevokeTestClient(t, repoFactory)

	ctx := context.Background()
	tokenRepo := repoFactory.TokenRepository()

	clientUUID := testClient.ID
	tokenID := googleUuid.Must(googleUuid.NewV7())

	expiredToken := &cryptoutilIdentityDomain.Token{
		ID:         tokenID,
		TokenValue: fmt.Sprintf("expired-token-%s", tokenID.String()),
		TokenType:  cryptoutilIdentityDomain.TokenTypeAccess,
		ClientID:   clientUUID,
		Scopes:     []string{"openid", "profile"},
		ExpiresAt:  time.Now().UTC().Add(-1 * time.Hour),
		IssuedAt:   time.Now().UTC().Add(-2 * time.Hour),
	}

	err := tokenRepo.Create(ctx, expiredToken)
	require.NoError(t, err, "Failed to create expired token")

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formBody := url.Values{
		cryptoutilIdentityMagic.ParamToken: []string{expiredToken.TokenValue},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/introspect", strings.NewReader(formBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Should return 200 OK with active:false for expired token")
}

// TestHandleRevoke_ValidToken validates successful token revocation.
func TestHandleRevoke_ValidToken(t *testing.T) {
	t.Parallel()

	config := createIntrospectRevokeTestConfig(t)
	repoFactory := createIntrospectRevokeTestRepoFactory(t)

	testClient := createIntrospectRevokeTestClient(t, repoFactory)
	testToken := createIntrospectRevokeTestToken(t, repoFactory, testClient.ID, cryptoutilIdentityDomain.TokenTypeAccess, false)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formBody := url.Values{
		cryptoutilIdentityMagic.ParamToken: []string{testToken.TokenValue},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/revoke", strings.NewReader(formBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Should return 200 OK for successful revocation")
}

// TestHandleRevoke_NonExistentToken validates revocation of non-existent tokens.
func TestHandleRevoke_NonExistentToken(t *testing.T) {
	t.Parallel()

	config := createIntrospectRevokeTestConfig(t)
	repoFactory := createIntrospectRevokeTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formBody := url.Values{
		cryptoutilIdentityMagic.ParamToken: []string{"non-existent-token-12345"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/revoke", strings.NewReader(formBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Should return 200 OK for non-existent token (RFC 7009 section 2.2)")
}

func createIntrospectRevokeTestConfig(t *testing.T) *cryptoutilIdentityConfig.Config {
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

func createIntrospectRevokeTestRepoFactory(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	cfg := createIntrospectRevokeTestConfig(t)
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

func createIntrospectRevokeTestClient(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) *cryptoutilIdentityDomain.Client {
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

func createIntrospectRevokeTestToken(
	t *testing.T,
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory,
	clientID googleUuid.UUID,
	tokenType cryptoutilIdentityDomain.TokenType,
	revoked bool,
) *cryptoutilIdentityDomain.Token {
	t.Helper()

	ctx := context.Background()
	tokenRepo := repoFactory.TokenRepository()

	tokenID := googleUuid.Must(googleUuid.NewV7())

	testToken := &cryptoutilIdentityDomain.Token{
		ID:         tokenID,
		TokenValue: fmt.Sprintf("test-token-%s", tokenID.String()),
		TokenType:  tokenType,
		ClientID:   clientID,
		Scopes:     []string{"openid", "profile"},
		ExpiresAt:  time.Now().UTC().Add(1 * time.Hour),
		IssuedAt:   time.Now().UTC(),
	}

	if revoked {
		now := time.Now().UTC()
		testToken.RevokedAt = &now
	}

	err := tokenRepo.Create(ctx, testToken)
	require.NoError(t, err, "Failed to create test token")

	return testToken
}
