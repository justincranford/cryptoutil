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
	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	telemetryService "cryptoutil/internal/common/telemetry"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"
	cryptoutilBarrierService "cryptoutil/internal/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/server/barrier/unsealkeysservice"
	cryptoutilBusinessLogic "cryptoutil/internal/server/businesslogic"
	cryptoutilOpenapiHandler "cryptoutil/internal/server/handler"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"
	cryptoutilSqlRepository "cryptoutil/internal/server/repository/sqlrepository"

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

const shutdownRequestTimeout = 5 * time.Second
const livenessRequestTimeout = 3 * time.Second
const serverShutdownStartTimeout = 50 * time.Millisecond
const serverShutdownFinishTimeout = 3 * time.Second
const apiAdminShutdownPath = "/api/admin/shutdown"
const fiberAppIdKey = "fiberAppId"
const fiberAppIdService = "service"
const fiberAppIdAdmin = "admin"

var ready atomic.Bool

func SendServerShutdownRequest(settings *cryptoutilConfig.Settings) error {
	shutdownEndpoint := fmt.Sprintf("%s://%s:%d%s", settings.BindAdminProtocol, settings.BindAdminAddress, settings.BindAdminPort, apiAdminShutdownPath)
	shutdownRequestCtx, shutdownRequestCancel := context.WithTimeout(context.Background(), shutdownRequestTimeout)
	defer shutdownRequestCancel()
	shutdownRequest, err := http.NewRequestWithContext(shutdownRequestCtx, http.MethodPost, shutdownEndpoint, nil)
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

	time.Sleep(serverShutdownStartTimeout)
	livenessEndpoint := fmt.Sprintf("%s://%s:%d/livez", settings.BindAdminProtocol, settings.BindAdminAddress, settings.BindAdminPort)
	livenessRequestCtx, livenessRequestCancel := context.WithTimeout(context.Background(), livenessRequestTimeout)
	defer livenessRequestCancel()
	livenessRequest, _ := http.NewRequestWithContext(livenessRequestCtx, http.MethodGet, livenessEndpoint, nil)
	livenessResponse, err := http.DefaultClient.Do(livenessRequest)
	if err == nil && livenessResponse != nil {
		livenessResponse.Body.Close()
		return fmt.Errorf("server did not shut down properly")
	}
	return nil
}

