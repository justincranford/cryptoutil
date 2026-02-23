// Copyright (c) 2025 Justin Cranford
//
//

package listener

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceConfigTlsGenerator "cryptoutil/internal/apps/template/service/config/tls_generator"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// publicListenFn is injectable for testing the TCP listener creation in Start().
var publicListenFn = func(ctx context.Context, network, address string) (net.Listener, error) {
	return (&net.ListenConfig{}).Listen(ctx, network, address)
}

// publicAppListenerFn is injectable for testing the Fiber app.Listener call in Start().
var publicAppListenerFn = func(app *fiber.App, ln net.Listener) error {
	return app.Listener(ln)
}

// PublicHTTPServer implements the PublicServer interface for business logic APIs and UIs.
// Binds to configurable address and port from ServiceTemplateServerSettings.
//
// Request Path Prefixes:
// - /service/** : Service-to-service APIs (headless clients, IP allowlist, rate limiting)
// - /browser/** : Browser-to-service APIs/UI (sessions, CSRF, CORS, CSP headers)
//
// Both paths serve the SAME OpenAPI specification but with different middleware stacks.
type PublicHTTPServer struct {
	app         *fiber.App
	settings    *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
	actualPort  int
	listener    net.Listener
	tlsMaterial *cryptoutilAppsTemplateServiceConfig.TLSMaterial
	mu          sync.RWMutex
	shutdown    bool
}

// NewPublicHTTPServer creates a new public HTTPS server instance.
//
// The server starts in shutdown=false state and ready=false state.
// Applications must call SetReady(true) after initializing dependencies (database, cache, etc.).
//
// Parameters:
// - ctx: Context for initialization (must not be nil)
// - settings: ServiceTemplateServerSettings containing bind address, port, and paths (must not be nil)
// - tlsCfg: TLS configuration (mode, certificates, parameters)
//
// Returns:
// - *PublicHTTPServer: Server instance ready to Start()
// - error: Non-nil if initialization fails (nil context, TLS generation failure, Fiber setup failure).
func NewPublicHTTPServer(ctx context.Context, settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings, tlsCfg *cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings) (*PublicHTTPServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if settings == nil {
		return nil, fmt.Errorf("settings cannot be nil")
	} else if tlsCfg == nil {
		return nil, fmt.Errorf("TLS config cannot be nil")
	}

	// Generate TLS material based on configured mode.
	tlsMaterial, err := generateTLSMaterialFn(tlsCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS material: %w", err)
	}

	server := &PublicHTTPServer{
		settings:    settings,
		tlsMaterial: tlsMaterial,
	}

	server.app = fiber.New(fiber.Config{
		DisableStartupMessage: true,
		AppName:               "Public API",
		ReadTimeout:           cryptoutilSharedMagic.DefaultHTTPServerTimeoutSeconds * time.Second,
		WriteTimeout:          cryptoutilSharedMagic.DefaultHTTPServerTimeoutSeconds * time.Second,
		IdleTimeout:           cryptoutilSharedMagic.DefaultHTTPServerTimeoutSeconds * time.Second,
	})

	// Register public routes (placeholder - to be implemented by services).
	server.registerRoutes()

	return server, nil
}

// registerRoutes registers public HTTP endpoints.
// This is a placeholder - services will inject their own route handlers.
func (s *PublicHTTPServer) registerRoutes() {
	// Service-to-service paths.
	s.app.Get("/service/api/v1/health", s.handleServiceHealth)

	// Browser-to-service paths.
	s.app.Get("/browser/api/v1/health", s.handleBrowserHealth)
}

// handleServiceHealth returns health status for service-to-service clients.
func (s *PublicHTTPServer) handleServiceHealth(c *fiber.Ctx) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.shutdown {
		if err := c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "shutting down",
		}); err != nil {
			return fmt.Errorf("failed to send service health shutdown response: %w", err)
		}

		return nil
	}

	if err := c.JSON(fiber.Map{
		"status": "healthy",
	}); err != nil {
		return fmt.Errorf("failed to send service health response: %w", err)
	}

	return nil
}

// handleBrowserHealth returns health status for browser clients.
func (s *PublicHTTPServer) handleBrowserHealth(c *fiber.Ctx) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.shutdown {
		if err := c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "shutting down",
		}); err != nil {
			return fmt.Errorf("failed to send browser health shutdown response: %w", err)
		}

		return nil
	}

	if err := c.JSON(fiber.Map{
		"status": "healthy",
	}); err != nil {
		return fmt.Errorf("failed to send browser health response: %w", err)
	}

	return nil
}

// Start starts the public HTTPS server and blocks until shutdown or error.
//
// The server:
// 1. Uses TLS material generated during NewPublicHTTPServer (configured mode)
// 2. Creates TCP listener on configured address and port from ServiceTemplateServerSettings
// 3. Starts HTTPS server with Fiber app
// 4. Blocks until context cancelled or server error
// 5. Triggers graceful shutdown on context cancellation
//
// Parameters:
// - ctx: Context for server lifecycle (cancellation triggers shutdown)
//
// Returns:
// - error: Non-nil if server fails to start or encounters runtime error.
func (s *PublicHTTPServer) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	// Create TCP listener using address and port from ServiceTemplateServerSettings.
	listener, err := publicListenFn(ctx, "tcp", fmt.Sprintf("%s:%d", s.settings.BindPublicAddress, s.settings.BindPublicPort))
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	s.mu.Lock()
	s.listener = listener

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		s.mu.Unlock()

		return fmt.Errorf("listener address is not *net.TCPAddr")
	}

	s.actualPort = tcpAddr.Port
	s.mu.Unlock()

	// Create TLS listener using configured TLS material.
	tlsListener := tls.NewListener(listener, s.tlsMaterial.Config)

	// Start server in goroutine.
	errChan := make(chan error, 1)

	go func() {
		if err := publicAppListenerFn(s.app, tlsListener); err != nil {
			errChan <- fmt.Errorf("public server error: %w", err)
		} else {
			errChan <- nil
		}
	}()

	// Wait for either context cancellation or server error.
	select {
	case <-ctx.Done():
		// Context cancelled - trigger graceful shutdown.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultServerShutdownTimeout)
		defer cancel()

		_ = s.Shutdown(shutdownCtx)

		return fmt.Errorf("public server stopped: %w", ctx.Err())
	case err := <-errChan:
		return err
	}
}

// Shutdown gracefully shuts down the public server.
func (s *PublicHTTPServer) Shutdown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	s.mu.Lock()

	if s.shutdown {
		s.mu.Unlock()

		return fmt.Errorf("public server already shutdown")
	}

	s.shutdown = true
	s.mu.Unlock()

	// Shutdown Fiber app.
	if err := s.app.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("failed to shutdown public server: %w", err)
	}

	return nil
}

// ActualPort returns the actual port the server is listening on (after dynamic allocation).
func (s *PublicHTTPServer) ActualPort() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.actualPort
}

// PublicBaseURL returns the base URL for public API access.
func (s *PublicHTTPServer) PublicBaseURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return fmt.Sprintf("%s://%s:%d", s.settings.BindPublicProtocol, s.settings.BindPublicAddress, s.actualPort)
}

// App returns the underlying fiber.App for in-memory testing.
// This allows tests to use app.Test() without starting an HTTPS listener.
func (s *PublicHTTPServer) App() *fiber.App {
	return s.app
}
