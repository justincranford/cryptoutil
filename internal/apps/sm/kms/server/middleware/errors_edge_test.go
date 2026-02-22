// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package middleware

import (
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// TestAuthErrorResponder_DefaultFormat tests the default case in sendError format switch.
func TestAuthErrorResponder_DefaultFormat(t *testing.T) {
	t.Parallel()

	// Create responder with unknown format to trigger default case.
	responder := NewAuthErrorResponder(ErrorFormat("unknown-format"), "standard")

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/test", func(c *fiber.Ctx) error {
		return responder.SendUnauthorized(c, AuthErrorInvalidToken, "test default format")
	})

	req := httptest.NewRequest("GET", "/test", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	// Default case falls through to OAuth2 format.
	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestAuthErrorResponder_VerboseDetail tests the verbose detail level in adjustDescription.
func TestAuthErrorResponder_VerboseDetail(t *testing.T) {
	t.Parallel()

	responder := NewAuthErrorResponder(ErrorFormatOAuth2, "verbose")

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/test", func(c *fiber.Ctx) error {
		return responder.SendBadRequest(c, AuthErrorInvalidRequest, "verbose description")
	})

	req := httptest.NewRequest("GET", "/test", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestAuthErrorResponder_DebugDetail tests the debug detail level in adjustDescription.
func TestAuthErrorResponder_DebugDetail(t *testing.T) {
	t.Parallel()

	responder := NewAuthErrorResponder(ErrorFormatOAuth2, "debug")

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/test", func(c *fiber.Ctx) error {
		return responder.SendServerError(c, "debug description")
	})

	req := httptest.NewRequest("GET", "/test", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

// TestAuthErrorResponder_DefaultDetail tests the default case in adjustDescription.
func TestAuthErrorResponder_DefaultDetail(t *testing.T) {
	t.Parallel()

	responder := NewAuthErrorResponder(ErrorFormatOAuth2, "unknown-level")

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/test", func(c *fiber.Ctx) error {
		return responder.SendForbidden(c, AuthErrorInsufficientScope, "should be hidden")
	})

	req := httptest.NewRequest("GET", "/test", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	// Default detail level returns empty description (same as minimal).
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// TestAuthErrorResponder_ProblemDetailsNon401 tests problem details format with non-401 status.
func TestAuthErrorResponder_ProblemDetailsNon401(t *testing.T) {
	t.Parallel()

	responder := NewAuthErrorResponder(ErrorFormatProblem, "standard")

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/test", func(c *fiber.Ctx) error {
		return responder.SendForbidden(c, AuthErrorInsufficientScope, "insufficient scope")
	})

	req := httptest.NewRequest("GET", "/test", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

// TestAuthErrorResponder_HybridNon401 tests hybrid format with non-401 status.
func TestAuthErrorResponder_HybridNon401(t *testing.T) {
	t.Parallel()

	responder := NewAuthErrorResponder(ErrorFormatHybrid, "standard")

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/test", func(c *fiber.Ctx) error {
		return responder.SendBadRequest(c, AuthErrorInvalidRequest, "bad request")
	})

	req := httptest.NewRequest("GET", "/test", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestAuthErrorResponder_HybridUnauthorized tests hybrid format with 401 status.
func TestAuthErrorResponder_HybridUnauthorized(t *testing.T) {
	t.Parallel()

	responder := NewAuthErrorResponder(ErrorFormatHybrid, "standard")

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/test", func(c *fiber.Ctx) error {
		return responder.SendUnauthorized(c, AuthErrorInvalidToken, "unauthorized")
	})

	req := httptest.NewRequest("GET", "/test", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	require.Contains(t, resp.Header.Get("WWW-Authenticate"), "Bearer")
}
