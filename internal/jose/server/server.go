// Copyright (c) 2025 Justin Cranford
//
//

// Package server provides the JOSE Authority Server HTTP service.
package server

import (
	"context"
	"fmt"
	"net"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilJose "cryptoutil/internal/jose"

	"github.com/gofiber/fiber/v2"
)

// Server represents the JOSE Authority Server.
type Server struct {
	settings         *cryptoutilConfig.Settings
	telemetryService *cryptoutilTelemetry.TelemetryService
	jwkGenService    *cryptoutilJose.JWKGenService
	keyStore         *KeyStore
	fiberApp         *fiber.App
	listener         net.Listener
	actualPort       int // Actual port after dynamic allocation.
}

// New creates a new JOSE Authority Server instance using context.Background().
func New(settings *cryptoutilConfig.Settings) (*Server, error) {
	return NewServer(context.Background(), settings)
}

// NewServer creates a new JOSE Authority Server instance.
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
	}

	// Setup routes.
	server.setupRoutes()

	return server, nil
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

	s.telemetryService.Slogger.Info("Starting JOSE Authority Server",
		"address", s.settings.BindPublicAddress,
		"port", s.actualPort,
	)

	// Start server in goroutine.
	errChan := make(chan error, 1)

	go func() {
		if err := s.fiberApp.Listener(listener); err != nil {
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

	go func() {
		if err := s.fiberApp.Listener(listener); err != nil {
			s.telemetryService.Slogger.Error("Server error", "error", err)
		}
	}()

	return nil
}

// ActualPort returns the actual port the server is listening on.
func (s *Server) ActualPort() int {
	return s.actualPort
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

// setupRoutes configures all API routes.
func (s *Server) setupRoutes() {
	// Health endpoints.
	s.fiberApp.Get("/health", s.handleHealth)
	s.fiberApp.Get("/livez", s.handleLivez)
	s.fiberApp.Get("/readyz", s.handleReadyz)

	// Well-known endpoints.
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
