package application

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"
	cryptoutilBarrierService "cryptoutil/internal/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/server/barrier/unsealkeysservice"
	cryptoutilBusinessLogic "cryptoutil/internal/server/businesslogic"
	cryptoutilOpenapiHandler "cryptoutil/internal/server/handler"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"
	cryptoutilSqlRepository "cryptoutil/internal/server/repository/sqlrepository"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
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

var ready atomic.Bool

func StartServerApplication(settings *cryptoutilConfig.Settings) (func(), func(), error) {
	ctx := context.Background()

	telemetryService, err := cryptoutilTelemetry.NewTelemetryService(ctx, settings)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initailize telemetry: %w", err)
	}

	sqlRepository, err := cryptoutilSqlRepository.NewSqlRepository(ctx, telemetryService, settings)
	if err != nil {
		telemetryService.Slogger.Error("failed to connect to SQL DB", "error", err)
		stopServerFunc(telemetryService, nil, nil, nil, nil, nil, nil)()
		return nil, nil, fmt.Errorf("failed to connect to SQL DB: %w", err)
	}

	jwkGenService, err := cryptoutilJose.NewJwkGenService(ctx, telemetryService)
	if err != nil {
		telemetryService.Slogger.Error("failed to create JWK Gen Service", "error", err)
		stopServerFunc(telemetryService, sqlRepository, nil, nil, nil, nil, nil)()
		return nil, nil, fmt.Errorf("failed to create JWK Gen Service: %w", err)
	}

	ormRepository, err := cryptoutilOrmRepository.NewOrmRepository(ctx, telemetryService, sqlRepository, jwkGenService, settings)
	if err != nil {
		telemetryService.Slogger.Error("failed to create ORM repository", "error", err)
		stopServerFunc(telemetryService, sqlRepository, jwkGenService, nil, nil, nil, nil)()
		return nil, nil, fmt.Errorf("failed to create ORM repository: %w", err)
	}

	// Create unseal keys service based on configuration settings
	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	if err != nil {
		telemetryService.Slogger.Error("failed to create unseal repository", "error", err)
		stopServerFunc(telemetryService, sqlRepository, jwkGenService, ormRepository, nil, nil, nil)()
		return nil, nil, fmt.Errorf("failed to create unseal repository: %w", err)
	}

	barrierService, err := cryptoutilBarrierService.NewBarrierService(ctx, telemetryService, jwkGenService, ormRepository, unsealKeysService)
	if err != nil {
		telemetryService.Slogger.Error("failed to initialize barrier service", "error", err)
		stopServerFunc(telemetryService, sqlRepository, jwkGenService, ormRepository, unsealKeysService, nil, nil)()
		return nil, nil, fmt.Errorf("failed to create barrier service: %w", err)
	}

	businessLogicService, err := cryptoutilBusinessLogic.NewBusinessLogicService(ctx, telemetryService, jwkGenService, ormRepository, barrierService)
	if err != nil {
		telemetryService.Slogger.Error("failed to initialize business logic service", "error", err)
		stopServerFunc(telemetryService, sqlRepository, jwkGenService, ormRepository, unsealKeysService, barrierService, nil)()
		return nil, nil, fmt.Errorf("failed to initialize business logic service: %w", err)
	}

	swaggerApi, err := cryptoutilOpenapiServer.GetSwagger()
	if err != nil {
		telemetryService.Slogger.Error("failed to get swagger", "error", err)
		stopServerFunc(telemetryService, sqlRepository, jwkGenService, ormRepository, unsealKeysService, barrierService, nil)()
		return nil, nil, fmt.Errorf("failed to get swagger: %w", err)
	}

	fiberHandlerOpenAPISpec, err := cryptoutilOpenapiServer.FiberHandlerOpenAPISpec()
	if err != nil {
		telemetryService.Slogger.Error("failed to get fiber handler for OpenAPI spec", "error", err)
		stopServerFunc(telemetryService, sqlRepository, jwkGenService, ormRepository, unsealKeysService, barrierService, nil)()
		return nil, nil, fmt.Errorf("failed to get fiber handler for OpenAPI spec: %w", err)
	}

	app := fiber.New(fiber.Config{Immutable: true})
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(logger.New()) // TODO Remove this since it prints unstructured logs, and doesn't push to OpenTelemetry
	app.Use(fiberOtelLoggerMiddleware(telemetryService.Slogger))
	app.Use(ipFilterMiddleware(settings))
	app.Use(ipRateLimiterMiddleware(settings))
	app.Use(cacheControlMiddleware())
	app.Use(corsMiddleware(settings)) // Cross-Origin Resource Sharing
	app.Use(helmet.New())             // Cross-Site Scripting (XSS)
	app.Use(csrfMiddleware(settings)) // Cross-Site Request Forgery (CSRF)
	app.Use(healthcheck.New())

	app.Use(otelfiber.Middleware(
		otelfiber.WithTracerProvider(telemetryService.TracesProvider),
		otelfiber.WithMeterProvider(telemetryService.MetricsProvider),
		otelfiber.WithPropagators(*telemetryService.TextMapPropagator),
		otelfiber.WithServerName(settings.BindAddress),
		otelfiber.WithPort(int(settings.BindPort)),
	))
	app.Get("/swagger/doc.json", fiberHandlerOpenAPISpec)
	app.Get("/swagger/*", swagger.New(swagger.Config{
		Title:                  "Cryptoutil",
		TryItOutEnabled:        true,
		DisplayRequestDuration: true,
		ShowCommonExtensions:   true,
		// Add custom JavaScript to inject CSRF token into all non-GET requests
		CustomScript: `
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
		`,
	}))

	openapiStrictServer := cryptoutilOpenapiHandler.NewOpenapiStrictServer(businessLogicService)
	openapiStrictHandler := cryptoutilOpenapiServer.NewStrictHandler(openapiStrictServer, nil)
	fiberServerOptions := cryptoutilOpenapiServer.FiberServerOptions{
		Middlewares: []cryptoutilOpenapiServer.MiddlewareFunc{ // Defined as MiddlewareFunc => Fiber.Handler in generated code
			fibermiddleware.OapiRequestValidatorWithOptions(swaggerApi, &fibermiddleware.Options{}),
		},
	}
	cryptoutilOpenapiServer.RegisterHandlersWithOptions(app, openapiStrictHandler, fiberServerOptions)

	listenAddress := fmt.Sprintf("%s:%d", settings.BindAddress, settings.BindPort)

	startServer := startServerFunc(err, listenAddress, app, telemetryService)
	stopServer := stopServerFunc(telemetryService, sqlRepository, jwkGenService, ormRepository, unsealKeysService, barrierService, app)
	go stopServerSignalFunc(telemetryService, stopServer)() // listen for OS signals to gracefully shutdown the server

	return startServer, stopServer, nil
}

