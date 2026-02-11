// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

func TestHandleToken_UnsupportedGrantType(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := url.Values{
		cryptoutilIdentityMagic.ParamGrantType: []string{"invalid_grant_type"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(reqBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")
}

func TestHandleToken_AuthorizationCodeGrant_MissingCode(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := url.Values{
		cryptoutilIdentityMagic.ParamGrantType:   []string{cryptoutilIdentityMagic.GrantTypeAuthorizationCode},
		cryptoutilIdentityMagic.ParamRedirectURI: []string{"https://example.com/callback"},
		cryptoutilIdentityMagic.ParamClientID:    []string{"test-client"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(reqBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")
}

func TestHandleToken_AuthorizationCodeGrant_MissingRedirectURI(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := url.Values{
		cryptoutilIdentityMagic.ParamGrantType: []string{cryptoutilIdentityMagic.GrantTypeAuthorizationCode},
		cryptoutilIdentityMagic.ParamCode:      []string{"test-code"},
		cryptoutilIdentityMagic.ParamClientID:  []string{"test-client"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(reqBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")
}

func TestHandleToken_ClientCredentialsGrant_MissingClient(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := url.Values{
		cryptoutilIdentityMagic.ParamGrantType: []string{cryptoutilIdentityMagic.GrantTypeClientCredentials},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(reqBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode, "Should return 401 Unauthorized")
}

func TestHandleToken_RefreshTokenGrant_MissingRefreshToken(t *testing.T) {
	t.Parallel()

	config := createTokenTestConfig(t)
	repoFactory := createTokenTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := url.Values{
		cryptoutilIdentityMagic.ParamGrantType: []string{cryptoutilIdentityMagic.GrantTypeRefreshToken},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(reqBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")
}

func createTokenTestConfig(t *testing.T) *cryptoutilIdentityConfig.Config {
	t.Helper()

	return &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  "file::memory:?cache=private",
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: "https://localhost:8080",
		},
	}
}

func createTokenTestRepoFactory(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	cfg := createTokenTestConfig(t)
	ctx := context.Background()

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, cfg.Database)
	require.NoError(t, err, "Failed to create repository factory")
	require.NotNil(t, repoFactory, "Repository factory should not be nil")

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run auto migrations")

	return repoFactory
}
