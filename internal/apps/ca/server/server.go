// Copyright (c) 2025 Justin Cranford

// Package server implements the pki-ca HTTPS server using the service template.
package server

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"time"

	"gorm.io/gorm"

	"cryptoutil/internal/apps/ca/server/config"
	cryptoutilCAHandler "cryptoutil/internal/ca/api/handler"
	cryptoutilCACrypto "cryptoutil/internal/ca/crypto"
	cryptoutilCAServiceIssuer "cryptoutil/internal/ca/service/issuer"
	cryptoutilCAServiceRevocation "cryptoutil/internal/ca/service/revocation"
	cryptoutilCAStorage "cryptoutil/internal/ca/storage"
	cryptoutilTemplateServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilTemplateBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilTemplateBuilder "cryptoutil/internal/apps/template/service/server/builder"
	cryptoutilTemplateBusinessLogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilTemplateRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilTemplateService "cryptoutil/internal/apps/template/service/server/service"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
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

// CAServer represents the pki-ca service application.
type CAServer struct {
	app *cryptoutilTemplateServer.Application
	db  *gorm.DB

	// Template services.
	telemetryService      *cryptoutilTelemetry.TelemetryService
	jwkGenService         *cryptoutilJose.JWKGenService
	barrierService        *cryptoutilTemplateBarrier.BarrierService
	sessionManagerService *cryptoutilTemplateBusinessLogic.SessionManagerService
	realmService          cryptoutilTemplateService.RealmService

	// CA-specific services.
	issuer      *cryptoutilCAServiceIssuer.Issuer
	storage     cryptoutilCAStorage.Store
	crlService  *cryptoutilCAServiceRevocation.CRLService
	ocspService *cryptoutilCAServiceRevocation.OCSPService
	handler     *cryptoutilCAHandler.Handler

	// Template repositories.
	realmRepo cryptoutilTemplateRepository.TenantRealmRepository

	// Shutdown functions.
	shutdownCore      func()
	shutdownContainer func()
}

// NewFromConfig creates a new pki-ca server from CAServerSettings.
// Uses service-template builder for infrastructure initialization.
func NewFromConfig(ctx context.Context, cfg *config.CAServerSettings) (*CAServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Initialize CA-specific services before builder (needed for route registration).
	cryptoProvider := cryptoutilCACrypto.NewSoftwareProvider()

	// Create in-memory storage for certificates.
	storage := cryptoutilCAStorage.NewMemoryStore()

	// Create self-signed CA certificate for development.
	caCert, caKey, err := createSelfSignedCA(cryptoProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA: %w", err)
	}

	// Create issuer config.
	issuerConfig := &cryptoutilCAServiceIssuer.IssuingCAConfig{
		Name:        "pki-ca",
		Certificate: caCert,
		PrivateKey:  caKey,
	}

	// Create issuer service.
	issuer, err := cryptoutilCAServiceIssuer.NewIssuer(cryptoProvider, issuerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create issuer: %w", err)
	}

	// Safely cast private key to signer.
	caKeySigner, ok := caKey.(crypto.Signer)
	if !ok {
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
		return nil, fmt.Errorf("failed to create handler: %w", err)
	}

	// Wire up revocation services.
	handler.SetCRLService(crlService)
	handler.SetOCSPService(ocspService)

	// Create server builder with template config.
	// Note: CA uses in-memory certificate storage, but still uses template database for sessions/barrier.
	builder := cryptoutilTemplateBuilder.NewServerBuilder(ctx, cfg.ServiceTemplateServerSettings)

	// Register pki-ca specific public routes.
	builder.WithPublicRouteRegistration(func(
		base *cryptoutilTemplateServer.PublicServerBase,
		_ *cryptoutilTemplateBuilder.ServiceResources,
	) error {
		// Create public server with CA handlers.
		publicServer := NewPublicServer(base, handler, crlService, ocspService, cfg)

		// Register all routes (standard + CA-specific).
		if err := publicServer.registerRoutes(); err != nil {
			return fmt.Errorf("failed to register public routes: %w", err)
		}

		return nil
	})

	// Build complete service infrastructure.
	resources, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build pki-ca service: %w", err)
	}

	// Create pki-ca server wrapper.
	server := &CAServer{
		app:                   resources.Application,
		db:                    resources.DB,
		telemetryService:      resources.TelemetryService,
		jwkGenService:         resources.JWKGenService,
		barrierService:        resources.BarrierService,
		sessionManagerService: resources.SessionManager,
		realmService:          resources.RealmService,
		issuer:                issuer,
		storage:               storage,
		crlService:            crlService,
		ocspService:           ocspService,
		handler:               handler,
		realmRepo:             resources.RealmRepository,
		shutdownCore:          resources.ShutdownCore,
		shutdownContainer:     resources.ShutdownContainer,
	}

	return server, nil
}

// Start begins serving both public and admin HTTPS endpoints.
// Blocks until context is cancelled or an unrecoverable error occurs.
func (s *CAServer) Start(ctx context.Context) error {
	if err := s.app.Start(ctx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down all servers and closes resources.
func (s *CAServer) Shutdown(ctx context.Context) error {
	if err := s.app.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown application: %w", err)
	}

	return nil
}

// DB returns the GORM database connection (for tests).
func (s *CAServer) DB() *gorm.DB {
	return s.db
}

// App returns the application wrapper (for tests).
func (s *CAServer) App() *cryptoutilTemplateServer.Application {
	return s.app
}

// JWKGen returns the JWK generation service (for tests).
func (s *CAServer) JWKGen() *cryptoutilJose.JWKGenService {
	return s.jwkGenService
}

// Telemetry returns the telemetry service (for tests).
func (s *CAServer) Telemetry() *cryptoutilTelemetry.TelemetryService {
	return s.telemetryService
}

// Barrier returns the barrier service (for tests).
func (s *CAServer) Barrier() *cryptoutilTemplateBarrier.BarrierService {
	return s.barrierService
}

// Issuer returns the CA issuer service (for tests).
func (s *CAServer) Issuer() *cryptoutilCAServiceIssuer.Issuer {
	return s.issuer
}

// Storage returns the certificate storage (for tests).
func (s *CAServer) Storage() cryptoutilCAStorage.Store {
	return s.storage
}

// CRLService returns the CRL service (for tests).
func (s *CAServer) CRLService() *cryptoutilCAServiceRevocation.CRLService {
	return s.crlService
}

// OCSPService returns the OCSP service (for tests).
func (s *CAServer) OCSPService() *cryptoutilCAServiceRevocation.OCSPService {
	return s.ocspService
}

// PublicPort returns the actual port the public server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *CAServer) PublicPort() int {
	return s.app.PublicPort()
}

// AdminPort returns the actual port the admin server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *CAServer) AdminPort() int {
	return s.app.AdminPort()
}

// SetReady marks the server as ready (enables /admin/api/v1/readyz to return 200 OK).
func (s *CAServer) SetReady(ready bool) {
	s.app.SetReady(ready)
}

// PublicBaseURL returns the public server base URL (for tests).
func (s *CAServer) PublicBaseURL() string {
	return s.app.PublicBaseURL()
}

// AdminBaseURL returns the admin server base URL (for tests).
func (s *CAServer) AdminBaseURL() string {
	return s.app.AdminBaseURL()
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
			CommonName:   "PKI-CA Development CA",
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
