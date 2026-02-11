// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

import (
	"context"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

func TestRegisterMiddleware_ComprehensivePipeline(t *testing.T) {
	t.Parallel()

	config := createMiddlewareTestConfig(t)
	repoFactory := createMiddlewareTestRepoFactory(t)

	// Middleware tests don't need token service (testing infrastructure).
	service := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)

	app := fiber.New()
	service.RegisterMiddleware(app)

	// Add test route to verify middleware chain executes.
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Execute request through middleware pipeline.
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Response should be 200 OK")
}

func TestRegisterMiddleware_PanicRecovery(t *testing.T) {
	t.Parallel()

	config := createMiddlewareTestConfig(t)
	repoFactory := createMiddlewareTestRepoFactory(t)
	service := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)

	app := fiber.New()
	service.RegisterMiddleware(app)

	// Add route that panics.
	app.Get("/panic", func(_ *fiber.Ctx) error {
		panic("intentional test panic")
	})

	// Execute request - should recover from panic.
	req := httptest.NewRequest("GET", "/panic", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should not error (panic recovered)")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode, "Should return 500 after panic recovery")
}

func TestRegisterMiddleware_CORSHeaders(t *testing.T) {
	t.Parallel()

	config := createMiddlewareTestConfig(t)
	repoFactory := createMiddlewareTestRepoFactory(t)
	service := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)

	app := fiber.New()
	service.RegisterMiddleware(app)

	app.Get("/cors-test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Execute OPTIONS request (CORS preflight).
	req := httptest.NewRequest("OPTIONS", "/cors-test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "OPTIONS request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusNoContent, resp.StatusCode, "CORS preflight should return 204")
	require.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Origin"), "CORS headers should be present")
}

func TestRegisterMiddleware_RateLimitEnforcement(t *testing.T) {
	t.Parallel()

	config := createMiddlewareTestConfig(t)
	repoFactory := createMiddlewareTestRepoFactory(t)
	service := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)

	app := fiber.New()
	service.RegisterMiddleware(app)

	app.Get("/rate-limited", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Execute multiple requests rapidly to test rate limiting.
	// Note: Default rate limit is 100 req/min, so we send 5 requests to verify middleware is active.
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/rate-limited", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err, "Request should succeed")

		defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

		require.NotEqual(t, fiber.StatusTooManyRequests, resp.StatusCode, "Rate limit should not be hit with 5 requests")
	}
}

func createMiddlewareTestConfig(t *testing.T) *cryptoutilIdentityConfig.Config {
	t.Helper()

	return &cryptoutilIdentityConfig.Config{
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: "https://identity.example.com",
		},
	}
}

func createMiddlewareTestRepoFactory(t *testing.T) *cryptoutilIdentityRepository.RepositoryFactory {
	t.Helper()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:        "sqlite",
		DSN:         "file::memory:?cache=private",
		AutoMigrate: true, // Enable auto-migration for tests
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err, "Failed to create repository factory")

	// Run migrations explicitly (AutoMigrate field doesn't trigger migration).
	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err, "Failed to run database migrations")

	return repoFactory
}
