// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"fmt"
	"html/template"
	"strings"
	"sync"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedTelemetry "cryptoutil/internal/apps/template/service/telemetry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"go.opentelemetry.io/otel/metric"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/helmet"
)

func performConcurrentReadinessChecks(serverApplicationCore *ServerApplicationCore) map[string]any {
	results := make(map[string]any)

	// Channel to collect results
	resultsChan := make(chan struct {
		name   string
		result any
	})

	// WaitGroup to wait for all checks to complete
	var wg sync.WaitGroup

	// Helper function to perform a check and send the result to the channel
	doCheck := func(name string, checkFunc func() any) {
		defer wg.Done()

		result := checkFunc()
		resultsChan <- struct {
			name   string
			result any
		}{name, result}
	}

	// Number of concurrent readiness checks to perform.
	const numReadinessChecks = 4

	// Add readiness checks here
	wg.Add(numReadinessChecks)

	go doCheck("database", func() any {
		return checkDatabaseHealth(serverApplicationCore)
	})
	go doCheck("memory", func() any {
		return checkMemoryHealth()
	})
	go doCheck("sidecar", func() any {
		return checkSidecarHealth(serverApplicationCore)
	})
	go doCheck("dependencies", func() any {
		return checkDependenciesHealth(serverApplicationCore)
	})

	// Close the results channel once all checks are done
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results from the channel
	for result := range resultsChan {
		results[result.name] = result.result
	}

	return results
}

func commonSetFiberRequestAttribute(fiberAppIDValue fiberAppID) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Locals(cryptoutilSharedMagic.FiberAppIDRequestAttribute, string(fiberAppIDValue))

		return c.Next()
	}
}

func publicBrowserCORSMiddlewareFunction(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) fiber.Handler {
	return cors.New(cors.Config{ // Cross-Origin Resource Sharing (CORS)
		AllowOrigins: strings.Join(settings.CORSAllowedOrigins, ","), // cryptoutilConfig.defaultAllowedCORSOrigins
		AllowMethods: strings.Join(settings.CORSAllowedMethods, ","), // cryptoutilConfig.defaultAllowedCORSMethods
		AllowHeaders: strings.Join(settings.CORSAllowedHeaders, ","), // cryptoutilConfig.defaultAllowedCORSHeaders
		MaxAge:       int(settings.CORSMaxAge),
		Next:         isNonBrowserUserAPIRequestFunc(settings), // Skip CORS for /service/api/v1/*, /oauth2/v1/*, /openid/v1/* (non-browser clients)
	})
}

func publicBrowserXSSMiddlewareFunction(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) fiber.Handler {
	// Content Security Policy for enhanced XSS protection
	// This CSP is specifically designed to work with Swagger UI while maintaining security
	csp := buildContentSecurityPolicy(settings)

	return helmet.New(helmet.Config{
		Next: isNonBrowserUserAPIRequestFunc(settings), // Skip XSS check for /service/api/v1/*, /oauth2/v1/*, /openid/v1/* (non-browser clients)

		// Content Security Policy implementation
		ContentSecurityPolicy: csp,

		// Additional security headers (using available Helmet fields)
		XFrameOptions: "DENY",          // Prevent clickjacking
		XSSProtection: "1; mode=block", // Enable XSS filter

		// Allow same-origin referrers for CSRF protection
		ReferrerPolicy: "same-origin",
	})
}

