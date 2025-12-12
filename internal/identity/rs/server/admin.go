// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
)

// AdminServer represents the private admin API server for resource server.
type AdminServer struct {
	config   *cryptoutilIdentityConfig.Config
	app      *fiber.App
	listener net.Listener
	mu       sync.RWMutex
	ready    bool
	shutdown bool
}

// NewAdminServer creates a new admin server instance for private administrative operations.
func NewAdminServer(
	ctx context.Context,
	config *cryptoutilIdentityConfig.Config,
) (*AdminServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	server := &AdminServer{
		config:   config,
		ready:    false,
		shutdown: false,
	}

	// Create Fiber app with minimal configuration.
	const defaultTimeout = 30

	server.app = fiber.New(fiber.Config{
		DisableStartupMessage: true,
		AppName:               "Identity RS Admin API",
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
	api := s.app.Group("/admin/v1")

	// Health check endpoints.
	api.Get("/livez", s.handleLivez)
	api.Get("/readyz", s.handleReadyz)

	// Graceful shutdown endpoint.
	api.Post("/shutdown", s.handleShutdown)
}

// handleLivez returns liveness status (200 if server is running).
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

	// Trigger async shutdown.
	const shutdownDelay = 200

	go func() {
		time.Sleep(shutdownDelay * time.Millisecond)

		_ = s.Shutdown(context.Background())
	}()

	return nil
}

// Start begins listening for admin API requests on the configured port.
func (s *AdminServer) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	// Load TLS configuration.
	tlsConfig, err := s.loadTLSConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load TLS configuration: %w", err)
	}

	// Create TLS listener on admin bind address and port.
	bindAddr := fmt.Sprintf("%s:%d",
		s.config.RS.AdminBindAddress,
		s.config.RS.AdminPort)

	s.listener, err = tls.Listen("tcp", bindAddr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to create TLS listener on %s: %w", bindAddr, err)
	}

	// Mark server as ready.
	s.mu.Lock()
	s.ready = true
	s.mu.Unlock()

	// Start Fiber server with TLS listener.
	if err := s.app.Listener(s.listener); err != nil {
		return fmt.Errorf("admin server failed: %w", err)
	}

	return nil
}

// loadTLSConfig loads TLS configuration from files specified in config.
func (s *AdminServer) loadTLSConfig(ctx context.Context) (*tls.Config, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	certFile := s.config.RS.TLSCertFile
	keyFile := s.config.RS.TLSKeyFile

	if certFile == "" || keyFile == "" {
		return nil, fmt.Errorf("TLS cert file and key file must be configured")
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate and key: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
	}, nil
}

// Shutdown gracefully shuts down the admin server.
func (s *AdminServer) Shutdown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	s.mu.Lock()
	s.shutdown = true
	s.ready = false
	s.mu.Unlock()

	if s.app != nil {
		if err := s.app.ShutdownWithContext(ctx); err != nil {
			return fmt.Errorf("admin server shutdown failed: %w", err)
		}
	}

	return nil
}

// ActualPort returns the actual port the admin server is listening on.
func (s *AdminServer) ActualPort() (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.listener == nil {
		return 0, fmt.Errorf("admin server listener not initialized")
	}

	tcpAddr, ok := s.listener.Addr().(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("admin listener address is not TCP: %T", s.listener.Addr())
	}

	return tcpAddr.Port, nil
}
