// Copyright (c) 2025 Justin Cranford
//
//

// Package server implements the sm-im HTTPS server using the service template.
package server

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	cryptoutilAppsSmImRepository "cryptoutil/internal/apps/sm/im/repository"
	cryptoutilAppsSmImServerConfig "cryptoutil/internal/apps/sm/im/server/config"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilAppsTemplateServiceServerBuilder "cryptoutil/internal/apps/template/service/server/builder"
	cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilAppsTemplateServiceServerService "cryptoutil/internal/apps/template/service/server/service"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// SmIMServer represents the sm-im service application.
type SmIMServer struct {
	app *cryptoutilAppsTemplateServiceServer.Application
	db  *gorm.DB

	// Services.
	telemetryService      *cryptoutilSharedTelemetry.TelemetryService
	jwkGenService         *cryptoutilSharedCryptoJose.JWKGenService
	barrierService        *cryptoutilAppsTemplateServiceServerBarrier.Service
	sessionManagerService *cryptoutilAppsTemplateServiceServerBusinesslogic.SessionManagerService
	realmService          cryptoutilAppsTemplateServiceServerService.RealmService
	registrationService   *cryptoutilAppsTemplateServiceServerBusinesslogic.TenantRegistrationService

	// Repositories.
	userRepo                *cryptoutilAppsSmImRepository.UserRepository
	messageRepo             *cryptoutilAppsSmImRepository.MessageRepository
	messageRecipientJWKRepo *cryptoutilAppsSmImRepository.MessageRecipientJWKRepository
	realmRepo               cryptoutilAppsTemplateServiceServerRepository.TenantRealmRepository // Uses service-template repository.
}

// NewFromConfig creates a new sm-im server from SmIMServerSettings only.
// Uses service-template builder for infrastructure initialization.
func NewFromConfig(ctx context.Context, cfg *cryptoutilAppsSmImServerConfig.SmIMServerSettings) (*SmIMServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create server builder with template config.
	builder := cryptoutilAppsTemplateServiceServerBuilder.NewServerBuilder(ctx, cfg.ServiceTemplateServerSettings)

	// Register sm-im specific migrations.
	builder.WithDomainMigrations(cryptoutilAppsSmImRepository.MigrationsFS, "migrations")

	// Register sm-im specific public routes.
	builder.WithPublicRouteRegistration(func(
		base *cryptoutilAppsTemplateServiceServer.PublicServerBase,
		res *cryptoutilAppsTemplateServiceServerBuilder.ServiceResources,
	) error {
		// Create sm-im specific repositories.
		userRepo := cryptoutilAppsSmImRepository.NewUserRepository(res.DB)
		messageRepo := cryptoutilAppsSmImRepository.NewMessageRepository(res.DB)
		messageRecipientJWKRepo := cryptoutilAppsSmImRepository.NewMessageRecipientJWKRepository(res.DB, res.BarrierService)

		// Create public server with sm-im handlers.
		publicServer, err := NewPublicServer(
			base,
			res.SessionManager,
			res.RealmService,
			res.RegistrationService,
			userRepo,
			messageRepo,
			messageRecipientJWKRepo,
			res.JWKGenService,
			res.BarrierService,
		)
		if err != nil {
			return fmt.Errorf("failed to create public server: %w", err)
		}

		// Register all routes (standard + domain-specific).
		if err := publicServer.registerRoutes(); err != nil {
			return fmt.Errorf("failed to register public routes: %w", err)
		}

		return nil
	})

	// Build complete service infrastructure.
	resources, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build sm-im service: %w", err)
	}

	// Create sm-im specific repositories for server struct.
	userRepo := cryptoutilAppsSmImRepository.NewUserRepository(resources.DB)
	messageRepo := cryptoutilAppsSmImRepository.NewMessageRepository(resources.DB)
	messageRecipientJWKRepo := cryptoutilAppsSmImRepository.NewMessageRecipientJWKRepository(resources.DB, resources.BarrierService)

	// Create sm-im server wrapper.
	server := &SmIMServer{
		app:                     resources.Application,
		db:                      resources.DB,
		telemetryService:        resources.TelemetryService,
		jwkGenService:           resources.JWKGenService,
		barrierService:          resources.BarrierService,
		sessionManagerService:   resources.SessionManager,
		realmService:            resources.RealmService,
		registrationService:     resources.RegistrationService,
		userRepo:                userRepo,
		messageRepo:             messageRepo,
		messageRecipientJWKRepo: messageRecipientJWKRepo,
		realmRepo:               resources.RealmRepository,
	}

	return server, nil
}

