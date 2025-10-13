package application

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os/signal"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	cryptoutilOpenapiServer "cryptoutil/api/server"
	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilCertificate "cryptoutil/internal/common/crypto/certificate"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilOpenapiHandler "cryptoutil/internal/server/handler"

	"go.opentelemetry.io/otel/metric"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/swagger"
	fibermiddleware "github.com/oapi-codegen/fiber-middleware"
)

const (
	clientShutdownRequestTimeout = 5 * time.Second
	clientLivenessStartTimeout   = 200 * time.Millisecond
	clientLivenessRequestTimeout = 3 * time.Second
	errorStr                     = "error"
	statusStr                    = "status"
	protocolHTTPS                = "https"
)

// TODO Add separate timeouts for different shutdown phases (drain, force close, etc.)
const serverShutdownRequestPath = "/shutdown"

const fiberAppIDRequestAttribute = "fiberAppID"

type fiberAppID string

const (
	fiberAppIDPublic  fiberAppID = "public"
	fiberAppIDPrivate fiberAppID = "private"
)

var ready atomic.Bool

type ServerApplicationListener struct {
	StartFunction     func()
	ShutdownFunction  func()
	PublicTLSServer   *TLSServerConfig
	PrivateTLSServer  *TLSServerConfig
	ActualPublicPort  uint16
	ActualPrivatePort uint16
}

type TLSServerConfig struct {
	Certificate         *tls.Certificate
	RootCAsPool         *x509.CertPool
	IntermediateCAsPool *x509.CertPool
	Config              *tls.Config
}

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
	if settings.BindPrivateProtocol == protocolHTTPS {
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: settings.DevMode, // Only skip verification in dev mode
					MinVersion:         tls.VersionTLS12,
				},
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
		defer func() {
			if closeErr := shutdownResponse.Body.Close(); closeErr != nil {
				fmt.Printf("Warning: failed to close shutdown response body: %v\n", closeErr)
			}
		}()
		if err != nil {
			return fmt.Errorf("shutdown request failed: %s (could not read response body: %w)", shutdownResponse.Status, err)
		}
		return fmt.Errorf("shutdown request failed, status: %s, body: %s", shutdownResponse.Status, string(shutdownResponseBody))
	}

	time.Sleep(clientLivenessStartTimeout)
	livenessRequestCtx, livenessRequestCancel := context.WithTimeout(context.Background(), clientLivenessRequestTimeout)
	defer livenessRequestCancel()
	livenessRequest, err := http.NewRequestWithContext(livenessRequestCtx, http.MethodGet, privateBaseURL+"/livez", nil)
	if err != nil {
		return fmt.Errorf("failed to create liveness request: %w", err)
	}
	livenessResponse, err := client.Do(livenessRequest)
	if err == nil && livenessResponse != nil {
		if closeErr := livenessResponse.Body.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close liveness response body: %v\n", closeErr)
		}
		return fmt.Errorf("server did not shut down properly")
	}
	return nil
}

