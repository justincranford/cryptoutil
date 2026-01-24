// Copyright (c) 2025 Justin Cranford
//

// Package server implements the JOSE Authority Server HTTPS service using the service template.
package server

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilTemplateBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilTemplateBuilder "cryptoutil/internal/apps/template/service/server/builder"
	cryptoutilTemplateBusinessLogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilTemplateService "cryptoutil/internal/apps/template/service/server/service"
	cryptoutilJoseConfig "cryptoutil/internal/jose/config"
	"cryptoutil/internal/jose/repository"
	"cryptoutil/internal/jose/server/middleware"
	"cryptoutil/internal/jose/service"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

// JoseServer represents the JOSE Authority Server using the service template.
// This wraps the service-template Application with JOSE-specific functionality.
type JoseServer struct {
	app                *cryptoutilAppsTemplateServiceServer.Application
	db                 *gorm.DB
	telemetryService   *cryptoutilTelemetry.TelemetryService
	jwkGenService      *cryptoutilJose.JWKGenService
	barrierService     *cryptoutilTemplateBarrier.BarrierService
	sessionManager     *cryptoutilTemplateBusinessLogic.SessionManagerService
	realmService       cryptoutilTemplateService.RealmService
	realmRepo          cryptoutilAppsTemplateServiceServerRepository.TenantRealmRepository
	elasticJWKRepo     repository.ElasticJWKRepository
	materialJWKRepo    repository.MaterialJWKRepository
	elasticJWKService  *service.ElasticJWKService
	auditConfigRepo    repository.AuditConfigRepository
	auditLogRepo       repository.AuditLogRepository
	auditConfigService *service.AuditConfigService
	keyStore           *KeyStore // In-memory key store for legacy API compatibility.
	cfg                *cryptoutilJoseConfig.JoseServerSettings
}

// NewFromConfig creates a new JOSE server from JoseServerSettings.
// Uses service-template builder for infrastructure initialization.
func NewFromConfig(ctx context.Context, cfg *cryptoutilJoseConfig.JoseServerSettings) (*JoseServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create server builder with template config.
	builder := cryptoutilTemplateBuilder.NewServerBuilder(ctx, cfg.ServiceTemplateServerSettings)

	// Register JOSE-specific migrations (2001-2004).
	builder.WithDomainMigrations(repository.MigrationsFS, "migrations")

	// Capture repositories and services from route registration closure.
	var (
		elasticJWKRepo     repository.ElasticJWKRepository
		materialJWKRepo    repository.MaterialJWKRepository
		elasticJWKService  *service.ElasticJWKService
		auditConfigRepo    repository.AuditConfigRepository
		auditLogRepo       repository.AuditLogRepository
		auditConfigService *service.AuditConfigService
		keyStore           *KeyStore
	)

	// Register JOSE-specific public routes.
	builder.WithPublicRouteRegistration(func(
		base *cryptoutilAppsTemplateServiceServer.PublicServerBase,
		res *cryptoutilTemplateBuilder.ServiceResources,
	) error {
		// Create JOSE-specific repositories.
		elasticJWKRepo = repository.NewElasticJWKRepository(res.DB)
		materialJWKRepo = repository.NewMaterialJWKRepository(res.DB)
		auditConfigRepo = repository.NewAuditConfigGormRepository(res.DB)
		auditLogRepo = repository.NewAuditLogGormRepository(res.DB)

		// Create AuditConfigService.
		auditConfigService = service.NewAuditConfigService(auditConfigRepo)

		// Create ElasticJWKService.
		elasticJWKService = service.NewElasticJWKService(
			elasticJWKRepo,
			materialJWKRepo,
			res.JWKGenService,
			res.BarrierService,
		)

		// Create in-memory key store for legacy API compatibility.
		keyStore = NewKeyStore()

		// Register routes.
		return registerJosePublicRoutes(
			base.App(),
			res.TelemetryService,
			res.JWKGenService,
			elasticJWKService,
			auditConfigService,
			keyStore,
			cfg,
		)
	})

	// Build complete service infrastructure.
	resources, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build jose-ja service: %w", err)
	}

	// Create JOSE server wrapper.
	server := &JoseServer{
		app:                resources.Application,
		db:                 resources.DB,
		telemetryService:   resources.TelemetryService,
		jwkGenService:      resources.JWKGenService,
		barrierService:     resources.BarrierService,
		sessionManager:     resources.SessionManager,
		realmService:       resources.RealmService,
		realmRepo:          resources.RealmRepository,
		elasticJWKRepo:     elasticJWKRepo,
		materialJWKRepo:    materialJWKRepo,
		elasticJWKService:  elasticJWKService,
		auditConfigRepo:    auditConfigRepo,
		auditLogRepo:       auditLogRepo,
		auditConfigService: auditConfigService,
		keyStore:           keyStore,
		cfg:                cfg,
	}

	return server, nil
}

