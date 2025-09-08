package application

import (
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilCertificate "cryptoutil/internal/common/crypto/certificate"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"
	cryptoutilOpenapiHandler "cryptoutil/internal/server/handler"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/swagger"
	fibermiddleware "github.com/oapi-codegen/fiber-middleware"
)

const clientShutdownRequestTimeout = 5 * time.Second
const clientLivenessStartTimeout = 200 * time.Millisecond
const clientLivenessRequestTimeout = 3 * time.Second

const serverShutdownFinishTimeout = 5 * time.Second
const serverShutdownRequestPath = "/shutdown"

const fiberAppIDRequestAttribute = "fiberAppID"

type fiberAppID string

const (
	fiberAppIDPublic  fiberAppID = "public"
	fiberAppIDPrivate fiberAppID = "private"
)

var ready atomic.Bool

func SendServerListenerShutdownRequest(settings *cryptoutilConfig.Settings) error {
	privateBaseURL := fmt.Sprintf("%s://%s:%d", settings.BindPrivateProtocol, settings.BindPrivateAddress, settings.BindPrivatePort)
	shutdownRequestCtx, shutdownRequestCancel := context.WithTimeout(context.Background(), clientShutdownRequestTimeout)
	defer shutdownRequestCancel()
	shutdownRequest, err := http.NewRequestWithContext(shutdownRequestCtx, http.MethodPost, privateBaseURL+serverShutdownRequestPath, nil)
	if err != nil {
		return fmt.Errorf("failed to create shutdown request: %w", err)
	}

	// TODO Only use InsecureSkipVerify for DevMode
	// Create HTTP client that accepts self-signed certificates for local testing
	var client *http.Client
	if settings.BindPrivateProtocol == "https" {
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	} else {
		client = http.DefaultClient
	}

	shutdownResponse, err := client.Do(shutdownRequest)
	if err != nil {
		return fmt.Errorf("failed to send shutdown request: %w", err)
	} else if shutdownResponse.StatusCode != http.StatusOK {
		shutdownResponseBody, err := io.ReadAll(shutdownResponse.Body)
		defer shutdownResponse.Body.Close()
		if err != nil {
			return fmt.Errorf("shutdown request failed: %s (could not read response body: %v)", shutdownResponse.Status, err)
		}
		return fmt.Errorf("shutdown request failed, status: %s, body: %s", shutdownResponse.Status, string(shutdownResponseBody))
	}

	time.Sleep(clientLivenessStartTimeout)
	livenessRequestCtx, livenessRequestCancel := context.WithTimeout(context.Background(), clientLivenessRequestTimeout)
	defer livenessRequestCancel()
	livenessRequest, _ := http.NewRequestWithContext(livenessRequestCtx, http.MethodGet, privateBaseURL+"/livez", nil)
	livenessResponse, err := client.Do(livenessRequest)
	if err == nil && livenessResponse != nil {
		livenessResponse.Body.Close()
		return fmt.Errorf("server did not shut down properly")
	}
	return nil
}

