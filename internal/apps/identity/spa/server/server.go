// Copyright (c) 2025 Justin Cranford

// Package server provides the identity-spa server implementation.
// SPA serves static files for the Single Page Application frontend (reference implementation).
package server

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	cryptoutilSPAConfig "cryptoutil/internal/apps/identity/spa/server/config"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilTemplateBuilder "cryptoutil/internal/apps/template/service/server/builder"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

// SPAServer wraps the template Application with SPA-specific functionality.
// SPA serves static files for Single Page Application frontends.
type SPAServer struct {
	cfg              *cryptoutilSPAConfig.IdentitySPAServerSettings
	app              *cryptoutilAppsTemplateServiceServer.Application
	db               *gorm.DB
	barrierService   *cryptoutilBarrier.Service
	jwkGenService    *cryptoutilJose.JWKGenService
	telemetryService *cryptoutilTelemetry.TelemetryService
	shutdownCore     func()
}

// NewFromConfig creates a new SPA server from configuration.
func NewFromConfig(ctx context.Context, cfg *cryptoutilSPAConfig.IdentitySPAServerSettings) (*SPAServer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	// Create server builder with template configuration.
	builder := cryptoutilTemplateBuilder.NewServerBuilder(ctx, cfg.ServiceTemplateServerSettings)

	// No domain-specific migrations for SPA (static file server).

	// Register public route registration for SPA endpoints.
	builder.WithPublicRouteRegistration(func(
		base *cryptoutilAppsTemplateServiceServer.PublicServerBase,
		_ *cryptoutilTemplateBuilder.ServiceResources,
	) error {
		// Create SPA public server.
		publicServer := NewPublicServer(base, cfg)

		// Register SPA-specific routes.
		publicServer.RegisterRoutes()

		return nil
	})

	// Build the server infrastructure.
	resources, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build server infrastructure: %w", err)
	}

	server := &SPAServer{
		cfg:              cfg,
		app:              resources.Application,
		db:               resources.DB,
		barrierService:   resources.BarrierService,
		jwkGenService:    resources.JWKGenService,
		telemetryService: resources.TelemetryService,
		shutdownCore:     resources.ShutdownCore,
	}

	return server, nil
}

// Start starts the SPA server (blocking).
func (s *SPAServer) Start(ctx context.Context) error {
	if err := s.app.Start(ctx); err != nil {
		return fmt.Errorf("failed to start SPA server: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the SPA server.
func (s *SPAServer) Shutdown(ctx context.Context) error {
	var errs []error

	// Shutdown application (servers).
	if s.app != nil {
		if err := s.app.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("app shutdown: %w", err))
		}
	}

	// Shutdown core resources (barrier, session manager, telemetry).
	if s.shutdownCore != nil {
		s.shutdownCore()
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	return nil
}

// DB returns the database connection for tests.
func (s *SPAServer) DB() *gorm.DB {
	return s.db
}

// App returns the template Application for tests.
func (s *SPAServer) App() *cryptoutilAppsTemplateServiceServer.Application {
	return s.app
}

// JWKGen returns the JWK generation service for tests.
func (s *SPAServer) JWKGen() *cryptoutilJose.JWKGenService {
	return s.jwkGenService
}

// Telemetry returns the telemetry service for tests.
func (s *SPAServer) Telemetry() *cryptoutilTelemetry.TelemetryService {
	return s.telemetryService
}

// Barrier returns the barrier encryption service for tests.
func (s *SPAServer) Barrier() *cryptoutilBarrier.Service {
	return s.barrierService
}

// Config returns the server configuration for tests.
func (s *SPAServer) Config() *cryptoutilSPAConfig.IdentitySPAServerSettings {
	return s.cfg
}

// PublicPort returns the actual public port (for dynamic port allocation in tests).
func (s *SPAServer) PublicPort() int {
	if s.app == nil {
		return 0
	}

	return s.app.PublicPort()
}

// AdminPort returns the actual admin port (for dynamic port allocation in tests).
func (s *SPAServer) AdminPort() int {
	if s.app == nil {
		return 0
	}

	return s.app.AdminPort()
}

// SetReady sets the server ready state (for health checks).
func (s *SPAServer) SetReady(ready bool) {
	if s.app != nil {
		s.app.SetReady(ready)
	}
}

// PublicBaseURL returns the base URL for public API endpoints.
func (s *SPAServer) PublicBaseURL() string {
	if s.app == nil {
		return ""
	}

	return s.app.PublicBaseURL()
}

// AdminBaseURL returns the base URL for admin API endpoints.
func (s *SPAServer) AdminBaseURL() string {
	if s.app == nil {
		return ""
	}

	return s.app.AdminBaseURL()
}
