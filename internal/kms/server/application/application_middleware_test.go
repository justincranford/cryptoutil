// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"encoding/base64"
	"io"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// TestSwaggerUIBasicAuthMiddleware_NoAuthConfigured tests middleware with no auth configured.
func TestSwaggerUIBasicAuthMiddleware_NoAuthConfigured(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/test", swaggerUIBasicAuthMiddleware("", ""), func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "success", string(body))
}

// TestSwaggerUIBasicAuthMiddleware_MissingAuthHeader tests middleware when Authorization header is missing.
func TestSwaggerUIBasicAuthMiddleware_MissingAuthHeader(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/test", swaggerUIBasicAuthMiddleware("admin", "secret"), func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	require.Equal(t, `Basic realm="Swagger UI"`, resp.Header.Get("WWW-Authenticate"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "Authentication required")
}

// TestSwaggerUIBasicAuthMiddleware_InvalidAuthMethod tests middleware with non-Basic auth method.
func TestSwaggerUIBasicAuthMiddleware_InvalidAuthMethod(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/test", swaggerUIBasicAuthMiddleware("admin", "secret"), func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "Invalid authentication method")
}

// TestSwaggerUIBasicAuthMiddleware_InvalidBase64Encoding tests middleware with malformed base64.
func TestSwaggerUIBasicAuthMiddleware_InvalidBase64Encoding(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/test", swaggerUIBasicAuthMiddleware("admin", "secret"), func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Basic not-valid-base64!!!")
	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "Invalid authentication encoding")
}

// TestSwaggerUIBasicAuthMiddleware_InvalidCredentialFormat tests middleware with credentials missing colon.
func TestSwaggerUIBasicAuthMiddleware_InvalidCredentialFormat(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/test", swaggerUIBasicAuthMiddleware("admin", "secret"), func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	// Encode "invalidformat" (no colon) as base64
	encodedCreds := base64.StdEncoding.EncodeToString([]byte("invalidformat"))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Basic "+encodedCreds)
	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "Invalid authentication format")
}

// TestSwaggerUIBasicAuthMiddleware_InvalidCredentials tests middleware with wrong username/password.
func TestSwaggerUIBasicAuthMiddleware_InvalidCredentials(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/test", swaggerUIBasicAuthMiddleware("admin", "secret"), func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	// Encode "wrong:credentials" as base64
	encodedCreds := base64.StdEncoding.EncodeToString([]byte("wrong:credentials"))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Basic "+encodedCreds)
	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "Invalid credentials")
}

// TestSwaggerUIBasicAuthMiddleware_ValidCredentials tests middleware with correct username/password.
func TestSwaggerUIBasicAuthMiddleware_ValidCredentials(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/test", swaggerUIBasicAuthMiddleware("admin", "secret"), func(c *fiber.Ctx) error {
		return c.SendString("success")
	})

	// Encode "admin:secret" as base64
	encodedCreds := base64.StdEncoding.EncodeToString([]byte("admin:secret"))
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Basic "+encodedCreds)
	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "success", string(body))
}
