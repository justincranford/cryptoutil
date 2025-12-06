// Copyright (c) 2025 Justin Cranford

// Package server provides the CA Server HTTP service.
package server

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/gofiber/fiber/v2"

	cryptoutilCAServer "cryptoutil/api/ca/server"
	cryptoutilCAHandler "cryptoutil/internal/ca/api/handler"
	cryptoutilCACrypto "cryptoutil/internal/ca/crypto"
	cryptoutilCAServiceIssuer "cryptoutil/internal/ca/service/issuer"
	cryptoutilCAServiceRevocation "cryptoutil/internal/ca/service/revocation"
	cryptoutilCAStorage "cryptoutil/internal/ca/storage"
	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
)

// Default validity durations for development CA.
const (
	defaultCAValidity          = 10 * 365 * 24 * time.Hour // 10 years.
	defaultCRLValidity         = 7 * 24 * time.Hour        // 7 days.
	defaultCRLNextUpdateBuffer = 1 * 24 * time.Hour        // 1 day.
	defaultOCSPValidity        = 1 * 24 * time.Hour        // 1 day.
	serialNumberBitSize        = 128
	defaultRootCAPathLength    = 2 // Allow up to 2 levels of subordinate CAs.
)

// Server represents the CA Server.
type Server struct {
	settings         *cryptoutilConfig.Settings
	telemetryService *cryptoutilTelemetry.TelemetryService
	issuer           *cryptoutilCAServiceIssuer.Issuer
	storage          cryptoutilCAStorage.Store
	crlService       *cryptoutilCAServiceRevocation.CRLService
	ocspService      *cryptoutilCAServiceRevocation.OCSPService
	handler          *cryptoutilCAHandler.Handler
	fiberApp         *fiber.App
	listener         net.Listener
	actualPort       int
}

// New creates a new CA Server instance using context.Background().
func New(settings *cryptoutilConfig.Settings) (*Server, error) {
	return NewServer(context.Background(), settings)
}

