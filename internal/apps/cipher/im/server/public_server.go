// Copyright (c) 2025 Justin Cranford
//

package server

import (
	"context"
	"fmt"

	googleUuid "github.com/google/uuid"

	cryptoutilAppsCipherImRepository "cryptoutil/internal/apps/cipher/im/repository"
	"cryptoutil/internal/apps/cipher/im/server/apis"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilTemplateAPIs "cryptoutil/internal/apps/template/service/server/apis"
	cryptoutilBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilTemplateBusinessLogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilTemplateMiddleware "cryptoutil/internal/apps/template/service/server/middleware"
	cryptoutilTemplateRealms "cryptoutil/internal/apps/template/service/server/realms"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilTemplateService "cryptoutil/internal/apps/template/service/server/service"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// PublicServer implements the cipher-im public server by embedding PublicServerBase.
type PublicServer struct {
	base *cryptoutilAppsTemplateServiceServer.PublicServerBase // Reusable server infrastructure

	userRepo                *cryptoutilAppsCipherImRepository.UserRepository
	messageRepo             *cryptoutilAppsCipherImRepository.MessageRepository
	messageRecipientJWKRepo *cryptoutilAppsCipherImRepository.MessageRecipientJWKRepository // Per-recipient decryption keys
	jwkGenService           *cryptoutilJose.JWKGenService                                   // JWK generation for message encryption
	sessionManagerService   *cryptoutilTemplateBusinessLogic.SessionManagerService          // Session management service
	realmService            cryptoutilTemplateService.RealmService                          // Realm management service
	registrationService     *cryptoutilTemplateBusinessLogic.TenantRegistrationService      // Tenant registration service

	// Cipher-IM demo state (auto-created tenant on first registration).
	demoTenantID *googleUuid.UUID

	// Handlers (composition pattern).
	authnHandler   *cryptoutilTemplateRealms.UserServiceImpl
	messageHandler *apis.MessageHandler
}

// NewPublicServer creates a new cipher-im public server using builder-provided infrastructure.
// Used by ServerBuilder during route registration.
func NewPublicServer(
	base *cryptoutilAppsTemplateServiceServer.PublicServerBase,
	sessionManagerService *cryptoutilTemplateBusinessLogic.SessionManagerService,
	realmService cryptoutilTemplateService.RealmService,
	registrationService *cryptoutilTemplateBusinessLogic.TenantRegistrationService,
	userRepo *cryptoutilAppsCipherImRepository.UserRepository,
	messageRepo *cryptoutilAppsCipherImRepository.MessageRepository,
	messageRecipientJWKRepo *cryptoutilAppsCipherImRepository.MessageRecipientJWKRepository,
	jwkGenService *cryptoutilJose.JWKGenService,
	barrierService *cryptoutilBarrier.Service,
) (*PublicServer, error) {
	if base == nil {
		return nil, fmt.Errorf("public server base cannot be nil")
	} else if sessionManagerService == nil {
		return nil, fmt.Errorf("session manager service cannot be nil")
	} else if realmService == nil {
		return nil, fmt.Errorf("realm service cannot be nil")
	} else if registrationService == nil {
		return nil, fmt.Errorf("registration service cannot be nil")
	} else if userRepo == nil {
		return nil, fmt.Errorf("user repository cannot be nil")
	} else if messageRepo == nil {
		return nil, fmt.Errorf("message repository cannot be nil")
	} else if messageRecipientJWKRepo == nil {
		return nil, fmt.Errorf("message recipient JWK repository cannot be nil")
	} else if jwkGenService == nil {
		return nil, fmt.Errorf("JWK generation service cannot be nil")
	} else if barrierService == nil {
		return nil, fmt.Errorf("barrier service cannot be nil")
	}

	s := &PublicServer{
		base:                    base,
		userRepo:                userRepo,
		messageRepo:             messageRepo,
		messageRecipientJWKRepo: messageRecipientJWKRepo,
		jwkGenService:           jwkGenService,
		sessionManagerService:   sessionManagerService,
		realmService:            realmService,
		registrationService:     registrationService,
	}

	// Create repository adapter for template realms.
	userRepoAdapter := cryptoutilAppsCipherImRepository.NewUserRepositoryAdapter(userRepo)

	// Create user factory for template realms.
	// For cipher-im demo: Creates tenant dynamically on first user registration.
	// All subsequent users share the same demo tenant.
	userFactory := func() cryptoutilTemplateRealms.UserModel {
		// Check if demo tenant already created.
		if s.demoTenantID != nil {
			return &cryptoutilAppsTemplateServiceServerRepository.User{
				TenantID: *s.demoTenantID,
			}
		}

		// First user registration - create demo tenant.
		ctx := context.Background()
		dummyUserID := googleUuid.New() // Temporary user ID for tenant creation.

		tenant, err := s.registrationService.RegisterUserWithTenant(
			ctx,
			dummyUserID,
			"Cipher-IM Demo Tenant",
			true, // createTenant = true
		)
		if err != nil {
			// Log error but continue with zero UUID (will fail later with better error).
			fmt.Printf("Warning: Failed to create demo tenant: %v\n", err)

			return &cryptoutilAppsTemplateServiceServerRepository.User{
				TenantID: googleUuid.UUID{},
			}
		}

		// Store tenant ID for reuse.
		s.demoTenantID = &tenant.ID

		return &cryptoutilAppsTemplateServiceServerRepository.User{
			TenantID: tenant.ID,
		}
	}

	// Create realms handler using template service (authentication/authorization).
	s.authnHandler = cryptoutilTemplateRealms.NewUserService(userRepoAdapter, userFactory)

	// Create message handler (business logic).
	s.messageHandler = apis.NewMessageHandler(messageRepo, messageRecipientJWKRepo, jwkGenService, barrierService)

	return s, nil
}

// registerRoutes sets up the API endpoints.
// Called by ServerBuilder after NewPublicServer returns.
func (s *PublicServer) registerRoutes() error {
	// Create session handler.
	sessionHandler := apis.NewSessionHandler(s.sessionManagerService)

	// Create session middleware for browser and service paths using template middleware directly.
	browserSessionMiddleware := cryptoutilTemplateMiddleware.BrowserSessionMiddleware(s.sessionManagerService)
	serviceSessionMiddleware := cryptoutilTemplateMiddleware.ServiceSessionMiddleware(s.sessionManagerService)

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

	// Register tenant registration routes (from template) with default rate limit.
	cryptoutilTemplateAPIs.RegisterRegistrationRoutes(app, s.registrationService, cryptoutilMagic.RateLimitDefaultRequestsPerMin)

	return nil
}

// PublicBaseURL returns the base URL for public API access by delegating to PublicServerBase.
func (s *PublicServer) PublicBaseURL() string {
	return s.base.PublicBaseURL()
}
