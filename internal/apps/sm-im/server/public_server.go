// Copyright (c) 2025 Justin Cranford
//

package server

import (
	"context"
	"fmt"

	googleUuid "github.com/google/uuid"

	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps/framework/service/server"
	cryptoutilAppsFrameworkServiceServerApis "cryptoutil/internal/apps/framework/service/server/apis"
	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps/framework/service/server/barrier"
	cryptoutilAppsFrameworkServiceServerBusinesslogic "cryptoutil/internal/apps/framework/service/server/businesslogic"
	cryptoutilAppsFrameworkServiceServerMiddleware "cryptoutil/internal/apps/framework/service/server/middleware"
	cryptoutilAppsFrameworkServiceServerRealms "cryptoutil/internal/apps/framework/service/server/realms"
	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps/framework/service/server/repository"
	cryptoutilAppsFrameworkServiceServerService "cryptoutil/internal/apps/framework/service/server/service"
	cryptoutilAppsSmImRepository "cryptoutil/internal/apps/sm-im/repository"
	cryptoutilAppsSmImServerApis "cryptoutil/internal/apps/sm-im/server/apis"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// PublicServer implements the sm-im public server by embedding PublicServerBase.
type PublicServer struct {
	base *cryptoutilAppsFrameworkServiceServer.PublicServerBase // Reusable server infrastructure

	userRepo                *cryptoutilAppsSmImRepository.UserRepository
	messageRepo             *cryptoutilAppsSmImRepository.MessageRepository
	messageRecipientJWKRepo *cryptoutilAppsSmImRepository.MessageRecipientJWKRepository                  // Per-recipient decryption keys
	jwkGenService           *cryptoutilSharedCryptoJose.JWKGenService                                    // JWK generation for message encryption
	sessionManagerService   *cryptoutilAppsFrameworkServiceServerBusinesslogic.SessionManagerService     // Session management service
	realmService            cryptoutilAppsFrameworkServiceServerService.RealmService                     // Realm management service
	registrationService     *cryptoutilAppsFrameworkServiceServerBusinesslogic.TenantRegistrationService // Tenant registration service

	// SM-IM state (auto-created tenant on first registration).
	autoTenantID *googleUuid.UUID

	// Handlers (composition pattern).
	authnHandler   *cryptoutilAppsFrameworkServiceServerRealms.UserServiceImpl
	messageHandler *cryptoutilAppsSmImServerApis.MessageHandler
}

// NewPublicServer creates a new sm-im public server using builder-provided infrastructure.
// Used by ServerBuilder during route registration.
func NewPublicServer(
	base *cryptoutilAppsFrameworkServiceServer.PublicServerBase,
	sessionManagerService *cryptoutilAppsFrameworkServiceServerBusinesslogic.SessionManagerService,
	realmService cryptoutilAppsFrameworkServiceServerService.RealmService,
	registrationService *cryptoutilAppsFrameworkServiceServerBusinesslogic.TenantRegistrationService,
	userRepo *cryptoutilAppsSmImRepository.UserRepository,
	messageRepo *cryptoutilAppsSmImRepository.MessageRepository,
	messageRecipientJWKRepo *cryptoutilAppsSmImRepository.MessageRecipientJWKRepository,
	jwkGenService *cryptoutilSharedCryptoJose.JWKGenService,
	barrierService *cryptoutilAppsFrameworkServiceServerBarrier.Service,
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
	userRepoAdapter := cryptoutilAppsSmImRepository.NewUserRepositoryAdapter(userRepo)

	// Create user factory for template realms.
	// Creates tenant dynamically on first user registration.
	// All subsequent users share the same auto-created tenant.
	userFactory := func() cryptoutilAppsFrameworkServiceServerRealms.UserModel {
		// Check if tenant already created.
		if s.autoTenantID != nil {
			return &cryptoutilAppsFrameworkServiceServerRepository.User{
				TenantID: *s.autoTenantID,
			}
		}

		// First user registration - create tenant.
		ctx := context.Background()
		dummyUserID := googleUuid.New() // Temporary user ID for tenant creation.

		tenant, err := s.registrationService.RegisterUserWithTenant(
			ctx,
			dummyUserID,
			"auto-user",         // username
			"auto@sm-im.local",  // email
			"",                  // passwordHash (not used for auto tenant)
			"SM-IM Auto Tenant", // tenantName
			true,                // createTenant = true
		)
		if err != nil {
			// Log error but continue with zero UUID (will fail later with better error).
			fmt.Printf("Warning: Failed to create auto tenant: %v\n", err)

			return &cryptoutilAppsFrameworkServiceServerRepository.User{
				TenantID: googleUuid.UUID{},
			}
		}

		// Store tenant ID for reuse.
		s.autoTenantID = &tenant.ID

		return &cryptoutilAppsFrameworkServiceServerRepository.User{
			TenantID: tenant.ID,
		}
	}

	// Create realms handler using template service (authentication/authorization).
	s.authnHandler = cryptoutilAppsFrameworkServiceServerRealms.NewUserService(userRepoAdapter, userFactory)

	// Create message handler (business logic).
	s.messageHandler = cryptoutilAppsSmImServerApis.NewMessageHandler(messageRepo, messageRecipientJWKRepo, jwkGenService, barrierService)

	return s, nil
}

// registerRoutes sets up the API endpoints.
// Called by ServerBuilder after NewPublicServer returns.
func (s *PublicServer) registerRoutes() error {
	// Create session handler.
	sessionHandler := cryptoutilAppsSmImServerApis.NewSessionHandler(s.sessionManagerService)

	// Create session middleware for browser and service paths using template middleware directly.
	browserSessionMiddleware := cryptoutilAppsFrameworkServiceServerMiddleware.BrowserSessionMiddleware(s.sessionManagerService)
	serviceSessionMiddleware := cryptoutilAppsFrameworkServiceServerMiddleware.ServiceSessionMiddleware(s.sessionManagerService)

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
	cryptoutilAppsFrameworkServiceServerApis.RegisterRegistrationRoutes(app, s.registrationService, cryptoutilSharedMagic.RateLimitDefaultRequestsPerMin)

	return nil
}

// PublicBaseURL returns the base URL for public API access by delegating to PublicServerBase.
func (s *PublicServer) PublicBaseURL() string {
	return s.base.PublicBaseURL()
}
