// Copyright (c) 2025 Justin Cranford

// Package server implements the identity-authz HTTPS server using the service template.
package server

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"cryptoutil/internal/apps/identity/authz/server/config"
	cryptoutilTemplateServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilTemplateBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilTemplateBuilder "cryptoutil/internal/apps/template/service/server/builder"
	cryptoutilTemplateBusinessLogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilTemplateRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilTemplateService "cryptoutil/internal/apps/template/service/server/service"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

// AuthzServer represents the identity-authz service application.
// This is an OAuth 2.1 Authorization Server with OIDC Discovery support.
type AuthzServer struct {
	app *cryptoutilTemplateServer.Application
	db  *gorm.DB

	// Authz configuration.
	cfg *config.IdentityAuthzServerSettings

	// Template services.
	telemetryService      *cryptoutilTelemetry.TelemetryService
	jwkGenService         *cryptoutilJose.JWKGenService
	barrierService        *cryptoutilTemplateBarrier.BarrierService
	sessionManagerService *cryptoutilTemplateBusinessLogic.SessionManagerService
	realmService          cryptoutilTemplateService.RealmService

	// Template repositories.
	realmRepo cryptoutilTemplateRepository.TenantRealmRepository

	// Shutdown functions.
	shutdownCore      func()
	shutdownContainer func()
}

// NewFromConfig creates a new identity-authz server from IdentityAuthzServerSettings.
// Uses service-template builder for infrastructure initialization.
func NewFromConfig(ctx context.Context, cfg *config.IdentityAuthzServerSettings) (*AuthzServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create server builder with template config.
	// Note: Authz uses template database for sessions/barrier but has no domain-specific migrations yet.
	builder := cryptoutilTemplateBuilder.NewServerBuilder(ctx, cfg.ServiceTemplateServerSettings)

	// Register identity-authz specific public routes.
	builder.WithPublicRouteRegistration(func(
		base *cryptoutilTemplateServer.PublicServerBase,
		_ *cryptoutilTemplateBuilder.ServiceResources,
	) error {
		// Create public server with authz handlers.
		publicServer := NewPublicServer(base, cfg)

		// Register all routes (standard + authz-specific).
		if err := publicServer.registerRoutes(); err != nil {
			return fmt.Errorf("failed to register public routes: %w", err)
		}

		return nil
	})

	// Build complete service infrastructure.
	resources, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build identity-authz service: %w", err)
	}

	// Create identity-authz server wrapper.
	server := &AuthzServer{
		app:                   resources.Application,
		db:                    resources.DB,
		cfg:                   cfg,
		telemetryService:      resources.TelemetryService,
		jwkGenService:         resources.JWKGenService,
		barrierService:        resources.BarrierService,
		sessionManagerService: resources.SessionManager,
		realmService:          resources.RealmService,
		realmRepo:             resources.RealmRepository,
		shutdownCore:          resources.ShutdownCore,
		shutdownContainer:     resources.ShutdownContainer,
	}

	return server, nil
}

// Start begins serving both public and admin HTTPS endpoints.
// Blocks until context is cancelled or an unrecoverable error occurs.
func (s *AuthzServer) Start(ctx context.Context) error {
	if err := s.app.Start(ctx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down all servers and closes database connections.
func (s *AuthzServer) Shutdown(ctx context.Context) error {
	if err := s.app.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown application: %w", err)
	}

	// Invoke shutdown callbacks.
	if s.shutdownCore != nil {
		s.shutdownCore()
	}

	if s.shutdownContainer != nil {
		s.shutdownContainer()
	}

	return nil
}

// Config returns the server configuration (for tests).
func (s *AuthzServer) Config() *config.IdentityAuthzServerSettings {
	return s.cfg
}

// DB returns the GORM database connection (for tests).
func (s *AuthzServer) DB() *gorm.DB {
	return s.db
}

// App returns the application wrapper (for tests).
func (s *AuthzServer) App() *cryptoutilTemplateServer.Application {
	return s.app
}

// JWKGen returns the JWK generation service (for tests).
func (s *AuthzServer) JWKGen() *cryptoutilJose.JWKGenService {
	return s.jwkGenService
}

// Telemetry returns the telemetry service (for tests).
func (s *AuthzServer) Telemetry() *cryptoutilTelemetry.TelemetryService {
	return s.telemetryService
}

// Barrier returns the barrier service (for tests).
func (s *AuthzServer) Barrier() *cryptoutilTemplateBarrier.BarrierService {
	return s.barrierService
}

// PublicPort returns the actual port the public server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *AuthzServer) PublicPort() int {
	return s.app.PublicPort()
}

// AdminPort returns the actual port the admin server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *AuthzServer) AdminPort() int {
	return s.app.AdminPort()
}

// SetReady marks the server as ready (enables /admin/api/v1/readyz to return 200 OK).
func (s *AuthzServer) SetReady(ready bool) {
	s.app.SetReady(ready)
}

// PublicBaseURL returns the public server base URL (for tests).
func (s *AuthzServer) PublicBaseURL() string {
	return s.app.PublicBaseURL()
}

// AdminBaseURL returns the admin server base URL (for tests).
func (s *AuthzServer) AdminBaseURL() string {
	return s.app.AdminBaseURL()
}
