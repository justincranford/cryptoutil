// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	"cryptoutil/internal/learn/repository"
)

// PublicServer implements the template.PublicServer interface for learn-im.
type PublicServer struct {
	port        int
	userRepo    *repository.UserRepository
	messageRepo *repository.MessageRepository

	app        *fiber.App
	mu         sync.RWMutex
	shutdown   bool
	actualPort int
}

// NewPublicServer creates a new learn-im public server.
func NewPublicServer(ctx context.Context, port int, userRepo *repository.UserRepository, messageRepo *repository.MessageRepository) (*PublicServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	if userRepo == nil {
		return nil, fmt.Errorf("user repository cannot be nil")
	}

	if messageRepo == nil {
		return nil, fmt.Errorf("message repository cannot be nil")
	}

	s := &PublicServer{
		port:        port,
		userRepo:    userRepo,
		messageRepo: messageRepo,
		app:         fiber.New(fiber.Config{DisableStartupMessage: true}),
	}

	s.registerRoutes()

	return s, nil
}

// registerRoutes sets up the API endpoints.
func (s *PublicServer) registerRoutes() {
	// Health endpoints (required by template pattern).
	s.app.Get("/service/api/v1/health", s.handleServiceHealth)
	s.app.Get("/browser/api/v1/health", s.handleBrowserHealth)

	// Business logic endpoints.
	s.app.Put("/service/api/v1/messages/tx", s.handleSendMessage)
	s.app.Get("/service/api/v1/messages/rx", s.handleReceiveMessages)
	s.app.Delete("/service/api/v1/messages/:id", s.handleDeleteMessage)

	s.app.Put("/browser/api/v1/messages/tx", s.handleSendMessage)
	s.app.Get("/browser/api/v1/messages/rx", s.handleReceiveMessages)
	s.app.Delete("/browser/api/v1/messages/:id", s.handleDeleteMessage)
}

// handleServiceHealth returns health status for service-to-service clients.
func (s *PublicServer) handleServiceHealth(c *fiber.Ctx) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.shutdown {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "shutting down",
		})
	}

	//nolint:wrapcheck // Fiber framework error, wrapping not needed.
	return c.JSON(fiber.Map{
		"status": "healthy",
	})
}

// handleBrowserHealth returns health status for browser clients.
func (s *PublicServer) handleBrowserHealth(c *fiber.Ctx) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.shutdown {
		//nolint:wrapcheck // Fiber framework error, wrapping not needed.
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "shutting down",
		})
	}

	//nolint:wrapcheck // Fiber framework error, wrapping not needed.
	return c.JSON(fiber.Map{
		"status": "healthy",
	})
}

// handleSendMessage handles PUT /messages/tx.
func (s *PublicServer) handleSendMessage(c *fiber.Ctx) error {
	// TODO: Implement message sending.
	//nolint:wrapcheck // Fiber framework error, wrapping not needed.
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "not implemented",
	})
}

// handleReceiveMessages handles GET /messages/rx.
func (s *PublicServer) handleReceiveMessages(c *fiber.Ctx) error {
	// TODO: Implement message retrieval.
	//nolint:wrapcheck // Fiber framework error, wrapping not needed.
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "not implemented",
	})
}

// handleDeleteMessage handles DELETE /messages/:id.
func (s *PublicServer) handleDeleteMessage(c *fiber.Ctx) error {
	// TODO: Implement message deletion.
	//nolint:wrapcheck // Fiber framework error, wrapping not needed.
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "not implemented",
	})
}

// Start starts the HTTPS server (implements template.PublicServer).
func (s *PublicServer) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	// Generate self-signed TLS certificate.
	tlsConfig, err := s.generateTLSConfig()
	if err != nil {
		return fmt.Errorf("failed to generate TLS config: %w", err)
	}

	// Create TCP listener.
	const bindAddress = "127.0.0.1" // IPv4 loopback.

	listenConfig := &net.ListenConfig{}

	listener, err := listenConfig.Listen(ctx, "tcp", fmt.Sprintf("%s:%d", bindAddress, s.port))
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

	// Create TLS listener.
	tlsListener := tls.NewListener(listener, tlsConfig)

	// Start server in goroutine.
	errChan := make(chan error, 1)

	go func() {
		if err := s.app.Listener(tlsListener); err != nil {
			errChan <- fmt.Errorf("public server error: %w", err)
		} else {
			errChan <- nil
		}
	}()

	// Wait for either context cancellation or server error.
	select {
	case <-ctx.Done():
		// Context cancelled - trigger graceful shutdown.
		const shutdownTimeout = 5

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout*time.Second)
		defer cancel()

		_ = s.Shutdown(shutdownCtx)

		return fmt.Errorf("public server stopped: %w", ctx.Err())
	case err := <-errChan:
		return err
	}
}

// generateTLSConfig creates a TLS configuration using a self-signed certificate.
func (s *PublicServer) generateTLSConfig() (*tls.Config, error) {
	// Generate ECDSA P-384 key for TLS certificate.
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS private key: %w", err)
	}

	// Generate serial number.
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128)) //nolint:mnd // X.509 serial number bit length
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	now := time.Now().UTC()

	// Create self-signed certificate template.
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "learn-im-server",
			Organization: []string{"cryptoutil"},
			Country:      []string{"US"},
		},
		NotBefore:             now,
		NotAfter:              now.Add(365 * 24 * time.Hour), // 1 year.
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost", "learn-im-server"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	// Self-sign the certificate.
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, privateKey.Public(), privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS certificate: %w", err)
	}

	// Parse the created certificate.
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TLS certificate: %w", err)
	}

	// Create TLS certificate.
	tlsCert := tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  privateKey,
		Leaf:        cert,
	}

	// Create TLS config.
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		MinVersion:   tls.VersionTLS13,
		ClientAuth:   tls.NoClientCert,
	}

	return tlsConfig, nil
}

// Shutdown gracefully shuts down the server (implements template.PublicServer).
func (s *PublicServer) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.shutdown {
		return fmt.Errorf("public server already shutdown")
	}

	s.shutdown = true

	if s.app != nil {
		if err := s.app.Shutdown(); err != nil {
			return fmt.Errorf("failed to shutdown fiber app: %w", err)
		}
	}

	return nil
}

// ActualPort returns the actual port the server is listening on (implements template.PublicServer).
func (s *PublicServer) ActualPort() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.actualPort
}
