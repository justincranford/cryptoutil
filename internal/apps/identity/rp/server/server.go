// Copyright (c) 2025 Justin Cranford

// Package server implements the identity-rp HTTPS server using the service template.
package server

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	cryptoutilAppsIdentityRpServerConfig "cryptoutil/internal/apps/identity/rp/server/config"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilAppsTemplateServiceServerBuilder "cryptoutil/internal/apps/template/service/server/builder"
	cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilAppsTemplateServiceServerService "cryptoutil/internal/apps/template/service/server/service"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// RPServer represents the identity-rp service application.
// This is a Backend-for-Frontend (BFF) that proxies OAuth 2.1 flows for SPAs.
type RPServer struct {
	app *cryptoutilAppsTemplateServiceServer.Application
	db  *gorm.DB

	// RP configuration.
	cfg *cryptoutilAppsIdentityRpServerConfig.IdentityRPServerSettings

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

// NewFromConfig creates a new identity-rp server from IdentityRPServerSettings.
// Uses service-template builder for infrastructure initialization.
func NewFromConfig(ctx context.Context, cfg *cryptoutilAppsIdentityRpServerConfig.IdentityRPServerSettings) (*RPServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create server builder with template config.
	// Note: RP uses template database for sessions/barrier but has no domain-specific migrations.
	builder := cryptoutilAppsTemplateServiceServerBuilder.NewServerBuilder(ctx, cfg.ServiceTemplateServerSettings)

	// Register identity-rp specific public routes.
	builder.WithPublicRouteRegistration(func(
		base *cryptoutilAppsTemplateServiceServer.PublicServerBase,
		res *cryptoutilAppsTemplateServiceServerBuilder.ServiceResources,
	) error {
		// Create public server with BFF handlers.
		publicServer := NewPublicServer(base, cfg, res.SessionManager, res.RealmService)

		// Register all routes (standard + RP-specific).
		if err := publicServer.registerRoutes(); err != nil {
			return fmt.Errorf("failed to register public routes: %w", err)
		}

		return nil
	})

	// Build complete service infrastructure.
	resources, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build identity-rp service: %w", err)
	}

	// Create identity-rp server wrapper.
	server := &RPServer{
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
func (s *RPServer) Start(ctx context.Context) error {
	if err := s.app.Start(ctx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down all servers and closes database connections.
func (s *RPServer) Shutdown(ctx context.Context) error {
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

// DB returns the GORM database connection (for tests).
func (s *RPServer) DB() *gorm.DB {
	return s.db
}

// App returns the application wrapper (for tests).
func (s *RPServer) App() *cryptoutilAppsTemplateServiceServer.Application {
	return s.app
}

// JWKGen returns the JWK generation service (for tests).
func (s *RPServer) JWKGen() *cryptoutilSharedCryptoJose.JWKGenService {
	return s.jwkGenService
}

// Telemetry returns the telemetry service (for tests).
func (s *RPServer) Telemetry() *cryptoutilSharedTelemetry.TelemetryService {
	return s.telemetryService
}

// Barrier returns the barrier service (for tests).
func (s *RPServer) Barrier() *cryptoutilAppsTemplateServiceServerBarrier.Service {
	return s.barrierService
}

// PublicPort returns the actual port the public server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *RPServer) PublicPort() int {
	return s.app.PublicPort()
}

// AdminPort returns the actual port the admin server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *RPServer) AdminPort() int {
	return s.app.AdminPort()
}

// SetReady marks the server as ready (enables /admin/api/v1/readyz to return 200 OK).
func (s *RPServer) SetReady(ready bool) {
	s.app.SetReady(ready)
}

// PublicBaseURL returns the public server base URL (for tests).
func (s *RPServer) PublicBaseURL() string {
	return s.app.PublicBaseURL()
}

// AdminBaseURL returns the admin server base URL (for tests).
func (s *RPServer) AdminBaseURL() string {
	return s.app.AdminBaseURL()
}

// PublicServerActualPort returns the actual port the public server is listening on.
// Useful when configured with port 0 for dynamic allocation.
func (s *RPServer) PublicServerActualPort() int {
	return s.app.PublicPort()
}

// AdminServerActualPort returns the actual port the admin server is listening on.
// Useful when configured with port 0 for dynamic allocation.
func (s *RPServer) AdminServerActualPort() int {
	return s.app.AdminPort()
}
