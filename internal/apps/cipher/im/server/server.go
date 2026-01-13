// Copyright (c) 2025 Justin Cranford
//
//

// Package server implements the cipher-im HTTPS server using the service template.
package server

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"cryptoutil/internal/apps/cipher/im/repository"
	"cryptoutil/internal/apps/cipher/im/server/businesslogic"
	"cryptoutil/internal/apps/cipher/im/server/config"
	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	tlsGenerator "cryptoutil/internal/apps/template/service/config/tls_generator"
	cryptoutilTemplateServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilTemplateBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilTemplateServerListener "cryptoutil/internal/apps/template/service/server/listener"
	cryptoutilTemplateRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

// CipherIMServer represents the cipher-im service application.
type CipherIMServer struct {
	app *cryptoutilTemplateServer.Application
	db  *gorm.DB

	// Services.
	telemetryService      *cryptoutilTelemetry.TelemetryService
	jwkGenService         *cryptoutilJose.JWKGenService
	barrierService        *cryptoutilTemplateBarrier.BarrierService
	sessionManagerService *businesslogic.SessionManagerService
	realmService          businesslogic.RealmService

	// Repositories.
	userRepo    *repository.UserRepository
	messageRepo *repository.MessageRepository
	realmRepo   cryptoutilTemplateRepository.TenantRealmRepository // Uses service-template repository.
}

