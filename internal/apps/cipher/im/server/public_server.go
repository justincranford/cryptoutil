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

	cryptoutilCipherDomain "cryptoutil/internal/apps/cipher/im/domain"
	cryptoutilCipherRepository "cryptoutil/internal/apps/cipher/im/repository"
	"cryptoutil/internal/apps/cipher/im/server/apis"
	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilTLSGenerator "cryptoutil/internal/shared/config/tls_generator"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilBarrier "cryptoutil/internal/template/server/barrier"
	cryptoutilTemplateRealms "cryptoutil/internal/template/server/realms"
)

// PublicServer implements the template.PublicServer interface for cipher-im.
type PublicServer struct {
	port                    int
	userRepo                *cryptoutilCipherRepository.UserRepository
	messageRepo             *cryptoutilCipherRepository.MessageRepository
	messageRecipientJWKRepo *cryptoutilCipherRepository.MessageRecipientJWKRepository // Per-recipient decryption keys
	jwkGenService           *cryptoutilJose.JWKGenService                             // JWK generation for message encryption
	jwtSecret               string                                                    // JWT signing secret for authentication

	// Handlers (composition pattern).
	authnHandler   *cryptoutilTemplateRealms.UserServiceImpl
	messageHandler *apis.MessageHandler

	app         *fiber.App
	mu          sync.RWMutex
	shutdown    bool
	actualPort  int
	tlsMaterial *cryptoutilConfig.TLSMaterial
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewPublicServer creates a new cipher-im public server.
func NewPublicServer(
	ctx context.Context,
	port int,
	userRepo *cryptoutilCipherRepository.UserRepository,
	messageRepo *cryptoutilCipherRepository.MessageRepository,
	messageRecipientJWKRepo *cryptoutilCipherRepository.MessageRecipientJWKRepository,
	jwkGenService *cryptoutilJose.JWKGenService,
	barrierService *cryptoutilBarrier.BarrierService,
	jwtSecret string,
	tlsCfg *cryptoutilTLSGenerator.TLSGeneratedSettings,
) (*PublicServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if userRepo == nil {
		return nil, fmt.Errorf("user repository cannot be nil")
	} else if messageRepo == nil {
		return nil, fmt.Errorf("message repository cannot be nil")
	} else if messageRecipientJWKRepo == nil {
		return nil, fmt.Errorf("message recipient JWK repository cannot be nil")
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
		port:                    port,
		userRepo:                userRepo,
		messageRepo:             messageRepo,
		messageRecipientJWKRepo: messageRecipientJWKRepo,
		jwkGenService:           jwkGenService,
		jwtSecret:               jwtSecret,
		app:                     fiber.New(fiber.Config{DisableStartupMessage: true}),
		tlsMaterial:             tlsMaterial,
	}

	// Create repository adapter for template realms.
	userRepoAdapter := cryptoutilCipherRepository.NewUserRepositoryAdapter(userRepo)

	// Create user factory for template realms.
	userFactory := func() cryptoutilTemplateRealms.UserModel {
		return &cryptoutilCipherDomain.User{}
	}

	// Create realms handler using template service (authentication/authorization).
	s.authnHandler = cryptoutilTemplateRealms.NewUserService(userRepoAdapter, userFactory)

	// Create apis handler (business logic).
	s.messageHandler = apis.NewMessageHandler(messageRepo, messageRecipientJWKRepo, jwkGenService, barrierService)

	s.registerRoutes()

	return s, nil
}

// registerRoutes sets up the API endpoints.
func (s *PublicServer) registerRoutes() {
	// Health endpoints (required by template pattern).
	s.app.Get("/service/api/v1/health", s.handleServiceHealth)
	s.app.Get("/browser/api/v1/health", s.handleBrowserHealth)

	// User management endpoints (authentication - no JWT required).
	s.app.Post("/service/api/v1/users/register", s.authnHandler.HandleRegisterUser())
	s.app.Post("/service/api/v1/users/login", s.authnHandler.HandleLoginUser(s.jwtSecret))
	s.app.Post("/browser/api/v1/users/register", s.authnHandler.HandleRegisterUser())
	s.app.Post("/browser/api/v1/users/login", s.authnHandler.HandleLoginUser(s.jwtSecret))

	// Business logic endpoints (message operations - JWT required).
	s.app.Put("/service/api/v1/messages/tx", cryptoutilTemplateRealms.JWTMiddleware(s.jwtSecret), s.messageHandler.HandleSendMessage())
	s.app.Get("/service/api/v1/messages/rx", cryptoutilTemplateRealms.JWTMiddleware(s.jwtSecret), s.messageHandler.HandleReceiveMessages())
	s.app.Delete("/service/api/v1/messages/:id", cryptoutilTemplateRealms.JWTMiddleware(s.jwtSecret), s.messageHandler.HandleDeleteMessage())

	s.app.Put("/browser/api/v1/messages/tx", cryptoutilTemplateRealms.JWTMiddleware(s.jwtSecret), s.messageHandler.HandleSendMessage())
	s.app.Get("/browser/api/v1/messages/rx", cryptoutilTemplateRealms.JWTMiddleware(s.jwtSecret), s.messageHandler.HandleReceiveMessages())
	s.app.Delete("/browser/api/v1/messages/:id", cryptoutilTemplateRealms.JWTMiddleware(s.jwtSecret), s.messageHandler.HandleDeleteMessage())
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

	// Create cancellable context for server lifecycle management.
	s.mu.Lock()
	s.ctx, s.cancel = context.WithCancel(ctx)
	serverCtx := s.ctx
	s.mu.Unlock()

	// Create TCP listener.
	listenConfig := &net.ListenConfig{}

	listener, err := listenConfig.Listen(serverCtx, "tcp", fmt.Sprintf("%s:%d", cryptoutilMagic.IPv4Loopback, s.port))
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
	case <-serverCtx.Done():
		// Context cancelled - trigger graceful shutdown.
		const shutdownTimeout = 5

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout*time.Second)
		defer cancel()

		_ = s.Shutdown(shutdownCtx)

		return fmt.Errorf("public server stopped: %w", serverCtx.Err())
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

	// Cancel the server context to unblock Start() method.
	if s.cancel != nil {
		s.cancel()
	}

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

// PublicBaseURL returns the base URL for public API access.
func (s *PublicServer) PublicBaseURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return fmt.Sprintf("https://127.0.0.1:%d", s.actualPort)
}
