// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedTelemetry "cryptoutil/internal/apps/template/service/telemetry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/gofiber/contrib/otelfiber"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func stopServerFuncWithListeners(serverApplicationCore *ServerApplicationCore, publicFiberApp, privateFiberApp *fiber.App, publicListener, privateListener net.Listener, settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) func() {
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

func stopServerSignalFunc(telemetryService *cryptoutilSharedTelemetry.TelemetryService, stopServerFunc func()) func() {
	return func() {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
		defer stop()

		<-ctx.Done() // blocks until signal is received
		telemetryService.Slogger.Warn("received stop server signal")
		stopServerFunc()
	}
}

func commonOtelFiberTelemetryMiddleware(telemetryService *cryptoutilSharedTelemetry.TelemetryService, settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) fiber.Handler {
	return otelfiber.Middleware(
		otelfiber.WithTracerProvider(telemetryService.TracesProvider),
		otelfiber.WithMeterProvider(telemetryService.MetricsProvider),
		otelfiber.WithPropagators(*telemetryService.TextMapPropagator),
		otelfiber.WithServerName(settings.BindPublicAddress),
		otelfiber.WithPort(int(settings.BindPublicPort)),
	)
}

func commonIPFilterMiddleware(telemetryService *cryptoutilSharedTelemetry.TelemetryService, settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) func(c *fiber.Ctx) error {
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
		switch c.Locals(cryptoutilSharedMagic.FiberAppIDRequestAttribute) {
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
			telemetryService.Slogger.Error("Unexpected app ID:", c.Locals(cryptoutilSharedMagic.FiberAppIDRequestAttribute))

			return c.Status(fiber.StatusInternalServerError).SendString("Internal server error")
		}
	}
}

func commonIPRateLimiterMiddleware(telemetryService *cryptoutilSharedTelemetry.TelemetryService, ipRateLimit int) fiber.Handler {
	return limiter.New(limiter.Config{ // Mitigate DOS by throttling clients
		Max:        ipRateLimit,
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
	return func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")

		return c.Next()
	}
}

func checkDatabaseHealth(serverApplicationCore *ServerApplicationCore) map[string]any {
	if serverApplicationCore.OrmRepository == nil {
		return map[string]any{
			"status": "error",
			"error":  "ORM repository not initialized",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DatabaseHealthCheckTimeout)
	defer cancel()

	health, err := serverApplicationCore.OrmRepository.HealthCheck(ctx)
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
		"heap_alloc":     m.Alloc,
		"num_goroutines": runtime.NumGoroutine(),
	}
}

func checkSidecarHealth(serverApplicationCore *ServerApplicationCore) map[string]any {
	// Only check sidecar health if OTLP is enabled
	if !serverApplicationCore.Settings.OTLPEnabled {
		return map[string]any{
			"status": "disabled",
			"note":   "OTLP export is disabled",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.OtelCollectorHealthCheckTimeout)
	defer cancel()

	// Check sidecar connectivity using the telemetry service method
	err := serverApplicationCore.ServerApplicationBasic.TelemetryService.CheckSidecarHealth(ctx)
	if err != nil {
		return map[string]any{
			"status": "error",
			"error":  fmt.Sprintf("sidecar connectivity check failed: %v", err),
		}
	}

	return map[string]any{
		"status":   "ok",
		"endpoint": serverApplicationCore.Settings.OTLPEndpoint,
	}
}

func checkDependenciesHealth(_ *ServerApplicationCore) map[string]any {
	// No external dependencies (APIs, message queues, etc.) are currently used by kms cryptoutil server
	// Database and telemetry sidecar health are checked separately in their own health endpoints
	services := map[string]any{}

	return map[string]any{
		"status":   "ok",
		"services": services,
		"note":     "No external dependencies configured",
	}
}

func commonUnsupportedHTTPMethodsMiddleware(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) fiber.Handler {
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

func swaggerUIBasicAuthMiddleware(username, password string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// If no username/password configured, skip authentication
		if username == "" && password == "" {
			return c.Next()
		}

		// Check for Authorization header
		auth := c.Get("Authorization")
		if auth == "" {
			c.Set("WWW-Authenticate", `Basic realm="Swagger UI"`)

			return c.Status(fiber.StatusUnauthorized).SendString("Authentication required")
		}

		// Parse Basic Auth
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

		// Check credentials
		if reqUsername != username || reqPassword != password {
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid credentials")
		}

		return c.Next()
	}
}

func privateHealthCheckMiddlewareFunction(serverApplicationCore *ServerApplicationCore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if this is a liveness or readiness probe
		path := c.Path()
		adminContextPath := serverApplicationCore.Settings.PrivateAdminAPIContextPath
		isReadiness := strings.HasSuffix(path, adminContextPath+cryptoutilSharedMagic.PrivateAdminReadyzRequestPath)
		isLiveness := strings.HasSuffix(path, adminContextPath+cryptoutilSharedMagic.PrivateAdminLivezRequestPath)

		// If not a health check path, continue to next middleware
		if !isReadiness && !isLiveness {
			return c.Next()
		}

		healthStatus := map[string]any{
			cryptoutilSharedMagic.StringStatus: "ok",
			"timestamp":                        time.Now().UTC().Format(time.RFC3339),
			"service":                          "cryptoutil",
			"version":                          cryptoutilSharedMagic.ServiceVersion,
			"probe":                            "liveness",
		}

		if isReadiness {
			// Readiness checks framework is in place: database, memory, sidecar, dependencies
			// Additional checks (e.g., crypto operations, key generation) can be added here as needed
			healthStatus["probe"] = "readiness"

			// Perform readiness checks concurrently for better performance
			readinessResults := performConcurrentReadinessChecks(serverApplicationCore)

			// Add results to health status
			for checkName, result := range readinessResults {
				healthStatus[checkName] = result
			}

			// Check if any component is unhealthy for readiness
			if dbStatus, ok := healthStatus["database"].(map[string]any); ok {
				if status, ok := dbStatus[cryptoutilSharedMagic.StringStatus].(string); ok && status != cryptoutilSharedMagic.StringStatusOK {
					healthStatus[cryptoutilSharedMagic.StringStatus] = cryptoutilSharedMagic.StringStatusDegraded
				}
			}

			if depsStatus, ok := healthStatus["dependencies"].(map[string]any); ok {
				if status, ok := depsStatus[cryptoutilSharedMagic.StringStatus].(string); ok && status != cryptoutilSharedMagic.StringStatusOK {
					healthStatus[cryptoutilSharedMagic.StringStatus] = cryptoutilSharedMagic.StringStatusDegraded
				}
			}

			if sidecarStatus, ok := healthStatus["sidecar"].(map[string]any); ok {
				if status, ok := sidecarStatus[cryptoutilSharedMagic.StringStatus].(string); ok && status == "error" {
					healthStatus[cryptoutilSharedMagic.StringStatus] = cryptoutilSharedMagic.StringStatusDegraded
				}
			}
		}

		statusCode := fiber.StatusOK
		if healthStatus[cryptoutilSharedMagic.StringStatus] != cryptoutilSharedMagic.StringStatusOK {
			statusCode = fiber.StatusServiceUnavailable
		}

		return c.Status(statusCode).JSON(healthStatus)
	}
}
