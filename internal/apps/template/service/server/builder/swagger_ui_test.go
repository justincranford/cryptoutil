// Copyright (c) 2025 Justin Cranford
//

package builder

import (
	"encoding/base64"
	"io"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestRegisterSwaggerUI_NilConfig(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	err := RegisterSwaggerUI(app, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config is required")
}

func TestRegisterSwaggerUI_EmptySpec(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	err := RegisterSwaggerUI(app, &SwaggerUIConfig{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "OpenAPI spec JSON is required")
}

func TestRegisterSwaggerUI_DocJSON(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	spec := []byte(`{"openapi":"3.0.3","info":{"title":"Test","version":"1.0"}}`)

	err := RegisterSwaggerUI(app, &SwaggerUIConfig{
		OpenAPISpecJSON:       spec,
		BrowserAPIContextPath: "/browser/api/v1",
		CSRFTokenName:         "csrf_token",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/ui/swagger/doc.json", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, spec, body)
}

func TestRegisterSwaggerUI_CSRFTokenEndpoint(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	err := RegisterSwaggerUI(app, &SwaggerUIConfig{
		OpenAPISpecJSON:       []byte(`{"openapi":"3.0.3"}`),
		BrowserAPIContextPath: "/browser/api/v1",
		CSRFTokenName:         "csrf_token",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/browser/api/v1/csrf-token", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, 200, resp.StatusCode)
	require.Contains(t, string(body), "csrf_token")
}

func TestRegisterSwaggerUI_AuthNoHeader(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	err := RegisterSwaggerUI(app, &SwaggerUIConfig{
		OpenAPISpecJSON:       []byte(`{"openapi":"3.0.3"}`),
		BrowserAPIContextPath: "/browser/api/v1",
		CSRFTokenName:         "csrf_token",
		Username:              "admin",
		Password:              "secret",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/ui/swagger/doc.json", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, 401, resp.StatusCode)
	require.Equal(t, `Basic realm="Swagger UI"`, resp.Header.Get("WWW-Authenticate"))
}

func TestRegisterSwaggerUI_AuthInvalidMethod(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	err := RegisterSwaggerUI(app, &SwaggerUIConfig{
		OpenAPISpecJSON:       []byte(`{"openapi":"3.0.3"}`),
		BrowserAPIContextPath: "/browser/api/v1",
		CSRFTokenName:         "csrf_token",
		Username:              "admin",
		Password:              "secret",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/ui/swagger/doc.json", nil)
	req.Header.Set("Authorization", "Bearer token123")
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, 401, resp.StatusCode)
}

func TestRegisterSwaggerUI_AuthInvalidBase64(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	err := RegisterSwaggerUI(app, &SwaggerUIConfig{
		OpenAPISpecJSON:       []byte(`{"openapi":"3.0.3"}`),
		BrowserAPIContextPath: "/browser/api/v1",
		CSRFTokenName:         "csrf_token",
		Username:              "admin",
		Password:              "secret",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/ui/swagger/doc.json", nil)
	req.Header.Set("Authorization", "Basic not!!valid!!base64")
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, 401, resp.StatusCode)
}

func TestRegisterSwaggerUI_AuthNoColon(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	err := RegisterSwaggerUI(app, &SwaggerUIConfig{
		OpenAPISpecJSON:       []byte(`{"openapi":"3.0.3"}`),
		BrowserAPIContextPath: "/browser/api/v1",
		CSRFTokenName:         "csrf_token",
		Username:              "admin",
		Password:              "secret",
	})
	require.NoError(t, err)

	// No colon in decoded credentials.
	encoded := base64.StdEncoding.EncodeToString([]byte("adminnosecret"))
	req := httptest.NewRequest("GET", "/ui/swagger/doc.json", nil)
	req.Header.Set("Authorization", "Basic "+encoded)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, 401, resp.StatusCode)
}

func TestRegisterSwaggerUI_AuthWrongPassword(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	err := RegisterSwaggerUI(app, &SwaggerUIConfig{
		OpenAPISpecJSON:       []byte(`{"openapi":"3.0.3"}`),
		BrowserAPIContextPath: "/browser/api/v1",
		CSRFTokenName:         "csrf_token",
		Username:              "admin",
		Password:              "secret",
	})
	require.NoError(t, err)

	encoded := base64.StdEncoding.EncodeToString([]byte("admin:wrongpass"))
	req := httptest.NewRequest("GET", "/ui/swagger/doc.json", nil)
	req.Header.Set("Authorization", "Basic "+encoded)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, 401, resp.StatusCode)
}

func TestRegisterSwaggerUI_AuthSuccess(t *testing.T) {
	t.Parallel()

	spec := []byte(`{"openapi":"3.0.3"}`)
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	err := RegisterSwaggerUI(app, &SwaggerUIConfig{
		OpenAPISpecJSON:       spec,
		BrowserAPIContextPath: "/browser/api/v1",
		CSRFTokenName:         "csrf_token",
		Username:              "admin",
		Password:              "secret",
	})
	require.NoError(t, err)

	encoded := base64.StdEncoding.EncodeToString([]byte("admin:secret"))
	req := httptest.NewRequest("GET", "/ui/swagger/doc.json", nil)
	req.Header.Set("Authorization", "Basic "+encoded)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, 200, resp.StatusCode)
	require.Equal(t, spec, body)
}

func TestRegisterSwaggerUI_SwaggerPage(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	err := RegisterSwaggerUI(app, &SwaggerUIConfig{
		OpenAPISpecJSON:       []byte(`{"openapi":"3.0.3"}`),
		BrowserAPIContextPath: "/browser/api/v1",
		CSRFTokenName:         "csrf_token",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/ui/swagger/index.html", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, 200, resp.StatusCode)
}

func TestRegisterSwaggerUI_SwaggerPageContentTypeHTML(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	err := RegisterSwaggerUI(app, &SwaggerUIConfig{
		OpenAPISpecJSON:       []byte(`{"openapi":"3.0.3"}`),
		BrowserAPIContextPath: "/browser/api/v1",
		CSRFTokenName:         "csrf_token",
	})
	require.NoError(t, err)

	// Sending Content-Type: text/html request header triggers the charset code path.
	req := httptest.NewRequest("GET", "/ui/swagger/index.html", nil)
	req.Header.Set("Content-Type", "text/html")
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, 200, resp.StatusCode)
}