func startServerFunc(err error, listenAddress string, app *fiber.App, telemetryService *cryptoutilTelemetry.TelemetryService) func() {
	return func() {
		telemetryService.Slogger.Debug("starting fiber listener")
		ready.Store(true)
		err = app.Listen(listenAddress) // blocks until fiber app is stopped (e.g. stopServerFunc called by unit test or stopServerSignalFunc)
		if err != nil {
			telemetryService.Slogger.Error("failed to start fiber listener", "error", err)
		}
		telemetryService.Slogger.Debug("listener fiber stopped")
	}
}

func stopServerFunc(telemetryService *cryptoutilTelemetry.TelemetryService, sqlRepository *cryptoutilSqlRepository.SqlRepository, jwkGenService *cryptoutilJose.JwkGenService, ormRepository *cryptoutilOrmRepository.OrmRepository, unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService, barrierService *cryptoutilBarrierService.BarrierService, app *fiber.App) func() {
	return func() {
		if telemetryService != nil {
			telemetryService.Slogger.Debug("stopping server")
		}
		if app != nil {
			err := app.Shutdown()
			if err != nil {
				telemetryService.Slogger.Error("failed to stop fiber server", "error", err)
			}
		}
		if barrierService != nil {
			barrierService.Shutdown() // does its own logging
		}
		if unsealKeysService != nil {
			unsealKeysService.Shutdown() // does its own logging
		}
		if ormRepository != nil {
			ormRepository.Shutdown() // does its own logging
		}
		if jwkGenService != nil {
			jwkGenService.Shutdown() // does its own logging
		}
		if sqlRepository != nil {
			sqlRepository.Shutdown() // does its own logging
		}
		if telemetryService != nil {
			telemetryService.Shutdown() // does its own logging
		}
	}
}