// registerJosePublicRoutes registers all JOSE-specific routes on the public server.
// Uses the existing handler functions from handlers.go with adapter pattern.
func registerJosePublicRoutes(
	app *fiber.App,
	telemetryService *cryptoutilTelemetry.TelemetryService,
	jwkGenService *cryptoutilJose.JWKGenService,
	elasticJWKService *service.ElasticJWKService,
	auditConfigService *service.AuditConfigService,
	keyStore *KeyStore,
	_ *cryptoutilJoseConfig.JoseServerSettings,
) error {
	// Create handler adapter that wraps existing handler functions.
	h := &joseHandlerAdapter{
		telemetryService:  telemetryService,
		jwkGenService:     jwkGenService,
		elasticJWKService: elasticJWKService,
		keyStore:          keyStore,
	}

	// Create audit config handlers.
	auditH := newAuditConfigHandlers(auditConfigService)

	// Well-known endpoints (no auth required for public key discovery).
	app.Get("/.well-known/jwks.json", h.handleJWKS)

	// Create rate limiter middleware.
	rateLimiter := middleware.NewRateLimiter(&middleware.RateLimitConfig{
		Max:              middleware.DefaultRateLimit, // 100 requests per second per IP.
		Expiration:       middleware.DefaultRateLimitExpiration,
		TelemetryService: telemetryService,
	})

	// Service API v1 group (headless clients).
	serviceV1 := app.Group("/service/api/v1/jose")
	serviceV1.Use(rateLimiter)
	serviceV1.Post("/jwk/generate", h.handleJWKGenerate)
	serviceV1.Get("/jwk/:kid", h.handleJWKGet)
	serviceV1.Delete("/jwk/:kid", h.handleJWKDelete)
	serviceV1.Get("/jwk", h.handleJWKList)
	serviceV1.Get("/jwks", h.handleJWKS)
	serviceV1.Post("/jws/sign", h.handleJWSSign)
	serviceV1.Post("/jws/verify", h.handleJWSVerify)
	serviceV1.Post("/jwe/encrypt", h.handleJWEEncrypt)
	serviceV1.Post("/jwe/decrypt", h.handleJWEDecrypt)
	serviceV1.Post("/jwt/sign", h.handleJWTSign)
	serviceV1.Post("/jwt/verify", h.handleJWTVerify)

	// Elastic JWK JWKS endpoint - returns public keys for verification/encryption.
	serviceV1.Get("/elastic-jwks/:kid/.well-known/jwks.json", h.handleElasticJWKS)

	// Browser API v1 group (browser clients).
	browserV1 := app.Group("/browser/api/v1/jose")
	browserV1.Use(rateLimiter)
	browserV1.Post("/jwk/generate", h.handleJWKGenerate)
	browserV1.Get("/jwk/:kid", h.handleJWKGet)
	browserV1.Delete("/jwk/:kid", h.handleJWKDelete)
	browserV1.Get("/jwk", h.handleJWKList)
	browserV1.Get("/jwks", h.handleJWKS)
	browserV1.Post("/jws/sign", h.handleJWSSign)
	browserV1.Post("/jws/verify", h.handleJWSVerify)
	browserV1.Post("/jwe/encrypt", h.handleJWEEncrypt)
	browserV1.Post("/jwe/decrypt", h.handleJWEDecrypt)
	browserV1.Post("/jwt/sign", h.handleJWTSign)
	browserV1.Post("/jwt/verify", h.handleJWTVerify)

	// Elastic JWK JWKS endpoint for browser clients.
	browserV1.Get("/elastic-jwks/:kid/.well-known/jwks.json", h.handleElasticJWKS)

	// Admin API routes (browser clients only).
	// TODO: Add admin permission middleware.
	browserAdminV1 := app.Group("/browser/api/v1/admin")
	browserAdminV1.Use(rateLimiter)
	browserAdminV1.Get("/audit-config", auditH.handleGetAuditConfig)
	browserAdminV1.Get("/audit-config/:operation", auditH.handleGetAuditConfigByOperation)
	browserAdminV1.Put("/audit-config", auditH.handleSetAuditConfig)

	// Legacy API routes (for backward compatibility).
	legacyV1 := app.Group("/jose/v1")
	legacyV1.Post("/jwk/generate", h.handleJWKGenerate)
	legacyV1.Get("/jwk/:kid", h.handleJWKGet)
	legacyV1.Delete("/jwk/:kid/delete", h.handleJWKDelete)
	legacyV1.Get("/jwk", h.handleJWKList)
	legacyV1.Get("/jwks", h.handleJWKS)
	legacyV1.Post("/jws/sign", h.handleJWSSign)
	legacyV1.Post("/jws/verify", h.handleJWSVerify)
	legacyV1.Post("/jwe/encrypt", h.handleJWEEncrypt)
	legacyV1.Post("/jwe/decrypt", h.handleJWEDecrypt)
	legacyV1.Post("/jwt/sign", h.handleJWTSign)
	legacyV1.Post("/jwt/verify", h.handleJWTVerify)

	return nil
}