// TODO Refactor ServerApplicationBasic vs StartServerApplicationCore vs StartServerListenerApplication?
func StartServerListenerApplication(settings *cryptoutilConfig.Settings) (func(), func(), error) {
	ctx := context.Background()

	serverApplicationCore, err := StartServerApplicationCore(ctx, settings)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize server application: %w", err)
	}

	var publicTLSServerConfig *tls.Config
	var privateTLSServerConfig *tls.Config
	if settings.BindPublicProtocol == "https" || settings.BindPrivateProtocol == "https" {
		publicTLSServerSubject, privateTLSServerSubject, err := generateTLSServerSubjects(settings, serverApplicationCore.ServerApplicationBasic)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to run new function: %w", err)
		}

		publicTLSServerCertificate, _, _, err := cryptoutilCertificate.BuildTLSCertificate(publicTLSServerSubject)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to build TLS server certificate: %w", err)
		}
		privateTLSServerCertificate, _, _, err := cryptoutilCertificate.BuildTLSCertificate(privateTLSServerSubject)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to build TLS client certificate: %w", err)
		}

		publicTLSServerConfig = &tls.Config{Certificates: []tls.Certificate{publicTLSServerCertificate}, ClientAuth: tls.NoClientCert}
		privateTLSServerConfig = &tls.Config{Certificates: []tls.Certificate{privateTLSServerCertificate}, ClientAuth: tls.NoClientCert}
	}

	// Middlewares

	commonMiddlewares := []fiber.Handler{
		recover.New(),
		requestid.New(),
		logger.New(), // TODO Replace this with improved otelFiberTelemetryMiddleware; unstructured logs and no OpenTelemetry are undesirable
		commonOtelFiberTelemetryMiddleware(serverApplicationCore.ServerApplicationBasic.TelemetryService, settings),
		commonOtelFiberRequestLoggerMiddleware(serverApplicationCore.ServerApplicationBasic.TelemetryService),
		commonIPFilterMiddleware(serverApplicationCore.ServerApplicationBasic.TelemetryService, settings),
		commonIPRateLimiterMiddleware(serverApplicationCore.ServerApplicationBasic.TelemetryService, settings),
		commonHTTPGETCacheControlMiddleware(), // TODO Limit this to Swagger GET APIs, not Swagger UI static content
	}

	privateMiddlewares := append([]fiber.Handler{commonSetFiberRequestAttribute(fiberAppIDPrivate)}, commonMiddlewares...)
	privateMiddlewares = append(privateMiddlewares, privateHealthCheckMiddlewareFunction()) // /livez, /readyz
	privateFiberApp := fiber.New(fiber.Config{Immutable: true})
	for _, middleware := range privateMiddlewares {
		privateFiberApp.Use(middleware)
	}

	publicMiddlewares := append([]fiber.Handler{commonSetFiberRequestAttribute(fiberAppIDPublic)}, commonMiddlewares...)
	publicMiddlewares = append(publicMiddlewares, publicBrowserCORSMiddlewareFunction(settings)) // Browser-specific: Cross-Origin Resource Sharing (CORS)
	publicMiddlewares = append(publicMiddlewares, publicBrowserXSSMiddlewareFunction(settings))  // Browser-specific: Cross-Site Scripting (XSS)
	publicMiddlewares = append(publicMiddlewares, publicBrowserCSRFMiddlewareFunction(settings)) // Browser-specific: Cross-Site Request Forgery (CSRF)
	publicFiberApp := fiber.New(fiber.Config{Immutable: true})
	for _, middleware := range publicMiddlewares {
		publicFiberApp.Use(middleware)
	}

	// shutdownServerFunction stops privateFiberApp and publicFiberApp, it is called via /shutdown hosted by privateFiberApp
	var shutdownServerFunction func()

	// Private APIs
	privateFiberApp.Post(serverShutdownRequestPath, func(c *fiber.Ctx) error {
		serverApplicationCore.ServerApplicationBasic.TelemetryService.Slogger.Info("shutdown requested via API endpoint")
		if shutdownServerFunction != nil {
			defer func() {
				time.Sleep(clientLivenessStartTimeout) // allow server small amount of time to finish sending response to client
				shutdownServerFunction()
			}()
		}
		return c.SendString("Server shutdown initiated")
	})

	// Public Swagger UI
	swaggerApi, err := cryptoutilOpenapiServer.GetSwagger()
	if err != nil {
		serverApplicationCore.ServerApplicationBasic.TelemetryService.Slogger.Error("failed to get swagger", "error", err)
		serverApplicationCore.Shutdown()
		return nil, nil, fmt.Errorf("failed to get swagger: %w", err)
	}

	swaggerApi.Servers = []*openapi3.Server{
		{URL: settings.PublicBrowserAPIContextPath}, // Browser users will access the APIs via this context path, with browser middlewares (CORS, CSRF, etc)
		{URL: settings.PublicServiceAPIContextPath}, // Service clients will access the APIs via this context path, without browser middlewares
	}
	swaggerSpecBytes, err := swaggerApi.MarshalJSON() // Serialize OpenAPI 3 spec to JSON with the added public server context path
	if err != nil {
		serverApplicationCore.ServerApplicationBasic.TelemetryService.Slogger.Error("failed to get fiber handler for OpenAPI spec", "error", err)
		serverApplicationCore.Shutdown()
		return nil, nil, fmt.Errorf("failed to marshal OpenAPI spec: %w", err)
	}

	publicFiberApp.Get("/ui/swagger/doc.json", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.Send(swaggerSpecBytes)
	})
	publicFiberApp.Get("/ui/swagger/*", swagger.New(swagger.Config{
		Title:                  "Cryptoutil API",
		URL:                    "/ui/swagger/doc.json",
		TryItOutEnabled:        true,
		DisplayRequestDuration: true,
		ShowCommonExtensions:   true,
		CustomScript:           swaggerUICustomCSRFScript(settings.CSRFTokenName, settings.PublicBrowserAPIContextPath),
	}))
	publicFiberApp.Get(settings.PublicBrowserAPIContextPath+"/csrf-token", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":         "CSRF token set in cookie",
			"csrf_token_name": settings.CSRFTokenName,
			"cookie_secure":   settings.CSRFTokenCookieSecure,
			"same_site":       settings.CSRFTokenSameSite,
		})
	})

	// Swagger APIs, will be double exposed on publicFiberApp, but with different security middlewares (i.e. browser user vs machine client)
	openapiStrictServer := cryptoutilOpenapiHandler.NewOpenapiStrictServer(serverApplicationCore.BusinessLogicService)
	openapiStrictHandler := cryptoutilOpenapiServer.NewStrictHandler(openapiStrictServer, nil)
	commonOapiMiddlewareFiberRequestValidators := []cryptoutilOpenapiServer.MiddlewareFunc{
		fibermiddleware.OapiRequestValidatorWithOptions(swaggerApi, &fibermiddleware.Options{}),
	}
	publicBrowserFiberServerOptions := cryptoutilOpenapiServer.FiberServerOptions{
		BaseURL:     settings.PublicBrowserAPIContextPath,
		Middlewares: commonOapiMiddlewareFiberRequestValidators,
	}
	publicServiceFiberServerOptions := cryptoutilOpenapiServer.FiberServerOptions{
		BaseURL:     settings.PublicServiceAPIContextPath,
		Middlewares: commonOapiMiddlewareFiberRequestValidators,
	}
	cryptoutilOpenapiServer.RegisterHandlersWithOptions(publicFiberApp, openapiStrictHandler, publicBrowserFiberServerOptions)
	cryptoutilOpenapiServer.RegisterHandlersWithOptions(publicFiberApp, openapiStrictHandler, publicServiceFiberServerOptions)

	publicBinding := fmt.Sprintf("%s:%d", settings.BindPublicAddress, settings.BindPublicPort)
	privateBinding := fmt.Sprintf("%s:%d", settings.BindPrivateAddress, settings.BindPrivatePort)

	startServerFunction := startServerFunc(
		publicBinding, privateBinding,
		publicFiberApp, privateFiberApp,
		settings.BindPublicProtocol, settings.BindPrivateProtocol,
		publicTLSServerConfig, privateTLSServerConfig,
		serverApplicationCore.ServerApplicationBasic.TelemetryService)
	shutdownServerFunction = stopServerFunc(serverApplicationCore, publicFiberApp, privateFiberApp)

	go stopServerSignalFunc(serverApplicationCore.ServerApplicationBasic.TelemetryService, shutdownServerFunction)() // listen for OS signals to gracefully shutdown the server

	return startServerFunction, shutdownServerFunction, nil
}