func StartServerApplication(settings *cryptoutilConfig.Settings) (func(), func(), error) {
	ctx := context.Background()

	telemetryService, err := cryptoutilTelemetry.NewTelemetryService(ctx, settings)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initailize telemetry: %w", err)
	}

	sqlRepository, err := cryptoutilSqlRepository.NewSqlRepository(ctx, telemetryService, settings)
	if err != nil {
		telemetryService.Slogger.Error("failed to connect to SQL DB", "error", err)
		stopServerFunc(telemetryService, nil, nil, nil, nil, nil, nil, nil)()
		return nil, nil, fmt.Errorf("failed to connect to SQL DB: %w", err)
	}

	jwkGenService, err := cryptoutilJose.NewJwkGenService(ctx, telemetryService)
	if err != nil {
		telemetryService.Slogger.Error("failed to create JWK Gen Service", "error", err)
		stopServerFunc(telemetryService, sqlRepository, nil, nil, nil, nil, nil, nil)()
		return nil, nil, fmt.Errorf("failed to create JWK Gen Service: %w", err)
	}

	ormRepository, err := cryptoutilOrmRepository.NewOrmRepository(ctx, telemetryService, sqlRepository, jwkGenService, settings)
	if err != nil {
		telemetryService.Slogger.Error("failed to create ORM repository", "error", err)
		stopServerFunc(telemetryService, sqlRepository, jwkGenService, nil, nil, nil, nil, nil)()
		return nil, nil, fmt.Errorf("failed to create ORM repository: %w", err)
	}

	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	if err != nil {
		telemetryService.Slogger.Error("failed to create unseal repository", "error", err)
		stopServerFunc(telemetryService, sqlRepository, jwkGenService, ormRepository, nil, nil, nil, nil)()
		return nil, nil, fmt.Errorf("failed to create unseal repository: %w", err)
	}

	barrierService, err := cryptoutilBarrierService.NewBarrierService(ctx, telemetryService, jwkGenService, ormRepository, unsealKeysService)
	if err != nil {
		telemetryService.Slogger.Error("failed to initialize barrier service", "error", err)
		stopServerFunc(telemetryService, sqlRepository, jwkGenService, ormRepository, unsealKeysService, nil, nil, nil)()
		return nil, nil, fmt.Errorf("failed to create barrier service: %w", err)
	}

	businessLogicService, err := cryptoutilBusinessLogic.NewBusinessLogicService(ctx, telemetryService, jwkGenService, ormRepository, barrierService)
	if err != nil {
		telemetryService.Slogger.Error("failed to initialize business logic service", "error", err)
		stopServerFunc(telemetryService, sqlRepository, jwkGenService, ormRepository, unsealKeysService, barrierService, nil, nil)()
		return nil, nil, fmt.Errorf("failed to initialize business logic service: %w", err)
	}

	swaggerApi, err := cryptoutilOpenapiServer.GetSwagger()
	if err != nil {
		telemetryService.Slogger.Error("failed to get swagger", "error", err)
		stopServerFunc(telemetryService, sqlRepository, jwkGenService, ormRepository, unsealKeysService, barrierService, nil, nil)()
		return nil, nil, fmt.Errorf("failed to get swagger: %w", err)
	}

	openapiStrictServer := cryptoutilOpenapiHandler.NewOpenapiStrictServer(businessLogicService)
	openapiStrictHandler := cryptoutilOpenapiServer.NewStrictHandler(openapiStrictServer, nil)
	fiberServerOptions := cryptoutilOpenapiServer.FiberServerOptions{
		Middlewares: []cryptoutilOpenapiServer.MiddlewareFunc{ // Defined as MiddlewareFunc => Fiber.Handler in generated code
			fibermiddleware.OapiRequestValidatorWithOptions(swaggerApi, &fibermiddleware.Options{}),
		},
	}

	fiberHandlerOpenAPISpec, err := cryptoutilOpenapiServer.FiberHandlerOpenAPISpec()
	if err != nil {
		telemetryService.Slogger.Error("failed to get fiber handler for OpenAPI spec", "error", err)
		stopServerFunc(telemetryService, sqlRepository, jwkGenService, ormRepository, unsealKeysService, barrierService, nil, nil)()
		return nil, nil, fmt.Errorf("failed to get fiber handler for OpenAPI spec: %w", err)
	}

	commonMiddlewares := []fiber.Handler{
		recover.New(),
		requestid.New(),
		logger.New(), // TODO Remove this since it prints unstructured logs, and doesn't push to OpenTelemetry
		otelFiberTelemetryMiddleware(telemetryService, settings),
		otelFiberRequestLoggerMiddleware(telemetryService),
		ipFilterMiddleware(telemetryService, settings),
		ipRateLimiterMiddleware(telemetryService, settings),
		httpGetCacheControlMiddleware(),
	}

	serviceFiberApp := fiber.New(fiber.Config{Immutable: true})
	serviceFiberApp.Use(setFiberAppId(fiberAppIdService))
	for _, middleware := range commonMiddlewares {
		serviceFiberApp.Use(middleware)
	}
	serviceFiberApp.Use(corsMiddleware(settings)) // Browser-specific: Cross-Origin Resource Sharing (CORS)
	serviceFiberApp.Use(helmet.New())             // Browser-specific: Cross-Site Scripting (XSS)
	serviceFiberApp.Use(csrfMiddleware(settings)) // Browser-specific: Cross-Site Request Forgery (CSRF)
	serviceFiberApp.Get("/swagger/doc.json", fiberHandlerOpenAPISpec)
	serviceFiberApp.Get("/swagger/*", swagger.New(swagger.Config{
		Title:                  "Cryptoutil",
		TryItOutEnabled:        true,
		DisplayRequestDuration: true,
		ShowCommonExtensions:   true,
		CustomScript:           swaggerUICustomCSRFScript, // Custom JavaScript to inject CSRF token into all non-GET requests
	}))
	cryptoutilOpenapiServer.RegisterHandlersWithOptions(serviceFiberApp, openapiStrictHandler, fiberServerOptions)

	var stopServer func() // circular dependency: adminFiberApp -> stopServer -> adminFiberApp
	adminFiberApp := fiber.New(fiber.Config{Immutable: true})
	adminFiberApp.Use(setFiberAppId(fiberAppIdAdmin))
	for _, middleware := range commonMiddlewares {
		adminFiberApp.Use(middleware)
	}
	adminFiberApp.Use(healthcheck.New()) // /livez, /readyz
	adminFiberApp.Post(apiAdminShutdownPath, func(c *fiber.Ctx) error {
		telemetryService.Slogger.Info("shutdown requested via API endpoint")
		if stopServer != nil {
			go func() {
				time.Sleep(serverShutdownStartTimeout)
				stopServer()
			}()
		}
		return c.SendString("Server shutdown initiated")
	})

	serviceBinding := fmt.Sprintf("%s:%d", settings.BindServiceAddress, settings.BindServicePort)
	adminBinding := fmt.Sprintf("%s:%d", settings.BindAdminAddress, settings.BindAdminPort)
	startServer := startServerFunc(serviceBinding, adminBinding, serviceFiberApp, adminFiberApp, telemetryService)
	stopServer = stopServerFunc(telemetryService, sqlRepository, jwkGenService, ormRepository, unsealKeysService, barrierService, serviceFiberApp, adminFiberApp)

	go stopServerSignalFunc(telemetryService, stopServer)() // listen for OS signals to gracefully shutdown the server

	return startServer, stopServer, nil
}

