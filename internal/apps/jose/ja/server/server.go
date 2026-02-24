// Copyright (c) 2025 Justin Cranford
//
//

// Package server implements the jose-ja HTTPS server using the service template.
package server

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	cryptoutilAppsJoseJaRepository "cryptoutil/internal/apps/jose/ja/repository"
	cryptoutilAppsJoseJaServerConfig "cryptoutil/internal/apps/jose/ja/server/config"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilAppsTemplateServiceServerBuilder "cryptoutil/internal/apps/template/service/server/builder"
	cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilAppsTemplateServiceServerService "cryptoutil/internal/apps/template/service/server/service"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// JoseJAServer represents the jose-ja service application.
type JoseJAServer struct {
	app *cryptoutilAppsTemplateServiceServer.Application
	db  *gorm.DB

	// Services.
	telemetryService      *cryptoutilSharedTelemetry.TelemetryService
	jwkGenService         *cryptoutilSharedCryptoJose.JWKGenService
	barrierService        *cryptoutilAppsTemplateServiceServerBarrier.Service
	sessionManagerService *cryptoutilAppsTemplateServiceServerBusinesslogic.SessionManagerService
	realmService          cryptoutilAppsTemplateServiceServerService.RealmService

	// Repositories.
	elasticJWKRepo  cryptoutilAppsJoseJaRepository.ElasticJWKRepository
	materialJWKRepo cryptoutilAppsJoseJaRepository.MaterialJWKRepository
	auditConfigRepo cryptoutilAppsJoseJaRepository.AuditConfigRepository
	auditLogRepo    cryptoutilAppsJoseJaRepository.AuditLogRepository
	realmRepo       cryptoutilAppsTemplateServiceServerRepository.TenantRealmRepository // Uses service-template repository.
}

// NewFromConfig creates a new jose-ja server from JoseJAServerSettings only.
// Uses service-template builder for infrastructure initialization.
func NewFromConfig(ctx context.Context, cfg *cryptoutilAppsJoseJaServerConfig.JoseJAServerSettings) (*JoseJAServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create server builder with template config.
	builder := cryptoutilAppsTemplateServiceServerBuilder.NewServerBuilder(ctx, cfg.ServiceTemplateServerSettings)

	// Register jose-ja specific migrations.
	builder.WithDomainMigrations(cryptoutilAppsJoseJaRepository.MigrationsFS, "migrations")

	// Register jose-ja specific public routes.
	builder.WithPublicRouteRegistration(func(
		base *cryptoutilAppsTemplateServiceServer.PublicServerBase,
		res *cryptoutilAppsTemplateServiceServerBuilder.ServiceResources,
	) error {
		// Create jose-ja specific repositories.
		elasticJWKRepo := cryptoutilAppsJoseJaRepository.NewElasticJWKRepository(res.DB)
		materialJWKRepo := cryptoutilAppsJoseJaRepository.NewMaterialJWKRepository(res.DB)
		auditConfigRepo := cryptoutilAppsJoseJaRepository.NewAuditConfigRepository(res.DB)
		auditLogRepo := cryptoutilAppsJoseJaRepository.NewAuditLogRepository(res.DB)

		// Create public server with jose-ja handlers.
		publicServer, err := NewPublicServer(
			base,
			res.SessionManager,
			res.RealmService,
			elasticJWKRepo,
			materialJWKRepo,
			auditConfigRepo,
			auditLogRepo,
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
		return nil, fmt.Errorf("failed to build jose-ja service: %w", err)
	}

	// Create jose-ja specific repositories for server struct.
	elasticJWKRepo := cryptoutilAppsJoseJaRepository.NewElasticJWKRepository(resources.DB)
	materialJWKRepo := cryptoutilAppsJoseJaRepository.NewMaterialJWKRepository(resources.DB)
	auditConfigRepo := cryptoutilAppsJoseJaRepository.NewAuditConfigRepository(resources.DB)
	auditLogRepo := cryptoutilAppsJoseJaRepository.NewAuditLogRepository(resources.DB)

	// Create jose-ja server wrapper.
	server := &JoseJAServer{
		app:                   resources.Application,
		db:                    resources.DB,
		telemetryService:      resources.TelemetryService,
		jwkGenService:         resources.JWKGenService,
		barrierService:        resources.BarrierService,
		sessionManagerService: resources.SessionManager,
		realmService:          resources.RealmService,
		elasticJWKRepo:        elasticJWKRepo,
		materialJWKRepo:       materialJWKRepo,
		auditConfigRepo:       auditConfigRepo,
		auditLogRepo:          auditLogRepo,
		realmRepo:             resources.RealmRepository,
	}

	return server, nil
}

// Start begins serving both public and admin HTTPS endpoints.
// Blocks until context is cancelled or an unrecoverable error occurs.
func (s *JoseJAServer) Start(ctx context.Context) error {
	if err := s.app.Start(ctx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down all servers and closes database connections.
func (s *JoseJAServer) Shutdown(ctx context.Context) error {
	if err := s.app.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown application: %w", err)
	}

	return nil
}

// DB returns the GORM database connection (for tests).
func (s *JoseJAServer) DB() *gorm.DB {
	return s.db
}

// App returns the application wrapper (for tests).
func (s *JoseJAServer) App() *cryptoutilAppsTemplateServiceServer.Application {
	return s.app
}

// JWKGen returns the JWK generation service (for tests).
func (s *JoseJAServer) JWKGen() *cryptoutilSharedCryptoJose.JWKGenService {
	return s.jwkGenService
}

// Telemetry returns the telemetry service (for tests).
func (s *JoseJAServer) Telemetry() *cryptoutilSharedTelemetry.TelemetryService {
	return s.telemetryService
}

// Barrier returns the barrier service (for tests).
func (s *JoseJAServer) Barrier() *cryptoutilAppsTemplateServiceServerBarrier.Service {
	return s.barrierService
}

// PublicPort returns the actual port the public server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *JoseJAServer) PublicPort() int {
	return s.app.PublicPort()
}

// AdminPort returns the actual port the admin server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *JoseJAServer) AdminPort() int {
	return s.app.AdminPort()
}

// SetReady marks the server as ready (enables /admin/v1/readyz to return 200 OK).
func (s *JoseJAServer) SetReady(ready bool) {
	s.app.SetReady(ready)
}

// PublicBaseURL returns the public server base URL (for tests).
func (s *JoseJAServer) PublicBaseURL() string {
	return s.app.PublicBaseURL()
}

// AdminBaseURL returns the admin server base URL (for tests).
func (s *JoseJAServer) AdminBaseURL() string {
	return s.app.AdminBaseURL()
}

// PublicServerActualPort returns the actual port the public server is listening on.
// Useful when configured with port 0 for dynamic allocation.
func (s *JoseJAServer) PublicServerActualPort() int {
	return s.app.PublicPort()
}

// AdminServerActualPort returns the actual port the admin server is listening on.
// Useful when configured with port 0 for dynamic allocation.
func (s *JoseJAServer) AdminServerActualPort() int {
	return s.app.AdminPort()
}
