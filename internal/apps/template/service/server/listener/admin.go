// Copyright (c) 2025 Justin Cranford
//
//

// Package listener provides a reusable template for dual HTTPS server pattern used across all cryptoutil services.
// AdminServer implements the private admin API server with health check endpoints and graceful shutdown.
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
)

// AdminServer represents the private admin API server for health checks and graceful shutdown.
// Binds to address and port from ServiceTemplateServerSettings.
type AdminServer struct {
	app         *fiber.App
	listener    net.Listener
	settings    *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
	actualPort  uint16
	tlsMaterial *cryptoutilAppsTemplateServiceConfig.TLSMaterial
	mu          sync.RWMutex
	ready       bool
	shutdown    bool
}

// NewAdminHTTPServer creates a new admin server instance for private administrative operations.
// settings: ServiceTemplateServerSettings containing bind address, port, and paths (MUST NOT be nil).
// tlsCfg: TLS configuration (mode + parameters) for HTTPS server. MUST NOT be nil.
func NewAdminHTTPServer(ctx context.Context, settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings, tlsCfg *cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings) (*AdminServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	if settings == nil {
		return nil, fmt.Errorf("settings cannot be nil")
	}

	if tlsCfg == nil {
		return nil, fmt.Errorf("TLS configuration cannot be nil")
	}

	// Generate TLS material based on configured mode.
	tlsMaterial, err := cryptoutilAppsTemplateServiceConfigTlsGenerator.GenerateTLSMaterial(tlsCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS material: %w", err)
	}

	server := &AdminServer{
		settings:    settings,
		tlsMaterial: tlsMaterial,
		ready:       false,
		shutdown:    false,
	}

	// Create Fiber app with minimal configuration.
	const defaultTimeout = 30

	server.app = fiber.New(fiber.Config{
		DisableStartupMessage: true,
		AppName:               "Admin API",
		ReadTimeout:           defaultTimeout * time.Second,
		WriteTimeout:          defaultTimeout * time.Second,
		IdleTimeout:           defaultTimeout * time.Second,
	})

	// Register admin routes.
	server.registerRoutes()

	return server, nil
}

// registerRoutes sets up admin API endpoints.
func (s *AdminServer) registerRoutes() {
	api := s.app.Group("/admin/api/v1")

	// Health check endpoints.
	api.Get("/livez", s.handleLivez)
	api.Get("/readyz", s.handleReadyz)

	// Graceful shutdown endpoint.
	api.Post("/shutdown", s.handleShutdown)
}

// handleLivez returns liveness status (200 if server is running).
// Liveness check: Is the process alive? Failure action: restart container.
func (s *AdminServer) handleLivez(c *fiber.Ctx) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.shutdown {
		if err := c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "shutting down",
		}); err != nil {
			return fmt.Errorf("failed to send livez shutdown response: %w", err)
		}

		return nil
	}

	if err := c.JSON(fiber.Map{
		"status": "alive",
	}); err != nil {
		return fmt.Errorf("failed to send livez response: %w", err)
	}

	return nil
}

// handleReadyz returns readiness status (200 if server is ready to accept traffic).
// Readiness check: Is the service ready? Failure action: remove from load balancer (do NOT restart).
func (s *AdminServer) handleReadyz(c *fiber.Ctx) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.shutdown {
		if err := c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "shutting down",
		}); err != nil {
			return fmt.Errorf("failed to send readyz shutdown response: %w", err)
		}

		return nil
	}

	if !s.ready {
		if err := c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "not ready",
		}); err != nil {
			return fmt.Errorf("failed to send readyz not-ready response: %w", err)
		}

		return nil
	}

	if err := c.JSON(fiber.Map{
		"status": "ready",
	}); err != nil {
		return fmt.Errorf("failed to send readyz response: %w", err)
	}

	return nil
}