// NewFromConfig creates a new cipher-im server from CipherImServerSettings only.
// This is the PREFERRED constructor - automatically provisions database (SQLite, PostgreSQL testcontainer, or external).
// Uses service-template application layer for infrastructure management.
func NewFromConfig(ctx context.Context, cfg *config.CipherImServerSettings) (*CipherIMServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Generate TLS configuration for admin server based on config settings.
	// Uses TLSPrivateMode (static, mixed, auto) from ServiceTemplateServerSettings.
	// For demo/dev environments, typically uses TLSModeAuto (auto-generated certificates).
	var adminTLSCfg *tlsGenerator.TLSGeneratedSettings

	var err error

	// Use TLSPrivateMode to determine how to generate admin TLS configuration.
	// Empty string defaults to auto mode for backward compatibility.
	tlsPrivateMode := cfg.TLSPrivateMode
	if tlsPrivateMode == "" {
		tlsPrivateMode = cryptoutilConfig.TLSModeAuto
	}

	switch tlsPrivateMode {
	case cryptoutilConfig.TLSModeStatic:
		// Static mode: Use pre-provided certificates from config.
		adminTLSCfg = &tlsGenerator.TLSGeneratedSettings{
			StaticCertPEM: cfg.TLSStaticCertPEM,
			StaticKeyPEM:  cfg.TLSStaticKeyPEM,
		}
	case cryptoutilConfig.TLSModeMixed:
		// Mixed mode: Generate server cert from CA.
		adminTLSCfg, err = tlsGenerator.GenerateServerCertFromCA(
			cfg.TLSMixedCACertPEM,
			cfg.TLSMixedCAKeyPEM,
			cfg.TLSPrivateDNSNames,
			cfg.TLSPrivateIPAddresses,
			cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to generate admin TLS config (mixed mode): %w", err)
		}
	case cryptoutilConfig.TLSModeAuto:
		// Auto mode: Fully auto-generate CA hierarchy and server certificate.
		adminTLSCfg, err = tlsGenerator.GenerateAutoTLSGeneratedSettings(
			cfg.TLSPrivateDNSNames,
			cfg.TLSPrivateIPAddresses,
			cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to generate admin TLS config (auto mode): %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported TLS private mode: %s", tlsPrivateMode)
	}

	adminServer, err := cryptoutilTemplateServerListener.NewAdminHTTPServer(ctx, cfg.ServiceTemplateServerSettings, adminTLSCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin server: %w", err)
	}

	// Start application core (telemetry, JWK gen, unseal, database).
	// This automatically provisions database based on cfg.DatabaseURL and cfg.DatabaseContainer.
	core, err := cryptoutilTemplateServer.StartApplicationCore(ctx, cfg.ServiceTemplateServerSettings)
	if err != nil {
		return nil, fmt.Errorf("failed to start application core: %w", err)
	}

	// Determine database type from core.
	var dbType repository.DatabaseType

	// GORM doesn't expose dialect name directly, so check DatabaseURL instead.
	if cfg.DatabaseURL == "" || cfg.DatabaseURL == "file::memory:?cache=shared" || cfg.DatabaseURL == ":memory:" || (len(cfg.DatabaseURL) >= 7 && cfg.DatabaseURL[:7] == "file://") {
		dbType = repository.DatabaseTypeSQLite
	} else {
		dbType = repository.DatabaseTypePostgreSQL
	}

	// Apply database migrations.
	sqlDB, err := core.DB.DB()
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to get sql.DB from GORM: %w", err)
	}

	err = repository.ApplyCipherIMMigrations(sqlDB, dbType)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	// Create GORM barrier repository adapter.
	barrierRepo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(core.DB)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to create barrier repository: %w", err)
	}

	// Drop stale barrier tables completely before AutoMigrate creates clean schema.
	// GORM AutoMigrate doesn't drop removed columns, and GORM Migrator.DropColumn
	// generates invalid SQL for SQLite (comments cause "incomplete input" errors).
	// Dropping tables ensures clean schema matching current struct definitions.
	if err := core.DB.Migrator().DropTable(
		&cryptoutilTemplateBarrier.BarrierRootKey{},
		&cryptoutilTemplateBarrier.BarrierIntermediateKey{},
		&cryptoutilTemplateBarrier.BarrierContentKey{},
	); err != nil {
		// Ignore error if tables don't exist (first run).
		core.Basic.TelemetryService.Slogger.Info("dropping barrier tables (if exist)", "error", err)
	}

	// Now AutoMigrate barrier tables with clean schema.
	if err := core.DB.AutoMigrate(
		&cryptoutilTemplateBarrier.BarrierRootKey{},
		&cryptoutilTemplateBarrier.BarrierIntermediateKey{},
		&cryptoutilTemplateBarrier.BarrierContentKey{},
	); err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to migrate barrier tables: %w", err)
	}

	// Drop stale session JWK tables completely before AutoMigrate creates clean schema.
	// PostgreSQL strict typing rejects boolean for int4 columns (active column).
	// SQLite type affinity masks this issue. Drop tables to ensure clean schema.
	if err := core.DB.Migrator().DropTable(
		&cryptoutilTemplateRepository.BrowserSessionJWK{},
		&cryptoutilTemplateRepository.ServiceSessionJWK{},
	); err != nil {
		// Ignore error if tables don't exist (first run).
		core.Basic.TelemetryService.Slogger.Info("dropping session JWK tables (if exist)", "error", err)
	}

	// Now AutoMigrate session JWK tables with clean schema.
	if err := core.DB.AutoMigrate(
		&cryptoutilTemplateRepository.BrowserSessionJWK{},
		&cryptoutilTemplateRepository.ServiceSessionJWK{},
	); err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to migrate session JWK tables: %w", err)
	}

	// Create barrier service with GORM repository.
	// For cipher-im demo service, unseal keys are already initialized in ApplicationBasic.
	barrierService, err := cryptoutilTemplateBarrier.NewBarrierService(
		ctx,
		core.Basic.TelemetryService,
		core.Basic.JWKGenService,
		barrierRepo,
		core.Basic.UnsealKeysService,
	)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to create barrier service: %w", err)
	}

	// Initialize repositories.
	userRepo := repository.NewUserRepository(core.DB)
	messageRepo := repository.NewMessageRepository(core.DB)
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(core.DB, barrierService)

	// Initialize realm repository and service (using service-template implementation).
	realmRepo := cryptoutilTemplateRepository.NewTenantRealmRepository(core.DB)
	realmService := businesslogic.NewRealmService(realmRepo)

	// Initialize SessionManager service for session management.
	sessionManagerService, err := businesslogic.NewSessionManagerService(
		ctx,
		core.DB,
		core.Basic.TelemetryService,
		core.Basic.JWKGenService,
		barrierService,
		cfg.ServiceTemplateServerSettings,
	)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to create session manager service: %w", err)
	}

	// Generate TLS configuration for public server based on config settings.
	// Uses TLSPublicMode (static, mixed, auto) from ServiceTemplateServerSettings.
	var publicTLSCfg *tlsGenerator.TLSGeneratedSettings

	// Use TLSPublicMode to determine how to generate public TLS configuration.
	// Empty string defaults to auto mode for backward compatibility.
	tlsPublicMode := cfg.TLSPublicMode
	if tlsPublicMode == "" {
		tlsPublicMode = cryptoutilConfig.TLSModeAuto
	}

	switch tlsPublicMode {
	case cryptoutilConfig.TLSModeStatic:
		// Static mode: Use pre-provided certificates from config.
		publicTLSCfg = &tlsGenerator.TLSGeneratedSettings{
			StaticCertPEM: cfg.TLSStaticCertPEM,
			StaticKeyPEM:  cfg.TLSStaticKeyPEM,
		}
	case cryptoutilConfig.TLSModeMixed:
		// Mixed mode: Generate server cert from CA.
		publicTLSCfg, err = tlsGenerator.GenerateServerCertFromCA(
			cfg.TLSMixedCACertPEM,
			cfg.TLSMixedCAKeyPEM,
			cfg.TLSPublicDNSNames,
			cfg.TLSPublicIPAddresses,
			cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
		)
		if err != nil {
			core.Shutdown()

			return nil, fmt.Errorf("failed to generate public TLS config (mixed mode): %w", err)
		}
	case cryptoutilConfig.TLSModeAuto:
		// Auto mode: Fully auto-generate CA hierarchy and server certificate.
		publicTLSCfg, err = tlsGenerator.GenerateAutoTLSGeneratedSettings(
			cfg.TLSPublicDNSNames,
			cfg.TLSPublicIPAddresses,
			cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
		)
		if err != nil {
			core.Shutdown()

			return nil, fmt.Errorf("failed to generate public TLS config (auto mode): %w", err)
		}
	default:
		core.Shutdown()

		return nil, fmt.Errorf("unsupported TLS public mode: %s", tlsPublicMode)
	}

	// Create public server with handlers.
	publicServer, err := NewPublicServer(ctx, cfg.BindPublicAddress, int(cfg.BindPublicPort), userRepo, messageRepo, messageRecipientJWKRepo, core.Basic.JWKGenService, barrierService, sessionManagerService, publicTLSCfg)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to create public server: %w", err)
	}

	// Create rotation service for manual key rotation admin endpoints.
	rotationService, err := cryptoutilTemplateBarrier.NewRotationService(
		core.Basic.JWKGenService,
		barrierRepo,
		core.Basic.UnsealKeysService,
	)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to create rotation service: %w", err)
	}

	// Register rotation endpoints on admin server.
	cryptoutilTemplateBarrier.RegisterRotationRoutes(adminServer.App(), rotationService)

	// Create status service for barrier keys status endpoint.
	statusService, err := cryptoutilTemplateBarrier.NewStatusService(barrierRepo)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to create status service: %w", err)
	}

	// Register status endpoint on admin server.
	cryptoutilTemplateBarrier.RegisterStatusRoutes(adminServer.App(), statusService)

	// Create application with both servers.
	app, err := cryptoutilTemplateServer.NewApplication(ctx, publicServer, adminServer)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	return &CipherIMServer{
		app:                   app,
		db:                    core.DB,
		telemetryService:      core.Basic.TelemetryService,
		jwkGenService:         core.Basic.JWKGenService,
		barrierService:        barrierService,
		sessionManagerService: sessionManagerService,
		realmService:          realmService,
		userRepo:              userRepo,
		messageRepo:           messageRepo,
		realmRepo:             realmRepo,
	}, nil
}

