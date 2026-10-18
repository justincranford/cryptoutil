// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	json "encoding/json"
	"io"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestServeOpenAPISpec_HandlerResponse(t *testing.T) {
	t.Parallel()

	handler, err := ServeOpenAPISpec()
	require.NoError(t, err)
	require.NotNil(t, handler)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/openapi.json", handler)

	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/openapi.json", nil), -1)

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()

	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, "application/json", response.Header.Get("Content-Type"))

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	var spec struct {
		OpenAPI string         `json:"openapi"`
		Info    map[string]any `json:"info"`
		Paths   map[string]any `json:"paths"`
	}
	require.NoError(t, json.Unmarshal(body, &spec))
	require.Equal(t, "3.0.3", spec.OpenAPI)
	require.Equal(t, "OIDC Relying Party API", spec.Info["title"])
	require.Equal(t, cryptoutilSharedMagic.ServiceVersion, spec.Info[cryptoutilSharedMagic.CLIVersionCommand])
	require.Empty(t, spec.Paths)
}
