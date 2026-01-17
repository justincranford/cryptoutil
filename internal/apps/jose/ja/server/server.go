// Copyright (c) 2025 Justin Cranford
//
//

// Package server implements the jose-ja HTTPS server using the service template.
package server

import (
	"context"
	"fmt"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	"cryptoutil/internal/apps/jose/ja/repository"
	"cryptoutil/internal/apps/jose/ja/server/config"
	cryptoutilTemplateServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilTemplateBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilTemplateBuilder "cryptoutil/internal/apps/template/service/server/builder"
	cryptoutilTemplateBusinessLogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilTemplateRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilTemplateService "cryptoutil/internal/apps/template/service/server/service"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

// JoseJADefaultTenantID is the default tenant ID for jose-ja demo.
// TODO: Move to magic constants once feature is stable.
var JoseJADefaultTenantID = googleUuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d479")

// JoseJADefaultRealmID is the default realm ID for jose-ja demo.
// TODO: Move to magic constants once feature is stable.
var JoseJADefaultRealmID = googleUuid.MustParse("7c9e6679-7425-40de-944b-e07fc1f90ae7")

// JoseJAServer represents the jose-ja service application.
type JoseJAServer struct {
	app *cryptoutilTemplateServer.Application
	db  *gorm.DB

	// Services.
	telemetryService      *cryptoutilTelemetry.TelemetryService
	jwkGenService         *cryptoutilJose.JWKGenService
	barrierService        *cryptoutilTemplateBarrier.BarrierService
	sessionManagerService *cryptoutilTemplateBusinessLogic.SessionManagerService
	realmService          cryptoutilTemplateService.RealmService

	// Repositories.
	elasticJWKRepo  repository.ElasticJWKRepository
	materialJWKRepo repository.MaterialJWKRepository
	auditConfigRepo repository.AuditConfigRepository
	auditLogRepo    repository.AuditLogRepository
	realmRepo       cryptoutilTemplateRepository.TenantRealmRepository // Uses service-template repository.
}

// NewFromConfig creates a new jose-ja server from JoseJAServerSettings only.
// Uses service-template builder for infrastructure initialization.
func NewFromConfig(ctx context.Context, cfg *config.JoseJAServerSettings) (*JoseJAServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create server builder with template config.
	builder := cryptoutilTemplateBuilder.NewServerBuilder(ctx, cfg.ServiceTemplateServerSettings)

	// Register jose-ja specific migrations.
	builder.WithDomainMigrations(repository.MigrationsFS, "migrations")

	// Register jose-ja specific public routes.
	builder.WithPublicRouteRegistration(func(
		base *cryptoutilTemplateServer.PublicServerBase,
		res *cryptoutilTemplateBuilder.ServiceResources,
	) error {
		// Create jose-ja specific repositories.
		elasticJWKRepo := repository.NewElasticJWKRepository(res.DB)
		materialJWKRepo := repository.NewMaterialJWKRepository(res.DB)
		auditConfigRepo := repository.NewAuditConfigRepository(res.DB)
		auditLogRepo := repository.NewAuditLogRepository(res.DB)

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

	// Ensure default tenant exists for jose-ja demo.
	// This is a simplified pattern for educational demos - production services should use
	// the tenant join request workflow (TenantRegistrationService).
	if err := ensureDefaultTenant(ctx, resources.DB, resources.RealmRepository); err != nil {
		return nil, fmt.Errorf("failed to ensure default tenant: %w", err)
	}

	// Create jose-ja specific repositories for server struct.
	elasticJWKRepo := repository.NewElasticJWKRepository(resources.DB)
	materialJWKRepo := repository.NewMaterialJWKRepository(resources.DB)
	auditConfigRepo := repository.NewAuditConfigRepository(resources.DB)
	auditLogRepo := repository.NewAuditLogRepository(resources.DB)

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
func (s *JoseJAServer) App() *cryptoutilTemplateServer.Application {
	return s.app
}

// JWKGen returns the JWK generation service (for tests).
func (s *JoseJAServer) JWKGen() *cryptoutilJose.JWKGenService {
	return s.jwkGenService
}

// Telemetry returns the telemetry service (for tests).
func (s *JoseJAServer) Telemetry() *cryptoutilTelemetry.TelemetryService {
	return s.telemetryService
}

// Barrier returns the barrier service (for tests).
func (s *JoseJAServer) Barrier() *cryptoutilTemplateBarrier.BarrierService {
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

// ensureDefaultTenant creates the default tenant and realm for jose-ja if they don't exist.
// This is a simplified pattern for educational demos - production services should use
// the tenant join request workflow (TenantRegistrationService).
func ensureDefaultTenant(
	ctx context.Context,
	db *gorm.DB,
	realmRepo cryptoutilTemplateRepository.TenantRealmRepository,
) error {
	// Create tenant repository.
	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(db)

	// Check if default tenant exists.
	_, err := tenantRepo.GetByID(ctx, JoseJADefaultTenantID)
	if err == nil {
		// Tenant already exists, nothing to do.
		return nil
	}

	// Create default tenant.
	tenant := &cryptoutilTemplateRepository.Tenant{
		ID:   JoseJADefaultTenantID,
		Name: "jose-ja-default",
	}

	if err := tenantRepo.Create(ctx, tenant); err != nil {
		return fmt.Errorf("failed to create default tenant: %w", err)
	}

	// Create default realm for tenant.
	realm := &cryptoutilTemplateRepository.TenantRealm{
		ID:       googleUuid.New(),
		TenantID: JoseJADefaultTenantID,
		RealmID:  JoseJADefaultRealmID,
		Type:     "api_key", // API key authentication for JOSE services.
		Active:   true,
		Source:   "db",
	}

	if err := realmRepo.Create(ctx, realm); err != nil {
		return fmt.Errorf("failed to create default realm: %w", err)
	}

	return nil
}
