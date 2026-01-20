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

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTLSGenerator "cryptoutil/internal/apps/template/service/config/tls_generator"
	cryptoutilTemplateServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilTemplateAPIs "cryptoutil/internal/apps/template/service/server/apis"
	cryptoutilBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilTemplateBusinessLogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilTemplateServerListener "cryptoutil/internal/apps/template/service/server/listener"
	cryptoutilTemplateRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilTemplateService "cryptoutil/internal/apps/template/service/server/service"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

// ServiceResources contains all initialized service resources available to domain-specific code.
type ServiceResources struct {
	// Infrastructure.
	DB                  *gorm.DB
	TelemetryService    *cryptoutilTelemetry.TelemetryService
	JWKGenService       *cryptoutilJose.JWKGenService
	BarrierService      *cryptoutilBarrier.BarrierService
	SessionManager      *cryptoutilTemplateBusinessLogic.SessionManagerService
	RegistrationService *cryptoutilTemplateBusinessLogic.TenantRegistrationService
	RealmService        cryptoutilTemplateService.RealmService
	RealmRepository     cryptoutilTemplateRepository.TenantRealmRepository

	// Application wrapper.
	Application *cryptoutilTemplateServer.Application

	// Shutdown functions.
	ShutdownCore      func()
	ShutdownContainer func()
}

// ServerBuilder provides fluent API for constructing complete service applications.
// Handles ALL common initialization: TLS, admin/public servers, database, migrations, sessions, barrier.
type ServerBuilder struct {
	ctx                 context.Context
	config              *cryptoutilConfig.ServiceTemplateServerSettings
	migrationFS         fs.FS
	migrationsPath      string
	publicRouteRegister func(*cryptoutilTemplateServer.PublicServerBase, *ServiceResources) error
	err                 error // Accumulates errors during fluent chain.
}

// NewServerBuilder creates a new server builder with base configuration.
func NewServerBuilder(ctx context.Context, config *cryptoutilConfig.ServiceTemplateServerSettings) *ServerBuilder {
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

// WithPublicRouteRegistration provides callback for domain-specific route registration.
// Callback receives initialized PublicServerBase and ServiceResources for handler creation.
func (b *ServerBuilder) WithPublicRouteRegistration(registerFunc func(*cryptoutilTemplateServer.PublicServerBase, *ServiceResources) error) *ServerBuilder {
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
	adminServer, err := cryptoutilTemplateServerListener.NewAdminHTTPServer(b.ctx, b.config, adminTLSCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin server: %w", err)
	}

	// Start application core (telemetry, JWK gen, unseal, database).
	core, err := cryptoutilTemplateServer.StartApplicationCore(b.ctx, b.config)
	if err != nil {
		return nil, fmt.Errorf("failed to start application core: %w", err)
	}

	// Apply migrations (template + domain merged into single migration stream).
	sqlDB, err := core.DB.DB()
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to get sql.DB from GORM: %w", err)
	}

	if err := b.applyMigrations(sqlDB); err != nil {
		core.Shutdown()

		return nil, err
	}

	// Create barrier repository and service.
	barrierRepo, err := cryptoutilBarrier.NewGormBarrierRepository(core.DB)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to create barrier repository: %w", err)
	}

	barrierService, err := cryptoutilBarrier.NewBarrierService(
		b.ctx,
		core.Basic.TelemetryService,
		core.Basic.JWKGenService,
		barrierRepo,
		core.Basic.UnsealKeysService,
	)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to create barrier service: %w", err)
	}

	// Create realm repository and service.
	realmRepo := cryptoutilTemplateRepository.NewTenantRealmRepository(core.DB)
	realmService := cryptoutilTemplateService.NewRealmService(realmRepo)

	// Create session manager service.
	sessionManager, err := cryptoutilTemplateBusinessLogic.NewSessionManagerService(
		b.ctx,
		core.DB,
		core.Basic.TelemetryService,
		core.Basic.JWKGenService,
		barrierService,
		b.config,
	)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to create session manager service: %w", err)
	}

	// Create tenant registration service and dependencies.
	tenantRepo := cryptoutilTemplateRepository.NewTenantRepository(core.DB)
	userRepo := cryptoutilTemplateRepository.NewUserRepository(core.DB)
	joinRequestRepo := cryptoutilTemplateRepository.NewTenantJoinRequestRepository(core.DB)
	registrationService := cryptoutilTemplateBusinessLogic.NewTenantRegistrationService(
		core.DB,
		tenantRepo,
		userRepo,
		joinRequestRepo,
	)

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
		core.Shutdown()

		return nil, err
	}

	// Generate TLS material for public server.
	publicTLSMaterial, err := cryptoutilTLSGenerator.GenerateTLSMaterial(publicTLSCfg)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to generate public TLS material: %w", err)
	}

	// Create public server base.
	publicServerBase, err := cryptoutilTemplateServer.NewPublicServerBase(&cryptoutilTemplateServer.PublicServerConfig{
		BindAddress: b.config.BindPublicAddress,
		Port:        int(b.config.BindPublicPort),
		TLSMaterial: publicTLSMaterial,
	})
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to create public server base: %w", err)
	}

	// Prepare service resources for domain-specific initialization.
	resources := &ServiceResources{
		DB:                  core.DB,
		TelemetryService:    core.Basic.TelemetryService,
		JWKGenService:       core.Basic.JWKGenService,
		BarrierService:      barrierService,
		SessionManager:      sessionManager,
		RegistrationService: registrationService,
		RealmService:        realmService,
		RealmRepository:     realmRepo,
		ShutdownCore:        core.Shutdown,
		ShutdownContainer:   core.ShutdownDBContainer,
	}

	// Register domain-specific public routes if provided.
	if b.publicRouteRegister != nil {
		if err := b.publicRouteRegister(publicServerBase, resources); err != nil {
			core.Shutdown()

			return nil, fmt.Errorf("failed to register public routes: %w", err)
		}
	}

	// Register tenant registration routes on PUBLIC server (unauthenticated user registration).
	cryptoutilTemplateAPIs.RegisterRegistrationRoutes(publicServerBase.App(), registrationService)

	// Register join request management routes on ADMIN server (authenticated admin operations).
	cryptoutilTemplateAPIs.RegisterJoinRequestManagementRoutes(adminServer.App(), registrationService)

	// Register barrier admin endpoints (key rotation, status) on ADMIN server.
	rotationService, err := cryptoutilBarrier.NewRotationService(
		core.Basic.JWKGenService,
		barrierRepo,
		core.Basic.UnsealKeysService,
	)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to create rotation service: %w", err)
	}

	cryptoutilBarrier.RegisterRotationRoutes(adminServer.App(), rotationService)

	statusService, err := cryptoutilBarrier.NewStatusService(barrierRepo)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to create status service: %w", err)
	}

	cryptoutilBarrier.RegisterStatusRoutes(adminServer.App(), statusService)

	// Create application wrapper with both servers.
	app, err := cryptoutilTemplateServer.NewApplication(b.ctx, publicServerBase, adminServer)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	resources.Application = app

	return resources, nil
}