// Start begins serving both public and admin HTTPS endpoints.
// Blocks until context is cancelled or an unrecoverable error occurs.
func (s *JoseServer) Start(ctx context.Context) error {
	return s.app.Start(ctx)
}

// Shutdown gracefully shuts down all servers and closes database connections.
func (s *JoseServer) Shutdown(ctx context.Context) error {
	return s.app.Shutdown(ctx)
}

// DB returns the GORM database connection (for tests).
func (s *JoseServer) DB() *gorm.DB {
	return s.db
}

// App returns the application wrapper (for tests).
func (s *JoseServer) App() *cryptoutilAppsTemplateServiceServer.Application {
	return s.app
}

// JWKGen returns the JWK generation service (for tests).
func (s *JoseServer) JWKGen() *cryptoutilJose.JWKGenService {
	return s.jwkGenService
}

// Telemetry returns the telemetry service (for tests).
func (s *JoseServer) Telemetry() *cryptoutilTelemetry.TelemetryService {
	return s.telemetryService
}

// PublicPort returns the actual port the public server is listening on (for tests).
func (s *JoseServer) PublicPort() int {
	return s.app.PublicPort()
}

// AdminPort returns the actual port the admin server is listening on (for tests).
func (s *JoseServer) AdminPort() int {
	return s.app.AdminPort()
}

// SetReady marks the server as ready (enables /admin/api/v1/readyz to return 200 OK).
func (s *JoseServer) SetReady(ready bool) {
	s.app.SetReady(ready)
}

// PublicBaseURL returns the public server base URL (for tests).
func (s *JoseServer) PublicBaseURL() string {
	return s.app.PublicBaseURL()
}

// AdminBaseURL returns the admin server base URL (for tests).
func (s *JoseServer) AdminBaseURL() string {
	return s.app.AdminBaseURL()
}
