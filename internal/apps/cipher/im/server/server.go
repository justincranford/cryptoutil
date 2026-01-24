// Copyright (c) 2025 Justin Cranford
//
//

// Package server implements the cipher-im HTTPS server using the service template.
package server

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"cryptoutil/internal/apps/cipher/im/repository"
	"cryptoutil/internal/apps/cipher/im/server/config"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilTemplateBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilTemplateBuilder "cryptoutil/internal/apps/template/service/server/builder"
	cryptoutilTemplateBusinessLogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilTemplateService "cryptoutil/internal/apps/template/service/server/service"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

// CipherIMServer represents the cipher-im service application.
type CipherIMServer struct {
	app *cryptoutilAppsTemplateServiceServer.Application
	db  *gorm.DB

	// Services.
	telemetryService      *cryptoutilTelemetry.TelemetryService
	jwkGenService         *cryptoutilJose.JWKGenService
	barrierService        *cryptoutilTemplateBarrier.BarrierService
	sessionManagerService *cryptoutilTemplateBusinessLogic.SessionManagerService
	realmService          cryptoutilTemplateService.RealmService
	registrationService   *cryptoutilTemplateBusinessLogic.TenantRegistrationService

	// Repositories.
	userRepo                *repository.UserRepository
	messageRepo             *repository.MessageRepository
	messageRecipientJWKRepo *repository.MessageRecipientJWKRepository
	realmRepo               cryptoutilAppsTemplateServiceServerRepository.TenantRealmRepository // Uses service-template repository.
}

// NewFromConfig creates a new cipher-im server from CipherImServerSettings only.
// Uses service-template builder for infrastructure initialization.
func NewFromConfig(ctx context.Context, cfg *config.CipherImServerSettings) (*CipherIMServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create server builder with template config.
	builder := cryptoutilTemplateBuilder.NewServerBuilder(ctx, cfg.ServiceTemplateServerSettings)

	// Register cipher-im specific migrations.
	builder.WithDomainMigrations(repository.MigrationsFS, "migrations")

	// Register cipher-im specific public routes.
	builder.WithPublicRouteRegistration(func(
		base *cryptoutilAppsTemplateServiceServer.PublicServerBase,
		res *cryptoutilTemplateBuilder.ServiceResources,
	) error {
		// Create cipher-im specific repositories.
		userRepo := repository.NewUserRepository(res.DB)
		messageRepo := repository.NewMessageRepository(res.DB)
		messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(res.DB, res.BarrierService)

		// Create public server with cipher-im handlers.
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
		return nil, fmt.Errorf("failed to build cipher-im service: %w", err)
	}

	// Create cipher-im specific repositories for server struct.
	userRepo := repository.NewUserRepository(resources.DB)
	messageRepo := repository.NewMessageRepository(resources.DB)
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(resources.DB, resources.BarrierService)

	// Create cipher-im server wrapper.
	server := &CipherIMServer{
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
func (s *CipherIMServer) Start(ctx context.Context) error {
	if err := s.app.Start(ctx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down all servers and closes database connections.
func (s *CipherIMServer) Shutdown(ctx context.Context) error {
	if err := s.app.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown application: %w", err)
	}

	return nil
}

// DB returns the GORM database connection (for tests).
func (s *CipherIMServer) DB() *gorm.DB {
	return s.db
}

// App returns the application wrapper (for tests).
func (s *CipherIMServer) App() *cryptoutilAppsTemplateServiceServer.Application {
	return s.app
}

// JWKGen returns the JWK generation service (for tests).
func (s *CipherIMServer) JWKGen() *cryptoutilJose.JWKGenService {
	return s.jwkGenService
}

// Telemetry returns the telemetry service (for tests).
func (s *CipherIMServer) Telemetry() *cryptoutilTelemetry.TelemetryService {
	return s.telemetryService
}

// PublicPort returns the actual port the public server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *CipherIMServer) PublicPort() int {
	return s.app.PublicPort()
}

// AdminPort returns the actual port the admin server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *CipherIMServer) AdminPort() int {
	return s.app.AdminPort()
}

// SetReady marks the server as ready (enables /admin/v1/readyz to return 200 OK).
func (s *CipherIMServer) SetReady(ready bool) {
	s.app.SetReady(ready)
}

// PublicBaseURL returns the public server base URL (for tests).
func (s *CipherIMServer) PublicBaseURL() string {
	return s.app.PublicBaseURL()
}

// AdminBaseURL returns the admin server base URL (for tests).
func (s *CipherIMServer) AdminBaseURL() string {
	return s.app.AdminBaseURL()
}

// PublicServerActualPort returns the actual port the public server is listening on.
// Useful when configured with port 0 for dynamic allocation.
func (s *CipherIMServer) PublicServerActualPort() int {
	return s.app.PublicPort()
}

// AdminServerActualPort returns the actual port the admin server is listening on.
// Useful when configured with port 0 for dynamic allocation.
func (s *CipherIMServer) AdminServerActualPort() int {
	return s.app.AdminPort()
}

// SessionManager returns the session manager service (for tests).
func (s *CipherIMServer) SessionManager() *cryptoutilTemplateBusinessLogic.SessionManagerService {
	return s.sessionManagerService
}

// RealmService returns the realm service (for tests).
func (s *CipherIMServer) RealmService() cryptoutilTemplateService.RealmService {
	return s.realmService
}

// RegistrationService returns the tenant registration service (for tests).
func (s *CipherIMServer) RegistrationService() *cryptoutilTemplateBusinessLogic.TenantRegistrationService {
	return s.registrationService
}

// BarrierService returns the barrier service (for tests).
func (s *CipherIMServer) BarrierService() *cryptoutilTemplateBarrier.BarrierService {
	return s.barrierService
}

// UserRepo returns the user repository (for tests).
func (s *CipherIMServer) UserRepo() *repository.UserRepository {
	return s.userRepo
}

// MessageRepo returns the message repository (for tests).
func (s *CipherIMServer) MessageRepo() *repository.MessageRepository {
	return s.messageRepo
}

// MessageRecipientJWKRepo returns the message recipient JWK repository (for tests).
func (s *CipherIMServer) MessageRecipientJWKRepo() *repository.MessageRecipientJWKRepository {
	return s.messageRecipientJWKRepo
}

// PublicServerBase returns the public server base for testing NewPublicServer.
// This extracts the base from the Application's public server.
func (s *CipherIMServer) PublicServerBase() *cryptoutilAppsTemplateServiceServer.PublicServerBase {
	return s.app.PublicServerBase()
}
