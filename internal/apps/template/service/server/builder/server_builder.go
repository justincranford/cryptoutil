// Copyright (c) 2025 Justin Cranford
//

// Package builder provides fluent API for constructing service applications.
// Eliminates 48,000+ lines of boilerplate server initialization per service.
package builder

import (
	"context"
	"fmt"
	"io/fs"

	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps/template/service/server/barrier/unsealkeysservice"
	cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilAppsTemplateServiceServerService "cryptoutil/internal/apps/template/service/server/service"
	cryptoutilSharedTelemetry "cryptoutil/internal/apps/template/service/telemetry"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
)

// ServiceResources contains all initialized service resources available to domain-specific code.
type ServiceResources struct {
	// Infrastructure.
	DB                  *gorm.DB
	DatabaseConnection  *DatabaseConnection // Multi-mode database access (GORM, raw SQL, or both).
	TelemetryService    *cryptoutilSharedTelemetry.TelemetryService
	JWKGenService       *cryptoutilSharedCryptoJose.JWKGenService
	UnsealKeysService   cryptoutilUnsealKeysService.UnsealKeysService
	BarrierService      *cryptoutilAppsTemplateServiceServerBarrier.Service
	SessionManager      *cryptoutilAppsTemplateServiceServerBusinesslogic.SessionManagerService
	RegistrationService *cryptoutilAppsTemplateServiceServerBusinesslogic.TenantRegistrationService
	RealmService        cryptoutilAppsTemplateServiceServerService.RealmService
	RealmRepository     cryptoutilAppsTemplateServiceServerRepository.TenantRealmRepository
	JWTAuthConfig       *JWTAuthConfig      // JWT authentication config (nil = session-based auth).
	StrictServerConfig  *StrictServerConfig // OpenAPI strict server config (nil = not registered).
	BarrierConfig       *BarrierConfig      // Barrier config (nil = default template barrier).
	MigrationConfig     *MigrationConfig    // Migration config (nil = default template+domain).

	// Application wrapper.
	Application *cryptoutilAppsTemplateServiceServer.Application

	// Shutdown functions.
	ShutdownCore      func()
	ShutdownContainer func()
}

// ServerBuilder provides fluent API for constructing complete service applications.
// Handles ALL common initialization: TLS, admin/public servers, database, migrations, sessions, barrier.
type ServerBuilder struct {
	ctx                 context.Context
	config              *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
	migrationFS         fs.FS
	migrationsPath      string
	migrationConfig     *MigrationConfig // Migration configuration (nil = use legacy WithDomainMigrations).
	publicRouteRegister func(*cryptoutilAppsTemplateServiceServer.PublicServerBase, *ServiceResources) error
	swaggerUIConfig     *SwaggerUIConfig    // Swagger UI configuration (nil = disabled).
	jwtAuthConfig       *JWTAuthConfig      // JWT authentication configuration (nil = use session-based auth).
	strictServerConfig  *StrictServerConfig // OpenAPI strict server configuration (nil = not registered).
	barrierConfig       *BarrierConfig      // Barrier configuration (nil = use default template barrier).
	err                 error               // Accumulates errors during fluent chain.
}

// NewServerBuilder creates a new server builder with base configuration.
func NewServerBuilder(ctx context.Context, config *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) *ServerBuilder {
	if ctx == nil {
		return &ServerBuilder{err: fmt.Errorf("context cannot be nil")}
	} else if config == nil {
		return &ServerBuilder{err: fmt.Errorf("config cannot be nil")}
	}

	return &ServerBuilder{
		ctx:    ctx,
		config: config,
	}
}

// WithDomainMigrations registers domain-specific migrations (e.g., message tables, topic tables).
// migrationFS should be embed.FS containing *.up.sql and *.down.sql files.
// migrationsPath is the path within the FS (e.g., "migrations").
func (b *ServerBuilder) WithDomainMigrations(migrationFS fs.FS, migrationsPath string) *ServerBuilder {
	if b.err != nil {
		return b
	}

	if migrationFS == nil {
		b.err = fmt.Errorf("migration FS cannot be nil")

		return b
	}

	if migrationsPath == "" {
		b.err = fmt.Errorf("migrations path cannot be empty")

		return b
	}

	b.migrationFS = migrationFS
	b.migrationsPath = migrationsPath

	return b
}