func startServerFunc(publicBinding string, privateBinding string, publicFiberApp *fiber.App, privateFiberApp *fiber.App, publicProtocol string, privateProtocol string, publicTLSConfig *tls.Config, privateTLSConfig *tls.Config, telemetryService *cryptoutilTelemetry.TelemetryService) func() {
	return func() {
		telemetryService.Slogger.Debug("starting fiber listeners")

		go func() {
			telemetryService.Slogger.Debug("starting private fiber listener", "binding", privateBinding, "protocol", privateProtocol)
			var err error
			if privateProtocol == "https" && privateTLSConfig != nil {
				err = privateFiberApp.ListenTLSWithCertificate(privateBinding, privateTLSConfig.Certificates[0])
			} else {
				err = privateFiberApp.Listen(privateBinding)
			}
			if err != nil {
				telemetryService.Slogger.Error("failed to start private fiber listener", "error", err)
			}
			telemetryService.Slogger.Debug("private fiber listener stopped")
		}()

		telemetryService.Slogger.Debug("starting public fiber listener", "binding", publicBinding, "protocol", publicProtocol)
		var err error
		if publicProtocol == "https" && publicTLSConfig != nil {
			err = publicFiberApp.ListenTLSWithCertificate(publicBinding, publicTLSConfig.Certificates[0])
		} else {
			err = publicFiberApp.Listen(publicBinding)
		}
		if err != nil {
			telemetryService.Slogger.Error("failed to start public fiber listener", "error", err)
		}
		telemetryService.Slogger.Debug("public fiber listener stopped")

		ready.Store(true)
	}
}

