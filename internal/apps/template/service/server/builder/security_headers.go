// Package builder provides a fluent API for constructing service applications.
package builder

import (
	"fmt"
	"strings"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// SecurityHeadersConfig contains configuration for security headers middleware.
type SecurityHeadersConfig struct {
	// Settings provides access to service-level configuration (dev mode, verbose mode, etc.).
	Settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings

	// CSP overrides - if set, these replace the default CSP directives.
	CustomCSP string

	// DisableHelmet disables the helmet middleware entirely.
	DisableHelmet bool

	// DisableAdditionalHeaders disables the additional security headers middleware.
	DisableAdditionalHeaders bool
}

// NewDefaultSecurityHeadersConfig creates a SecurityHeadersConfig with sensible defaults.
func NewDefaultSecurityHeadersConfig(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) *SecurityHeadersConfig {
	return &SecurityHeadersConfig{
		Settings: settings,
	}
}

// CreateHelmetMiddleware creates a helmet middleware configured for browser requests.
// This provides CSP, X-Frame-Options, XSS-Protection, and Referrer-Policy headers.
func (c *SecurityHeadersConfig) CreateHelmetMiddleware() fiber.Handler {
	if c.DisableHelmet {
		return func(ctx *fiber.Ctx) error {
			return ctx.Next()
		}
	}

	csp := c.CustomCSP
	if csp == "" {
		csp = buildContentSecurityPolicy(c.Settings)
	}

	return helmet.New(helmet.Config{
		Next:                  isNonBrowserAPIRequest(c.Settings),
		ContentSecurityPolicy: csp,
		XFrameOptions:         "DENY",
		XSSProtection:         "1; mode=block",
		ReferrerPolicy:        cryptoutilSharedMagic.CrossOriginOpenerPolicy,
	})
}

// CreateAdditionalSecurityHeadersMiddleware creates middleware for headers not covered by Helmet.
// This includes HSTS, Permissions-Policy, Cross-Origin-* headers, and Clear-Site-Data.
func (c *SecurityHeadersConfig) CreateAdditionalSecurityHeadersMiddleware() fiber.Handler {
	if c.DisableAdditionalHeaders {
		return func(ctx *fiber.Ctx) error {
			return ctx.Next()
		}
	}

	return func(ctx *fiber.Ctx) error {
		// Common security headers for all requests.
		ctx.Set("X-Content-Type-Options", cryptoutilSharedMagic.ContentTypeOptions)
		ctx.Set("Referrer-Policy", cryptoutilSharedMagic.ReferrerPolicy)

		// HSTS for HTTPS connections.
		if ctx.Protocol() == cryptoutilSharedMagic.ProtocolHTTPS {
			if c.Settings != nil && c.Settings.DevMode {
				ctx.Set("Strict-Transport-Security", cryptoutilSharedMagic.HSTSMaxAgeDev)
			} else {
				ctx.Set("Strict-Transport-Security", cryptoutilSharedMagic.HSTSMaxAge)
			}
		}

		// Browser-specific security headers.
		if !isNonBrowserAPIRequest(c.Settings)(ctx) {
			ctx.Set("Permissions-Policy", cryptoutilSharedMagic.PermissionsPolicy)
			ctx.Set("Cross-Origin-Opener-Policy", cryptoutilSharedMagic.CrossOriginOpenerPolicy)
			ctx.Set("Cross-Origin-Embedder-Policy", cryptoutilSharedMagic.CrossOriginEmbedderPolicy)
			ctx.Set("Cross-Origin-Resource-Policy", cryptoutilSharedMagic.CrossOriginResourcePolicy)
			ctx.Set("X-Permitted-Cross-Domain-Policies", cryptoutilSharedMagic.XPermittedCrossDomainPolicies)

			// Clear-Site-Data for logout endpoints.
			if ctx.Method() == fiber.MethodPost && strings.HasSuffix(ctx.OriginalURL(), cryptoutilSharedMagic.PathLogout) {
				ctx.Set("Clear-Site-Data", cryptoutilSharedMagic.ClearSiteDataLogout)
			}
		}

		return ctx.Next()
	}
}

// BrowserSecurityMiddlewares returns the full stack of security middlewares for browser paths.
func (c *SecurityHeadersConfig) BrowserSecurityMiddlewares() []fiber.Handler {
	return []fiber.Handler{
		c.CreateHelmetMiddleware(),
		c.CreateAdditionalSecurityHeadersMiddleware(),
	}
}

// buildContentSecurityPolicy creates a CSP tailored for the cryptoutil application.
// This CSP is specifically designed to work with Swagger UI while maintaining security.
func buildContentSecurityPolicy(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) string {
	// Base CSP - very restrictive.
	csp := "default-src 'none';"

	// Scripts: Allow self and necessary inline/eval for Swagger UI.
	// 'unsafe-inline' and 'unsafe-eval' are required for Swagger UI to function.
	csp += " script-src 'self' 'unsafe-inline' 'unsafe-eval';"

	// Styles: Allow self and inline styles (required for Swagger UI).
	csp += " style-src 'self' 'unsafe-inline';"

	// Images: Allow self and data URIs (for inline images/icons).
	csp += " img-src 'self' data:;"

	// Fonts: Allow self only.
	csp += " font-src 'self';"

	// Connections: Allow self for API calls.
	csp += " connect-src 'self';"

	// Forms: Allow self only.
	csp += " form-action 'self';"

	// Frames: Deny all framing (prevent clickjacking).
	csp += " frame-ancestors 'none';"

	// Base URI: Restrict to self.
	csp += " base-uri 'self';"

	// Object/embed: Block all plugins.
	csp += " object-src 'none';"

	// Media: Allow self for any video/audio.
	csp += " media-src 'self';"

	// Worker: Allow self for web workers.
	csp += " worker-src 'self';"

	// Manifest: Allow self for web app manifests.
	csp += " manifest-src 'self';"

	// In development mode, add localhost variations for flexible development.
	if settings != nil && settings.DevMode {
		localhostSources := " http://localhost:* https://localhost:* http://127.0.0.1:* https://127.0.0.1:*"
		csp = strings.ReplaceAll(csp, " 'self';", " 'self'"+localhostSources+";")

		if settings.VerboseMode {
			fmt.Printf("Content Security Policy (Dev Mode): %s\n", csp)
		}
	}

	return csp
}

// isNonBrowserAPIRequest returns a function that checks if a request is for a non-browser API.
// Non-browser APIs include /service/api/v1/*, /oauth2/v1/*, /openid/v1/*.
func isNonBrowserAPIRequest(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) func(*fiber.Ctx) bool {
	servicePrefix := "/service/"
	if settings != nil && settings.PublicServiceAPIContextPath != "" {
		servicePrefix = settings.PublicServiceAPIContextPath
	}

	return func(c *fiber.Ctx) bool {
		path := c.Path()

		return strings.HasPrefix(path, servicePrefix) ||
			strings.HasPrefix(path, "/oauth2/") ||
			strings.HasPrefix(path, "/openid/") ||
			strings.HasPrefix(path, "/.well-known/")
	}
}

// ExpectedBrowserHeaders returns the expected browser security headers for validation.
func ExpectedBrowserHeaders() map[string]string {
	return map[string]string{
		"X-Content-Type-Options":            cryptoutilSharedMagic.ContentTypeOptions,
		"Referrer-Policy":                   cryptoutilSharedMagic.ReferrerPolicy,
		"Permissions-Policy":                cryptoutilSharedMagic.PermissionsPolicy,
		"Cross-Origin-Opener-Policy":        cryptoutilSharedMagic.CrossOriginOpenerPolicy,
		"Cross-Origin-Embedder-Policy":      cryptoutilSharedMagic.CrossOriginEmbedderPolicy,
		"Cross-Origin-Resource-Policy":      cryptoutilSharedMagic.CrossOriginResourcePolicy,
		"X-Permitted-Cross-Domain-Policies": cryptoutilSharedMagic.XPermittedCrossDomainPolicies,
	}
}

// ValidateSecurityHeaders checks that all expected security headers are present in a response.
// Returns a list of missing or incorrect header names.
func ValidateSecurityHeaders(c *fiber.Ctx) []string {
	var missing []string

	for header, expectedValue := range ExpectedBrowserHeaders() {
		if actualValue := string(c.Response().Header.Peek(header)); actualValue != expectedValue {
			missing = append(missing, header)
		}
	}

	// Check HSTS is present if HTTPS.
	if c.Protocol() == cryptoutilSharedMagic.ProtocolHTTPS {
		if hsts := string(c.Response().Header.Peek("Strict-Transport-Security")); hsts == "" {
			missing = append(missing, "Strict-Transport-Security")
		}
	}

	return missing
}
