// Copyright (c) 2025 Justin Cranford
//
//

// Package server provides the JOSE Authority Server HTTP service.
package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTLSGenerator "cryptoutil/internal/apps/template/service/config/tls_generator"
	cryptoutilJoseServerMiddleware "cryptoutil/internal/jose/server/middleware"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"

	"github.com/gofiber/fiber/v2"
)

// Server represents the JOSE Authority Server.
type Server struct {
	settings         *cryptoutilConfig.ServiceTemplateServerSettings
	telemetryService *cryptoutilTelemetry.TelemetryService
	jwkGenService    *cryptoutilJose.JWKGenService
	keyStore         *KeyStore
	fiberApp         *fiber.App
	listener         net.Listener
	actualPort       int // Actual port after dynamic allocation.
	apiKeyMiddleware *cryptoutilJoseServerMiddleware.APIKeyMiddleware
	tlsMaterial      *cryptoutilConfig.TLSMaterial
}

// New creates a new JOSE Authority Server instance using context.Background().
// Deprecated: Use NewServer with explicit TLSConfig instead.
func New(settings *cryptoutilConfig.ServiceTemplateServerSettings) (*Server, error) {
	// Create default TLS config for backward compatibility.
	tlsCfg, err := cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{"localhost", "jose-server"},
		[]string{"127.0.0.1", "::1"},
		cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate default TLS config: %w", err)
	}

	return NewServer(context.Background(), settings, tlsCfg)
}

// NewServer creates a new JOSE Authority Server instance.
func NewServer(ctx context.Context, settings *cryptoutilConfig.ServiceTemplateServerSettings, tlsCfg *cryptoutilTLSGenerator.TLSGeneratedSettings) (*Server, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if settings == nil {
		return nil, fmt.Errorf("settings cannot be nil")
	} else if tlsCfg == nil {
		return nil, fmt.Errorf("tlsCfg cannot be nil")
	}

	// Generate TLS material.
	tlsMaterial, err := cryptoutilTLSGenerator.GenerateTLSMaterial(tlsCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS material: %w", err)
	}

	// Initialize telemetry.
	telemetryService, err := cryptoutilTelemetry.NewTelemetryService(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	// Initialize JWK generation service.
	jwkGenService, err := cryptoutilJose.NewJWKGenService(ctx, telemetryService, settings.VerboseMode)
	if err != nil {
		telemetryService.Shutdown()

		return nil, fmt.Errorf("failed to initialize JWK generation service: %w", err)
	}

	// Create in-memory key store.
	keyStore := NewKeyStore()

	// Create Fiber app.
	fiberApp := fiber.New(fiber.Config{
		AppName:       "JOSE Authority Server",
		ServerHeader:  "JOSE-Authority",
		StrictRouting: true,
		CaseSensitive: true,
	})

	server := &Server{
		settings:         settings,
		telemetryService: telemetryService,
		jwkGenService:    jwkGenService,
		keyStore:         keyStore,
		fiberApp:         fiberApp,
		tlsMaterial:      tlsMaterial,
	}

	// Setup routes.
	server.setupRoutes()

	return server, nil
}

// Start begins listening for HTTPS requests.
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

	s.telemetryService.Slogger.Info("Starting JOSE Authority Server",
		"address", s.settings.BindPublicAddress,
		"port", s.actualPort,
	)

	// Wrap listener with TLS.
	tlsListener := tls.NewListener(listener, s.tlsMaterial.Config)

	s.telemetryService.Slogger.Info("JOSE Authority Server listening with TLS", "addr", listener.Addr().String())

	// Start server in goroutine.
	errChan := make(chan error, 1)

	go func() {
		if err := s.fiberApp.Listener(tlsListener); err != nil {
			errChan <- err
		}
	}()

	// Wait for context cancellation or error.
	select {
	case <-ctx.Done():
		s.telemetryService.Slogger.Info("Context cancelled, shutting down server")

		return nil
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	}
}