func stopServerFunc(serverApplicationCore *ServerApplicationCore, publicFiberApp *fiber.App, privateFiberApp *fiber.App) func() {
	return func() {
		if serverApplicationCore.ServerApplicationBasic.TelemetryService != nil {
			serverApplicationCore.ServerApplicationBasic.TelemetryService.Slogger.Debug("stopping servers")
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), serverShutdownFinishTimeout)
		defer cancel() // perform shutdown respecting timeout

		if publicFiberApp != nil {
			serverApplicationCore.ServerApplicationBasic.TelemetryService.Slogger.Debug("shutting down public fiber app")
			if err := publicFiberApp.ShutdownWithContext(shutdownCtx); err != nil {
				serverApplicationCore.ServerApplicationBasic.TelemetryService.Slogger.Error("failed to stop public fiber server", "error", err)
			}
		}
		if privateFiberApp != nil {
			serverApplicationCore.ServerApplicationBasic.TelemetryService.Slogger.Debug("shutting down private fiber app")
			if err := privateFiberApp.ShutdownWithContext(shutdownCtx); err != nil {
				serverApplicationCore.ServerApplicationBasic.TelemetryService.Slogger.Error("failed to stop private fiber server", "error", err)
			}
		}
		serverApplicationCore.Shutdown()
	}
}

func stopServerSignalFunc(telemetryService *cryptoutilTelemetry.TelemetryService, stopServerFunc func()) func() {
	return func() {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
		defer stop()

		<-ctx.Done() // blocks until signal is received
		telemetryService.Slogger.Warn("received stop server signal")
		stopServerFunc()
	}
}

func commonOtelFiberTelemetryMiddleware(telemetryService *cryptoutilTelemetry.TelemetryService, settings *cryptoutilConfig.Settings) fiber.Handler {
	return otelfiber.Middleware(
		otelfiber.WithTracerProvider(telemetryService.TracesProvider),
		otelfiber.WithMeterProvider(telemetryService.MetricsProvider),
		otelfiber.WithPropagators(*telemetryService.TextMapPropagator),
		otelfiber.WithServerName(settings.BindPublicAddress),
		otelfiber.WithPort(int(settings.BindPublicPort)),
	)
}

