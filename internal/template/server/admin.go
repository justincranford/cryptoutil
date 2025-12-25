// Copyright (c) 2025 Justin Cranford
//
//

// Package server provides a reusable template for dual HTTPS server pattern used across all cryptoutil services.
// AdminServer implements the private admin API server with health check endpoints and graceful shutdown.
package server

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// AdminServer represents the private admin API server for health checks and graceful shutdown.
// Port 0 for tests (dynamic allocation avoids TIME_WAIT), 9090 for production containers.
type AdminServer struct {
	app      *fiber.App
	listener net.Listener
	port     uint16
	mu       sync.RWMutex
	ready    bool
	shutdown bool
}

// NewAdminServer creates a new admin server instance for private administrative operations.
// port: 0 for tests (dynamic allocation), 9090 for production containers, other for non-container production.
func NewAdminServer(ctx context.Context, port uint16) (*AdminServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	server := &AdminServer{
		port:     port,
		ready:    false,
		shutdown: false,
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
	api := s.app.Group("/admin/v1")

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

// Start begins listening on 127.0.0.1 with configured port for admin API requests.
// This method blocks until shutdown is called or context is cancelled.
func (s *AdminServer) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	// Bind to localhost only (127.0.0.1 explicit, not localhost due to IPv6 issues).
	// Port 0 for tests (dynamic allocation), 9090 for production containers.
	addr := fmt.Sprintf("%s:%d", cryptoutilMagic.IPv4Loopback, s.port)

	// Create listener.
	var lc net.ListenConfig

	listener, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to create admin listener: %w", err)
	}

	s.listener = listener

	// Store actual port if dynamic allocation was used.
	if s.port == 0 {
		tcpAddr, ok := listener.Addr().(*net.TCPAddr)
		if !ok {
			_ = listener.Close()

			return fmt.Errorf("listener address is not a TCP address")
		}

		if tcpAddr.Port < 0 || tcpAddr.Port > 65535 {
			_ = listener.Close()

			return fmt.Errorf("invalid port number: %d", tcpAddr.Port)
		}

		s.port = uint16(tcpAddr.Port) //nolint:gosec // Port range validated above.
	}

	// Generate self-signed TLS certificate.
	tlsConfig, err := s.generateTLSConfig()
	if err != nil {
		_ = listener.Close()

		return fmt.Errorf("failed to generate TLS config: %w", err)
	}

	// Wrap with TLS.
	tlsListener := tls.NewListener(listener, tlsConfig)

	// Mark server as ready.
	s.mu.Lock()
	s.ready = true
	s.mu.Unlock()

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

	return int(s.port)
}

// generateTLSConfig creates a self-signed certificate for admin server.
// Uses ECDSA P-256 for FIPS 140-3 compliance, 1-year validity.
func (s *AdminServer) generateTLSConfig() (*tls.Config, error) {
	// Generate private key.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template.
	const (
		validityDays     = 365
		hoursPerDay      = 24
		serialNumberBits = 128
	)

	validityDuration := validityDays * hoursPerDay * time.Hour

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), serialNumberBits))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "cryptoutil-admin",
			Organization: []string{"cryptoutil"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(validityDuration),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP(cryptoutilMagic.IPv4Loopback)},
	}

	// Self-sign certificate.
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Encode certificate and key to PEM.
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	privKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	privKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privKeyBytes})

	// Load certificate.
	cert, err := tls.X509KeyPair(certPEM, privKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to load key pair: %w", err)
	}

	// Create TLS configuration with modern security settings.
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13, // Enforce TLS 1.3+.
		CipherSuites: []uint16{ // FIPS 140-3 approved cipher suites (TLS 1.3 mandatory).
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
	}, nil
}
