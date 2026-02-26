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

func TestNewDefaultSecurityHeadersConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
	}{
		{
			name:     "nil settings",
			settings: nil,
		},
		{
			name: "production settings",
			settings: &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				DevMode: false,
			},
		},
		{
			name: "development settings",
			settings: &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				DevMode: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := NewDefaultSecurityHeadersConfig(tt.settings)
			require.NotNil(t, config)
			require.Equal(t, tt.settings, config.Settings)
			require.False(t, config.DisableHelmet)
			require.False(t, config.DisableAdditionalHeaders)
			require.Empty(t, config.CustomCSP)
		})
	}
}

func TestSecurityHeadersConfig_CreateHelmetMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		disableHelmet  bool
		customCSP      string
		path           string
		wantCSPApplied bool
	}{
		{
			name:           "helmet enabled on browser path",
			disableHelmet:  false,
			path:           "/browser/api/v1/test",
			wantCSPApplied: true,
		},
		{
			name:           "helmet disabled",
			disableHelmet:  true,
			path:           "/browser/api/v1/test",
			wantCSPApplied: false,
		},
		{
			name:           "helmet skipped for service path",
			disableHelmet:  false,
			path:           "/service/api/v1/test",
			wantCSPApplied: false,
		},
		{
			name:           "custom CSP",
			disableHelmet:  false,
			customCSP:      "default-src 'self';",
			path:           "/browser/api/v1/test",
			wantCSPApplied: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				PublicServiceAPIContextPath: "/service/",
				PublicBrowserAPIContextPath: "/browser/",
			}
			config := &SecurityHeadersConfig{
				Settings:      settings,
				DisableHelmet: tt.disableHelmet,
				CustomCSP:     tt.customCSP,
			}

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Use(config.CreateHelmetMiddleware())
			app.Get(tt.path, func(c *fiber.Ctx) error {
				return c.SendString("OK")
			})

			req := httptest.NewRequest("GET", tt.path, nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { require.NoError(t, resp.Body.Close()) }()

			require.Equal(t, 200, resp.StatusCode)

			csp := resp.Header.Get("Content-Security-Policy")
			if tt.wantCSPApplied {
				require.NotEmpty(t, csp, "CSP should be present")

				if tt.customCSP != "" {
					require.Equal(t, tt.customCSP, csp)
				}
			} else {
				require.Empty(t, csp, "CSP should not be present")
			}
		})
	}
}

func TestSecurityHeadersConfig_CreateAdditionalSecurityHeadersMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		disableAdditional bool
		devMode           bool
		path              string
		wantContentType   bool
		wantPermissions   bool
		wantCrossOrigin   bool
		wantHSTS          bool
	}{
		{
			name:              "browser path production",
			disableAdditional: false,
			devMode:           false,
			path:              "/browser/api/v1/test",
			wantContentType:   true,
			wantPermissions:   true,
			wantCrossOrigin:   true,
			wantHSTS:          false, // No HSTS in test (not HTTPS)
		},
		{
			name:              "service path production",
			disableAdditional: false,
			devMode:           false,
			path:              "/service/api/v1/test",
			wantContentType:   true,
			wantPermissions:   false, // Browser-specific
			wantCrossOrigin:   false, // Browser-specific
			wantHSTS:          false,
		},
		{
			name:              "disabled additional headers",
			disableAdditional: true,
			path:              "/browser/api/v1/test",
			wantContentType:   false,
			wantPermissions:   false,
			wantCrossOrigin:   false,
			wantHSTS:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				PublicServiceAPIContextPath: "/service/",
				PublicBrowserAPIContextPath: "/browser/",
				DevMode:                     tt.devMode,
			}
			config := &SecurityHeadersConfig{
				Settings:                 settings,
				DisableAdditionalHeaders: tt.disableAdditional,
			}

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Use(config.CreateAdditionalSecurityHeadersMiddleware())
			app.Get(tt.path, func(c *fiber.Ctx) error {
				return c.SendString("OK")
			})

			req := httptest.NewRequest("GET", tt.path, nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { require.NoError(t, resp.Body.Close()) }()

			require.Equal(t, 200, resp.StatusCode)

			// Check X-Content-Type-Options.
			contentType := resp.Header.Get("X-Content-Type-Options")
			if tt.wantContentType {
				require.Equal(t, cryptoutilSharedMagic.ContentTypeOptions, contentType)
			} else {
				require.Empty(t, contentType)
			}

			// Check Permissions-Policy.
			permissions := resp.Header.Get("Permissions-Policy")
			if tt.wantPermissions {
				require.Equal(t, cryptoutilSharedMagic.PermissionsPolicy, permissions)
			} else {
				require.Empty(t, permissions)
			}

			// Check Cross-Origin headers.
			crossOriginOpener := resp.Header.Get("Cross-Origin-Opener-Policy")
			if tt.wantCrossOrigin {
				require.Equal(t, cryptoutilSharedMagic.CrossOriginOpenerPolicy, crossOriginOpener)
			} else {
				require.Empty(t, crossOriginOpener)
			}
		})
	}
}

