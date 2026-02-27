// Copyright (c) 2025 Justin Cranford
//

// Package server implements the skeleton-template HTTPS server using the service template.
package server

import (
"context"
"fmt"

"gorm.io/gorm"

cryptoutilAppsSkeletonTemplateRepository "cryptoutil/internal/apps/skeleton/template/repository"
cryptoutilAppsSkeletonTemplateServerConfig "cryptoutil/internal/apps/skeleton/template/server/config"
cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
cryptoutilAppsTemplateServiceServerBuilder "cryptoutil/internal/apps/template/service/server/builder"
cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
cryptoutilAppsTemplateServiceServerService "cryptoutil/internal/apps/template/service/server/service"
cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// SkeletonTemplateServer represents the skeleton-template service application.
type SkeletonTemplateServer struct {
app *cryptoutilAppsTemplateServiceServer.Application
db  *gorm.DB

// Services.
telemetryService      *cryptoutilSharedTelemetry.TelemetryService
jwkGenService         *cryptoutilSharedCryptoJose.JWKGenService
barrierService        *cryptoutilAppsTemplateServiceServerBarrier.Service
sessionManagerService *cryptoutilAppsTemplateServiceServerBusinesslogic.SessionManagerService
realmService          cryptoutilAppsTemplateServiceServerService.RealmService

// Repositories.
realmRepo cryptoutilAppsTemplateServiceServerRepository.TenantRealmRepository // Uses service-template repository.
}

// NewFromConfig creates a new skeleton-template server from SkeletonTemplateServerSettings only.
// Uses service-template builder for infrastructure initialization.
func NewFromConfig(ctx context.Context, cfg *cryptoutilAppsSkeletonTemplateServerConfig.SkeletonTemplateServerSettings) (*SkeletonTemplateServer, error) {
if ctx == nil {
return nil, fmt.Errorf("context cannot be nil")
} else if cfg == nil {
return nil, fmt.Errorf("config cannot be nil")
}

// Create server builder with template config.
builder := cryptoutilAppsTemplateServiceServerBuilder.NewServerBuilder(ctx, cfg.ServiceTemplateServerSettings)

// Register skeleton-template specific migrations.
builder.WithDomainMigrations(cryptoutilAppsSkeletonTemplateRepository.MigrationsFS, "migrations")

// Register skeleton-template specific public routes.
// The skeleton-template has no domain-specific business routes — only health endpoints from the template.
builder.WithPublicRouteRegistration(func(
_ *cryptoutilAppsTemplateServiceServer.PublicServerBase,
_ *cryptoutilAppsTemplateServiceServerBuilder.ServiceResources,
) error {
// No domain-specific routes for skeleton-template.
// Health endpoints (/browser/api/v1/health, /service/api/v1/health) are registered by the template.
return nil
})

// Build complete service infrastructure.
resources, err := builder.Build()
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
func (s *SkeletonTemplateServer) App() *cryptoutilAppsTemplateServiceServer.Application {
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
func (s *SkeletonTemplateServer) Barrier() *cryptoutilAppsTemplateServiceServerBarrier.Service {
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