// buildContentSecurityPolicy creates a CSP tailored for the cryptoutil application.
func buildContentSecurityPolicy(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) string {
	// Base CSP - very restrictive
	csp := "default-src 'none';"

	// Scripts: Allow self and necessary inline/eval for Swagger UI
	// 'unsafe-inline' and 'unsafe-eval' are required for Swagger UI to function
	csp += " script-src 'self' 'unsafe-inline' 'unsafe-eval';"

	// Styles: Allow self and inline styles (required for Swagger UI)
	csp += " style-src 'self' 'unsafe-inline';"

	// Images: Allow self and data URIs (for inline images/icons)
	csp += " img-src 'self' data:;"

	// Fonts: Allow self only
	csp += " font-src 'self';"

	// Connections: Allow self for API calls
	csp += " connect-src 'self';"

	// Forms: Allow self only
	csp += " form-action 'self';"

	// Frames: Deny all framing (prevent clickjacking)
	csp += " frame-ancestors 'none';"

	// Base URI: Restrict to self
	csp += " base-uri 'self';"

	// Object/embed: Block all plugins
	csp += " object-src 'none';"

	// Media: Allow self for any video/audio
	csp += " media-src 'self';"

	// Worker: Allow self for web workers
	csp += " worker-src 'self';"

	// Manifest: Allow self for web app manifests
	csp += " manifest-src 'self';"

	// In development mode, add localhost variations for flexible development
	if settings.DevMode {
		// Add localhost variations for development
		localhostSources := " http://localhost:* https://localhost:* http://127.0.0.1:* https://127.0.0.1:*"
		csp = strings.ReplaceAll(csp, " 'self';", " 'self'"+localhostSources+";")

		// Log CSP in development mode for debugging
		if settings.VerboseMode {
			fmt.Printf("Content Security Policy (Dev Mode): %s\n", csp)
		}
	}

	return csp
}

// Security header policy constants - Last reviewed: 2025-10-01.
const (
	hstsMaxAge                    = cryptoutilSharedMagic.HSTSMaxAge
	hstsMaxAgeDev                 = cryptoutilSharedMagic.HSTSMaxAgeDev
	referrerPolicy                = cryptoutilSharedMagic.ReferrerPolicy
	permissionsPolicy             = cryptoutilSharedMagic.PermissionsPolicy
	crossOriginOpenerPolicy       = cryptoutilSharedMagic.CrossOriginOpenerPolicy
	crossOriginEmbedderPolicy     = cryptoutilSharedMagic.CrossOriginEmbedderPolicy
	crossOriginResourcePolicy     = cryptoutilSharedMagic.CrossOriginResourcePolicy
	xPermittedCrossDomainPolicies = cryptoutilSharedMagic.XPermittedCrossDomainPolicies
	contentTypeOptions            = cryptoutilSharedMagic.ContentTypeOptions
	clearSiteDataLogout           = cryptoutilSharedMagic.ClearSiteDataLogout
)

// Expected browser security headers for runtime validation.
var expectedBrowserHeaders = map[string]string{
	"X-Content-Type-Options":            cryptoutilSharedMagic.ContentTypeOptions,
	"Referrer-Policy":                   cryptoutilSharedMagic.ReferrerPolicy,
	"Permissions-Policy":                cryptoutilSharedMagic.PermissionsPolicy,
	"Cross-Origin-Opener-Policy":        cryptoutilSharedMagic.CrossOriginOpenerPolicy,
	"Cross-Origin-Embedder-Policy":      cryptoutilSharedMagic.CrossOriginEmbedderPolicy,
	"Cross-Origin-Resource-Policy":      cryptoutilSharedMagic.CrossOriginResourcePolicy,
	"X-Permitted-Cross-Domain-Policies": cryptoutilSharedMagic.XPermittedCrossDomainPolicies,
}