// StartServerListenerApplication creates and starts a new server application listener.
func StartServerListenerApplication(settings *cryptoutilConfig.Settings) (*ServerApplicationListener, error) {
	ctx := context.Background()

	serverApplicationCore, err := StartServerApplicationCore(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize server application: %w", err)
	}

	var publicTLSServer *TLSServerConfig
	var privateTLSServer *TLSServerConfig
	if settings.BindPublicProtocol == protocolHTTPS || settings.BindPrivateProtocol == protocolHTTPS {
		publicTLSServerSubject, privateTLSServerSubject, err := generateTLSServerSubjects(settings, serverApplicationCore.ServerApplicationBasic)
		if err != nil {
			return nil, fmt.Errorf("failed to run new function: %w", err)
		}

		publicTLSServerCertificate, publicTLSServerRootCACertsPool, publicTLSServerIntermediateCertsPool, err := cryptoutilCertificate.BuildTLSCertificate(publicTLSServerSubject)
		if err != nil {
			return nil, fmt.Errorf("failed to build TLS server certificate: %w", err)
		}
		privateTLSServerCertificate, privateTLSServerRootCACertsPool, privateTLSServerIntermediateCertsPool, err := cryptoutilCertificate.BuildTLSCertificate(privateTLSServerSubject)
		if err != nil {
			return nil, fmt.Errorf("failed to build TLS client certificate: %w", err)
		}

		publicTLSServer = &TLSServerConfig{
			Certificate:         publicTLSServerCertificate,
			RootCAsPool:         publicTLSServerRootCACertsPool,
			IntermediateCAsPool: publicTLSServerIntermediateCertsPool,
			Config: &tls.Config{
				Certificates: []tls.Certificate{*publicTLSServerCertificate},
				ClientAuth:   tls.NoClientCert,
				MinVersion:   tls.VersionTLS12,
			},
		}
		privateTLSServer = &TLSServerConfig{
			Certificate:         privateTLSServerCertificate,
			RootCAsPool:         privateTLSServerRootCACertsPool,
			IntermediateCAsPool: privateTLSServerIntermediateCertsPool,
			Config: &tls.Config{
				Certificates: []tls.Certificate{*privateTLSServerCertificate},
				ClientAuth:   tls.NoClientCert,
				MinVersion:   tls.VersionTLS12,
			},
		}
	}

	// Middlewares

	commonMiddlewares := []fiber.Handler{
		recover.New(),
		requestid.New(),
		logger.New(),   // TODO Replace this with improved otelFiberTelemetryMiddleware; unstructured logs and no OpenTelemetry are undesirable
		compress.New(), // Enable response compression for better performance
		commonOtelFiberTelemetryMiddleware(serverApplicationCore.ServerApplicationBasic.TelemetryService, settings),
		commonOtelFiberRequestLoggerMiddleware(serverApplicationCore.ServerApplicationBasic.TelemetryService),
		commonIPFilterMiddleware(serverApplicationCore.ServerApplicationBasic.TelemetryService, settings),
		commonIPRateLimiterMiddleware(serverApplicationCore.ServerApplicationBasic.TelemetryService, settings),
		commonHTTPGETCacheControlMiddleware(), // TODO Limit this to Swagger GET APIs, not Swagger UI static content
		commonUnsupportedHTTPMethodsMiddleware(settings),
	}

	privateMiddlewares := append([]fiber.Handler{commonSetFiberRequestAttribute(fiberAppIDPrivate)}, commonMiddlewares...)
	privateMiddlewares = append(privateMiddlewares, privateHealthCheckMiddlewareFunction(serverApplicationCore)) // /livez, /readyz
	privateFiberApp := fiber.New(fiber.Config{Immutable: true, BodyLimit: settings.RequestBodyLimit})
	for _, middleware := range privateMiddlewares {
		privateFiberApp.Use(middleware)
	}

	publicMiddlewares := append([]fiber.Handler{commonSetFiberRequestAttribute(fiberAppIDPublic)}, commonMiddlewares...)
	publicMiddlewares = append(publicMiddlewares, publicBrowserCORSMiddlewareFunction(settings))                                                                             // Browser-specific: Cross-Origin Resource Sharing (CORS)
	publicMiddlewares = append(publicMiddlewares, publicBrowserXSSMiddlewareFunction(settings))                                                                              // Browser-specific: Cross-Site Scripting (XSS)
	publicMiddlewares = append(publicMiddlewares, publicBrowserAdditionalSecurityHeadersMiddleware(serverApplicationCore.ServerApplicationBasic.TelemetryService, settings)) // Additional security headers
	publicMiddlewares = append(publicMiddlewares, publicBrowserCSRFMiddlewareFunction(settings))                                                                             // Browser-specific: Cross-Site Request Forgery (CSRF)
	publicFiberApp := fiber.New(fiber.Config{Immutable: true, BodyLimit: settings.RequestBodyLimit})
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
	// TODO Disable Swagger UI in production environments (check settings.DevMode or add settings.Environment)
	// TODO Add authentication middleware for Swagger UI access
	// TODO Add specific rate limiting for Swagger UI endpoints
	swaggerAPI, err := cryptoutilOpenapiServer.GetSwagger()
	if err != nil {
		serverApplicationCore.ServerApplicationBasic.TelemetryService.Slogger.Error("failed to get swagger", "error", err)
		serverApplicationCore.Shutdown()
		return nil, fmt.Errorf("failed to get swagger: %w", err)
	}

	swaggerAPI.Servers = []*openapi3.Server{
		{URL: settings.PublicBrowserAPIContextPath}, // Browser users will access the APIs via this context path, with browser middlewares (CORS, CSRF, etc)
		{URL: settings.PublicServiceAPIContextPath}, // Service clients will access the APIs via this context path, without browser middlewares
	}
	swaggerSpecBytes, err := swaggerAPI.MarshalJSON() // Serialize OpenAPI 3 spec to JSON with the added public server context path
	if err != nil {
		serverApplicationCore.ServerApplicationBasic.TelemetryService.Slogger.Error("failed to get fiber handler for OpenAPI spec", "error", err)
		serverApplicationCore.Shutdown()
		return nil, fmt.Errorf("failed to marshal OpenAPI spec: %w", err)
	}

	publicFiberApp.Get("/ui/swagger/doc.json", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.Send(swaggerSpecBytes)
	})
	publicFiberApp.Get("/ui/swagger/*", func(c *fiber.Ctx) error {
		swaggerHandler := swagger.New(swagger.Config{
			Title:                  "Cryptoutil API",
			URL:                    "/ui/swagger/doc.json",
			TryItOutEnabled:        true,
			DisplayRequestDuration: true,
			ShowCommonExtensions:   true,
			CustomScript:           swaggerUICustomCSRFScript(settings.CSRFTokenName, settings.PublicBrowserAPIContextPath),
		})
		err := swaggerHandler(c)
		if err != nil {
			return err
		}
		// Ensure Content-Type includes charset for HTML responses to satisfy security scanners
		if c.Get("Content-Type") == "text/html" {
			c.Set("Content-Type", "text/html; charset=utf-8")
		}
		return nil
	})
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
		fibermiddleware.OapiRequestValidatorWithOptions(swaggerAPI, &fibermiddleware.Options{}),
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

	// Create listeners - use port 0 for testing (to get OS-assigned ports), configured ports for production
	publicBinding := fmt.Sprintf("%s:%d", settings.BindPublicAddress, settings.BindPublicPort)
	privateBinding := fmt.Sprintf("%s:%d", settings.BindPrivateAddress, settings.BindPrivatePort)

	// Create net listeners to get actual assigned ports (port 0 for tests, configured ports for production)
	publicListener, err := net.Listen("tcp", publicBinding)
	if err != nil {
		serverApplicationCore.Shutdown()
		return nil, fmt.Errorf("failed to create public listener: %w", err)
	}

	privateListener, err := net.Listen("tcp", privateBinding)
	if err != nil {
		if closeErr := publicListener.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close public listener during cleanup: %v\n", closeErr)
		}
		serverApplicationCore.Shutdown()
		return nil, fmt.Errorf("failed to create private listener: %w", err)
	}

	// Extract actual assigned ports
	publicAddr, ok := publicListener.Addr().(*net.TCPAddr)
	if !ok {
		return nil, fmt.Errorf("failed to get public listener address")
	}
	if publicAddr.Port < 0 || publicAddr.Port > 65535 {
		return nil, fmt.Errorf("invalid public port: %d", publicAddr.Port)
	}
	actualPublicPort := uint16(publicAddr.Port)

	privateAddr, ok := privateListener.Addr().(*net.TCPAddr)
	if !ok {
		return nil, fmt.Errorf("failed to get private listener address")
	}
	if privateAddr.Port < 0 || privateAddr.Port > 65535 {
		return nil, fmt.Errorf("invalid private port: %d", privateAddr.Port)
	}
	actualPrivatePort := uint16(privateAddr.Port)

	serverApplicationCore.ServerApplicationBasic.TelemetryService.Slogger.Info("assigned ports",
		"public", actualPublicPort, "private", actualPrivatePort)

	startServerFunction := startServerFuncWithListeners(
		publicListener, privateListener,
		publicFiberApp, privateFiberApp,
		settings.BindPublicProtocol, settings.BindPrivateProtocol,
		publicTLSServer.Config, privateTLSServer.Config,
		serverApplicationCore.ServerApplicationBasic.TelemetryService)
	shutdownServerFunction = stopServerFuncWithListeners(serverApplicationCore, publicFiberApp, privateFiberApp, publicListener, privateListener, settings)

	go stopServerSignalFunc(serverApplicationCore.ServerApplicationBasic.TelemetryService, shutdownServerFunction)() // listen for OS signals to gracefully shutdown the server

	return &ServerApplicationListener{
		StartFunction:     startServerFunction,
		ShutdownFunction:  shutdownServerFunction,
		PublicTLSServer:   publicTLSServer,
		PrivateTLSServer:  privateTLSServer,
		ActualPublicPort:  actualPublicPort,
		ActualPrivatePort: actualPrivatePort,
	}, nil
}

