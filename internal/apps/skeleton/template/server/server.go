// Copyright (c) 2025 Justin Cranford
//
// TEMPLATE: Copy and rename 'skeleton' → your-service-name before use.

// Package server implements the skeleton-template HTTPS server using the service template.
package server

import (
	"context"
	"crypto/x509"
	"fmt"

	"gorm.io/gorm"

	cryptoutilSkeletonTemplateServer "cryptoutil/api/skeleton-template/server"
	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps/framework/service/server"
	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps/framework/service/server/barrier"
	cryptoutilAppsFrameworkServiceServerBuilder "cryptoutil/internal/apps/framework/service/server/builder"
	cryptoutilAppsFrameworkServiceServerBusinesslogic "cryptoutil/internal/apps/framework/service/server/businesslogic"
	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps/framework/service/server/repository"
	cryptoutilAppsFrameworkServiceServerService "cryptoutil/internal/apps/framework/service/server/service"
	cryptoutilAppsSkeletonTemplateRepository "cryptoutil/internal/apps/skeleton/template/repository"
	cryptoutilAppsSkeletonTemplateServerConfig "cryptoutil/internal/apps/skeleton/template/server/config"
	cryptoutilAppsSkeletonTemplateServerHandler "cryptoutil/internal/apps/skeleton/template/server/handler"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// SkeletonTemplateServer represents the skeleton-template service application.
type SkeletonTemplateServer struct {
	app *cryptoutilAppsFrameworkServiceServer.Application
	db  *gorm.DB

	// Services.
	telemetryService      *cryptoutilSharedTelemetry.TelemetryService
	jwkGenService         *cryptoutilSharedCryptoJose.JWKGenService
	barrierService        *cryptoutilAppsFrameworkServiceServerBarrier.Service
	sessionManagerService *cryptoutilAppsFrameworkServiceServerBusinesslogic.SessionManagerService
	realmService          cryptoutilAppsFrameworkServiceServerService.RealmService

	// Repositories.
	realmRepo cryptoutilAppsFrameworkServiceServerRepository.TenantRealmRepository // Uses service-framework repository.
}

// NewFromConfig creates a new skeleton-template server from SkeletonTemplateServerSettings only.
// Uses service-framework builder for infrastructure initialization.
func NewFromConfig(ctx context.Context, cfg *cryptoutilAppsSkeletonTemplateServerConfig.SkeletonTemplateServerSettings) (*SkeletonTemplateServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create server builder with template config.
	resources, err := cryptoutilAppsFrameworkServiceServerBuilder.Build(ctx, cfg.ServiceFrameworkServerSettings, &cryptoutilAppsFrameworkServiceServerBuilder.DomainConfig{
		MigrationsFS:   cryptoutilAppsSkeletonTemplateRepository.MigrationsFS,
		MigrationsPath: "migrations",
		RouteRegistration: func(base *cryptoutilAppsFrameworkServiceServer.PublicServerBase, res *cryptoutilAppsFrameworkServiceServerBuilder.ServiceResources) error {
			return registerItemRoutes(base, res)
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build skeleton-template service: %w", err)
	}

	server := &SkeletonTemplateServer{
		app:                   resources.Application,
		db:                    resources.DB,
		telemetryService:      resources.TelemetryService,
		jwkGenService:         resources.JWKGenService,
		barrierService:        resources.BarrierService,
		sessionManagerService: resources.SessionManager,
		realmService:          resources.RealmService,
		realmRepo:             resources.RealmRepository,
	}

	return server, nil
}

// Start begins serving both public and admin HTTPS endpoints.
// Blocks until context is cancelled or an unrecoverable error occurs.
func (s *SkeletonTemplateServer) Start(ctx context.Context) error {
	if err := s.app.Start(ctx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down all servers and closes database connections.
func (s *SkeletonTemplateServer) Shutdown(ctx context.Context) error {
	if err := s.app.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown application: %w", err)
	}

	return nil
}

// DB returns the GORM database connection (for tests).
func (s *SkeletonTemplateServer) DB() *gorm.DB {
	return s.db
}

// App returns the application wrapper (for tests).
func (s *SkeletonTemplateServer) App() *cryptoutilAppsFrameworkServiceServer.Application {
	return s.app
}

// JWKGen returns the JWK generation service (for tests).
func (s *SkeletonTemplateServer) JWKGen() *cryptoutilSharedCryptoJose.JWKGenService {
	return s.jwkGenService
}

// Telemetry returns the telemetry service (for tests).
func (s *SkeletonTemplateServer) Telemetry() *cryptoutilSharedTelemetry.TelemetryService {
	return s.telemetryService
}

// Barrier returns the barrier service (for tests).
func (s *SkeletonTemplateServer) Barrier() *cryptoutilAppsFrameworkServiceServerBarrier.Service {
	return s.barrierService
}

// PublicPort returns the actual port the public server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *SkeletonTemplateServer) PublicPort() int {
	return s.app.PublicPort()
}

// AdminPort returns the actual port the admin server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *SkeletonTemplateServer) AdminPort() int {
	return s.app.AdminPort()
}

// SetReady marks the server as ready (enables /admin/v1/readyz to return 200 OK).
func (s *SkeletonTemplateServer) SetReady(ready bool) {
	s.app.SetReady(ready)
}

// PublicBaseURL returns the public server base URL (for tests).
func (s *SkeletonTemplateServer) PublicBaseURL() string {
	return s.app.PublicBaseURL()
}

// AdminBaseURL returns the admin server base URL (for tests).
func (s *SkeletonTemplateServer) AdminBaseURL() string {
	return s.app.AdminBaseURL()
}

// PublicServerActualPort returns the actual port the public server is listening on.
// Useful when configured with port 0 for dynamic allocation.
func (s *SkeletonTemplateServer) PublicServerActualPort() int {
	return s.app.PublicPort()
}

// AdminServerActualPort returns the actual port the admin server is listening on.
// Useful when configured with port 0 for dynamic allocation.
func (s *SkeletonTemplateServer) AdminServerActualPort() int {
	return s.app.AdminPort()
}

// TLSRootCAPool returns the root CA pool for test client TLS configuration (public server).
func (s *SkeletonTemplateServer) TLSRootCAPool() *x509.CertPool {
	return s.app.TLSRootCAPool()
}

// AdminTLSRootCAPool returns the admin TLS root CA pool for test client TLS configuration.
func (s *SkeletonTemplateServer) AdminTLSRootCAPool() *x509.CertPool {
	return s.app.AdminTLSRootCAPool()
}

// Compile-time assertion: SkeletonTemplateServer must implement ServiceServer.
var _ cryptoutilAppsFrameworkServiceServer.ServiceServer = (*SkeletonTemplateServer)(nil)

// registerItemRoutes sets up the Item CRUD routes using the OpenAPI strict server pattern.
func registerItemRoutes(base *cryptoutilAppsFrameworkServiceServer.PublicServerBase, res *cryptoutilAppsFrameworkServiceServerBuilder.ServiceResources) error {
	// Create domain repository.
	itemRepo := cryptoutilAppsSkeletonTemplateRepository.NewItemRepository(res.DB)

	// Create OpenAPI strict server handler.
	strictServer := cryptoutilAppsSkeletonTemplateServerHandler.NewStrictServer(itemRepo)
	strictHandler := cryptoutilSkeletonTemplateServer.NewStrictHandler(strictServer, nil)

	// Register handlers on both browser and service paths.
	app := base.App()

	cryptoutilSkeletonTemplateServer.RegisterHandlersWithOptions(app, strictHandler, cryptoutilSkeletonTemplateServer.FiberServerOptions{
		BaseURL: cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath,
	})

	cryptoutilSkeletonTemplateServer.RegisterHandlersWithOptions(app, strictHandler, cryptoutilSkeletonTemplateServer.FiberServerOptions{
		BaseURL: cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath,
	})

	return nil
}