// StartNonBlocking starts the server without blocking.
func (s *Server) StartNonBlocking() error {
	ctx := context.Background()
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

	s.telemetryService.Slogger.Info("Starting JOSE Authority Server",
		"address", s.settings.BindPublicAddress,
		"port", s.actualPort,
	)

	// Wrap listener with TLS.
	tlsListener := tls.NewListener(listener, s.tlsMaterial.Config)

	s.telemetryService.Slogger.Info("JOSE Authority Server listening with TLS", "addr", listener.Addr().String())

	go func() {
		if err := s.fiberApp.Listener(tlsListener); err != nil {
			s.telemetryService.Slogger.Error("Server error", "error", err)
		}
	}()

	return nil
}

// ActualPort returns the actual port the server is listening on.
func (s *Server) ActualPort() int {
	return s.actualPort
}

// PublicBaseURL returns the base URL for public API access.
func (s *Server) PublicBaseURL() string {
	return fmt.Sprintf("%s://%s:%d", s.settings.BindPublicProtocol, s.settings.BindPublicAddress, s.actualPort)
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown() error {
	s.telemetryService.Slogger.Info("Shutting down JOSE Authority Server")

	var shutdownErr error

	if s.fiberApp != nil {
		if err := s.fiberApp.Shutdown(); err != nil {
			s.telemetryService.Slogger.Error("Error shutting down Fiber app", "error", err)
			shutdownErr = err
		}
	}

	if s.jwkGenService != nil {
		s.jwkGenService.Shutdown()
	}

	if s.telemetryService != nil {
		s.telemetryService.Shutdown()
	}

	return shutdownErr
}

// ConfigureAPIKeyAuth configures API key authentication middleware.
// This should be called before Start() to enable authentication.
func (s *Server) ConfigureAPIKeyAuth(config *cryptoutilJoseServerMiddleware.APIKeyConfig) {
	if config == nil {
		config = cryptoutilJoseServerMiddleware.DefaultAPIKeyConfig()
	}

	s.apiKeyMiddleware = cryptoutilJoseServerMiddleware.NewAPIKeyMiddleware(config)
}

// GetAPIKeyMiddleware returns the configured API key middleware handler.
// This can be used to apply authentication to specific routes.
func (s *Server) GetAPIKeyMiddleware() fiber.Handler {
	if s.apiKeyMiddleware == nil {
		return nil
	}

	return s.apiKeyMiddleware.Handler()
}

// setupRoutes configures all API routes.
func (s *Server) setupRoutes() {
	// Health endpoints (no auth required).
	s.fiberApp.Get("/health", s.handleHealth)
	s.fiberApp.Get("/livez", s.handleLivez)
	s.fiberApp.Get("/readyz", s.handleReadyz)

	// Well-known endpoints (no auth required for public key discovery).
	s.fiberApp.Get("/.well-known/jwks.json", s.handleJWKS)

	// API v1 group.
	v1 := s.fiberApp.Group("/jose/v1")

	// JWK endpoints.
	v1.Post("/jwk/generate", s.handleJWKGenerate)
	v1.Get("/jwk/:kid", s.handleJWKGet)
	v1.Delete("/jwk/:kid/delete", s.handleJWKDelete)
	v1.Get("/jwk", s.handleJWKList)
	v1.Get("/jwks", s.handleJWKS)

	// JWS endpoints.
	v1.Post("/jws/sign", s.handleJWSSign)
	v1.Post("/jws/verify", s.handleJWSVerify)

	// JWE endpoints.
	v1.Post("/jwe/encrypt", s.handleJWEEncrypt)
	v1.Post("/jwe/decrypt", s.handleJWEDecrypt)

	// JWT endpoints.
	v1.Post("/jwt/sign", s.handleJWTCreate)
	v1.Post("/jwt/verify", s.handleJWTVerify)
}
