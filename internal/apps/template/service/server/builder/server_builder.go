// Copyright (c) 2025 Justin Cranford
//

// Package builder provides fluent API for constructing service applications.
// Eliminates 260+ lines of boilerplate server initialization per service.
package builder

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceConfigTlsGenerator "cryptoutil/internal/apps/template/service/config/tls_generator"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilAppsTemplateServiceServerApis "cryptoutil/internal/apps/template/service/server/apis"
	cryptoutilAppsTemplateServiceServerApplication "cryptoutil/internal/apps/template/service/server/application"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerListener "cryptoutil/internal/apps/template/service/server/listener"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilAppsTemplateServiceServerService "cryptoutil/internal/apps/template/service/server/service"
	cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
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
func (b *ServerBuilder) Build() (*ServiceResources, error) {
	if b.err != nil {
		return nil, b.err
	}

	// Generate admin TLS configuration.
	adminTLSCfg, err := b.generateTLSConfig(
		b.config.TLSPrivateMode,
		b.config.TLSStaticCertPEM,
		b.config.TLSStaticKeyPEM,
		b.config.TLSMixedCACertPEM,
		b.config.TLSMixedCAKeyPEM,
		b.config.TLSPrivateDNSNames,
		b.config.TLSPrivateIPAddresses,
		"admin",
	)
	if err != nil {
		return nil, err
	}

	// Create admin server.
	adminServer, err := cryptoutilAppsTemplateServiceServerListener.NewAdminHTTPServer(b.ctx, b.config, adminTLSCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin server: %w", err)
	}

	// Phase W.1: Initialize application core WITHOUT services (DB + telemetry only).
	// CRITICAL: Must run migrations BEFORE initializing services (BarrierService needs barrier_root_keys table).
	applicationCore, err := cryptoutilAppsTemplateServiceServerApplication.StartCore(b.ctx, b.config)
	if err != nil {
		return nil, fmt.Errorf("failed to start application core: %w", err)
	}

	// Phase W.2: Apply migrations (always enabled - GORM database is mandatory).
	// Migration mode determines which migrations to apply (TemplateWithDomain or DomainOnly).
	{
		sqlDB, err := applicationCore.DB.DB()
		if err != nil {
			applicationCore.Shutdown()

			return nil, fmt.Errorf("failed to get sql.DB from GORM: %w", err)
		}

		if err := b.applyMigrations(sqlDB); err != nil {
			applicationCore.Shutdown()

			return nil, err
		}
	}

	// Phase W.3: Initialize services - barrier is always enabled.
	// The optional BarrierConfig only controls endpoint exposure (rotation, status).
	barrierEnabled := true

	var services *cryptoutilAppsTemplateServiceServerApplication.CoreWithServices

	if barrierEnabled {
		services, err = cryptoutilAppsTemplateServiceServerApplication.InitializeServicesOnCore(
			b.ctx,
			applicationCore,
			b.config,
		)
		if err != nil {
			applicationCore.Shutdown()

			return nil, fmt.Errorf("failed to initialize services on core: %w", err)
		}
	}

	// Generate public TLS configuration.
	publicTLSCfg, err := b.generateTLSConfig(
		b.config.TLSPublicMode,
		b.config.TLSStaticCertPEM,
		b.config.TLSStaticKeyPEM,
		b.config.TLSMixedCACertPEM,
		b.config.TLSMixedCAKeyPEM,
		b.config.TLSPublicDNSNames,
		b.config.TLSPublicIPAddresses,
		"public",
	)
	if err != nil {
		applicationCore.Shutdown()

		return nil, err
	}

	// Generate TLS material for public server.
	publicTLSMaterial, err := cryptoutilAppsTemplateServiceConfigTlsGenerator.GenerateTLSMaterial(publicTLSCfg)
	if err != nil {
		applicationCore.Shutdown()

		return nil, fmt.Errorf("failed to generate public TLS material: %w", err)
	}

	// Create public server base.
	publicServerBase, err := cryptoutilAppsTemplateServiceServer.NewPublicServerBase(&cryptoutilAppsTemplateServiceServer.PublicServerConfig{
		BindAddress: b.config.BindPublicAddress,
		Port:        int(b.config.BindPublicPort),
		TLSMaterial: publicTLSMaterial,
	})
	if err != nil {
		applicationCore.Shutdown()

		return nil, fmt.Errorf("failed to create public server base: %w", err)
	}

	// Create DatabaseConnection wrapper.
	dbConn, err := NewDatabaseConnection(applicationCore.DB)
	if err != nil {
		applicationCore.Shutdown()

		return nil, fmt.Errorf("failed to create database connection wrapper: %w", err)
	}

	// Prepare service resources for domain-specific initialization.
	// When barrier is disabled, service-specific fields (BarrierService, SessionManager, etc.) are nil.
	// Domain services (like KMS) provide their own implementations via route registration callback.
	resources := &ServiceResources{
		DB:                 applicationCore.DB,
		DatabaseConnection: dbConn,
		TelemetryService:   applicationCore.Basic.TelemetryService,
		JWKGenService:      applicationCore.Basic.JWKGenService,
		UnsealKeysService:  applicationCore.Basic.UnsealKeysService,
		JWTAuthConfig:      b.jwtAuthConfig,
		StrictServerConfig: b.strictServerConfig,
		BarrierConfig:      b.barrierConfig,
		MigrationConfig:    b.migrationConfig,
		ShutdownCore:       applicationCore.Shutdown,
		ShutdownContainer:  applicationCore.ShutdownDBContainer,
	}

	// If barrier is enabled, populate service resources from initialized services.
	if services != nil {
		resources.BarrierService = services.BarrierService
		resources.SessionManager = services.SessionManager
		resources.RegistrationService = services.RegistrationService
		resources.RealmService = services.RealmService
		resources.RealmRepository = services.RealmRepository
	}

	// Register domain-specific public routes if provided.
	if b.publicRouteRegister != nil {
		if err := b.publicRouteRegister(publicServerBase, resources); err != nil {
			applicationCore.Shutdown()

			return nil, fmt.Errorf("failed to register public routes: %w", err)
		}
	}

	// Register Swagger UI if configured.
	if b.swaggerUIConfig != nil {
		if err := RegisterSwaggerUI(publicServerBase.App(), b.swaggerUIConfig); err != nil {
			applicationCore.Shutdown()

			return nil, fmt.Errorf("failed to register swagger UI: %w", err)
		}
	}

	// Register template-specific routes ONLY if barrier/services are enabled.
	// KMS and other services with disabled barrier handle their own routes.
	if services != nil {
		// Register tenant registration routes on PUBLIC server (unauthenticated user registration).
		// Default rate limit configured via magic constant.
		cryptoutilAppsTemplateServiceServerApis.RegisterRegistrationRoutes(publicServerBase.App(), services.RegistrationService, cryptoutilSharedMagic.RateLimitDefaultRequestsPerMin)

		// Register join request management routes on ADMIN server (authenticated admin operations).
		// SessionManager implements SessionValidator interface for session validation.
		cryptoutilAppsTemplateServiceServerApis.RegisterJoinRequestManagementRoutes(adminServer.App(), services.RegistrationService, services.SessionManager)

		// Register barrier admin endpoints (key rotation, status) on ADMIN server.
		cryptoutilAppsTemplateServiceServerBarrier.RegisterRotationRoutes(adminServer.App(), services.RotationService)
		cryptoutilAppsTemplateServiceServerBarrier.RegisterStatusRoutes(adminServer.App(), services.StatusService)
	}

	// Create application wrapper with both servers.
	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(b.ctx, publicServerBase, adminServer)
	if err != nil {
		applicationCore.Shutdown()

		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	resources.Application = app

	return resources, nil
}

// generateTLSConfig handles TLS configuration generation for admin or public server.
// Supports three modes: static (pre-provided certs), mixed (generate from CA), auto (fully auto-generate).
func (b *ServerBuilder) generateTLSConfig(
	mode cryptoutilAppsTemplateServiceConfig.TLSMode,
	staticCertPEM []byte,
	staticKeyPEM []byte,
	mixedCACertPEM []byte,
	mixedCAKeyPEM []byte,
	dnsNames []string,
	ipAddresses []string,
	serverType string,
) (*cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings, error) {
	// Default to auto mode if not specified.
	if mode == "" {
		mode = cryptoutilAppsTemplateServiceConfig.TLSModeAuto
	}

	switch mode {
	case cryptoutilAppsTemplateServiceConfig.TLSModeStatic:
		return &cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings{
			StaticCertPEM: staticCertPEM,
			StaticKeyPEM:  staticKeyPEM,
		}, nil

	case cryptoutilAppsTemplateServiceConfig.TLSModeMixed:
		tlsCfg, err := cryptoutilAppsTemplateServiceConfigTlsGenerator.GenerateServerCertFromCA(
			mixedCACertPEM,
			mixedCAKeyPEM,
			dnsNames,
			ipAddresses,
			cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to generate %s TLS config (mixed mode): %w", serverType, err)
		}

		return tlsCfg, nil

	case cryptoutilAppsTemplateServiceConfig.TLSModeAuto:
		tlsCfg, err := cryptoutilAppsTemplateServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings(
			dnsNames,
			ipAddresses,
			cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to generate %s TLS config (auto mode): %w", serverType, err)
		}

		return tlsCfg, nil

	default:
		return nil, fmt.Errorf("unsupported TLS %s mode: %s", serverType, mode)
	}
}

// applyMigrations runs template + domain migrations as a merged stream.
func (b *ServerBuilder) applyMigrations(sqlDB *sql.DB) error {
	// Determine database type from URL.
	var databaseType string
	if b.config.DatabaseURL == "" ||
		b.config.DatabaseURL == "file::memory:?cache=shared" ||
		b.config.DatabaseURL == ":memory:" ||
		(len(b.config.DatabaseURL) >= 7 && b.config.DatabaseURL[:7] == "file://") ||
		(len(b.config.DatabaseURL) >= 9 && b.config.DatabaseURL[:9] == "sqlite://") {
		databaseType = "sqlite"
	} else {
		databaseType = "postgres"
	}

	// Merge template migrations with domain migrations (if provided).
	var migrationsFS fs.FS = cryptoutilAppsTemplateServiceServerRepository.MigrationsFS

	migrationsPath := "migrations"

	if b.migrationFS != nil {
		// Create merged FS combining template + domain migrations.
		migrationsFS = &mergedMigrations{
			templateFS:   cryptoutilAppsTemplateServiceServerRepository.MigrationsFS,
			templatePath: "migrations",
			domainFS:     b.migrationFS,
			domainPath:   b.migrationsPath,
		}
		migrationsPath = "" // Root of merged FS
	}

	// Apply migrations using merged FS.
	if err := cryptoutilAppsTemplateServiceServerRepository.ApplyMigrationsFromFS(sqlDB, migrationsFS, migrationsPath, databaseType); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

// mergedMigrations combines template and domain migrations into a single filesystem view.
type mergedMigrations struct {
	templateFS   fs.FS
	templatePath string
	domainFS     fs.FS
	domainPath   string
}

const (
	currentDir = "."
	pathSep    = "/"
)

func (m *mergedMigrations) Open(name string) (fs.File, error) {
	// Try domain migrations first (they have higher version numbers).
	if m.domainFS != nil {
		fullPath := m.domainPath
		if name != currentDir && name != "" {
			fullPath = m.domainPath + pathSep + name
		}

		if f, err := m.domainFS.Open(fullPath); err == nil {
			return f, nil
		}
	}

	// Fall back to template migrations.
	fullPath := m.templatePath
	if name != currentDir && name != "" {
		fullPath = m.templatePath + pathSep + name
	}

	f, err := m.templateFS.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open migration file %s: %w", name, err)
	}

	return f, nil
}

func (m *mergedMigrations) ReadDir(name string) ([]fs.DirEntry, error) {
	var entries []fs.DirEntry

	// Read template migrations.
	templatePath := m.templatePath
	if name != "." && name != "" {
		templatePath = m.templatePath + "/" + name
	}

	if templateEntries, err := fs.ReadDir(m.templateFS, templatePath); err == nil {
		entries = append(entries, templateEntries...)
	}

	// Read domain migrations.
	if m.domainFS != nil {
		domainPath := m.domainPath
		if name != currentDir && name != "" {
			domainPath = m.domainPath + pathSep + name
		}

		if domainEntries, err := fs.ReadDir(m.domainFS, domainPath); err == nil {
			entries = append(entries, domainEntries...)
		}
	}

	return entries, nil
}

func (m *mergedMigrations) ReadFile(name string) ([]byte, error) {
	// Try domain migrations first.
	if m.domainFS != nil {
		fullPath := m.domainPath + pathSep + name

		if data, err := fs.ReadFile(m.domainFS, fullPath); err == nil {
			return data, nil
		}
	}

	// Fall back to template migrations.
	fullPath := m.templatePath + pathSep + name

	data, err := fs.ReadFile(m.templateFS, fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration file %s: %w", name, err)
	}

	return data, nil
}

func (m *mergedMigrations) Stat(name string) (fs.FileInfo, error) {
	// Try domain migrations first.
	if m.domainFS != nil {
		fullPath := m.domainPath + pathSep + name

		if info, err := fs.Stat(m.domainFS, fullPath); err == nil {
			return info, nil
		}
	}

	// Fall back to template migrations.
	fullPath := m.templatePath + pathSep + name

	info, err := fs.Stat(m.templateFS, fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat migration file %s: %w", name, err)
	}

	return info, nil
}