// publicBrowserAdditionalSecurityHeadersMiddleware adds security headers not covered by Helmet.
func publicBrowserAdditionalSecurityHeadersMiddleware(telemetryService *cryptoutilSharedTelemetry.TelemetryService, settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) fiber.Handler {
	// Setup metrics for header validation
	meter := telemetryService.MetricsProvider.Meter("security-headers")

	missingHeaderCounter, err := meter.Int64Counter(
		"security_headers_missing_total",
		metric.WithDescription("Number of requests with missing expected security headers"),
		metric.WithUnit("1"),
	)
	if err != nil {
		telemetryService.Slogger.Error("Failed to create security headers metric", "error", err)
	}

	// Log active security policy on startup
	logger := telemetryService.Slogger.With("component", "security-headers")
	logger.Debug("Active browser security header policy",
		"referrer_policy", referrerPolicy,
		"permissions_policy", permissionsPolicy,
		"isolation_enabled", true,
		"hsts_preload", !settings.DevMode,
		"clear_site_data_logout", true,
	)

	return func(c *fiber.Ctx) error {
		// Apply common security headers to all requests
		c.Set("X-Content-Type-Options", contentTypeOptions)
		c.Set("Referrer-Policy", referrerPolicy)

		if c.Protocol() == cryptoutilSharedMagic.ProtocolHTTPS {
			if settings.DevMode {
				c.Set("Strict-Transport-Security", hstsMaxAgeDev)
			} else {
				c.Set("Strict-Transport-Security", hstsMaxAge)
			}
		}

		// Skip browser-specific headers for non-browser API requests
		if !isNonBrowserUserAPIRequestFunc(settings)(c) {
			// Apply browser-specific security headers
			c.Set("Permissions-Policy", permissionsPolicy)
			c.Set("Cross-Origin-Opener-Policy", crossOriginOpenerPolicy)
			c.Set("Cross-Origin-Embedder-Policy", crossOriginEmbedderPolicy)
			c.Set("Cross-Origin-Resource-Policy", crossOriginResourcePolicy)
			c.Set("X-Permitted-Cross-Domain-Policies", xPermittedCrossDomainPolicies)

			// Clear-Site-Data for logout endpoints only
			if c.Method() == fiber.MethodPost && strings.HasSuffix(c.OriginalURL(), "/logout") {
				c.Set("Clear-Site-Data", clearSiteDataLogout)
			}
		}

		// Process the request
		err := c.Next()

		// Runtime self-check: validate expected headers are present in response (only for browser requests)
		if !isNonBrowserUserAPIRequestFunc(settings)(c) {
			missingHeaders := validateSecurityHeaders(c)
			if len(missingHeaders) > 0 {
				logger.Warn("Security headers missing in response",
					"missing_headers", missingHeaders,
					"request_path", c.OriginalURL(),
					"request_id", c.Locals("requestid"),
				)
				// Increment metric for missing headers
				if missingHeaderCounter != nil {
					missingHeaderCounter.Add(c.UserContext(), int64(len(missingHeaders)))
				}
			}
		}

		// Return the error from c.Next() - in Fiber middleware, errors from c.Next() should be returned as-is
		// to maintain the middleware chain behavior
		return err //nolint:wrapcheck
	}
}

// validateSecurityHeaders checks that all expected security headers are present.
func validateSecurityHeaders(c *fiber.Ctx) []string {
	var missing []string

	for header, expectedValue := range expectedBrowserHeaders {
		if actualValue := c.Get(header); actualValue != expectedValue {
			missing = append(missing, header)
		}
	}

	// Check HSTS is present if HTTPS
	if c.Protocol() == cryptoutilSharedMagic.ProtocolHTTPS {
		if hsts := c.Get("Strict-Transport-Security"); hsts == "" {
			missing = append(missing, "Strict-Transport-Security")
		}
	}

	return missing
}

func publicBrowserCSRFMiddlewareFunction(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) fiber.Handler {
	csrfConfig := csrf.Config{
		CookieName:        settings.CSRFTokenName,
		CookieSameSite:    settings.CSRFTokenSameSite,
		Expiration:        settings.CSRFTokenMaxAge,
		CookieSecure:      settings.CSRFTokenCookieSecure,
		CookieHTTPOnly:    settings.CSRFTokenCookieHTTPOnly,
		CookieSessionOnly: settings.CSRFTokenCookieSessionOnly,
		SingleUseToken:    settings.CSRFTokenSingleUseToken,
		Next:              isNonBrowserUserAPIRequestFunc(settings), // Skip CSRF for /service/api/v1/*, /oauth2/v1/*, /openid/v1/* (non-browser clients)
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if settings.DevMode {
				cookieToken := c.Cookies(settings.CSRFTokenName)

				headerToken := c.Get("X-CSRF-Token")
				if headerToken == "" {
					headerToken = c.Get("X-Csrf-Token")
				}

				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error":           "CSRF token validation failed",
					"details":         err.Error(),
					"url":             c.OriginalURL(),
					"method":          c.Method(),
					"headers":         c.GetReqHeaders(),
					"cookies":         c.GetReqHeaders()["Cookie"],
					"csrf_token_name": settings.CSRFTokenName,
					"origin":          c.Get("Origin"),
					"referer":         c.Get("Referer"),
					"cookie_token":    cookieToken,
					"header_token":    headerToken,
					"tokens_match":    cookieToken == headerToken,
					"user_agent":      c.Get("User-Agent"),
					"content_type":    c.Get("Content-Type"),
					"request_id":      c.Locals("requestid"),
				})
			}

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "CSRF token validation failed",
			})
		},
	}

	return csrf.New(csrfConfig)
}

