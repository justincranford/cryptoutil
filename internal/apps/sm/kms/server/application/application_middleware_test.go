//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford
//
// NOTE: These tests require a PostgreSQL database and are skipped in CI without the integration tag.
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

// TestSwaggerUIBasicAuthMiddleware tests all middleware scenarios.
func TestSwaggerUIBasicAuthMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		username   string
		password   string
		authHeader string
		wantStatus int
		wantBody   string
		wantAuth   string
	}{
		{
			name:       "no auth configured",
			username:   "",
			password:   "",
			authHeader: "",
			wantStatus: fiber.StatusOK,
			wantBody:   "success",
		},
		{
			name:       "missing auth header",
			username:   "admin",
			password:   "secret",
			authHeader: "",
			wantStatus: fiber.StatusUnauthorized,
			wantBody:   "Authentication required",
			wantAuth:   `Basic realm="Swagger UI"`,
		},
		{
			name:       "invalid auth method",
			username:   "admin",
			password:   "secret",
			authHeader: "Bearer invalid-token",
			wantStatus: fiber.StatusUnauthorized,
			wantBody:   "Invalid authentication method",
		},
		{
			name:       "invalid base64 encoding",
			username:   "admin",
			password:   "secret",
			authHeader: "Basic not-valid-base64!!!",
			wantStatus: fiber.StatusUnauthorized,
			wantBody:   "Invalid authentication encoding",
		},
		{
			name:       "invalid credential format",
			username:   "admin",
			password:   "secret",
			authHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte("invalidformat")),
			wantStatus: fiber.StatusUnauthorized,
			wantBody:   "Invalid authentication format",
		},
		{
			name:       "invalid credentials",
			username:   "admin",
			password:   "secret",
			authHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte("wrong:credentials")),
			wantStatus: fiber.StatusUnauthorized,
			wantBody:   "Invalid credentials",
		},
		{
			name:       "valid credentials",
			username:   "admin",
			password:   "secret",
			authHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:secret")),
			wantStatus: fiber.StatusOK,
			wantBody:   "success",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			app.Get("/test", swaggerUIBasicAuthMiddleware(tc.username, tc.password), func(c *fiber.Ctx) error {
				return c.SendString("success")
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			resp, err := app.Test(req)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tc.wantStatus, resp.StatusCode)

			if tc.wantAuth != "" {
				require.Equal(t, tc.wantAuth, resp.Header.Get("WWW-Authenticate"))
			}

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Contains(t, string(body), tc.wantBody)
		})
	}
}