// handleShutdown initiates graceful shutdown of the admin server.
func (s *AdminServer) handleShutdown(c *fiber.Ctx) error {
	s.mu.Lock()
	s.shutdown = true
	s.mu.Unlock()

	// Acknowledge shutdown request.
	_ = c.JSON(fiber.Map{
		"status": "shutdown initiated",
	})

	// Trigger shutdown in background to avoid blocking response.
	go func() {
		// Wait for response to be sent.
		const shutdownDelay = 100 * time.Millisecond

		time.Sleep(shutdownDelay)

		// Shutdown server gracefully.
		const shutdownTimeout = 5 * time.Second

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		_ = s.Shutdown(ctx)
	}()

	return nil
}

// Start begins listening on configured address and port from ServiceTemplateServerSettings for admin API requests.
// This method blocks until shutdown is called or context is cancelled.
func (s *AdminServer) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	// Bind to address and port from ServiceTemplateServerSettings.
	addr := fmt.Sprintf("%s:%d", s.settings.BindPrivateAddress, s.settings.BindPrivatePort)

	// Create listener.
	var lc net.ListenConfig

	listener, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to create admin listener: %w", err)
	}

	s.listener = listener

	// Store actual port if dynamic allocation was used (port 0).
	// Use mutex to protect actualPort writes for concurrent access safety.
	s.mu.Lock()
	if s.settings.BindPrivatePort == 0 {
		tcpAddr, ok := listener.Addr().(*net.TCPAddr)
		if !ok {
			s.mu.Unlock()
			_ = listener.Close()

			return fmt.Errorf("listener address is not a TCP address")
		}

		if tcpAddr.Port < 0 || tcpAddr.Port > 65535 {
			s.mu.Unlock()
			_ = listener.Close()

			return fmt.Errorf("invalid port number: %d", tcpAddr.Port)
		}

		s.actualPort = uint16(tcpAddr.Port) //nolint:gosec // Port range validated above.
	} else {
		s.actualPort = s.settings.BindPrivatePort
	}
	s.mu.Unlock()

	// Wrap with TLS using pre-generated TLS configuration.
	tlsListener := tls.NewListener(listener, s.tlsMaterial.Config)

	// Note: Server starts with ready=false. Application should call SetReady(true) after initializing dependencies.

	// Start Fiber server in goroutine and monitor context cancellation.
	errChan := make(chan error, 1)

	go func() {
		if err := s.app.Listener(tlsListener); err != nil {
			errChan <- fmt.Errorf("admin server error: %w", err)
		} else {
			errChan <- nil
		}
	}()

	// Wait for either context cancellation or server error.
	select {
	case <-ctx.Done():
		// Context cancelled - trigger graceful shutdown.
		const shutdownTimeout = 5 * time.Second

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		_ = s.Shutdown(shutdownCtx)

		return fmt.Errorf("admin server stopped: %w", ctx.Err())
	case err := <-errChan:
		return err
	}
}

// Shutdown gracefully shuts down the admin server.
func (s *AdminServer) Shutdown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	s.mu.Lock()

	if s.shutdown {
		s.mu.Unlock()

		return nil // Already shutdown, return success.
	}

	s.shutdown = true
	s.ready = false
	s.mu.Unlock()

	// Shutdown Fiber app (this automatically closes the listener).
	if s.app != nil {
		if err := s.app.ShutdownWithContext(ctx); err != nil {
			return fmt.Errorf("failed to shutdown admin app: %w", err)
		}
	}

	// Do NOT explicitly close listener - Fiber's Shutdown already did this.
	// Attempting to close again causes "use of closed network connection" errors.

	return nil
}

// ActualPort returns the actual port the admin server is listening on.
// Returns 0 before Start() is called, or the dynamically allocated port after Start().
func (s *AdminServer) ActualPort() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return int(s.actualPort)
}

// AdminBaseURL returns the base URL for admin API access.
func (s *AdminServer) AdminBaseURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return fmt.Sprintf("%s://%s:%d", s.settings.BindPrivateProtocol, s.settings.BindPrivateAddress, s.actualPort)
}

// App returns the underlying fiber.App for custom route registration.
// This allows callers to register additional admin endpoints before calling Start().
// Thread-safe with read lock.
func (s *AdminServer) App() *fiber.App {
	return s.app
}

// SetReady marks the server as ready to accept traffic.
// This is called by the application after dependencies are initialized.
// Thread-safe with full Lock.
func (s *AdminServer) SetReady(ready bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ready = ready
}