func startServerFuncWithListeners(publicListener, privateListener net.Listener, publicFiberApp, privateFiberApp *fiber.App, publicProtocol, privateProtocol string, publicTLSConfig, privateTLSConfig *tls.Config, telemetryService *cryptoutilTelemetry.TelemetryService) func() {
	return func() {
		telemetryService.Slogger.Debug("starting fiber listeners with pre-created listeners")

		go func() {
			telemetryService.Slogger.Debug("starting private fiber listener", "addr", privateListener.Addr().String(), "protocol", privateProtocol)
			var err error
			if privateProtocol == protocolHTTPS && privateTLSConfig != nil {
				// Wrap the listener with TLS
				tlsListener := tls.NewListener(privateListener, privateTLSConfig)
				err = privateFiberApp.Listener(tlsListener)
			} else {
				err = privateFiberApp.Listener(privateListener)
			}
			if err != nil {
				telemetryService.Slogger.Error("failed to start private fiber listener", "error", err)
			}
			telemetryService.Slogger.Debug("private fiber listener stopped")
		}()

		telemetryService.Slogger.Debug("starting public fiber listener", "addr", publicListener.Addr().String(), "protocol", publicProtocol)
		var err error
		if publicProtocol == protocolHTTPS && publicTLSConfig != nil {
			// Wrap the listener with TLS
			tlsListener := tls.NewListener(publicListener, publicTLSConfig)
			err = publicFiberApp.Listener(tlsListener)
		} else {
			err = publicFiberApp.Listener(publicListener)
		}
		if err != nil {
			telemetryService.Slogger.Error("failed to start public fiber listener", "error", err)
		}
		telemetryService.Slogger.Debug("public fiber listener stopped")

		ready.Store(true)
	}
}

