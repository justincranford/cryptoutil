// Copyright (c) 2025 Justin Cranford

package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// appListenerFn is an injectable var for testing the app.Listener error path.
var appListenerFn = func(app *fiber.App, ln net.Listener) error {
	//nolint:wrapcheck // Pass-through to Fiber framework.
	return app.Listener(ln)
}

// PublicServerBase provides reusable public server infrastructure.
type PublicServerBase struct {
	bindAddress string
	port        int
	tlsMaterial *cryptoutilAppsTemplateServiceConfig.TLSMaterial
	app         *fiber.App
	mu          sync.RWMutex
	shutdown    bool
	actualPort  int
	ctx         context.Context
	cancel      context.CancelFunc
}

// PublicServerConfig holds configuration for PublicServerBase.
type PublicServerConfig struct {
	BindAddress string
	Port        int
	TLSMaterial *cryptoutilAppsTemplateServiceConfig.TLSMaterial
}

// NewPublicServerBase creates a new PublicServerBase.
func NewPublicServerBase(cfg *PublicServerConfig) (*PublicServerBase, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration cannot be nil")
	}

	if cfg.BindAddress == "" {
		return nil, fmt.Errorf("bind address cannot be empty")
	}

	if cfg.TLSMaterial == nil {
		return nil, fmt.Errorf("TLS material cannot be nil")
	}

	s := &PublicServerBase{
		bindAddress: cfg.BindAddress,
		port:        cfg.Port,
		tlsMaterial: cfg.TLSMaterial,
		app:         fiber.New(fiber.Config{DisableStartupMessage: true}),
	}

	s.registerHealthEndpoints()

	return s, nil
}

// registerHealthEndpoints registers standard health check endpoints.
func (s *PublicServerBase) registerHealthEndpoints() {
	s.app.Get("/service/api/v1/health", s.handleServiceHealth)
	s.app.Get("/browser/api/v1/health", s.handleBrowserHealth)
}

// handleServiceHealth returns health status for service-to-service clients.
func (s *PublicServerBase) handleServiceHealth(c *fiber.Ctx) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.shutdown {
		//nolint:wrapcheck // Fiber framework error.
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "shutting down",
		})
	}

	//nolint:wrapcheck // Fiber framework error.
	return c.JSON(fiber.Map{
		"status": "healthy",
	})
}

// handleBrowserHealth returns health status for browser clients.
func (s *PublicServerBase) handleBrowserHealth(c *fiber.Ctx) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.shutdown {
		//nolint:wrapcheck // Fiber framework error.
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "shutting down",
		})
	}

	//nolint:wrapcheck // Fiber framework error.
	return c.JSON(fiber.Map{
		"status": "healthy",
	})
}

// Start starts the HTTPS server.
func (s *PublicServerBase) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	s.mu.Lock()
	s.ctx, s.cancel = context.WithCancel(ctx)
	serverCtx := s.ctx
	s.mu.Unlock()

	listenConfig := &net.ListenConfig{}

	listener, err := listenConfig.Listen(serverCtx, "tcp", fmt.Sprintf("%s:%d", s.bindAddress, s.port))
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	s.mu.Lock()

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		s.mu.Unlock()

		return fmt.Errorf("listener address is not *net.TCPAddr")
	}

	s.actualPort = tcpAddr.Port
	s.mu.Unlock()

	tlsListener := tls.NewListener(listener, s.tlsMaterial.Config)

	errChan := make(chan error, 1)

	go func() {
		if err := appListenerFn(s.app, tlsListener); err != nil {
			errChan <- fmt.Errorf("public server error: %w", err)
		} else {
			errChan <- nil
		}
	}()

	select {
	case <-serverCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultServerShutdownTimeout)
		defer cancel()

		_ = s.Shutdown(shutdownCtx)

		return fmt.Errorf("public server stopped: %w", serverCtx.Err())
	case err := <-errChan:
		return err
	}
}

// Shutdown gracefully shuts down the server.
func (s *PublicServerBase) Shutdown(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.shutdown {
		return fmt.Errorf("public server already shutdown")
	}

	s.shutdown = true

	if s.cancel != nil {
		s.cancel()
	}

	if s.app != nil {
		if err := s.app.Shutdown(); err != nil {
			return fmt.Errorf("failed to shutdown fiber app: %w", err)
		}
	}

	return nil
}

// ActualPort returns the actual port the server is listening on.
func (s *PublicServerBase) ActualPort() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.actualPort
}

// PublicBaseURL returns the base URL for public API access.
func (s *PublicServerBase) PublicBaseURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return fmt.Sprintf("https://127.0.0.1:%d", s.actualPort)
}

// App returns the underlying Fiber app for route registration.
func (s *PublicServerBase) App() *fiber.App {
	return s.app
}