// NewServer creates a new CA Server instance.
func NewServer(ctx context.Context, settings *cryptoutilConfig.Settings) (*Server, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if settings == nil {
		return nil, fmt.Errorf("settings cannot be nil")
	}

	// Initialize telemetry.
	telemetryService, err := cryptoutilTelemetry.NewTelemetryService(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	// Initialize crypto provider.
	cryptoProvider := cryptoutilCACrypto.NewSoftwareProvider()

	// Create in-memory storage.
	storage := cryptoutilCAStorage.NewMemoryStore()

	// Create self-signed CA certificate for development.
	caCert, caKey, err := createSelfSignedCA(cryptoProvider)
	if err != nil {
		telemetryService.Shutdown()

		return nil, fmt.Errorf("failed to create CA: %w", err)
	}

	// Create issuer config.
	issuerConfig := &cryptoutilCAServiceIssuer.IssuingCAConfig{
		Name:        "ca-server",
		Certificate: caCert,
		PrivateKey:  caKey,
	}

	// Create issuer service.
	issuer, err := cryptoutilCAServiceIssuer.NewIssuer(cryptoProvider, issuerConfig)
	if err != nil {
		telemetryService.Shutdown()

		return nil, fmt.Errorf("failed to create issuer: %w", err)
	}

	// Safely cast private key to signer.
	caKeySigner, ok := caKey.(crypto.Signer)
	if !ok {
		telemetryService.Shutdown()

		return nil, fmt.Errorf("CA private key is not a signer")
	}

	// Create CRL service.
	crlService, err := cryptoutilCAServiceRevocation.NewCRLService(&cryptoutilCAServiceRevocation.CRLConfig{
		Issuer:           caCert,
		PrivateKey:       caKeySigner,
		Provider:         cryptoProvider,
		Validity:         defaultCRLValidity,
		NextUpdateBuffer: defaultCRLNextUpdateBuffer,
	})
	if err != nil {
		telemetryService.Shutdown()

		return nil, fmt.Errorf("failed to create CRL service: %w", err)
	}

	// Create OCSP service.
	ocspService, err := cryptoutilCAServiceRevocation.NewOCSPService(&cryptoutilCAServiceRevocation.OCSPConfig{
		Issuer:       caCert,
		Responder:    caCert,
		ResponderKey: caKeySigner,
		Provider:     cryptoProvider,
		Validity:     defaultOCSPValidity,
	}, crlService)
	if err != nil {
		telemetryService.Shutdown()

		return nil, fmt.Errorf("failed to create OCSP service: %w", err)
	}

	// Create default profile configuration.
	profiles := map[string]*cryptoutilCAHandler.ProfileConfig{
		"tls-server": {
			ID:          "tls-server",
			Name:        "TLS Server Certificate",
			Description: "Standard TLS server certificate for HTTPS",
			Category:    "tls",
		},
		"tls-client": {
			ID:          "tls-client",
			Name:        "TLS Client Certificate",
			Description: "TLS client authentication certificate",
			Category:    "tls",
		},
	}

	// Create handler.
	handler, err := cryptoutilCAHandler.NewHandler(issuer, storage, profiles)
	if err != nil {
		telemetryService.Shutdown()

		return nil, fmt.Errorf("failed to create handler: %w", err)
	}

	// Wire up revocation services.
	handler.SetCRLService(crlService)
	handler.SetOCSPService(ocspService)

	// Create Fiber app.
	fiberApp := fiber.New(fiber.Config{
		AppName:       "CA Server",
		ServerHeader:  "CA-Server",
		StrictRouting: true,
		CaseSensitive: true,
	})

	server := &Server{
		settings:         settings,
		telemetryService: telemetryService,
		issuer:           issuer,
		storage:          storage,
		crlService:       crlService,
		ocspService:      ocspService,
		handler:          handler,
		fiberApp:         fiberApp,
	}

	// Setup routes.
	server.setupRoutes()

	return server, nil
}

// setupRoutes configures the API routes.
func (s *Server) setupRoutes() {
	// Health endpoints.
	s.fiberApp.Get("/health", s.handleHealth)
	s.fiberApp.Get("/livez", s.handleLivez)
	s.fiberApp.Get("/readyz", s.handleReadyz)

	// Register CA API handlers with base URL.
	cryptoutilCAServer.RegisterHandlersWithOptions(s.fiberApp, s.handler, cryptoutilCAServer.FiberServerOptions{
		BaseURL: "/api/v1/ca",
	})
}

// handleHealth returns server health status.
func (s *Server) handleHealth(c *fiber.Ctx) error {
	if err := c.JSON(fiber.Map{
		"status": "healthy",
		"time":   time.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		return fmt.Errorf("failed to send health response: %w", err)
	}

	return nil
}

// handleLivez returns liveness status.
func (s *Server) handleLivez(c *fiber.Ctx) error {
	if err := c.SendString("OK"); err != nil {
		return fmt.Errorf("failed to send liveness response: %w", err)
	}

	return nil
}

// handleReadyz returns readiness status.
func (s *Server) handleReadyz(c *fiber.Ctx) error {
	if err := c.SendString("OK"); err != nil {
		return fmt.Errorf("failed to send readiness response: %w", err)
	}

	return nil
}

// Start begins listening for HTTP requests.
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.settings.BindPublicAddress, s.settings.BindPublicPort)

	// Create listener for dynamic port allocation.
	var lc net.ListenConfig

	listener, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	s.listener = listener

	// Extract actual port.
	if tcpAddr, ok := listener.Addr().(*net.TCPAddr); ok {
		s.actualPort = tcpAddr.Port
	}

	s.telemetryService.Slogger.Info("Starting CA Server",
		"address", s.settings.BindPublicAddress,
		"port", s.actualPort)

	// Start server.
	if err := s.fiberApp.Listener(listener); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown() error {
	if s.fiberApp != nil {
		if err := s.fiberApp.Shutdown(); err != nil {
			return fmt.Errorf("failed to shutdown fiber app: %w", err)
		}
	}

	if s.telemetryService != nil {
		s.telemetryService.Shutdown()
	}

	return nil
}

// ActualPort returns the actual port the server is listening on.
func (s *Server) ActualPort() int {
	return s.actualPort
}

// createSelfSignedCA generates a self-signed CA certificate for development.
func createSelfSignedCA(provider cryptoutilCACrypto.Provider) (*x509.Certificate, crypto.PrivateKey, error) {
	// Generate ECDSA P-384 key for the CA.
	keySpec := cryptoutilCACrypto.KeySpec{
		Type:       cryptoutilCACrypto.KeyTypeECDSA,
		ECDSACurve: "P-384",
	}

	keyPair, err := provider.GenerateKeyPair(keySpec)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate CA key: %w", err)
	}

	// Generate serial number.
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), serialNumberBitSize))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	now := time.Now().UTC()

	// Create CA certificate template.
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "CA Server Development CA",
			Organization: []string{"cryptoutil"},
			Country:      []string{"US"},
		},
		NotBefore:             now,
		NotAfter:              now.Add(defaultCAValidity),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            defaultRootCAPathLength,
		MaxPathLenZero:        false,
	}

	// Get signature algorithm.
	signer, ok := keyPair.PrivateKey.(crypto.Signer)
	if !ok {
		return nil, nil, fmt.Errorf("private key is not a signer")
	}

	sigAlg, err := provider.GetSignatureAlgorithm(signer.Public())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get signature algorithm: %w", err)
	}

	template.SignatureAlgorithm = sigAlg

	// Self-sign the certificate.
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, signer.Public(), signer)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create CA certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	return cert, keyPair.PrivateKey, nil
}