func startServerFunc(serviceBinding string, adminBinding string, serviceFiberApp *fiber.App, adminFiberApp *fiber.App, telemetryService *cryptoutilTelemetry.TelemetryService) func() {
	return func() {
		telemetryService.Slogger.Debug("starting fiber listeners")

		go func() {
			telemetryService.Slogger.Debug("starting admin fiber listener", "binding", adminBinding)
			if err := adminFiberApp.Listen(adminBinding); err != nil {
				telemetryService.Slogger.Error("failed to start admin fiber listener", "error", err)
			}
			telemetryService.Slogger.Debug("admin fiber listener stopped")
		}()
		telemetryService.Slogger.Debug("starting service fiber listener", "binding", serviceBinding)
		if err := serviceFiberApp.Listen(serviceBinding); err != nil {
			telemetryService.Slogger.Error("failed to start service fiber listener", "error", err)
		}
		telemetryService.Slogger.Debug("service fiber listener stopped")

		ready.Store(true)
	}
}

func stopServerFunc(telemetryService *cryptoutilTelemetry.TelemetryService, sqlRepository *cryptoutilSqlRepository.SqlRepository, jwkGenService *cryptoutilJose.JwkGenService, ormRepository *cryptoutilOrmRepository.OrmRepository, unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService, barrierService *cryptoutilBarrierService.BarrierService, serviceFiberApp *fiber.App, adminFiberApp *fiber.App) func() {
	return func() {
		if telemetryService != nil {
			telemetryService.Slogger.Debug("stopping servers")
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), serverShutdownFinishTimeout)
		defer cancel() // perform shutdown respecting timeout

		if serviceFiberApp != nil {
			telemetryService.Slogger.Debug("shutting down service fiber app")
			if err := serviceFiberApp.ShutdownWithContext(shutdownCtx); err != nil {
				telemetryService.Slogger.Error("failed to stop service fiber server", "error", err)
			}
		}
		if adminFiberApp != nil {
			telemetryService.Slogger.Debug("shutting down admin fiber app")
			if err := adminFiberApp.ShutdownWithContext(shutdownCtx); err != nil {
				telemetryService.Slogger.Error("failed to stop admin fiber server", "error", err)
			}
		}

		// each service should do its own logging
		if barrierService != nil {
			barrierService.Shutdown()
		}
		if unsealKeysService != nil {
			unsealKeysService.Shutdown()
		}
		if ormRepository != nil {
			ormRepository.Shutdown()
		}
		if jwkGenService != nil {
			jwkGenService.Shutdown()
		}
		if sqlRepository != nil {
			sqlRepository.Shutdown()
		}
		if telemetryService != nil {
			telemetryService.Shutdown()
		}
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

func setFiberAppId(fiberAppIdValue string) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Locals(fiberAppIdKey, fiberAppIdValue)
		return c.Next()
	}
}