func commonIPFilterMiddleware(telemetryService *cryptoutilTelemetry.TelemetryService, settings *cryptoutilConfig.Settings) func(c *fiber.Ctx) error {
	allowedIPs := make(map[string]bool)
	if settings.AllowedIPs != "" {
		for allowedIP := range strings.SplitSeq(settings.AllowedIPs, ",") {
			parsedIP := net.ParseIP(allowedIP) // IPv4 (e.g.  192.0.2.1"), IPv6 (e.g. 2001:db8::68), or IPv4-mapped IPv6 (e.g. ::ffff:192.0.2.1)
			if parsedIP == nil {
				telemetryService.Slogger.Error("invalid allowed IP address:", "IP", allowedIP)
			} else {
				allowedIPs[allowedIP] = true
				if settings.VerboseMode {
					telemetryService.Slogger.Debug("Parsed IP successfully", "IP", allowedIP, "parsed", parsedIP.String())
				}
			}
		}
	}

	var allowedCIDRs []*net.IPNet
	if settings.AllowedCIDRs != "" {
		for allowedCIDR := range strings.SplitSeq(settings.AllowedCIDRs, ",") {
			_, network, err := net.ParseCIDR(allowedCIDR) // "192.0.2.1/24" => 192.0.2.1 (not useful) and 192.0.2.0/24 (useful)
			if err != nil {
				telemetryService.Slogger.Error("invalid allowed CIDR:", "CIDR", allowedCIDR, "error", err)
			} else {
				allowedCIDRs = append(allowedCIDRs, network)
				if settings.VerboseMode {
					telemetryService.Slogger.Debug("Parsed CIDR successfully", "CIDR", allowedCIDR, "network", network.String())
				}
			}
		}
	}

	return func(c *fiber.Ctx) error { // Mitigate against DDOS by allowlisting IP addresses and CIDRs
		switch c.Locals(fiberAppIDRequestAttribute) {
		case string(fiberAppIDPublic): // Apply IP/CIDR filtering for public app requests
			clientIP := c.IP()
			parsedIP := net.ParseIP(clientIP)
			if parsedIP == nil {
				telemetryService.Slogger.Debug("invalid IP", "#", c.Locals("requestid"), "method", c.Method(), "IP", clientIP, "URL", c.OriginalURL(), "Headers", c.GetReqHeaders())
				return c.Status(fiber.StatusForbidden).SendString("Invalid IP format")
			} else if _, allowed := allowedIPs[parsedIP.String()]; allowed {
				if settings.VerboseMode {
					telemetryService.Slogger.Debug("Allowed IP:", "#", c.Locals("requestid"), "method", c.Method(), "IP", clientIP, "URL", c.OriginalURL(), "Headers", c.GetReqHeaders())
				}
				return c.Next() // IP is contained in the allowed IPs set
			}
			for _, allowedCIDR := range allowedCIDRs {
				if allowedCIDR.Contains(parsedIP) {
					if settings.VerboseMode {
						telemetryService.Slogger.Debug("Allowed CIDR:", "#", c.Locals("requestid"), "method", c.Method(), "IP", clientIP, "URL", c.OriginalURL(), "Headers", c.GetReqHeaders())
					}
					return c.Next() // IP is contained in the allowed CIDRs list
				}
			}
			telemetryService.Slogger.Debug("Access denied:", "#", c.Locals("requestid"), "method", c.Method(), "IP", clientIP, "URL", c.OriginalURL(), "Headers", c.GetReqHeaders())
			return c.Status(fiber.StatusForbidden).SendString("Access denied: IP not allowed")
		case string(fiberAppIDPrivate): // Skip IP/CIDR filtering for private app requests
			return c.Next()
		default:
			telemetryService.Slogger.Error("Unexpected app ID:", c.Locals(fiberAppIDRequestAttribute))
			return c.Status(fiber.StatusInternalServerError).SendString("Internal server error")
		}
	}
}

