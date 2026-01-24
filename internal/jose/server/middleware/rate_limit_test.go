// Copyright (c) 2025 Justin Cranford
//

package middleware

import (
	"context"
	"io"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestNewRateLimiter_DefaultConfig(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(NewRateLimiter(nil))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/test", nil)
	require.NoError(t, err)

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, "OK", string(body))
}

func TestNewRateLimiter_CustomConfig(t *testing.T) {
	t.Parallel()

	cfg := &RateLimitConfig{
		Max:        10,
		Expiration: time.Second,
	}

	app := fiber.New()
	app.Use(NewRateLimiter(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/test", nil)
	require.NoError(t, err)

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestNewRateLimiter_ExceedsLimit(t *testing.T) {
	t.Parallel()

	cfg := &RateLimitConfig{
		Max:        2,
		Expiration: time.Minute, // Long expiration to ensure we hit the limit.
	}

	app := fiber.New()
	app.Use(NewRateLimiter(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First two requests should succeed.
	for i := range 2 {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/test", nil)
		require.NoError(t, err)

		resp, err := app.Test(req)
		require.NoError(t, err, "request %d failed", i+1)
		require.Equal(t, http.StatusOK, resp.StatusCode, "request %d should succeed", i+1)
		require.NoError(t, resp.Body.Close())
	}

	// Third request should be rate limited.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/test", nil)
	require.NoError(t, err)

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Contains(t, string(body), "Rate limit exceeded")
}

func TestNewRateLimiter_WithTelemetry(t *testing.T) {
	t.Parallel()

	telemetrySettings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	telemetryService, err := cryptoutilSharedTelemetry.NewTelemetryService(
		context.Background(),
		telemetrySettings,
	)
	require.NoError(t, err)

	defer telemetryService.Shutdown()

	cfg := &RateLimitConfig{
		Max:              1,
		Expiration:       time.Minute,
		TelemetryService: telemetryService,
	}

	app := fiber.New()
	app.Use(NewRateLimiter(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First request should succeed.
	req1, err := http.NewRequestWithContext(ctx, http.MethodGet, "/test", nil)
	require.NoError(t, err)

	resp1, err := app.Test(req1)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp1.StatusCode)
	require.NoError(t, resp1.Body.Close())

	// Second request should be rate limited (with telemetry logging).
	req2, err := http.NewRequestWithContext(ctx, http.MethodGet, "/test", nil)
	require.NoError(t, err)

	resp2, err := app.Test(req2)
	require.NoError(t, err)
	require.Equal(t, http.StatusTooManyRequests, resp2.StatusCode)
	require.NoError(t, resp2.Body.Close())
}

func TestNewRateLimiter_DifferentPaths(t *testing.T) {
	t.Parallel()

	cfg := &RateLimitConfig{
		Max:        2,
		Expiration: time.Minute,
	}

	app := fiber.New()
	app.Use(NewRateLimiter(cfg))
	app.Get("/path1", func(c *fiber.Ctx) error {
		return c.SendString("Path1")
	})
	app.Get("/path2", func(c *fiber.Ctx) error {
		return c.SendString("Path2")
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Request path1 twice (uses up limit).
	for range 2 {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/path1", nil)
		require.NoError(t, err)

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.NoError(t, resp.Body.Close())
	}

	// Third request to different path should still be rate limited (IP-based).
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/path2", nil)
	require.NoError(t, err)

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestRateLimitConfig_Defaults(t *testing.T) {
	t.Parallel()

	require.Equal(t, 100, DefaultRateLimit)
	require.Equal(t, time.Second, DefaultRateLimitExpiration)
}

func TestNewRateLimiter_ZeroMax(t *testing.T) {
	t.Parallel()

	cfg := &RateLimitConfig{
		Max:        0, // Should use default.
		Expiration: time.Second,
	}

	app := fiber.New()
	app.Use(NewRateLimiter(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/test", nil)
	require.NoError(t, err)

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestNewRateLimiter_ZeroExpiration(t *testing.T) {
	t.Parallel()

	cfg := &RateLimitConfig{
		Max:        10,
		Expiration: 0, // Should use default.
	}

	app := fiber.New()
	app.Use(NewRateLimiter(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/test", nil)
	require.NoError(t, err)

	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestNewRateLimiter_ResponseBody(t *testing.T) {
	t.Parallel()

	cfg := &RateLimitConfig{
		Max:        1,
		Expiration: time.Minute,
	}

	app := fiber.New()
	app.Use(NewRateLimiter(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First request succeeds.
	req1, err := http.NewRequestWithContext(ctx, http.MethodGet, "/test", nil)
	require.NoError(t, err)

	resp1, err := app.Test(req1)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp1.StatusCode)
	require.NoError(t, resp1.Body.Close())

	// Second request rate limited - check response body.
	req2, err := http.NewRequestWithContext(ctx, http.MethodGet, "/test", nil)
	require.NoError(t, err)

	resp2, err := app.Test(req2)
	require.NoError(t, err)
	require.Equal(t, http.StatusTooManyRequests, resp2.StatusCode)

	body, err := io.ReadAll(resp2.Body)
	require.NoError(t, err)
	require.NoError(t, resp2.Body.Close())
	require.Contains(t, string(body), "Rate limit exceeded")
	require.Contains(t, string(body), "Too many requests")
}

// Helper to ensure httptest import is used.
var _ = httptest.NewRecorder