func stopServerSignalFunc(telemetryService *cryptoutilTelemetry.TelemetryService, stopServerFunc func()) func() {
	return func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		telemetryService.Slogger.Info("received stop server signal")
		stopServerFunc()
	}
}

func ipFilterMiddleware(settings *cryptoutilConfig.Settings) func(c *fiber.Ctx) error {
	allowedIPs := make(map[string]bool)
	if settings.AllowedIPs != "" {
		for _, ip := range strings.Split(settings.AllowedIPs, ",") {
			parsedIP := net.ParseIP(ip)
			if parsedIP == nil {
				log.Fatal("Invalid IP address:", ip)
			}
			allowedIPs[ip] = true
		}
	}

	var allowedCIDRs []*net.IPNet
	if settings.AllowedCIDRs != "" {
		for _, cidr := range strings.Split(settings.AllowedCIDRs, ",") {
			_, network, err := net.ParseCIDR(cidr)
			if err != nil {
				log.Fatal("Invalid CIDR:", cidr)
			}
			allowedCIDRs = append(allowedCIDRs, network)
		}
	}

	return func(c *fiber.Ctx) error { // Mitigate against DDOS by allowlisting IP addresses and CIDRs
		clientIP := c.IP()
		parsedIP := net.ParseIP(clientIP)
		if parsedIP == nil {
			log.Debug("Invalid IP: #=", c.Locals("requestid"), ", method=", c.Method(), ", IP=", clientIP, ", URL=", c.OriginalURL(), " Headers=", c.GetReqHeaders())
			return c.Status(fiber.StatusForbidden).SendString("Invalid IP format")
		} else if _, allowed := allowedIPs[parsedIP.String()]; allowed {
			log.Debug("Allowed IP: #=", c.Locals("requestid"), ", method=", c.Method(), ", IP=", clientIP, ", URL=", c.OriginalURL(), " Headers=", c.GetReqHeaders())
			return c.Next() // IP is contained in the allowed IPs set
		}
		for _, cidr := range allowedCIDRs {
			if cidr.Contains(parsedIP) {
				log.Debug("Allowed CIDR: #=", c.Locals("requestid"), ", method=", c.Method(), ", IP=", clientIP, ", URL=", c.OriginalURL(), " Headers=", c.GetReqHeaders())
				return c.Next() // IP is contained in minGenreID of the allowed CIDRs
			}
		}
		log.Debug("Access denied: #=", c.Locals("requestid"), ", method=", c.Method(), ", IP=", clientIP, ", URL=", c.OriginalURL(), " Headers=", c.GetReqHeaders())
		return c.Status(fiber.StatusForbidden).SendString("Access denied")
	}
}

func ipRateLimiterMiddleware(settings *cryptoutilConfig.Settings) fiber.Handler {
	return limiter.New(limiter.Config{ // Mitigate DOS by throttling clients
		Max:        int(settings.IPRateLimit),
		Expiration: time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() // throttle by IP, could be improved in future (e.g. append JWTClaim.sub or JWTClaim.tenantid)
		},
		LimitReached: func(c *fiber.Ctx) error {
			log.Warn("Rate limited: #=", c.Locals("requestid"), ", method=", c.Method(), ", IP=", c.IP(), ", URL=", c.OriginalURL(), " Headers=", c.GetReqHeaders())
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
	// TODO update tests to enable CSRF protection in dev mode
	if settings.DevMode { // NOOP in dev mode
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}
	return csrf.New(csrf.Config{ // Cross-Site Request Forgery (CSRF)
		CookieName:     settings.CSRFTokenName,     // cryptoutilConfig.defaultCSRFTokenName
		CookieSameSite: settings.CSRFTokenSameSite, // cryptoutilConfig.defaultCSRFTokenSameSite
		Expiration:     settings.CSRFTokenMaxAge,   // cryptoutilConfig.defaultCSRFTokenMaxAge
	})
}

func cacheControlMiddleware() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error { // Disable caching of HTTP GET responses
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		return c.Next()
	}
}