func commonIPRateLimiterMiddleware(telemetryService *cryptoutilTelemetry.TelemetryService, settings *cryptoutilConfig.Settings) fiber.Handler {
	return limiter.New(limiter.Config{ // Mitigate DOS by throttling clients
		Max:        int(settings.IPRateLimit),
		Expiration: time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() // throttle by IP, could be improved in future (e.g. append JWTClaim.sub or JWTClaim.tenantid)
		},
		LimitReached: func(c *fiber.Ctx) error {
			telemetryService.Slogger.Warn("Rate limit exceeded", "requestid", c.Locals("requestid"), "method", c.Method(), "IP", c.IP(), "URL", c.OriginalURL(), "Headers", c.GetReqHeaders())
			return c.Status(fiber.StatusTooManyRequests).SendString("Rate limit exceeded")
		},
	})
}

func commonHTTPGETCacheControlMiddleware() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error { // Disable caching of HTTP GET responses
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		return c.Next()
	}
}

func privateHealthCheckMiddlewareFunction() fiber.Handler {
	return healthcheck.New()
}

func commonSetFiberRequestAttribute(fiberAppIdValue fiberAppID) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Locals(fiberAppIDRequestAttribute, string(fiberAppIdValue))
		return c.Next()
	}
}

func publicBrowserCORSMiddlewareFunction(settings *cryptoutilConfig.Settings) fiber.Handler {
	return cors.New(cors.Config{ // Cross-Origin Resource Sharing (CORS)
		AllowOrigins: settings.CORSAllowedOrigins, // cryptoutilConfig.defaultAllowedCORSOrigins
		AllowMethods: settings.CORSAllowedMethods, // cryptoutilConfig.defaultAllowedCORSMethods
		AllowHeaders: settings.CORSAllowedHeaders, // cryptoutilConfig.defaultAllowedCORSHeaders
		MaxAge:       int(settings.CORSMaxAge),
		Next:         isNonBrowserUserApiRequestFunc(settings), // Skip check for /service/api/v1/* requests by non-browser clients
	})
}

func publicBrowserXSSMiddlewareFunction(settings *cryptoutilConfig.Settings) fiber.Handler {
	return helmet.New(helmet.Config{
		Next: isNonBrowserUserApiRequestFunc(settings), // Skip check for /service/api/v1/* requests by non-browser clients
		// TODO Check if this is sufficient, or if we need to add Content-Security-Policy (CSP) headers
		// Allow same-origin referrers for CSRF protection
		ReferrerPolicy: "same-origin",
	})
}

func publicBrowserCSRFMiddlewareFunction(settings *cryptoutilConfig.Settings) fiber.Handler {
	csrfConfig := csrf.Config{
		CookieName:        settings.CSRFTokenName,
		CookieSameSite:    settings.CSRFTokenSameSite,
		Expiration:        settings.CSRFTokenMaxAge,
		CookieSecure:      settings.CSRFTokenCookieSecure,
		CookieHTTPOnly:    settings.CSRFTokenCookieHTTPOnly,
		CookieSessionOnly: settings.CSRFTokenCookieSessionOnly,
		SingleUseToken:    settings.CSRFTokenSingleUseToken,
		Next:              isNonBrowserUserApiRequestFunc(settings), // Skip check for /service/api/v1/* requests by non-browser clients
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
// FALSE => Enforce CSRF check for /browser/api/v1/* requests by browser clients (e.g. web apps, Swagger UI)
// ASSUME: UI Authentication only authorizes browser users to access /browser/api/v1/*
func isNonBrowserUserApiRequestFunc(settings *cryptoutilConfig.Settings) func(c *fiber.Ctx) bool {
	return func(c *fiber.Ctx) bool {
		return strings.HasPrefix(c.OriginalURL(), settings.PublicServiceAPIContextPath+"/")
	}
}

func swaggerUICustomCSRFScript(csrfTokenName string, browserAPIContextPath string) template.JS {
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
								// Add Referer header since CSRF middleware requires it
								options.headers['Referer'] = window.location.href;
								console.log('Added CSRF token and Referer to request headers:', options.method, url);
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
