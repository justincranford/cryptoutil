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

// TestServeOpenAPISpec_Success validates successful OpenAPI spec serving.
func TestServeOpenAPISpec_Success(t *testing.T) {
	t.Parallel()

	handler, err := cryptoutilIdentityAuthz.ServeOpenAPISpec()
	require.NoError(t, err, "ServeOpenAPISpec should succeed")
	require.NotNil(t, handler, "Handler should not be nil")

	app := fiber.New()
	app.Get("/spec", handler)

	req := httptest.NewRequest(http.MethodGet, "/spec", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Should return 200")
	require.Equal(t, "application/json", resp.Header.Get("Content-Type"), "Content-Type should be JSON")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Should read response body")
	require.NotEmpty(t, body, "Response body should not be empty")

	// Validate JSON contains expected OpenAPI fields.
	bodyStr := string(body)
	require.Contains(t, bodyStr, "\"openapi\"", "Should contain openapi field")
	require.Contains(t, bodyStr, "\"info\"", "Should contain info field")
	require.Contains(t, bodyStr, "\"paths\"", "Should contain paths field")
}

// TestServeOpenAPISpec_HandlerInvocation validates handler can be invoked multiple times.
func TestServeOpenAPISpec_HandlerInvocation(t *testing.T) {
	t.Parallel()

	handler, err := cryptoutilIdentityAuthz.ServeOpenAPISpec()
	require.NoError(t, err, "ServeOpenAPISpec should succeed")

	app := fiber.New()
	app.Get("/spec", handler)

	// Invoke handler multiple times.
	for range 3 {
		req := httptest.NewRequest(http.MethodGet, "/spec", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err, "Request should succeed")

		defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

		require.Equal(t, fiber.StatusOK, resp.StatusCode, "Should return 200")
	}
}
