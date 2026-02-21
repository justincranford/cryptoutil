// Copyright (c) 2025 Justin Cranford
//

package builder

import (
	"net/http/httptest"
	"strings"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestValidateSecurityHeaders_AllPresent(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	var missing []string

	app.Get("/test", func(c *fiber.Ctx) error {
		for header, value := range ExpectedBrowserHeaders() {
			c.Set(header, value)
		}

		missing = ValidateSecurityHeaders(c)

		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Empty(t, missing)
}

func TestValidateSecurityHeaders_MissingHeaders(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	var missing []string

	app.Get("/test", func(c *fiber.Ctx) error {
		missing = ValidateSecurityHeaders(c)

		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.NotEmpty(t, missing)
}

func TestValidateSecurityHeaders_HTTPSMissingHSTS(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ProxyHeader:           "X-Forwarded-Proto",
	})

	var missing []string

	app.Get("/test", func(c *fiber.Ctx) error {
		for header, value := range ExpectedBrowserHeaders() {
			c.Set(header, value)
		}

		// Intentionally do NOT set HSTS here.
		missing = ValidateSecurityHeaders(c)

		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-Proto", cryptoutilSharedMagic.ProtocolHTTPS)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Contains(t, missing, "Strict-Transport-Security")
}

func TestValidateSecurityHeaders_HTTPSWithHSTS(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ProxyHeader:           "X-Forwarded-Proto",
	})

	var missing []string

	app.Get("/test", func(c *fiber.Ctx) error {
		for header, value := range ExpectedBrowserHeaders() {
			c.Set(header, value)
		}

		c.Set("Strict-Transport-Security", cryptoutilSharedMagic.HSTSMaxAge)
		missing = ValidateSecurityHeaders(c)

		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-Proto", cryptoutilSharedMagic.ProtocolHTTPS)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Empty(t, missing)
}

func TestCreateAdditionalSecurityHeadersMiddleware_HTTPSDevMode(t *testing.T) {
	t.Parallel()

	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		PublicServiceAPIContextPath: "/service/",
		PublicBrowserAPIContextPath: "/browser/",
		DevMode:                     true,
	}
	config := &SecurityHeadersConfig{Settings: settings}

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ProxyHeader:           "X-Forwarded-Proto",
	})
	app.Use(config.CreateAdditionalSecurityHeadersMiddleware())
	app.Get("/browser/api/v1/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/browser/api/v1/test", nil)
	req.Header.Set("X-Forwarded-Proto", cryptoutilSharedMagic.ProtocolHTTPS)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, cryptoutilSharedMagic.HSTSMaxAgeDev, resp.Header.Get("Strict-Transport-Security"))
}

func TestCreateAdditionalSecurityHeadersMiddleware_HTTPSProduction(t *testing.T) {
	t.Parallel()

	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		PublicServiceAPIContextPath: "/service/",
		PublicBrowserAPIContextPath: "/browser/",
		DevMode:                     false,
	}
	config := &SecurityHeadersConfig{Settings: settings}

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ProxyHeader:           "X-Forwarded-Proto",
	})
	app.Use(config.CreateAdditionalSecurityHeadersMiddleware())
	app.Get("/browser/api/v1/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/browser/api/v1/test", nil)
	req.Header.Set("X-Forwarded-Proto", cryptoutilSharedMagic.ProtocolHTTPS)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, cryptoutilSharedMagic.HSTSMaxAge, resp.Header.Get("Strict-Transport-Security"))
}

func TestCreateAdditionalSecurityHeadersMiddleware_PostLogout(t *testing.T) {
	t.Parallel()

	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		PublicServiceAPIContextPath: "/service/",
		PublicBrowserAPIContextPath: "/browser/",
	}
	config := &SecurityHeadersConfig{Settings: settings}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(config.CreateAdditionalSecurityHeadersMiddleware())
	app.Post("/browser/api/v1/logout", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("POST", "/browser/api/v1/logout", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, cryptoutilSharedMagic.ClearSiteDataLogout, resp.Header.Get("Clear-Site-Data"))
}

func TestBuildContentSecurityPolicy_VerboseMode(t *testing.T) {
	t.Parallel()

	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:     true,
		VerboseMode: true,
	}

	csp := buildContentSecurityPolicy(settings)

	require.True(t, strings.Contains(csp, "localhost"))
}