// generateTLSConfig handles TLS configuration generation for admin or public server.
// Supports three modes: static (pre-provided certs), mixed (generate from CA), auto (fully auto-generate).
func (b *ServerBuilder) generateTLSConfig(
	mode cryptoutilConfig.TLSMode,
	staticCertPEM []byte,
	staticKeyPEM []byte,
	mixedCACertPEM []byte,
	mixedCAKeyPEM []byte,
	dnsNames []string,
	ipAddresses []string,
	serverType string,
) (*cryptoutilTLSGenerator.TLSGeneratedSettings, error) {
	// Default to auto mode if not specified.
	if mode == "" {
		mode = cryptoutilConfig.TLSModeAuto
	}

	switch mode {
	case cryptoutilConfig.TLSModeStatic:
		return &cryptoutilTLSGenerator.TLSGeneratedSettings{
			StaticCertPEM: staticCertPEM,
			StaticKeyPEM:  staticKeyPEM,
		}, nil

	case cryptoutilConfig.TLSModeMixed:
		tlsCfg, err := cryptoutilTLSGenerator.GenerateServerCertFromCA(
			mixedCACertPEM,
			mixedCAKeyPEM,
			dnsNames,
			ipAddresses,
			cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to generate %s TLS config (mixed mode): %w", serverType, err)
		}

		return tlsCfg, nil

	case cryptoutilConfig.TLSModeAuto:
		tlsCfg, err := cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(
			dnsNames,
			ipAddresses,
			cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
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
	if b.config.DatabaseURL == "" || b.config.DatabaseURL == "file::memory:?cache=shared" || b.config.DatabaseURL == ":memory:" || (len(b.config.DatabaseURL) >= 7 && b.config.DatabaseURL[:7] == "file://") {
		databaseType = "sqlite"
	} else {
		databaseType = "postgres"
	}

	// Merge template migrations with domain migrations (if provided).
	var migrationsFS fs.FS = cryptoutilTemplateRepository.MigrationsFS

	migrationsPath := "migrations"

	if b.migrationFS != nil {
		// Create merged FS combining template + domain migrations.
		migrationsFS = &mergedMigrations{
			templateFS:   cryptoutilTemplateRepository.MigrationsFS,
			templatePath: "migrations",
			domainFS:     b.migrationFS,
			domainPath:   b.migrationsPath,
		}
		migrationsPath = "" // Root of merged FS
	}

	// Apply migrations using merged FS.
	if err := cryptoutilTemplateRepository.ApplyMigrationsFromFS(sqlDB, migrationsFS, migrationsPath, databaseType); err != nil {
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

	if f, err := m.templateFS.Open(fullPath); err != nil {
		return nil, fmt.Errorf("failed to open migration file %s: %w", name, err)
	} else {
		return f, nil
	}
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
