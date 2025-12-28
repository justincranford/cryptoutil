// Copyright (c) 2025 Justin Cranford
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

	"cryptoutil/internal/learn/repository"
	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilTLSGenerator "cryptoutil/internal/shared/config/tls_generator"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// PublicServer implements the template.PublicServer interface for learn-im.
type PublicServer struct {
	port          int
	userRepo      *repository.UserRepository
	messageRepo   *repository.MessageRepository
	jwkGenService *cryptoutilJose.JWKGenService // JWK generation for message encryption
	jwtSecret     string                        // JWT signing secret for authentication

	// In-memory key cache for Phase 5a (no barrier service yet).
	// NOTE: Phase 5b will replace with barrier-encrypted database storage.
	messageKeysCache sync.Map // map[string]joseJwk.Key (keyID -> decryption JWK)

	app         *fiber.App
	mu          sync.RWMutex
	shutdown    bool
	actualPort  int
	tlsMaterial *cryptoutilConfig.TLSMaterial
}

// NewPublicServer creates a new learn-im public server.
func NewPublicServer(
	ctx context.Context,
	port int,
	userRepo *repository.UserRepository,
	messageRepo *repository.MessageRepository,
	jwkGenService *cryptoutilJose.JWKGenService,
	jwtSecret string,
	tlsCfg *cryptoutilTLSGenerator.TLSGeneratedSettings,
) (*PublicServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if userRepo == nil {
		return nil, fmt.Errorf("user repository cannot be nil")
	} else if messageRepo == nil {
		return nil, fmt.Errorf("message repository cannot be nil")
	} else if jwkGenService == nil {
		return nil, fmt.Errorf("JWK generation service cannot be nil")
	} else if tlsCfg == nil {
		return nil, fmt.Errorf("TLS configuration cannot be nil")
	}

	// Generate TLS material using centralized infrastructure.
	tlsMaterial, err := cryptoutilTLSGenerator.GenerateTLSMaterial(tlsCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS material: %w", err)
	}

	s := &PublicServer{
		port:          port,
		userRepo:      userRepo,
		messageRepo:   messageRepo,
		jwkGenService: jwkGenService,
		jwtSecret:     jwtSecret,
		app:           fiber.New(fiber.Config{DisableStartupMessage: true}),
		tlsMaterial:   tlsMaterial,
	}

	s.registerRoutes()

	return s, nil
}

// registerRoutes sets up the API endpoints.
func (s *PublicServer) registerRoutes() {
	// Health endpoints (required by template pattern).
	s.app.Get("/service/api/v1/health", s.handleServiceHealth)
	s.app.Get("/browser/api/v1/health", s.handleBrowserHealth)

	// User management endpoints (authentication - no JWT required).
	s.app.Post("/service/api/v1/users/register", s.handleRegisterUser)
	s.app.Post("/service/api/v1/users/login", s.handleLoginUser)
	s.app.Post("/browser/api/v1/users/register", s.handleRegisterUser)
	s.app.Post("/browser/api/v1/users/login", s.handleLoginUser)

	// Business logic endpoints (message operations - JWT required).
	s.app.Put("/service/api/v1/messages/tx", JWTMiddleware(s.jwtSecret), s.handleSendMessage)
	s.app.Get("/service/api/v1/messages/rx", JWTMiddleware(s.jwtSecret), s.handleReceiveMessages)
	s.app.Delete("/service/api/v1/messages/:id", JWTMiddleware(s.jwtSecret), s.handleDeleteMessage)

	s.app.Put("/browser/api/v1/messages/tx", JWTMiddleware(s.jwtSecret), s.handleSendMessage)
	s.app.Get("/browser/api/v1/messages/rx", JWTMiddleware(s.jwtSecret), s.handleReceiveMessages)
	s.app.Delete("/browser/api/v1/messages/:id", JWTMiddleware(s.jwtSecret), s.handleDeleteMessage)
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

// Start starts the HTTPS server (implements template.PublicServer).
func (s *PublicServer) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}

	// Create TCP listener.
	listenConfig := &net.ListenConfig{}

	listener, err := listenConfig.Listen(ctx, "tcp", fmt.Sprintf("%s:%d", cryptoutilMagic.IPv4Loopback, s.port))
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

	// Create TLS listener using centralized TLS material.
	tlsListener := tls.NewListener(listener, s.tlsMaterial.Config)

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
