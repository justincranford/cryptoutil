// Copyright (c) 2025 Justin Cranford

// Package server implements the identity-rs HTTPS server using the service template.
package server

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	cryptoutilAppsIdentityRsServerConfig "cryptoutil/internal/apps/identity/rs/server/config"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilAppsTemplateServiceServerBuilder "cryptoutil/internal/apps/template/service/server/builder"
	cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilAppsTemplateServiceServerService "cryptoutil/internal/apps/template/service/server/service"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// RSServer represents the identity-rs service application.
// This is a reference implementation Resource Server demonstrating protected API access.
type RSServer struct {
	app *cryptoutilAppsTemplateServiceServer.Application
	db  *gorm.DB

	// RS configuration.
	cfg *cryptoutilAppsIdentityRsServerConfig.IdentityRSServerSettings

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

// NewFromConfig creates a new identity-rs server from IdentityRSServerSettings.
// Uses service-template builder for infrastructure initialization.
func NewFromConfig(ctx context.Context, cfg *cryptoutilAppsIdentityRsServerConfig.IdentityRSServerSettings) (*RSServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create server builder with template config.
	// Note: RS uses template database for sessions/barrier but has no domain-specific migrations yet.
	builder := cryptoutilAppsTemplateServiceServerBuilder.NewServerBuilder(ctx, cfg.ServiceTemplateServerSettings)

	// Register identity-rs specific public routes.
	builder.WithPublicRouteRegistration(func(
		base *cryptoutilAppsTemplateServiceServer.PublicServerBase,
		_ *cryptoutilAppsTemplateServiceServerBuilder.ServiceResources,
	) error {
		// Create public server with rs handlers.
		publicServer := NewPublicServer(base, cfg)

		// Register all routes (standard + rs-specific).
		if err := publicServer.registerRoutes(); err != nil {
			return fmt.Errorf("failed to register public routes: %w", err)
		}

		return nil
	})

	// Build complete service infrastructure.
	resources, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build identity-rs service: %w", err)
	}

	// Create identity-rs server wrapper.
	server := &RSServer{
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
func (s *RSServer) Start(ctx context.Context) error {
	if err := s.app.Start(ctx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down all servers and closes database connections.
func (s *RSServer) Shutdown(ctx context.Context) error {
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
func (s *RSServer) Config() *cryptoutilAppsIdentityRsServerConfig.IdentityRSServerSettings {
	return s.cfg
}

// DB returns the GORM database connection (for tests).
func (s *RSServer) DB() *gorm.DB {
	return s.db
}

// App returns the application wrapper (for tests).
func (s *RSServer) App() *cryptoutilAppsTemplateServiceServer.Application {
	return s.app
}

// JWKGen returns the JWK generation service (for tests).
func (s *RSServer) JWKGen() *cryptoutilSharedCryptoJose.JWKGenService {
	return s.jwkGenService
}

// Telemetry returns the telemetry service (for tests).
func (s *RSServer) Telemetry() *cryptoutilSharedTelemetry.TelemetryService {
	return s.telemetryService
}

// Barrier returns the barrier service (for tests).
func (s *RSServer) Barrier() *cryptoutilAppsTemplateServiceServerBarrier.Service {
	return s.barrierService
}

// PublicPort returns the actual port the public server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *RSServer) PublicPort() int {
	return s.app.PublicPort()
}

// AdminPort returns the actual port the admin server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *RSServer) AdminPort() int {
	return s.app.AdminPort()
}

// SetReady marks the server as ready (enables /admin/api/v1/readyz to return 200 OK).
func (s *RSServer) SetReady(ready bool) {
	s.app.SetReady(ready)
}

// PublicBaseURL returns the public server base URL (for tests).
func (s *RSServer) PublicBaseURL() string {
	return s.app.PublicBaseURL()
}

// AdminBaseURL returns the admin server base URL (for tests).
func (s *RSServer) AdminBaseURL() string {
	return s.app.AdminBaseURL()
}
