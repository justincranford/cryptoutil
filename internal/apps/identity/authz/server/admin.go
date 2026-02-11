// Copyright (c) 2025 Justin Cranford
//
//

// Package server provides the authorization server implementation for OAuth 2.1.
package server

import (
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"sync"
	"time"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// AdminServer represents the private admin API server for OAuth 2.1 authorization service.
type AdminServer struct {
	config   *cryptoutilIdentityConfig.Config
	app      *fiber.App
	listener net.Listener
	mu       sync.RWMutex
	ready    bool
	shutdown bool
}

// NewAdminHTTPServer creates a new admin server instance for private administrative operations.
func NewAdminHTTPServer(
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
		AppName:               "Identity Admin API",
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
		s.config.AuthZ.AdminBindAddress,
		s.config.AuthZ.AdminPort)

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

// generateSelfSignedTLSConfig generates a self-signed TLS certificate for testing.
func generateSelfSignedTLSConfig() (*tls.Config, error) {
	// Generate ECDSA P-256 key pair.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template.
	serialNumber, err := crand.Int(crand.Reader, new(big.Int).Lsh(big.NewInt(1), cryptoutilSharedMagic.TLSSelfSignedCertSerialNumberBits))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"CryptoUtil Test"},
		},
		NotBefore:             time.Now().UTC(),
		NotAfter:              time.Now().UTC().Add(time.Duration(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Day*cryptoutilSharedMagic.HoursPerDay) * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
		DNSNames:              []string{"localhost"},
	}

	// Create self-signed certificate.
	certDER, err := x509.CreateCertificate(crand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Encode certificate and key to PEM.
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

	// Parse TLS certificate.
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TLS certificate: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
	}, nil
}

// loadTLSConfig loads TLS configuration from files specified in config.
// For testing, if cert/key files are not provided, generates a self-signed certificate.
func (s *AdminServer) loadTLSConfig(ctx context.Context) (*tls.Config, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	certFile := s.config.AuthZ.TLSCertFile
	keyFile := s.config.AuthZ.TLSKeyFile

	// If cert/key files not provided, generate self-signed cert for testing.
	if certFile == "" || keyFile == "" {
		return generateSelfSignedTLSConfig()
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
func (s *AdminServer) ActualPort() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.listener == nil {
		return 0
	}

	tcpAddr, ok := s.listener.Addr().(*net.TCPAddr)
	if !ok {
		return 0
	}

	return tcpAddr.Port
}

// AdminBaseURL returns the base URL for admin API access.
func (s *AdminServer) AdminBaseURL() string {
	port := s.ActualPort()

	return fmt.Sprintf("https://127.0.0.1:%d", port)
}
