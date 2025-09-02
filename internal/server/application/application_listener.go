package application

import (
	"context"
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
	shutdownResponse, err := http.DefaultClient.Do(shutdownRequest)
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
	livenessResponse, err := http.DefaultClient.Do(livenessRequest)
	if err == nil && livenessResponse != nil {
		livenessResponse.Body.Close()
		return fmt.Errorf("server did not shut down properly")
	}
	return nil
}

func StartServerListenerApplication(settings *cryptoutilConfig.Settings) (func(), func(), error) {
	ctx := context.Background()

	serverApplicationCore, err := StartServerApplicationCore(ctx, settings)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize server application: %w", err)
	}

	// Middlewares

	commonMiddlewares := []fiber.Handler{
		recover.New(),
		requestid.New(),
		logger.New(), // TODO Remove this since it prints unstructured logs, and doesn't push to OpenTelemetry
		otelFiberTelemetryMiddleware(serverApplicationCore.TelemetryService, settings),
		otelFiberRequestLoggerMiddleware(serverApplicationCore.TelemetryService),
		ipFilterMiddleware(serverApplicationCore.TelemetryService, settings),
		ipRateLimiterMiddleware(serverApplicationCore.TelemetryService, settings),
		httpGetCacheControlMiddleware(),
	}

	privateMiddlewares := append([]fiber.Handler{setFiberRequestAttribute(fiberAppIDPrivate)}, commonMiddlewares...)
	publicMiddlewares := append([]fiber.Handler{setFiberRequestAttribute(fiberAppIDPublic)}, commonMiddlewares...)
	privateMiddlewares = append(privateMiddlewares, healthcheck.New())      // /livez, /readyz
	publicMiddlewares = append(publicMiddlewares, corsMiddleware(settings)) // Browser-specific: Cross-Origin Resource Sharing (CORS)
	publicMiddlewares = append(publicMiddlewares, helmet.New())             // Browser-specific: Cross-Site Scripting (XSS)
	publicMiddlewares = append(publicMiddlewares, csrfMiddleware(settings)) // Browser-specific: Cross-Site Request Forgery (CSRF)

	privateFiberApp := fiber.New(fiber.Config{Immutable: true})
	for _, middleware := range privateMiddlewares {
		privateFiberApp.Use(middleware)
	}
	publicFiberApp := fiber.New(fiber.Config{Immutable: true})
	for _, middleware := range publicMiddlewares {
		publicFiberApp.Use(middleware)
	}

	// shutdownServerFunction stops privateFiberApp and publicFiberApp, it is called via /shutdown hosted by privateFiberApp
	var shutdownServerFunction func()

	// Private APIs
	privateFiberApp.Post(serverShutdownRequestPath, func(c *fiber.Ctx) error {
		serverApplicationCore.TelemetryService.Slogger.Info("shutdown requested via API endpoint")
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
		serverApplicationCore.TelemetryService.Slogger.Error("failed to get swagger", "error", err)
		serverApplicationCore.Shutdown()
		return nil, nil, fmt.Errorf("failed to get swagger: %w", err)
	}

	swaggerApi.Servers = []*openapi3.Server{{URL: settings.APIContextPath}}
	swaggerSpecBytes, err := swaggerApi.MarshalJSON() // Serialize OpenAPI 3 spec with updated context path
	if err != nil {
		serverApplicationCore.TelemetryService.Slogger.Error("failed to get fiber handler for OpenAPI spec", "error", err)
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
		CustomScript:           swaggerUICustomCSRFScript,
	}))

	// Public Swagger APIs
	openapiStrictServer := cryptoutilOpenapiHandler.NewOpenapiStrictServer(serverApplicationCore.BusinessLogicService)
	openapiStrictHandler := cryptoutilOpenapiServer.NewStrictHandler(openapiStrictServer, nil)
	fiberServerOptions := cryptoutilOpenapiServer.FiberServerOptions{
		BaseURL: settings.APIContextPath,
		Middlewares: []cryptoutilOpenapiServer.MiddlewareFunc{ // Defined as MiddlewareFunc => Fiber.Handler in generated code
			fibermiddleware.OapiRequestValidatorWithOptions(swaggerApi, &fibermiddleware.Options{}),
		},
	}
	cryptoutilOpenapiServer.RegisterHandlersWithOptions(publicFiberApp, openapiStrictHandler, fiberServerOptions)

	publicBinding := fmt.Sprintf("%s:%d", settings.BindPublicAddress, settings.BindPublicPort)
	privateBinding := fmt.Sprintf("%s:%d", settings.BindPrivateAddress, settings.BindPrivatePort)
	startServerFunction := startServerFunc(publicBinding, privateBinding, publicFiberApp, privateFiberApp, serverApplicationCore.TelemetryService)
	shutdownServerFunction = stopServerFunc(serverApplicationCore, publicFiberApp, privateFiberApp)

	go stopServerSignalFunc(serverApplicationCore.TelemetryService, shutdownServerFunction)() // listen for OS signals to gracefully shutdown the server

	return startServerFunction, shutdownServerFunction, nil
}