// TRUE  => Skip CSRF check for /service/api/v1/* requests by non-browser clients (e.g. curl, Postman, service-to-service calls)
// ASSUME: Non-browser Authentication only authorizes clients to access /service/api/v1/*
// TRUE  => Skip CSRF check for /oauth2/v1/* and /openid/v1/* OAuth 2.1 endpoints (machine-to-machine, never browser-based)
// FALSE => Enforce CSRF check for /browser/api/v1/* requests by browser clients (e.g. web apps, Swagger UI)
// ASSUME: UI Authentication only authorizes browser users to access /browser/api/v1/*.
func isNonBrowserUserAPIRequestFunc(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) func(c *fiber.Ctx) bool {
	return func(c *fiber.Ctx) bool {
		url := c.OriginalURL()

		return strings.HasPrefix(url, settings.PublicServiceAPIContextPath+"/") ||
			strings.HasPrefix(url, "/oauth2/v1/") ||
			strings.HasPrefix(url, "/openid/v1/")
	}
}

func swaggerUICustomCSRFScript(csrfTokenName, browserAPIContextPath string) template.JS {
	csrfTokenEndpoint := browserAPIContextPath + "/csrf-token"

	return template.JS(fmt.Sprintf(`
		// Wait for Swagger UI to fully load
		const interval = setInterval(function() {
			if (window.ui) {
				clearInterval(interval);

				let csrfTokenName = '%s'; // Use actual CSRF token name from settings

				// Get CSRF configuration from server
				fetch('%s', {
					method: 'GET',
					credentials: 'same-origin'
				}).then(response => response.json())
				.then(data => {
					csrfTokenName = data.csrf_token_name || '%s';
					console.log('CSRF Configuration:', data);
					console.log('Using CSRF token name:', csrfTokenName);
				}).catch(err => {
					console.warn('Could not fetch CSRF config:', err);
				});

				// Get CSRF token from cookie
				function getCSRFToken() {
					const cookies = document.cookie.split(';');
					console.log('All cookies:', document.cookie);
					for (let i = 0; i < cookies.length; i++) {
						const cookie = cookies[i].trim();
						if (cookie.startsWith(csrfTokenName + '=')) {
							const token = cookie.substring((csrfTokenName + '=').length);
							console.log('Found CSRF token:', token);
							return token;
						}
					}
					console.log('No CSRF token found in cookies');
					return null;
				}

				// Make a GET request to trigger CSRF cookie creation if needed
				function ensureCSRFToken() {
					return new Promise((resolve) => {
						let token = getCSRFToken();
						if (token) {
							console.log('CSRF token already available:', token);
							resolve(token);
							return;
						}

						console.log('Making request to get CSRF token...');
						// Make a GET request to trigger CSRF cookie creation
						fetch('%s', {
							method: 'GET',
							credentials: 'same-origin'
						}).then(() => {
							console.log('CSRF token request completed, checking cookies...');
							token = getCSRFToken();
							if (token) {
								console.log('CSRF token retrieved:', token);
							} else {
								console.warn('CSRF token still not available after request');
							}
							resolve(token);
						}).catch(err => {
							console.error('Failed to get CSRF token:', err);
							resolve(null);
						});
					});
				}

				// Add CSRF token to all non-GET requests
				const originalFetch = window.fetch;
				window.fetch = function(url, options) {
					options = options || {};

					if (options && options.method && options.method !== 'GET') {
						options.headers = options.headers || {};
						options.credentials = options.credentials || 'same-origin';

						console.log('Intercepted non-GET request:', options.method, url);

						// Get CSRF token and add to headers
						return ensureCSRFToken().then(token => {
							if (token) {
								options.headers['X-CSRF-Token'] = token;
								console.log('Added CSRF token to request headers:', options.method, url);
								console.log('Request headers:', options.headers);
							} else {
								console.error('No CSRF token available for request:', options.method, url);
							}
							return originalFetch.call(this, url, options);
						});
					}
					return originalFetch.call(this, url, options);
				};

				console.log('Enhanced CSRF token handling enabled for Swagger UI');
			}
		}, 100);
	`, csrfTokenName, csrfTokenEndpoint, csrfTokenName, csrfTokenEndpoint))
}
