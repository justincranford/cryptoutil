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
	"fmt"
	"math/big"
	"net"
	"time"

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
)

// PublicServer represents the public AuthZ HTTPS server.
type PublicServer struct {
	config     *cryptoutilIdentityConfig.Config
	app        *fiber.App
	listener   net.Listener
	actualPort int
}

// NewPublicServer creates a new public AuthZ HTTPS server.
func NewPublicServer(
	ctx context.Context,
	config *cryptoutilIdentityConfig.Config,
) (*PublicServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create Fiber app.
	app := fiber.New(fiber.Config{
		AppName:       "AuthZ Public Server",
		ServerHeader:  "AuthZ-Public",
		StrictRouting: true,
		CaseSensitive: true,
	})

	// TODO: Register middleware and routes here.
	// app.Use(middleware.CORS())
	// app.Use(middleware.CSRF())
	// setupPublicRoutes(app)

	// Basic health endpoint for now.
	app.Get("/browser/api/v1/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "authz",
		})
	})

	app.Get("/service/api/v1/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "authz",
		})
	})

	return &PublicServer{
		config: config,
		app:    app,
	}, nil
}

// Start begins listening for public HTTPS requests.
func (s *PublicServer) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.config.AuthZ.BindAddress, s.config.AuthZ.Port)

	// Create listener for dynamic port allocation.
	var lc net.ListenConfig

	listener, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to create listener on %s: %w", addr, err)
	}

	s.listener = listener

	// Extract actual port.
	if tcpAddr, ok := listener.Addr().(*net.TCPAddr); ok {
		s.actualPort = tcpAddr.Port
	} else {
		return fmt.Errorf("failed to extract actual port from listener")
	}

	fmt.Printf("Starting AuthZ Public Server on %s (actual port: %d)\n", addr, s.actualPort)

	// Generate TLS config.
	tlsConfig, err := s.generateTLSConfig()
	if err != nil {
		return fmt.Errorf("failed to generate TLS config: %w", err)
	}

	// Wrap listener with TLS.
	tlsListener := tls.NewListener(listener, tlsConfig)

	fmt.Printf("AuthZ Public Server listening with TLS on %s\n", listener.Addr().String())

	// Start server in goroutine.
	errChan := make(chan error, 1)

	go func() {
		if err := s.app.Listener(tlsListener); err != nil {
			errChan <- err
		}
	}()

	// Wait for context cancellation or error.
	select {
	case <-ctx.Done():
		fmt.Println("Context cancelled, shutting down AuthZ Public Server")

		return nil
	case err := <-errChan:
		return fmt.Errorf("public server error: %w", err)
	}
}

// Shutdown gracefully shuts down the public server.
func (s *PublicServer) Shutdown() error {
	if s.app == nil {
		return nil
	}

	if err := s.app.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown public server: %w", err)
	}

	return nil
}

// ActualPort returns the actual port the public server is listening on.
func (s *PublicServer) ActualPort() int {
	return s.actualPort
}

// PublicBaseURL returns the base URL for public API access.
func (s *PublicServer) PublicBaseURL() string {
	return fmt.Sprintf("https://127.0.0.1:%d", s.actualPort)
}

// generateTLSConfig creates a self-signed certificate for TLS.
// TODO: Replace with CA-signed certificates or Docker secrets for production.
func (s *PublicServer) generateTLSConfig() (*tls.Config, error) {
	// Generate ECDSA P-256 private key.
	priv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create self-signed certificate template.
	serialNumber, err := crand.Int(crand.Reader, new(big.Int).Lsh(big.NewInt(1), cryptoutilSharedMagic.TLSSelfSignedCertSerialNumberBits))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"cryptoutil"},
			CommonName:   "authz-public-server",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	// Create self-signed certificate.
	certDER, err := x509.CreateCertificate(crand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Create TLS config.
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{certDER},
				PrivateKey:  priv,
			},
		},
		MinVersion: tls.VersionTLS13, // TLS 1.3 minimum
	}

	return tlsConfig, nil
}