// Start starts both public and admin servers.
func (s *CipherIMServer) Start(ctx context.Context) error {
	//nolint:wrapcheck // Pass-through to template, wrapping not needed.
	return s.app.Start(ctx)
}

// Shutdown gracefully shuts down both servers.
func (s *CipherIMServer) Shutdown(ctx context.Context) error {
	//nolint:wrapcheck // Pass-through to template, wrapping not needed.
	return s.app.Shutdown(ctx)
}

// PublicPort returns the actual public server port.
func (s *CipherIMServer) PublicPort() int {
	return s.app.PublicPort()
}

// ActualPort returns the actual public server port (alias for PublicPort).
// Implements ServerWithActualPort interface for e2e testing utilities.
func (s *CipherIMServer) ActualPort() int {
	return s.PublicPort()
}

// AdminPort returns the actual admin server port.
func (s *CipherIMServer) AdminPort() int {
	return s.app.AdminPort()
}

// PublicBaseURL returns the base URL for the public server.
func (s *CipherIMServer) PublicBaseURL() string {
	return s.app.PublicBaseURL()
}

// AdminBaseURL returns the base URL for the admin server.
func (s *CipherIMServer) AdminBaseURL() string {
	return s.app.AdminBaseURL()
}

// DB returns the database instance.
func (s *CipherIMServer) DB() *gorm.DB {
	return s.db
}

// JWKGen returns the JWK generation service.
func (s *CipherIMServer) JWKGen() *cryptoutilJose.JWKGenService {
	return s.jwkGenService
}

// Telemetry returns the telemetry service.
func (s *CipherIMServer) Telemetry() *cryptoutilTelemetry.TelemetryService {
	return s.telemetryService
}

// SessionManager returns the session manager service.
func (s *CipherIMServer) SessionManager() *businesslogic.SessionManagerService {
	return s.sessionManagerService
}

// SetReady marks the server as ready to accept traffic.
//
// Applications should call SetReady(true) after initializing all dependencies
// but before starting the server. This enables the /admin/v1/readyz endpoint.
func (s *CipherIMServer) SetReady(ready bool) {
	s.app.SetReady(ready)
}
