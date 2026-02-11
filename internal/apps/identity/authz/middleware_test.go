// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

import (
	"io"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
)

// TestRegisterMiddleware_RecoverPanic validates panic recovery middleware.
func TestRegisterMiddleware_RecoverPanic(t *testing.T) {
	t.Parallel()

	app := fiber.New()

	// Create minimal service for middleware registration.
	service := &cryptoutilIdentityAuthz.Service{}
	service.RegisterMiddleware(app)

	// Add route that panics.
	app.Get("/panic", func(_ *fiber.Ctx) error {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed despite panic")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode, "Should return 500 from panic recovery")
}

// TestRegisterMiddleware_RateLimiting validates rate limiting middleware.
func TestRegisterMiddleware_RateLimiting(t *testing.T) {
	t.Parallel()

	app := fiber.New()

	// Create minimal service for middleware registration.
	service := &cryptoutilIdentityAuthz.Service{}
	service.RegisterMiddleware(app)

	// Add test route.
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// Make multiple requests quickly.
	const requests = 50

	successCount := 0

	for range requests {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		resp, err := app.Test(req)
		require.NoError(t, err, "Request should succeed")

		defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

		if resp.StatusCode == fiber.StatusOK {
			successCount++
		}
	}

	// Rate limiter should allow some requests.
	require.Greater(t, successCount, 0, "Rate limiter should allow some requests")
}

// TestRegisterMiddleware_CORS validates CORS middleware.
func TestRegisterMiddleware_CORS(t *testing.T) {
	t.Parallel()

	app := fiber.New()

	// Create minimal service for middleware registration.
	service := &cryptoutilIdentityAuthz.Service{}
	service.RegisterMiddleware(app)

	// Add test route.
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")

	resp, err := app.Test(req)
	require.NoError(t, err, "CORS preflight should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusNoContent, resp.StatusCode, "CORS preflight should return 204")

	// Verify CORS headers are present.
	require.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Origin"), "CORS origin header should be set")
}

// TestRegisterMiddleware_Logging validates logging middleware.
func TestRegisterMiddleware_Logging(t *testing.T) {
	t.Parallel()

	app := fiber.New()

	// Create minimal service for middleware registration.
	service := &cryptoutilIdentityAuthz.Service{}
	service.RegisterMiddleware(app)

	// Add test route.
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Should return 200")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Should read response body")
	require.Equal(t, "OK", string(body), "Response body should match")
}
