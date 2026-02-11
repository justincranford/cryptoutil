// Copyright (c) 2025 Justin Cranford
//
//

package builder

import (
	"encoding/base64"
	"io"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestRegisterSwaggerUI tests the RegisterSwaggerUI function.
func TestRegisterSwaggerUI(t *testing.T) {
	t.Parallel()

	sampleOpenAPISpec := []byte(`{"openapi":"3.0.0","info":{"title":"Test API","version":"1.0.0"},"paths":{}}`)

	tests := []struct {
		name    string
		cfg     *SwaggerUIConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: true,
			errMsg:  "swagger UI config is required",
		},
		{
			name:    "empty OpenAPI spec",
			cfg:     &SwaggerUIConfig{},
			wantErr: true,
			errMsg:  "OpenAPI spec JSON is required",
		},
		{
			name: "valid config without auth",
			cfg: &SwaggerUIConfig{
				OpenAPISpecJSON:       sampleOpenAPISpec,
				BrowserAPIContextPath: "/browser/api/v1",
				CSRFTokenName:         "csrf_token",
			},
			wantErr: false,
		},
		{
			name: "valid config with auth",
			cfg: &SwaggerUIConfig{
				Username:              "admin-" + googleUuid.Must(googleUuid.NewV7()).String(),
				Password:              "secret-" + googleUuid.Must(googleUuid.NewV7()).String(),
				OpenAPISpecJSON:       sampleOpenAPISpec,
				BrowserAPIContextPath: "/browser/api/v1",
				CSRFTokenName:         "csrf_token",
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			err := RegisterSwaggerUI(app, tc.cfg)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestSwaggerUIBasicAuthMiddleware tests all middleware scenarios.
func TestSwaggerUIBasicAuthMiddleware(t *testing.T) {
	t.Parallel()

	// Generate unique credentials for valid credentials test.
	validUsername := "admin-" + googleUuid.Must(googleUuid.NewV7()).String()
	validPassword := "secret-" + googleUuid.Must(googleUuid.NewV7()).String()
	validAuthHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(validUsername+":"+validPassword))

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
			username:   "admin-" + googleUuid.Must(googleUuid.NewV7()).String(),
			password:   "secret-" + googleUuid.Must(googleUuid.NewV7()).String(),
			authHeader: "",
			wantStatus: fiber.StatusUnauthorized,
			wantBody:   "Authentication required",
			wantAuth:   `Basic realm="Swagger UI"`,
		},
		{
			name:       "invalid auth method",
			username:   "admin-" + googleUuid.Must(googleUuid.NewV7()).String(),
			password:   "secret-" + googleUuid.Must(googleUuid.NewV7()).String(),
			authHeader: "Bearer invalid-token",
			wantStatus: fiber.StatusUnauthorized,
			wantBody:   "Invalid authentication method",
		},
		{
			name:       "invalid base64 encoding",
			username:   "admin-" + googleUuid.Must(googleUuid.NewV7()).String(),
			password:   "secret-" + googleUuid.Must(googleUuid.NewV7()).String(),
			authHeader: "Basic not-valid-base64!!!",
			wantStatus: fiber.StatusUnauthorized,
			wantBody:   "Invalid authentication encoding",
		},
		{
			name:       "invalid credential format",
			username:   "admin-" + googleUuid.Must(googleUuid.NewV7()).String(),
			password:   "secret-" + googleUuid.Must(googleUuid.NewV7()).String(),
			authHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte("invalidformat")),
			wantStatus: fiber.StatusUnauthorized,
			wantBody:   "Invalid authentication format",
		},
		{
			name:       "invalid credentials",
			username:   "admin-" + googleUuid.Must(googleUuid.NewV7()).String(),
			password:   "secret-" + googleUuid.Must(googleUuid.NewV7()).String(),
			authHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte("wrong:credentials")),
			wantStatus: fiber.StatusUnauthorized,
			wantBody:   "Invalid credentials",
		},
		{
			name:       "valid credentials",
			username:   validUsername,
			password:   validPassword,
			authHeader: validAuthHeader,
			wantStatus: fiber.StatusOK,
			wantBody:   "success",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})
			app.Get("/test", swaggerUIBasicAuthMiddleware(tc.username, tc.password), func(c *fiber.Ctx) error {
				return c.SendString("success")
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			resp, err := app.Test(req, -1)
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

// TestSwaggerUIEndpoints tests the registered endpoints.
func TestSwaggerUIEndpoints(t *testing.T) {
	t.Parallel()

	sampleOpenAPISpec := []byte(`{"openapi":"3.0.0","info":{"title":"Test API","version":"1.0.0"},"paths":{}}`)

	// Generate unique credentials for valid credentials test case.
	validUsername := "admin-" + googleUuid.Must(googleUuid.NewV7()).String()
	validPassword := "secret-" + googleUuid.Must(googleUuid.NewV7()).String()
	validAuthHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(validUsername+":"+validPassword))

	tests := []struct {
		name         string
		cfg          *SwaggerUIConfig
		method       string
		path         string
		authHeader   string
		wantStatus   int
		wantContains string
	}{
		{
			name: "doc.json without auth required",
			cfg: &SwaggerUIConfig{
				OpenAPISpecJSON:       sampleOpenAPISpec,
				BrowserAPIContextPath: "/browser/api/v1",
			},
			method:       http.MethodGet,
			path:         "/ui/swagger/doc.json",
			wantStatus:   fiber.StatusOK,
			wantContains: `"openapi":"3.0.0"`,
		},
		{
			name: "doc.json with auth required - missing auth",
			cfg: &SwaggerUIConfig{
				Username:              "admin-" + googleUuid.Must(googleUuid.NewV7()).String(),
				Password:              "secret-" + googleUuid.Must(googleUuid.NewV7()).String(),
				OpenAPISpecJSON:       sampleOpenAPISpec,
				BrowserAPIContextPath: "/browser/api/v1",
			},
			method:       http.MethodGet,
			path:         "/ui/swagger/doc.json",
			wantStatus:   fiber.StatusUnauthorized,
			wantContains: "Authentication required",
		},
		{
			name: "doc.json with auth required - valid auth",
			cfg: &SwaggerUIConfig{
				Username:              validUsername,
				Password:              validPassword,
				OpenAPISpecJSON:       sampleOpenAPISpec,
				BrowserAPIContextPath: "/browser/api/v1",
			},
			method:       http.MethodGet,
			path:         "/ui/swagger/doc.json",
			authHeader:   validAuthHeader,
			wantStatus:   fiber.StatusOK,
			wantContains: `"openapi":"3.0.0"`,
		},
		{
			name: "csrf-token endpoint",
			cfg: &SwaggerUIConfig{
				OpenAPISpecJSON:       sampleOpenAPISpec,
				BrowserAPIContextPath: "/browser/api/v1",
				CSRFTokenName:         "my_csrf_token",
			},
			method:       http.MethodGet,
			path:         "/browser/api/v1/csrf-token",
			wantStatus:   fiber.StatusOK,
			wantContains: `"csrf_token_name":"my_csrf_token"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
			})

			err := RegisterSwaggerUI(app, tc.cfg)
			require.NoError(t, err)

			req := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tc.wantStatus, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Contains(t, string(body), tc.wantContains)
		})
	}
}

// TestSwaggerUICustomCSRFScript tests the CSRF script generation.
func TestSwaggerUICustomCSRFScript(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		csrfTokenName         string
		browserAPIContextPath string
		wantContains          []string
	}{
		{
			name:                  "default token name",
			csrfTokenName:         "csrf_token",
			browserAPIContextPath: "/browser/api/v1",
			wantContains: []string{
				"csrf_token",
				"/browser/api/v1/csrf-token",
				"window.fetch",
				"X-CSRF-Token",
			},
		},
		{
			name:                  "custom token name",
			csrfTokenName:         "my_custom_token",
			browserAPIContextPath: "/api/browser",
			wantContains: []string{
				"my_custom_token",
				"/api/browser/csrf-token",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			script := swaggerUICustomCSRFScript(tc.csrfTokenName, tc.browserAPIContextPath)

			for _, want := range tc.wantContains {
				require.Contains(t, string(script), want)
			}
		})
	}
}
