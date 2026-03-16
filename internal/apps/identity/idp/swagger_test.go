// Copyright (c) 2025 Justin Cranford

package idp_test

import (
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityIdp "cryptoutil/internal/apps/identity/idp"
)

// TestServeOpenAPISpec_Success validates successful OpenAPI spec retrieval.
func TestServeOpenAPISpec_Success(t *testing.T) {
	t.Parallel()

	handler, err := cryptoutilIdentityIdp.ServeOpenAPISpec()
	require.NoError(t, err, "Should create handler successfully")
	require.NotNil(t, handler, "Handler should not be nil")

	app := fiber.New()
	app.Get("/swagger/doc.json", handler)

	req := httptest.NewRequest("GET", "/swagger/doc.json", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}
