// Copyright (c) 2025 Justin Cranford

// Package server implements the identity-idp HTTPS server using the service template.
package server

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	cryptoutilAppsIdentityIdpServerConfig "cryptoutil/internal/apps/identity/idp/server/config"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilAppsTemplateServiceServerBuilder "cryptoutil/internal/apps/template/service/server/builder"
	cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilAppsTemplateServiceServerService "cryptoutil/internal/apps/template/service/server/service"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// IDPServer represents the identity-idp service application.
// This is an OIDC 1.0 Identity Provider with login/consent UI and MFA enrollment.
type IDPServer struct {
	app *cryptoutilAppsTemplateServiceServer.Application
	db  *gorm.DB

	// IDP configuration.
	cfg *cryptoutilAppsIdentityIdpServerConfig.IdentityIDPServerSettings

	// Template services.
	telemetryService      *cryptoutilSharedTelemetry.TelemetryService
	jwkGenService         *cryptoutilSharedCryptoJose.JWKGenService
	barrierService        *cryptoutilAppsTemplateServiceServerBarrier.Service
	sessionManagerService *cryptoutilAppsTemplateServiceServerBusinesslogic.SessionManagerService
	realmService          cryptoutilAppsTemplateServiceServerService.RealmService

	// Template repositories.
	realmRepo cryptoutilAppsTemplateServiceServerRepository.TenantRealmRepository

	// Shutdown functions.
	shutdownCore      func()
	shutdownContainer func()
}

// NewFromConfig creates a new identity-idp server from IdentityIDPServerSettings.
// Uses service-template builder for infrastructure initialization.
func NewFromConfig(ctx context.Context, cfg *cryptoutilAppsIdentityIdpServerConfig.IdentityIDPServerSettings) (*IDPServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create server builder with template config.
	// Note: IDP uses template database for sessions/barrier but has no domain-specific migrations yet.
	builder := cryptoutilAppsTemplateServiceServerBuilder.NewServerBuilder(ctx, cfg.ServiceTemplateServerSettings)

	// Register identity-idp specific public routes.
	builder.WithPublicRouteRegistration(func(
		base *cryptoutilAppsTemplateServiceServer.PublicServerBase,
		_ *cryptoutilAppsTemplateServiceServerBuilder.ServiceResources,
	) error {
		// Create public server with idp handlers.
		publicServer := NewPublicServer(base, cfg)

		// Register all routes (standard + idp-specific).
		if err := publicServer.registerRoutes(); err != nil {
			return fmt.Errorf("failed to register public routes: %w", err)
		}

		return nil
	})

	// Build complete service infrastructure.
	resources, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build identity-idp service: %w", err)
	}

	// Create identity-idp server wrapper.
	server := &IDPServer{
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
func (s *IDPServer) Start(ctx context.Context) error {
	if err := s.app.Start(ctx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down all servers and closes database connections.
func (s *IDPServer) Shutdown(ctx context.Context) error {
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
func (s *IDPServer) Config() *cryptoutilAppsIdentityIdpServerConfig.IdentityIDPServerSettings {
	return s.cfg
}

// DB returns the GORM database connection (for tests).
func (s *IDPServer) DB() *gorm.DB {
	return s.db
}

// App returns the application wrapper (for tests).
func (s *IDPServer) App() *cryptoutilAppsTemplateServiceServer.Application {
	return s.app
}

// JWKGen returns the JWK generation service (for tests).
func (s *IDPServer) JWKGen() *cryptoutilSharedCryptoJose.JWKGenService {
	return s.jwkGenService
}

// Telemetry returns the telemetry service (for tests).
func (s *IDPServer) Telemetry() *cryptoutilSharedTelemetry.TelemetryService {
	return s.telemetryService
}

// Barrier returns the barrier service (for tests).
func (s *IDPServer) Barrier() *cryptoutilAppsTemplateServiceServerBarrier.Service {
	return s.barrierService
}

// PublicPort returns the actual port the public server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *IDPServer) PublicPort() int {
	return s.app.PublicPort()
}

// AdminPort returns the actual port the admin server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *IDPServer) AdminPort() int {
	return s.app.AdminPort()
}

// SetReady marks the server as ready (enables /admin/api/v1/readyz to return 200 OK).
func (s *IDPServer) SetReady(ready bool) {
	s.app.SetReady(ready)
}

// PublicBaseURL returns the public server base URL (for tests).
func (s *IDPServer) PublicBaseURL() string {
	return s.app.PublicBaseURL()
}

// AdminBaseURL returns the admin server base URL (for tests).
func (s *IDPServer) AdminBaseURL() string {
	return s.app.AdminBaseURL()
}