func otelFiberTelemetryMiddleware(telemetryService *cryptoutilTelemetry.TelemetryService, settings *cryptoutilConfig.Settings) fiber.Handler {
	return otelfiber.Middleware(
		otelfiber.WithTracerProvider(telemetryService.TracesProvider),
		otelfiber.WithMeterProvider(telemetryService.MetricsProvider),
		otelfiber.WithPropagators(*telemetryService.TextMapPropagator),
		otelfiber.WithServerName(settings.BindServiceAddress),
		otelfiber.WithPort(int(settings.BindServicePort)),
	)
}

func ipFilterMiddleware(telemetryService *telemetryService.TelemetryService, settings *cryptoutilConfig.Settings) func(c *fiber.Ctx) error {
	allowedIPs := make(map[string]bool)
	if settings.AllowedIPs != "" {
		for allowedIP := range strings.SplitSeq(settings.AllowedIPs, ",") {
			parsedIP := net.ParseIP(allowedIP) // IPv4 (e.g.  192.0.2.1"), IPv6 (e.g. 2001:db8::68), or IPv4-mapped IPv6 (e.g. ::ffff:192.0.2.1)
			if parsedIP == nil {
				telemetryService.Slogger.Error("invalid allowed IP address:", "IP", allowedIP)
			}
			allowedIPs[allowedIP] = true
		}
	}

	var allowedCIDRs []*net.IPNet
	if settings.AllowedCIDRs != "" {
		for allowedCIDR := range strings.SplitSeq(settings.AllowedCIDRs, ",") {
			_, network, err := net.ParseCIDR(allowedCIDR) // "192.0.2.1/24" => 192.0.2.1 (not useful) and 192.0.2.0/24 (useful)
			if err != nil {
				telemetryService.Slogger.Error("invalid allowed CIDR:", "CIDR", allowedCIDR, "error", err)
			}
			allowedCIDRs = append(allowedCIDRs, network)
		}
	}

	return func(c *fiber.Ctx) error { // Mitigate against DDOS by allowlisting IP addresses and CIDRs
		switch c.Locals(fiberAppIdKey) {
		case fiberAppIdService: // Apply IP/CIDR filtering for service app requests
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
		case fiberAppIdAdmin: // Skip IP/CIDR filtering for admin app requests
			return c.Next()
		default:
			telemetryService.Slogger.Error("Unexpected app ID:", c.Locals(fiberAppIdKey))
			return c.Status(fiber.StatusInternalServerError).SendString("Internal server error")
		}
	}
}

func ipRateLimiterMiddleware(telemetryService *telemetryService.TelemetryService, settings *cryptoutilConfig.Settings) fiber.Handler {
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
	return csrf.New(csrf.Config{ // Cross-Site Request Forgery (CSRF)
		CookieName:        settings.CSRFTokenName,              // cryptoutilConfig.defaultCSRFTokenName
		CookieSameSite:    settings.CSRFTokenSameSite,          // cryptoutilConfig.defaultCSRFTokenSameSite
		Expiration:        settings.CSRFTokenMaxAge,            // cryptoutilConfig.defaultCSRFTokenMaxAge
		CookieSecure:      settings.CSRFTokenCookieSecure,      // cryptoutilConfig.defaultCSRFTokenCookieSecure
		CookieHTTPOnly:    settings.CSRFTokenCookieHTTPOnly,    // cryptoutilConfig.defaultCSRFTokenCookieHTTPOnly
		CookieSessionOnly: settings.CSRFTokenCookieSessionOnly, // cryptoutilConfig.defaultCSRFTokenCookieSessionOnly
	})
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