func startServerFunc(publicBinding string, privateBinding string, publicFiberApp *fiber.App, privateFiberApp *fiber.App, telemetryService *cryptoutilTelemetry.TelemetryService) func() {
	return func() {
		telemetryService.Slogger.Debug("starting fiber listeners")

		go func() {
			telemetryService.Slogger.Debug("starting private fiber listener", "binding", privateBinding)
			if err := privateFiberApp.Listen(privateBinding); err != nil {
				telemetryService.Slogger.Error("failed to start private fiber listener", "error", err)
			}
			telemetryService.Slogger.Debug("private fiber listener stopped")
		}()
		telemetryService.Slogger.Debug("starting public fiber listener", "binding", publicBinding)
		if err := publicFiberApp.Listen(publicBinding); err != nil {
			telemetryService.Slogger.Error("failed to start public fiber listener", "error", err)
		}
		telemetryService.Slogger.Debug("public fiber listener stopped")

		ready.Store(true)
	}
}

func stopServerFunc(serverApplicationCore *ServerApplicationCore, publicFiberApp *fiber.App, privateFiberApp *fiber.App) func() {
	return func() {
		if serverApplicationCore.TelemetryService != nil {
			serverApplicationCore.TelemetryService.Slogger.Debug("stopping servers")
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), serverShutdownFinishTimeout)
		defer cancel() // perform shutdown respecting timeout

		if publicFiberApp != nil {
			serverApplicationCore.TelemetryService.Slogger.Debug("shutting down public fiber app")
			if err := publicFiberApp.ShutdownWithContext(shutdownCtx); err != nil {
				serverApplicationCore.TelemetryService.Slogger.Error("failed to stop public fiber server", "error", err)
			}
		}
		if privateFiberApp != nil {
			serverApplicationCore.TelemetryService.Slogger.Debug("shutting down private fiber app")
			if err := privateFiberApp.ShutdownWithContext(shutdownCtx); err != nil {
				serverApplicationCore.TelemetryService.Slogger.Error("failed to stop private fiber server", "error", err)
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

func setFiberRequestAttribute(fiberAppIdValue fiberAppID) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Locals(fiberAppIDRequestAttribute, string(fiberAppIdValue))
		return c.Next()
	}
}

func otelFiberTelemetryMiddleware(telemetryService *cryptoutilTelemetry.TelemetryService, settings *cryptoutilConfig.Settings) fiber.Handler {
	return otelfiber.Middleware(
		otelfiber.WithTracerProvider(telemetryService.TracesProvider),
		otelfiber.WithMeterProvider(telemetryService.MetricsProvider),
		otelfiber.WithPropagators(*telemetryService.TextMapPropagator),
		otelfiber.WithServerName(settings.BindPublicAddress),
		otelfiber.WithPort(int(settings.BindPublicPort)),
	)
}

func ipFilterMiddleware(telemetryService *cryptoutilTelemetry.TelemetryService, settings *cryptoutilConfig.Settings) func(c *fiber.Ctx) error {
	allowedIPs := make(map[string]bool)
	if settings.AllowedIPs != "" {
		for allowedIP := range strings.SplitSeq(settings.AllowedIPs, ",") {
			parsedIP := net.ParseIP(allowedIP) // IPv4 (e.g.  192.0.2.1"), IPv6 (e.g. 2001:db8::68), or IPv4-mapped IPv6 (e.g. ::ffff:192.0.2.1)
			if parsedIP == nil {
				telemetryService.Slogger.Error("invalid allowed IP address:", "IP", allowedIP)
			} else {
				allowedIPs[allowedIP] = true
				if settings.DevMode {
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
				if settings.DevMode {
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
			for _, cidr := range allowedCIDRs {
				if cidr.Contains(parsedIP) {
					if settings.VerboseMode {
						telemetryService.Slogger.Debug("Allowed CIDR:", "#", c.Locals("requestid"), "method", c.Method(), "IP", clientIP, "URL", c.OriginalURL(), "Headers", c.GetReqHeaders())
					}
					return c.Next() // IP is contained in the allowed CIDRs
				}
			}
			telemetryService.Slogger.Debug("Access denied:", "#", c.Locals("requestid"), "method", c.Method(), "IP", clientIP, "URL", c.OriginalURL(), "Headers", c.GetReqHeaders())
			return c.Status(fiber.StatusForbidden).SendString("Access denied")
		case string(fiberAppIDPrivate): // Skip IP/CIDR filtering for private app requests
			return c.Next()
		default:
			telemetryService.Slogger.Error("Unexpected app ID:", c.Locals(fiberAppIDRequestAttribute))
			return c.Status(fiber.StatusInternalServerError).SendString("Internal server error")
		}
	}
}

func ipRateLimiterMiddleware(telemetryService *cryptoutilTelemetry.TelemetryService, settings *cryptoutilConfig.Settings) fiber.Handler {
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

func corsMiddleware(settings *cryptoutilConfig.Settings) fiber.Handler {
	return cors.New(cors.Config{ // Cross-Origin Resource Sharing (CORS)
		AllowOrigins: settings.CORSAllowedOrigins, // cryptoutilConfig.defaultAllowedCORSOrigins
		AllowMethods: settings.CORSAllowedMethods, // cryptoutilConfig.defaultAllowedCORSMethods
		AllowHeaders: settings.CORSAllowedHeaders, // cryptoutilConfig.defaultAllowedCORSHeaders
		MaxAge:       int(settings.CORSMaxAge),
	})
}

func csrfMiddleware(settings *cryptoutilConfig.Settings) fiber.Handler {
	csrfConfig := csrf.Config{
		CookieName:        settings.CSRFTokenName,
		CookieSameSite:    settings.CSRFTokenSameSite,
		Expiration:        settings.CSRFTokenMaxAge,
		CookieSecure:      settings.CSRFTokenCookieSecure,
		CookieHTTPOnly:    settings.CSRFTokenCookieHTTPOnly,
		CookieSessionOnly: settings.CSRFTokenCookieSessionOnly,
		Next: func(c *fiber.Ctx) bool {
			if isApiEndpoint(c.OriginalURL(), settings.APIContextPath) {
				return !isBrowserClient(c) || settings.DevMode
			}
			return false
		},
	}
	return csrf.New(csrfConfig)
}

func isApiEndpoint(url string, apiContextPath string) bool {
	return strings.HasPrefix(url, apiContextPath+"/elastickey") ||
		strings.HasPrefix(url, apiContextPath+"/elastickeys") ||
		strings.HasPrefix(url, apiContextPath+"/materialkeys")
}

func isBrowserClient(c *fiber.Ctx) bool {
	origin := c.Get("Origin")
	referer := c.Get("Referer")
	accept := c.Get("Accept")
	contentType := c.Get("Content-Type")

	// Swagger UI requests - most reliable indicator
	if strings.Contains(referer, "/swagger") ||
		strings.Contains(origin, "swagger") {
		return true
	}

	// Browser-specific content types (form submissions)
	if strings.Contains(contentType, "application/x-www-form-urlencoded") ||
		strings.Contains(contentType, "multipart/form-data") {
		return true
	}

	// Browser requests typically have Origin or Referer headers
	// and accept HTML responses
	if (origin != "" || referer != "") &&
		strings.Contains(accept, "text/html") {
		return true
	}

	// API clients typically:
	// - Don't send Origin/Referer headers (unless explicitly set)
	// - Accept application/json primarily
	// - Don't include text/html in Accept header
	if origin == "" && referer == "" &&
		strings.Contains(accept, "application/json") &&
		!strings.Contains(accept, "text/html") {
		return false // Likely API client
	}

	// Fallback to User-Agent as last resort (least reliable)
	userAgent := c.Get("User-Agent")
	browserIndicators := []string{
		"Mozilla/", "Chrome/", "Safari/", "Edge/", "Firefox/", "Opera/",
	}

	for _, indicator := range browserIndicators {
		if strings.Contains(userAgent, indicator) {
			return true
		}
	}

	return false // Default to API client if unclear
}

func httpGetCacheControlMiddleware() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error { // Disable caching of HTTP GET responses
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		return c.Next()
	}
}

const swaggerUICustomCSRFScript = template.JS(`
		// Wait for Swagger UI to fully load
		const interval = setInterval(function() {
			if (window.ui) {
				clearInterval(interval);
				
				// Add CSRF token to all non-GET requests
				const originalFetch = window.fetch;
				window.fetch = function(url, options) {
					options = options || {};
					
					if (options && options.method && options.method !== 'GET') {
						options.headers = options.headers || {};
						// Extract CSRF token from cookies - using default cookie name "_csrf"
						const cookies = document.cookie.split(';');
						for (let i = 0; i < cookies.length; i++) {
							const cookie = cookies[i].trim();
							if (cookie.startsWith('_csrf=')) {
								options.headers['X-CSRF-Token'] = cookie.substring('_csrf='.length);
								console.log('Added CSRF token to request');
								break;
							}
						}
					}
					return originalFetch.call(this, url, options);
				};
				console.log('CSRF token handling enabled for Swagger UI');
			}
		}, 100);
	`)