func stopServerFuncWithListeners(serverApplicationCore *ServerApplicationCore, publicFiberApp, privateFiberApp *fiber.App, publicListener, privateListener net.Listener, settings *cryptoutilConfig.Settings) func() {
	return func() {
		if serverApplicationCore.ServerApplicationBasic.TelemetryService != nil {
			serverApplicationCore.ServerApplicationBasic.TelemetryService.Slogger.Debug("stopping servers")
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), settings.ServerShutdownTimeout)
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

		// Close the listeners if they're still open (they should be closed by Fiber, but just in case)
		if publicListener != nil {
			if err := publicListener.Close(); err != nil {
				serverApplicationCore.ServerApplicationBasic.TelemetryService.Slogger.Debug("public listener already closed", "error", err)
			}
		}
		if privateListener != nil {
			if err := privateListener.Close(); err != nil {
				serverApplicationCore.ServerApplicationBasic.TelemetryService.Slogger.Debug("private listener already closed", "error", err)
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
	if len(settings.AllowedIPs) > 0 {
		for _, allowedIP := range settings.AllowedIPs {
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
	if len(settings.AllowedCIDRs) > 0 {
		for _, allowedCIDR := range settings.AllowedCIDRs {
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

func checkDatabaseHealth(serverApplicationCore *ServerApplicationCore) map[string]any {
	if serverApplicationCore.SQLRepository == nil {
		return map[string]any{
			"status": "error",
			"error":  "SQL repository not initialized",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health, err := serverApplicationCore.SQLRepository.HealthCheck(ctx)
	if err != nil {
		return health // HealthCheck already returns the error details
	}

	return health
}

func checkMemoryHealth() map[string]any {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]any{
		"status":         "ok",
		"heap_alloc":     m.HeapAlloc,
		"heap_sys":       m.HeapSys,
		"heap_idle":      m.HeapIdle,
		"heap_released":  m.HeapReleased,
		"stack_inuse":    m.StackInuse,
		"stack_sys":      m.StackSys,
		"gc_cycles":      m.NumGC,
		"gc_pause_total": m.PauseTotalNs,
		"num_goroutines": runtime.NumGoroutine(),
	}
}

func checkDependenciesHealth(serverApplicationCore *ServerApplicationCore) map[string]any {
	deps := map[string]any{
		statusStr:  "ok",
		"services": map[string]any{},
	}

	services, ok := deps["services"].(map[string]any)
	if !ok {
		return map[string]any{
			statusStr: errorStr,
			errorStr:  "internal error: failed to create dependencies status map",
		}
	}

	// Check telemetry service
	if serverApplicationCore.ServerApplicationBasic.TelemetryService == nil {
		services["telemetry"] = map[string]any{statusStr: errorStr, errorStr: "not initialized"}
		deps[statusStr] = errorStr
	} else {
		services["telemetry"] = map[string]any{statusStr: "ok"}
	}

	// Check JWK gen service
	if serverApplicationCore.ServerApplicationBasic.JWKGenService == nil {
		services["jwk_generator"] = map[string]any{statusStr: errorStr, errorStr: "not initialized"}
		deps[statusStr] = errorStr
	} else {
		services["jwk_generator"] = map[string]any{statusStr: "ok"}
	}

	// Check barrier service
	if serverApplicationCore.BarrierService == nil {
		services["barrier"] = map[string]any{statusStr: errorStr, errorStr: "not initialized"}
		deps[statusStr] = errorStr
	} else {
		services["barrier"] = map[string]any{statusStr: "ok"}
	}

	// Check business logic service
	if serverApplicationCore.BusinessLogicService == nil {
		services["business_logic"] = map[string]any{statusStr: errorStr, errorStr: "not initialized"}
		deps[statusStr] = errorStr
	} else {
		services["business_logic"] = map[string]any{statusStr: "ok"}
	}

	// Check ORM repository
	if serverApplicationCore.OrmRepository == nil {
		services["orm_repository"] = map[string]any{statusStr: errorStr, errorStr: "not initialized"}
		deps[statusStr] = errorStr
	} else {
		services["orm_repository"] = map[string]any{statusStr: "ok"}
	}

	return deps
}

func commonUnsupportedHTTPMethodsMiddleware(settings *cryptoutilConfig.Settings) fiber.Handler {
	return func(c *fiber.Ctx) error {
		method := c.Method()
		for _, supported := range settings.CORSAllowedMethods {
			if method == supported {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Method not allowed")
	}
}

func privateHealthCheckMiddlewareFunction(serverApplicationCore *ServerApplicationCore) fiber.Handler {
	// Enhanced health checks with detailed status (database, dependencies, memory usage)
	return func(c *fiber.Ctx) error {
		// Check if this is a liveness or readiness probe
		path := c.Path()
		isReadiness := strings.HasSuffix(path, "/readyz")

		healthStatus := map[string]any{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"service":   "cryptoutil",
			"version":   "1.0.0", // TODO: Get from build info
			"probe":     "liveness",
		}

		if isReadiness {
			healthStatus["probe"] = "readiness"
			healthStatus["database"] = checkDatabaseHealth(serverApplicationCore)
			healthStatus["memory"] = checkMemoryHealth()
			healthStatus["dependencies"] = checkDependenciesHealth(serverApplicationCore)

			// Check if any component is unhealthy for readiness
			if dbStatus, ok := healthStatus["database"].(map[string]any); ok {
				if status, ok := dbStatus["status"].(string); ok && status != "ok" {
					healthStatus["status"] = "degraded"
				}
			}

			if depsStatus, ok := healthStatus["dependencies"].(map[string]any); ok {
				if status, ok := depsStatus["status"].(string); ok && status != "ok" {
					healthStatus["status"] = "degraded"
				}
			}
		}

		statusCode := fiber.StatusOK
		if healthStatus["status"] != "ok" {
			statusCode = fiber.StatusServiceUnavailable
		}

		return c.Status(statusCode).JSON(healthStatus)
	}
}

func commonSetFiberRequestAttribute(fiberAppIDValue fiberAppID) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Locals(fiberAppIDRequestAttribute, string(fiberAppIDValue))
		return c.Next()
	}
}

func publicBrowserCORSMiddlewareFunction(settings *cryptoutilConfig.Settings) fiber.Handler {
	return cors.New(cors.Config{ // Cross-Origin Resource Sharing (CORS)
		AllowOrigins: strings.Join(settings.CORSAllowedOrigins, ","), // cryptoutilConfig.defaultAllowedCORSOrigins
		AllowMethods: strings.Join(settings.CORSAllowedMethods, ","), // cryptoutilConfig.defaultAllowedCORSMethods
		AllowHeaders: strings.Join(settings.CORSAllowedHeaders, ","), // cryptoutilConfig.defaultAllowedCORSHeaders
		MaxAge:       int(settings.CORSMaxAge),
		Next:         isNonBrowserUserAPIRequestFunc(settings), // Skip check for /service/api/v1/* requests by non-browser clients
	})
}

func publicBrowserXSSMiddlewareFunction(settings *cryptoutilConfig.Settings) fiber.Handler {
	// Content Security Policy for enhanced XSS protection
	// This CSP is specifically designed to work with Swagger UI while maintaining security
	csp := buildContentSecurityPolicy(settings)

	return helmet.New(helmet.Config{
		Next: isNonBrowserUserAPIRequestFunc(settings), // Skip check for /service/api/v1/* requests by non-browser clients

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
func buildContentSecurityPolicy(settings *cryptoutilConfig.Settings) string {
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
	hstsMaxAge                    = "max-age=31536000; includeSubDomains; preload"
	hstsMaxAgeDev                 = "max-age=86400; includeSubDomains"
	referrerPolicy                = "strict-origin-when-cross-origin"
	permissionsPolicy             = "camera=(), microphone=(), geolocation=(), payment=(), usb=(), accelerometer=(), gyroscope=(), magnetometer=()"
	crossOriginOpenerPolicy       = "same-origin"
	crossOriginEmbedderPolicy     = "require-corp"
	crossOriginResourcePolicy     = "same-origin"
	xPermittedCrossDomainPolicies = "none"
	contentTypeOptions            = "nosniff"
	clearSiteDataLogout           = "\"cache\", \"cookies\", \"storage\""
)

// Expected browser security headers for runtime validation.
var expectedBrowserHeaders = map[string]string{
	"X-Content-Type-Options":            contentTypeOptions,
	"Referrer-Policy":                   referrerPolicy,
	"Permissions-Policy":                permissionsPolicy,
	"Cross-Origin-Opener-Policy":        crossOriginOpenerPolicy,
	"Cross-Origin-Embedder-Policy":      crossOriginEmbedderPolicy,
	"Cross-Origin-Resource-Policy":      crossOriginResourcePolicy,
	"X-Permitted-Cross-Domain-Policies": xPermittedCrossDomainPolicies,
}

// publicBrowserAdditionalSecurityHeadersMiddleware adds security headers not covered by Helmet.
func publicBrowserAdditionalSecurityHeadersMiddleware(telemetryService *cryptoutilTelemetry.TelemetryService, settings *cryptoutilConfig.Settings) fiber.Handler {
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
		// Apply common security headers
		c.Set("X-Content-Type-Options", contentTypeOptions)
		c.Set("Referrer-Policy", referrerPolicy)
		if c.Protocol() == protocolHTTPS {
			if settings.DevMode {
				c.Set("Strict-Transport-Security", hstsMaxAgeDev)
			} else {
				c.Set("Strict-Transport-Security", hstsMaxAge)
			}
		}

		// Skip for non-browser API requests
		if isNonBrowserUserAPIRequestFunc(settings)(c) {
			return c.Next()
		}

		// Apply common browser-only security headers
		c.Set("Permissions-Policy", permissionsPolicy)
		c.Set("Cross-Origin-Opener-Policy", crossOriginOpenerPolicy)
		c.Set("Cross-Origin-Embedder-Policy", crossOriginEmbedderPolicy)
		c.Set("Cross-Origin-Resource-Policy", crossOriginResourcePolicy)
		c.Set("X-Permitted-Cross-Domain-Policies", xPermittedCrossDomainPolicies)

		// Clear-Site-Data for logout endpoints only
		if c.Method() == fiber.MethodPost && strings.HasSuffix(c.OriginalURL(), "/logout") {
			c.Set("Clear-Site-Data", clearSiteDataLogout)
		}

		// Process the request
		err := c.Next()

		// Runtime self-check: validate expected headers are present in response
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
	if c.Protocol() == protocolHTTPS {
		if hsts := c.Get("Strict-Transport-Security"); hsts == "" {
			missing = append(missing, "Strict-Transport-Security")
		}
	}

	return missing
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
		Next:              isNonBrowserUserAPIRequestFunc(settings), // Skip check for /service/api/v1/* requests by non-browser clients
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
// ASSUME: UI Authentication only authorizes browser users to access /browser/api/v1/*.
func isNonBrowserUserAPIRequestFunc(settings *cryptoutilConfig.Settings) func(c *fiber.Ctx) bool {
	return func(c *fiber.Ctx) bool {
		return strings.HasPrefix(c.OriginalURL(), settings.PublicServiceAPIContextPath+"/")
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