// Start begins serving both public and admin HTTPS endpoints.
// Blocks until context is cancelled or an unrecoverable error occurs.
func (s *SmIMServer) Start(ctx context.Context) error {
	if err := s.app.Start(ctx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down all servers and closes database connections.
func (s *SmIMServer) Shutdown(ctx context.Context) error {
	if err := s.app.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown application: %w", err)
	}

	return nil
}

// DB returns the GORM database connection (for tests).
func (s *SmIMServer) DB() *gorm.DB {
	return s.db
}

// App returns the application wrapper (for tests).
func (s *SmIMServer) App() *cryptoutilAppsTemplateServiceServer.Application {
	return s.app
}

// JWKGen returns the JWK generation service (for tests).
func (s *SmIMServer) JWKGen() *cryptoutilSharedCryptoJose.JWKGenService {
	return s.jwkGenService
}

// Telemetry returns the telemetry service (for tests).
func (s *SmIMServer) Telemetry() *cryptoutilSharedTelemetry.TelemetryService {
	return s.telemetryService
}

// PublicPort returns the actual port the public server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *SmIMServer) PublicPort() int {
	return s.app.PublicPort()
}

// AdminPort returns the actual port the admin server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *SmIMServer) AdminPort() int {
	return s.app.AdminPort()
}

// SetReady marks the server as ready (enables /admin/v1/readyz to return 200 OK).
func (s *SmIMServer) SetReady(ready bool) {
	s.app.SetReady(ready)
}

// PublicBaseURL returns the public server base URL (for tests).
func (s *SmIMServer) PublicBaseURL() string {
	return s.app.PublicBaseURL()
}

// AdminBaseURL returns the admin server base URL (for tests).
func (s *SmIMServer) AdminBaseURL() string {
	return s.app.AdminBaseURL()
}

// PublicServerActualPort returns the actual port the public server is listening on.
// Useful when configured with port 0 for dynamic allocation.
func (s *SmIMServer) PublicServerActualPort() int {
	return s.app.PublicPort()
}

// AdminServerActualPort returns the actual port the admin server is listening on.
// Useful when configured with port 0 for dynamic allocation.
func (s *SmIMServer) AdminServerActualPort() int {
	return s.app.AdminPort()
}

// SessionManager returns the session manager service (for tests).
func (s *SmIMServer) SessionManager() *cryptoutilAppsTemplateServiceServerBusinesslogic.SessionManagerService {
	return s.sessionManagerService
}

// RealmService returns the realm service (for tests).
func (s *SmIMServer) RealmService() cryptoutilAppsTemplateServiceServerService.RealmService {
	return s.realmService
}

// RegistrationService returns the tenant registration service (for tests).
func (s *SmIMServer) RegistrationService() *cryptoutilAppsTemplateServiceServerBusinesslogic.TenantRegistrationService {
	return s.registrationService
}

// BarrierService returns the barrier service (for tests).
func (s *SmIMServer) BarrierService() *cryptoutilAppsTemplateServiceServerBarrier.Service {
	return s.barrierService
}

// UserRepo returns the user repository (for tests).
func (s *SmIMServer) UserRepo() *cryptoutilAppsSmImRepository.UserRepository {
	return s.userRepo
}

// MessageRepo returns the message repository (for tests).
func (s *SmIMServer) MessageRepo() *cryptoutilAppsSmImRepository.MessageRepository {
	return s.messageRepo
}

// MessageRecipientJWKRepo returns the message recipient JWK repository (for tests).
func (s *SmIMServer) MessageRecipientJWKRepo() *cryptoutilAppsSmImRepository.MessageRecipientJWKRepository {
	return s.messageRecipientJWKRepo
}

// PublicServerBase returns the public server base for testing NewPublicServer.
// This extracts the base from the Application's public server.
func (s *SmIMServer) PublicServerBase() *cryptoutilAppsTemplateServiceServer.PublicServerBase {
	return s.app.PublicServerBase()
}