func TestBuildContentSecurityPolicy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		settings       *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:     "nil settings",
			settings: nil,
			wantContains: []string{
				"default-src 'none'",
				"script-src 'self'",
				"style-src 'self'",
				"img-src 'self' data:",
				"frame-ancestors 'none'",
			},
			wantNotContain: []string{
				cryptoutilSharedMagic.DefaultOTLPHostnameDefault,
				cryptoutilSharedMagic.IPv4Loopback,
			},
		},
		{
			name: "production mode",
			settings: &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				DevMode: false,
			},
			wantContains: []string{
				"default-src 'none'",
			},
			wantNotContain: []string{
				cryptoutilSharedMagic.DefaultOTLPHostnameDefault,
				cryptoutilSharedMagic.IPv4Loopback,
			},
		},
		{
			name: "dev mode",
			settings: &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				DevMode: true,
			},
			wantContains: []string{
				"default-src 'none'",
				cryptoutilSharedMagic.DefaultOTLPHostnameDefault,
				cryptoutilSharedMagic.IPv4Loopback,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			csp := buildContentSecurityPolicy(tt.settings)

			for _, want := range tt.wantContains {
				require.True(t, strings.Contains(csp, want), "CSP should contain %q", want)
			}

			for _, notWant := range tt.wantNotContain {
				require.False(t, strings.Contains(csp, notWant), "CSP should not contain %q", notWant)
			}
		})
	}
}

func TestIsNonBrowserAPIRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		path           string
		wantNonBrowser bool
	}{
		{
			name:           "browser path",
			path:           "/browser/api/v1/users",
			wantNonBrowser: false,
		},
		{
			name:           "service path",
			path:           "/service/api/v1/keys",
			wantNonBrowser: true,
		},
		{
			name:           "oauth2 path",
			path:           "/oauth2/v1/token",
			wantNonBrowser: true,
		},
		{
			name:           "openid path",
			path:           "/openid/v1/userinfo",
			wantNonBrowser: true,
		},
		{
			name:           "well-known path",
			path:           cryptoutilSharedMagic.PathDiscovery,
			wantNonBrowser: true,
		},
		{
			name:           "root path",
			path:           "/",
			wantNonBrowser: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				PublicServiceAPIContextPath: "/service/",
			}
			checkFunc := isNonBrowserAPIRequest(settings)

			app := fiber.New(fiber.Config{DisableStartupMessage: true})

			var result bool

			app.Get(tt.path, func(c *fiber.Ctx) error {
				result = checkFunc(c)

				return c.SendString("OK")
			})

			req := httptest.NewRequest("GET", tt.path, nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { require.NoError(t, resp.Body.Close()) }()

			require.Equal(t, tt.wantNonBrowser, result)
		})
	}
}

func TestExpectedBrowserHeaders(t *testing.T) {
	t.Parallel()

	headers := ExpectedBrowserHeaders()

	require.NotEmpty(t, headers)
	require.Contains(t, headers, "X-Content-Type-Options")
	require.Contains(t, headers, "Referrer-Policy")
	require.Contains(t, headers, "Permissions-Policy")
	require.Contains(t, headers, "Cross-Origin-Opener-Policy")
	require.Contains(t, headers, "Cross-Origin-Embedder-Policy")
	require.Contains(t, headers, "Cross-Origin-Resource-Policy")
	require.Contains(t, headers, "X-Permitted-Cross-Domain-Policies")
}

func TestSecurityHeadersConfig_BrowserSecurityMiddlewares(t *testing.T) {
	t.Parallel()

	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		PublicServiceAPIContextPath: "/service/",
		PublicBrowserAPIContextPath: "/browser/",
	}
	config := NewDefaultSecurityHeadersConfig(settings)

	middlewares := config.BrowserSecurityMiddlewares()

	require.Len(t, middlewares, 2)
	require.NotNil(t, middlewares[0])
	require.NotNil(t, middlewares[1])
}
