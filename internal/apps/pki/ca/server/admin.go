// Copyright (c) 2025 Justin Cranford
//
//

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

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// AdminServer represents the private admin API server for CA service.
type AdminServer struct {
	settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
	app      *fiber.App
	listener net.Listener
	mu       sync.RWMutex
	ready    bool
	shutdown bool
}

// NewAdminHTTPServer creates a new admin server instance for private administrative operations.
func NewAdminHTTPServer(
	ctx context.Context,
	settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings,
) (*AdminServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if settings == nil {
		return nil, fmt.Errorf("settings cannot be nil")
	}

	server := &AdminServer{
		settings: settings,
		ready:    false,
		shutdown: false,
	}

	// Create Fiber app with minimal configuration.
	const defaultTimeout = 30

	server.app = fiber.New(fiber.Config{
		DisableStartupMessage: true,
		AppName:               "CA Admin API",
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

// Start begins listening on 127.0.0.1:9090 for admin API requests.
func (s *AdminServer) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	// Bind to localhost only (127.0.0.1 explicit, not localhost due to IPv6 issues).
	const defaultAdminPort = 9090

	addr := fmt.Sprintf("%s:%d", cryptoutilSharedMagic.IPv4Loopback, defaultAdminPort)

	// Create listener.
	var lc net.ListenConfig

	listener, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to create admin listener: %w", err)
	}

	s.listener = listener

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

	// Start Fiber server (blocks until shutdown).
	if err := s.app.Listener(tlsListener); err != nil {
		return fmt.Errorf("admin server error: %w", err)
	}

	return nil
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
			return fmt.Errorf("failed to shutdown admin app: %w", err)
		}
	}

	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return fmt.Errorf("failed to close admin listener: %w", err)
		}
	}

	return nil
}

// ActualPort returns the actual port the admin server is listening on.
func (s *AdminServer) ActualPort() int {
	if s.listener == nil {
		return 0
	}

	addr, ok := s.listener.Addr().(*net.TCPAddr)
	if !ok {
		return 0
	}

	return addr.Port
}

// AdminBaseURL returns the base URL for admin API access.
func (s *AdminServer) AdminBaseURL() string {
	port := s.ActualPort()

	return fmt.Sprintf("%s://%s:%d", s.settings.BindPrivateProtocol, s.settings.BindPrivateAddress, port)
}

// generateTLSConfig creates a self-signed certificate for admin server.
func (s *AdminServer) generateTLSConfig() (*tls.Config, error) {
	// Generate private key.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
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

	serialNumber, err := crand.Int(crand.Reader, new(big.Int).Lsh(big.NewInt(1), serialNumberBits))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "cryptoutil-ca-admin",
			Organization: []string{"cryptoutil"},
		},
		NotBefore:             time.Now().UTC(),
		NotAfter:              time.Now().UTC().Add(validityDuration),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP(cryptoutilSharedMagic.IPv4Loopback)},
	}

	// Self-sign certificate.
	certDER, err := x509.CreateCertificate(crand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
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
