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

func TestHandleIntrospect_MissingToken(t *testing.T) {
	t.Parallel()

	config := createIntrospectTestConfig(t)
	repoFactory := createIntrospectTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := url.Values{}

	req := httptest.NewRequest("POST", "/oauth2/v1/introspect", strings.NewReader(reqBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")
}

func TestHandleIntrospect_MissingClient(t *testing.T) {
	t.Parallel()

	config := createIntrospectTestConfig(t)
	repoFactory := createIntrospectTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := url.Values{
		cryptoutilIdentityMagic.ParamToken: []string{"test-token"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/introspect", strings.NewReader(reqBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Should return 200 OK with inactive token")
}

func TestHandleRevoke_MissingToken(t *testing.T) {
	t.Parallel()

	config := createIntrospectTestConfig(t)
	repoFactory := createIntrospectTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := url.Values{}

	req := httptest.NewRequest("POST", "/oauth2/v1/revoke", strings.NewReader(reqBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")
}

func TestHandleRevoke_MissingClient(t *testing.T) {
	t.Parallel()

	config := createIntrospectTestConfig(t)
	repoFactory := createIntrospectTestRepoFactory(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := url.Values{
		cryptoutilIdentityMagic.ParamToken: []string{"test-token"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/revoke", strings.NewReader(reqBody.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Should return 200 OK even without client auth")
}

func createIntrospectTestConfig(t *testing.T) *cryptoutilIdentityConfig.Config {
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

func createIntrospectTestRepoFactory(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	cfg := createIntrospectTestConfig(t)
	ctx := context.Background()

	// Each test uses cache=private which creates isolated in-memory database.
	// The migration logic already skips caching for in-memory databases.

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, cfg.Database)
	require.NoError(t, err, "Failed to create repository factory")
	require.NotNil(t, repoFactory, "Repository factory should not be nil")

	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run auto migrations")

	return repoFactory
}