// WithMigrationConfig configures migration handling with full flexibility.
// Use this instead of WithDomainMigrations() for services that need custom migration schemes.
// - NewDefaultMigrationConfig(): Template + domain migrations (default behavior)
// - NewDomainOnlyMigrationConfig(): Only domain migrations (KMS-style).
func (b *ServerBuilder) WithMigrationConfig(config *MigrationConfig) *ServerBuilder {
	if b.err != nil {
		return b
	}

	if config == nil {
		return b
	}

	if err := config.Validate(); err != nil {
		b.err = fmt.Errorf("invalid migration config: %w", err)

		return b
	}

	b.migrationConfig = config

	// Also set legacy fields for backward compatibility with applyMigrations().
	if config.DomainFS != nil {
		b.migrationFS = config.DomainFS
		b.migrationsPath = config.DomainPath
	}

	return b
}

// WithPublicRouteRegistration provides callback for domain-specific route registration.
// Callback receives initialized PublicServerBase and ServiceResources for handler creation.
func (b *ServerBuilder) WithPublicRouteRegistration(registerFunc func(*cryptoutilAppsTemplateServiceServer.PublicServerBase, *ServiceResources) error) *ServerBuilder {
	if b.err != nil {
		return b
	}

	if registerFunc == nil {
		b.err = fmt.Errorf("route registration function cannot be nil")

		return b
	}

	b.publicRouteRegister = registerFunc

	return b
}

// WithJWTAuth configures JWT authentication for the service.
// If not called, the service defaults to session-based authentication.
// Use NewKMSJWTAuthConfig() for KMS-style JWT authentication.
func (b *ServerBuilder) WithJWTAuth(config *JWTAuthConfig) *ServerBuilder {
	if b.err != nil {
		return b
	}

	if config == nil {
		b.err = fmt.Errorf("JWT auth config cannot be nil")

		return b
	}

	if err := config.Validate(); err != nil {
		b.err = fmt.Errorf("invalid JWT auth config: %w", err)

		return b
	}

	b.jwtAuthConfig = config

	return b
}

// WithStrictServer configures the OpenAPI strict server for handler registration.
// Domain services implement StrictServerInterface and provide registration functions
// that call the generated RegisterHandlersWithOptions().
func (b *ServerBuilder) WithStrictServer(config *StrictServerConfig) *ServerBuilder {
	if b.err != nil {
		return b
	}

	if config == nil {
		return b
	}

	if err := config.Validate(); err != nil {
		b.err = fmt.Errorf("invalid strict server config: %w", err)

		return b
	}

	b.strictServerConfig = config

	return b
}

// WithBarrierConfig configures the barrier service.
// If not called, the default barrier config (rotation + status endpoints enabled) is used.
func (b *ServerBuilder) WithBarrierConfig(config *BarrierConfig) *ServerBuilder {
	if b.err != nil {
		return b
	}

	if config == nil {
		return b
	}

	if err := config.Validate(); err != nil {
		b.err = fmt.Errorf("invalid barrier config: %w", err)

		return b
	}

	b.barrierConfig = config

	return b
}

// WithSwaggerUI enables Swagger UI with optional HTTP Basic Authentication.
// If username and password are both empty, Swagger UI is accessible without authentication.
// The openAPISpecJSON should be the serialized OpenAPI specification from oapi-codegen's GetSwagger().
func (b *ServerBuilder) WithSwaggerUI(username, password string, openAPISpecJSON []byte) *ServerBuilder {
	if b.err != nil {
		return b
	}

	if len(openAPISpecJSON) == 0 {
		b.err = fmt.Errorf("OpenAPI spec JSON cannot be empty")

		return b
	}

	b.swaggerUIConfig = &SwaggerUIConfig{
		Username:              username,
		Password:              password,
		CSRFTokenName:         b.config.CSRFTokenName,
		BrowserAPIContextPath: b.config.PublicBrowserAPIContextPath,
		OpenAPISpecJSON:       openAPISpecJSON,
	}

	return b
}

// Build constructs the complete service application.
// Returns ServiceResources containing all initialized infrastructure and application wrapper.
