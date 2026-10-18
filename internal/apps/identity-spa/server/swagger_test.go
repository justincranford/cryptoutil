// Copyright (c) 2025-2026 Justin Cranford.
//

package server_test

import (
	"io"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilIdentitySpaServer "cryptoutil/internal/apps/identity-spa/server"
)

// TestServeOpenAPISpec_Success validates successful OpenAPI spec serving.
func TestServeOpenAPISpec_Success(t *testing.T) {
	t.Parallel()

	handler, err := cryptoutilIdentitySpaServer.ServeOpenAPISpec()
	require.NoError(t, err)
	require.NotNil(t, handler)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/openapi.json", handler)
	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/openapi.json", nil), -1)

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, "application/json", response.Header.Get("Content-Type"))
	require.Contains(t, string(body), `"openapi":"3.0.3"`)
	require.Contains(t, string(body), "Single Page App Backend-for-Frontend API")
}
