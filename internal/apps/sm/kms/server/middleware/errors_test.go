// Copyright (c) 2025 Justin Cranford
//
//

package middleware

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"io"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestAuthErrorResponder_NewAuthErrorResponder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		format      ErrorFormat
		detailLevel string
	}{
		{
			name:        "oauth2 format minimal",
			format:      ErrorFormatOAuth2,
			detailLevel: "minimal",
		},
		{
			name:        "problem format standard",
			format:      ErrorFormatProblem,
			detailLevel: "standard",
		},
		{
			name:        "hybrid format verbose",
			format:      ErrorFormatHybrid,
			detailLevel: "verbose",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			responder := NewAuthErrorResponder(tc.format, tc.detailLevel)
			require.NotNil(t, responder)
			require.Equal(t, tc.format, responder.format)
			require.Equal(t, tc.detailLevel, responder.detailLevel)
		})
	}
}

func TestAuthErrorResponder_SendUnauthorized(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		format      ErrorFormat
		detailLevel string
		wantHeader  string
		wantContent string
	}{
		{
			name:        "oauth2 with standard detail",
			format:      ErrorFormatOAuth2,
			detailLevel: "standard",
			wantHeader:  `Bearer error="invalid_token"`,
			wantContent: "application/json",
		},
		{
			name:        "problem details",
			format:      ErrorFormatProblem,
			detailLevel: "standard",
			wantHeader:  "",
			wantContent: "application/problem+json",
		},
		{
			name:        "hybrid",
			format:      ErrorFormatHybrid,
			detailLevel: "standard",
			wantHeader:  `Bearer error="invalid_token"`,
			wantContent: "application/json",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			responder := NewAuthErrorResponder(tc.format, tc.detailLevel)

			app.Get("/test", func(c *fiber.Ctx) error {
				return responder.SendUnauthorized(c, AuthErrorInvalidToken, "Token has expired")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
			require.Contains(t, resp.Header.Get("Content-Type"), tc.wantContent)

			if tc.wantHeader != "" {
				require.Equal(t, tc.wantHeader, resp.Header.Get("WWW-Authenticate"))
			}
		})
	}
}

func TestAuthErrorResponder_SendForbidden(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	responder := NewAuthErrorResponder(ErrorFormatOAuth2, "standard")

	app.Get("/test", func(c *fiber.Ctx) error {
		return responder.SendForbidden(c, AuthErrorInsufficientScope, "Required scope: admin")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), cryptoutilSharedMagic.ErrorInsufficientScope)
}

func TestAuthErrorResponder_SendBadRequest(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	responder := NewAuthErrorResponder(ErrorFormatOAuth2, "standard")

	app.Get("/test", func(c *fiber.Ctx) error {
		return responder.SendBadRequest(c, AuthErrorInvalidRequest, "Missing required parameter")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), cryptoutilSharedMagic.ErrorInvalidRequest)
}

func TestAuthErrorResponder_SendServerError(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	responder := NewAuthErrorResponder(ErrorFormatOAuth2, "standard")

	app.Get("/test", func(c *fiber.Ctx) error {
		return responder.SendServerError(c, "Internal processing error")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), cryptoutilSharedMagic.ErrorServerError)
}

func TestAuthErrorResponder_MinimalDetail(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	responder := NewAuthErrorResponder(ErrorFormatOAuth2, "minimal")

	app.Get("/test", func(c *fiber.Ctx) error {
		return responder.SendUnauthorized(c, AuthErrorInvalidToken, "Should not appear")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Description should not appear in minimal mode.
	require.NotContains(t, string(body), "Should not appear")
}

func TestAuthErrorResponder_WithExtensions(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	responder := NewAuthErrorResponder(ErrorFormatHybrid, "standard")

	app.Get("/test", func(c *fiber.Ctx) error {
		extensions := map[string]any{
			"retry_after": cryptoutilSharedMagic.IdentityDefaultIdleTimeoutSeconds,
			"request_id":  "abc123",
		}

		return responder.SendErrorWithExtensions(c, fiber.StatusTooManyRequests, AuthErrorTemporarilyUnavail, "Rate limited", extensions)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusTooManyRequests, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "retry_after")
	require.Contains(t, string(body), "request_id")
}

func TestCodeToTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		code  AuthErrorCode
		title string
	}{
		{AuthErrorInvalidRequest, "Invalid Request"},
		{AuthErrorUnauthorizedClient, "Unauthorized Client"},
		{AuthErrorAccessDenied, "Access Denied"},
		{AuthErrorInvalidToken, "Invalid Token"},
		{AuthErrorInsufficientScope, "Insufficient Scope"},
		{AuthErrorServerError, "Server Error"},
		{AuthErrorTemporarilyUnavail, "Temporarily Unavailable"},
		{AuthErrorInvalidGrant, "Invalid Grant"},
		{AuthErrorUnsupportedGrantType, "Unsupported Grant Type"},
		{AuthErrorCode("unknown"), "Authentication Error"},
	}

	for _, tc := range tests {
		t.Run(string(tc.code), func(t *testing.T) {
			t.Parallel()

			result := codeToTitle(tc.code)
			require.Equal(t, tc.title, result)
		})
	}
}

func TestMakeProblemDetails(t *testing.T) {
	t.Parallel()

	pd := MakeProblemDetails(
		"https://example.com/errors/not-found",
		"Not Found",
		404,
		"The requested resource was not found",
		"/api/v1/users/123",
	)

	require.Equal(t, "https://example.com/errors/not-found", pd.Type)
	require.Equal(t, "Not Found", pd.Title)
	require.Equal(t, 404, pd.Status)
	require.Equal(t, "The requested resource was not found", pd.Detail)
	require.Equal(t, "/api/v1/users/123", pd.Instance)
}

func TestProblemDetails_WithExtension(t *testing.T) {
	t.Parallel()

	pd := MakeProblemDetails("", "Test", 400, "", "")
	pd = pd.WithExtension("field", "username")
	pd = pd.WithExtension("reason", "too_short")

	require.NotNil(t, pd.Extensions)
	require.Equal(t, "username", pd.Extensions["field"])
	require.Equal(t, "too_short", pd.Extensions["reason"])
}
