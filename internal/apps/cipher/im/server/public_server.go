// Copyright (c) 2025 Justin Cranford
//

package server

import (
	"context"
	"fmt"

	cryptoutilCipherRepository "cryptoutil/internal/apps/cipher/im/repository"
	"cryptoutil/internal/apps/cipher/im/server/apis"
	"cryptoutil/internal/apps/cipher/im/server/businesslogic"
	"cryptoutil/internal/apps/cipher/im/server/middleware"
	cryptoutilTLSGenerator "cryptoutil/internal/apps/template/service/config/tls_generator"
	cryptoutilBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilTemplateRealms "cryptoutil/internal/apps/template/service/server/realms"
	cryptoutilTemplateRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilTemplateServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
)

// PublicServer implements the cipher-im public server by embedding PublicServerBase.
type PublicServer struct {
	base *cryptoutilTemplateServer.PublicServerBase // Reusable server infrastructure

	userRepo                *cryptoutilCipherRepository.UserRepository
	messageRepo             *cryptoutilCipherRepository.MessageRepository
	messageRecipientJWKRepo *cryptoutilCipherRepository.MessageRecipientJWKRepository // Per-recipient decryption keys
	jwkGenService           *cryptoutilJose.JWKGenService                             // JWK generation for message encryption
	sessionManagerService   *businesslogic.SessionManagerService                      // Session management service

	// Handlers (composition pattern).
	authnHandler   *cryptoutilTemplateRealms.UserServiceImpl
	messageHandler *apis.MessageHandler
}

// NewPublicServer creates a new cipher-im public server.
// Exported for testing from external test packages.
func NewPublicServer(
	ctx context.Context,
	bindAddress string,
	port int,
	userRepo *cryptoutilCipherRepository.UserRepository,
	messageRepo *cryptoutilCipherRepository.MessageRepository,
	messageRecipientJWKRepo *cryptoutilCipherRepository.MessageRecipientJWKRepository,
	jwkGenService *cryptoutilJose.JWKGenService,
	barrierService *cryptoutilBarrier.BarrierService,
	sessionManagerService *businesslogic.SessionManagerService,
	tlsCfg *cryptoutilTLSGenerator.TLSGeneratedSettings,
) (*PublicServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if bindAddress == "" {
		return nil, fmt.Errorf("bind address cannot be empty")
	} else if userRepo == nil {
		return nil, fmt.Errorf("user repository cannot be nil")
	} else if messageRepo == nil {
		return nil, fmt.Errorf("message repository cannot be nil")
	} else if messageRecipientJWKRepo == nil {
		return nil, fmt.Errorf("message recipient JWK repository cannot be nil")
	} else if jwkGenService == nil {
		return nil, fmt.Errorf("JWK generation service cannot be nil")
	} else if sessionManagerService == nil {
		return nil, fmt.Errorf("session manager service cannot be nil")
	} else if tlsCfg == nil {
		return nil, fmt.Errorf("TLS configuration cannot be nil")
	}

	// Generate TLS material using centralized infrastructure.
	tlsMaterial, err := cryptoutilTLSGenerator.GenerateTLSMaterial(tlsCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS material: %w", err)
	}

	// Create PublicServerBase with reusable infrastructure.
	base, err := cryptoutilTemplateServer.NewPublicServerBase(&cryptoutilTemplateServer.PublicServerConfig{
		BindAddress: bindAddress,
		Port:        port,
		TLSMaterial: tlsMaterial,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create public server base: %w", err)
	}

	s := &PublicServer{
		base:                    base,
		userRepo:                userRepo,
		messageRepo:             messageRepo,
		messageRecipientJWKRepo: messageRecipientJWKRepo,
		jwkGenService:           jwkGenService,
		sessionManagerService:   sessionManagerService,
	}

	// Create repository adapter for template realms.
	userRepoAdapter := cryptoutilCipherRepository.NewUserRepositoryAdapter(userRepo)

	// Create user factory for template realms.
	// Generates fresh User model per request with NO hardcoded tenant.
	userFactory := func() cryptoutilTemplateRealms.UserModel {
		return &cryptoutilTemplateRepository.User{}
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
	// Create session handler.
	sessionHandler := apis.NewSessionHandler(s.sessionManagerService)

	// Create session middleware for browser and service paths.
	browserSessionMiddleware := middleware.BrowserSessionMiddleware(s.sessionManagerService)
	serviceSessionMiddleware := middleware.ServiceSessionMiddleware(s.sessionManagerService)

	// Get underlying Fiber app from base for route registration.
	app := s.base.App()

	// Session management endpoints (no middleware - these endpoints create/validate sessions).
	app.Post("/service/api/v1/sessions/issue", sessionHandler.IssueSession)
	app.Post("/service/api/v1/sessions/validate", sessionHandler.ValidateSession)
	app.Post("/browser/api/v1/sessions/issue", sessionHandler.IssueSession)
	app.Post("/browser/api/v1/sessions/validate", sessionHandler.ValidateSession)

	// User management endpoints (authentication - no middleware, returns session token on login).
	app.Post("/service/api/v1/users/register", s.authnHandler.HandleRegisterUser())
	app.Post("/service/api/v1/users/login", s.authnHandler.HandleLoginUserWithSession(s.sessionManagerService, false))
	app.Post("/browser/api/v1/users/register", s.authnHandler.HandleRegisterUser())
	app.Post("/browser/api/v1/users/login", s.authnHandler.HandleLoginUserWithSession(s.sessionManagerService, true))

	// Business logic endpoints (message operations - session required).
	app.Put("/service/api/v1/messages/tx", serviceSessionMiddleware, s.messageHandler.HandleSendMessage())
	app.Get("/service/api/v1/messages/rx", serviceSessionMiddleware, s.messageHandler.HandleReceiveMessages())
	app.Delete("/service/api/v1/messages/:id", serviceSessionMiddleware, s.messageHandler.HandleDeleteMessage())

	app.Put("/browser/api/v1/messages/tx", browserSessionMiddleware, s.messageHandler.HandleSendMessage())
	app.Get("/browser/api/v1/messages/rx", browserSessionMiddleware, s.messageHandler.HandleReceiveMessages())
	app.Delete("/browser/api/v1/messages/:id", browserSessionMiddleware, s.messageHandler.HandleDeleteMessage())
}

// Start starts the HTTPS server by delegating to PublicServerBase.
func (s *PublicServer) Start(ctx context.Context) error {
	return s.base.Start(ctx)
}

// Shutdown gracefully shuts down the server by delegating to PublicServerBase.
func (s *PublicServer) Shutdown(ctx context.Context) error {
	return s.base.Shutdown(ctx)
}

// ActualPort returns the actual port the server is listening on by delegating to PublicServerBase.
func (s *PublicServer) ActualPort() int {
	return s.base.ActualPort()
}

// PublicBaseURL returns the base URL for public API access by delegating to PublicServerBase.
func (s *PublicServer) PublicBaseURL() string {
	return s.base.PublicBaseURL()
}
