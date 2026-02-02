// Copyright (c) 2025 Justin Cranford
//
//

package builder

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"strings"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// SwaggerUIConfig holds configuration for Swagger UI setup.
type SwaggerUIConfig struct {
	// Username for HTTP Basic Authentication (empty means no auth).
	Username string
	// Password for HTTP Basic Authentication (empty means no auth).
	Password string
	// CSRFTokenName is the name of the CSRF token cookie.
	CSRFTokenName string
	// BrowserAPIContextPath is the base path for browser API endpoints.
	BrowserAPIContextPath string
	// OpenAPISpecJSON is the serialized OpenAPI specification.
	OpenAPISpecJSON []byte
}

// RegisterSwaggerUI sets up Swagger UI routes with optional Basic Auth protection.
// It registers:
// - /ui/swagger/doc.json - OpenAPI spec endpoint
// - /ui/swagger/* - Swagger UI interface
// - {BrowserAPIContextPath}/csrf-token - CSRF token endpoint for Swagger UI.
func RegisterSwaggerUI(app *fiber.App, cfg *SwaggerUIConfig) error {
	if cfg == nil {
		return fmt.Errorf("swagger UI config is required")
	}

	if len(cfg.OpenAPISpecJSON) == 0 {
		return fmt.Errorf("OpenAPI spec JSON is required")
	}

	authMiddleware := swaggerUIBasicAuthMiddleware(cfg.Username, cfg.Password)

	// Serve OpenAPI spec JSON.
	app.Get("/ui/swagger/doc.json", authMiddleware, func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")

		return c.Send(cfg.OpenAPISpecJSON)
	})

	// Serve Swagger UI.
	app.Get("/ui/swagger/*", authMiddleware, func(c *fiber.Ctx) error {
		swaggerHandler := swagger.New(swagger.Config{
			Title:                  "Cryptoutil API",
			URL:                    "/ui/swagger/doc.json",
			TryItOutEnabled:        true,
			DisplayRequestDuration: true,
			ShowCommonExtensions:   true,
			CustomScript:           swaggerUICustomCSRFScript(cfg.CSRFTokenName, cfg.BrowserAPIContextPath),
		})

		err := swaggerHandler(c)
		if err != nil {
			return err
		}

		// Ensure Content-Type includes charset for HTML responses to satisfy security scanners.
		if c.Get("Content-Type") == "text/html" {
			c.Set("Content-Type", "text/html; charset=utf-8")
		}

		return nil
	})

	// CSRF token endpoint for Swagger UI to fetch token name and configuration.
	csrfTokenEndpoint := cfg.BrowserAPIContextPath + "/csrf-token"
	app.Get(csrfTokenEndpoint, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":         "CSRF token set in cookie",
			"csrf_token_name": cfg.CSRFTokenName,
		})
	})

	return nil
}

// swaggerUIBasicAuthMiddleware creates HTTP Basic Auth middleware for Swagger UI.
// If username and password are both empty, authentication is skipped.
func swaggerUIBasicAuthMiddleware(username, password string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// If no username/password configured, skip authentication.
		if username == "" && password == "" {
			return c.Next()
		}

		// Check for Authorization header.
		auth := c.Get("Authorization")
		if auth == "" {
			c.Set("WWW-Authenticate", `Basic realm="Swagger UI"`)

			return c.Status(fiber.StatusUnauthorized).SendString("Authentication required")
		}

		// Parse Basic Auth.
		if !strings.HasPrefix(auth, "Basic ") {
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid authentication method")
		}

		encoded := strings.TrimPrefix(auth, "Basic ")

		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid authentication encoding")
		}

		credentials := string(decoded)

		colonIndex := strings.Index(credentials, ":")
		if colonIndex == -1 {
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid authentication format")
		}

		reqUsername := credentials[:colonIndex]
		reqPassword := credentials[colonIndex+1:]

		// Check credentials.
		if reqUsername != username || reqPassword != password {
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid credentials")
		}

		return c.Next()
	}
}

// swaggerUICustomCSRFScript generates JavaScript to inject CSRF tokens into Swagger UI requests.
func swaggerUICustomCSRFScript(csrfTokenName, browserAPIContextPath string) template.JS {
	csrfTokenEndpoint := browserAPIContextPath + "/csrf-token"

	//nolint:lll // JavaScript template requires long lines for readability.
	return template.JS(fmt.Sprintf(`
		// Wait for Swagger UI to fully load
		const interval = setInterval(function() {
			if (window.ui) {
				clearInterval(interval);

				let csrfTokenName = '%s';

				// Get CSRF configuration from server
				fetch('%s', {
					method: 'GET',
					credentials: 'same-origin'
				}).then(response => response.json())
				.then(data => {
					csrfTokenName = data.csrf_token_name || '%s';
					console.log('CSRF Configuration:', data);
				}).catch(err => {
					console.warn('Could not fetch CSRF config:', err);
				});

				// Get CSRF token from cookie
				function getCSRFToken() {
					const cookies = document.cookie.split(';');
					for (let i = 0; i < cookies.length; i++) {
						const cookie = cookies[i].trim();
						if (cookie.startsWith(csrfTokenName + '=')) {
							return cookie.substring((csrfTokenName + '=').length);
						}
					}
					return null;
				}

				// Make a GET request to trigger CSRF cookie creation if needed
				function ensureCSRFToken() {
					return new Promise((resolve) => {
						let token = getCSRFToken();
						if (token) {
							resolve(token);
							return;
						}
						fetch('%s', {
							method: 'GET',
							credentials: 'same-origin'
						}).then(() => {
							resolve(getCSRFToken());
						}).catch(() => {
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
						return ensureCSRFToken().then(token => {
							if (token) {
								options.headers['X-CSRF-Token'] = token;
							}
							return originalFetch.call(this, url, options);
						});
					}
					return originalFetch.call(this, url, options);
				};

				console.log('CSRF token handling enabled for Swagger UI');
			}
		}, 100);
	`, csrfTokenName, csrfTokenEndpoint, csrfTokenName, csrfTokenEndpoint))
}
