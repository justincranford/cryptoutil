// Copyright (c) 2025 Justin Cranford
//
//

// Package server implements the cipher-im HTTPS server using the service template.
package server

import (
	"context"
	"fmt"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"gorm.io/gorm"

	"cryptoutil/internal/cipher/repository"
	"cryptoutil/internal/cipher/server/config"
	cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
	tlsGenerator "cryptoutil/internal/shared/config/tls_generator"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilTemplateServer "cryptoutil/internal/template/server"
	cryptoutilTemplateBarrier "cryptoutil/internal/template/server/barrier"
	cryptoutilTemplateServerListener "cryptoutil/internal/template/server/listener"
	cryptoutilTemplateServerRepository "cryptoutil/internal/template/server/repository"
)

// CipherIMServer represents the cipher-im service application.
type CipherIMServer struct {
	app *cryptoutilTemplateServer.Application
	db  *gorm.DB

	// Services.
	telemetryService *cryptoutilTelemetry.TelemetryService
	jwkGenService    *cryptoutilJose.JWKGenService
	barrierService   *cryptoutilTemplateBarrier.BarrierService

	// Repositories.
	userRepo    *repository.UserRepository
	messageRepo *repository.MessageRepository
}

// NewFromConfig creates a new cipher-im server from AppConfig only.
// This is the PREFERRED constructor - automatically provisions database (SQLite, PostgreSQL testcontainer, or external).
// Uses service-template application layer for infrastructure management.
func NewFromConfig(ctx context.Context, cfg *config.AppConfig) (*CipherIMServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create admin server first (needed for ApplicationListener).
	adminTLSCfg, err := tlsGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{cryptoutilMagic.HostnameLocalhost},
		[]string{"127.0.0.1", "::1"},
		cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate admin TLS config: %w", err)
	}

	adminServer, err := cryptoutilTemplateServerListener.NewAdminHTTPServer(ctx, &cfg.ServerSettings, adminTLSCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin server: %w", err)
	}

	// Start application core (telemetry, JWK gen, unseal, database).
	// This automatically provisions database based on cfg.DatabaseURL and cfg.DatabaseContainer.
	core, err := cryptoutilTemplateServer.StartApplicationCore(ctx, &cfg.ServerSettings)
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

	err = repository.ApplyMigrations(sqlDB, dbType)
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

	// Create TLS config for public server using auto-generated certificates.
	publicTLSCfg, err := tlsGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{cryptoutilMagic.HostnameLocalhost, "cipher-im-server"},
		[]string{"127.0.0.1", "::1"},
		cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to generate public TLS config: %w", err)
	}

	// Create public server with handlers.
	publicServer, err := NewPublicServer(ctx, int(cfg.BindPublicPort), userRepo, messageRepo, messageRecipientJWKRepo, core.Basic.JWKGenService, barrierService, cfg.JWTSecret, publicTLSCfg)
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
		app:              app,
		db:               core.DB,
		telemetryService: core.Basic.TelemetryService,
		jwkGenService:    core.Basic.JWKGenService,
		barrierService:   barrierService,
		userRepo:         userRepo,
		messageRepo:      messageRepo,
	}, nil
}

// New creates a new cipher-im server using the template.
// Takes AppConfig (which embeds ServerSettings), database instance, and database type.
// DEPRECATED: Use NewFromConfig instead - this function requires manual database management.
func New(ctx context.Context, cfg *config.AppConfig, db *gorm.DB, dbType repository.DatabaseType) (*CipherIMServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	} else if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}

	// Apply database migrations.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from GORM: %w", err)
	}

	err = repository.ApplyMigrations(sqlDB, dbType)
	if err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	// Convert repository.DatabaseType to template.DatabaseType.
	var templateDBType cryptoutilTemplateServerRepository.DatabaseType

	switch dbType {
	case repository.DatabaseTypePostgreSQL:
		templateDBType = cryptoutilTemplateServerRepository.DatabaseTypePostgreSQL
	case repository.DatabaseTypeSQLite:
		templateDBType = cryptoutilTemplateServerRepository.DatabaseTypeSQLite
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	// Create ServiceTemplate with shared infrastructure (telemetry, JWK gen).
	template, err := cryptoutilTemplateServer.NewServiceTemplate(ctx, &cfg.ServerSettings, db, templateDBType)
	if err != nil {
		return nil, fmt.Errorf("failed to create service template: %w", err)
	}

	// Initialize Barrier Service for key encryption at rest.
	// For cipher-im demo service, create a simple in-memory unseal keys service using JWE encryption.
	// Production services should use NewUnsealKeysServiceFromSettings with proper HSM/KMS integration.
	_, unsealJWK, _, _, _, err := template.JWKGen().GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	if err != nil {
		return nil, fmt.Errorf("failed to generate unseal JWK: %w", err)
	}

	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	if err != nil {
		return nil, fmt.Errorf("failed to create unseal keys service: %w", err)
	}

	// Create GORM barrier repository adapter.
	barrierRepo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create barrier repository: %w", err)
	}

	// Create barrier service with GORM repository.
	barrierService, err := cryptoutilTemplateBarrier.NewBarrierService(
		ctx,
		template.Telemetry(),
		template.JWKGen(),
		barrierRepo,
		unsealKeysService,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create barrier service: %w", err)
	}

	// Initialize repositories.
	userRepo := repository.NewUserRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(db, barrierService)

	// Create TLS config for public server using auto-generated certificates.
	publicTLSCfg, err := tlsGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{cryptoutilMagic.HostnameLocalhost, "cipher-im-server"},
		[]string{"127.0.0.1", "::1"},
		cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate public TLS config: %w", err)
	}

	// Create public server with handlers.
	// Use BindPublicPort from embedded ServerSettings.
	publicServer, err := NewPublicServer(ctx, int(cfg.BindPublicPort), userRepo, messageRepo, messageRecipientJWKRepo, template.JWKGen(), barrierService, cfg.JWTSecret, publicTLSCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create public server: %w", err)
	}

	// Create admin server TLS config using auto-generated certificates.
	adminTLSCfg, err := tlsGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{cryptoutilMagic.HostnameLocalhost},
		[]string{"127.0.0.1", "::1"},
		cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate admin TLS config: %w", err)
	}

	// Create admin server using ServerSettings from AppConfig.
	adminServer, err := cryptoutilTemplateServerListener.NewAdminHTTPServer(ctx, &cfg.ServerSettings, adminTLSCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin server: %w", err)
	}

	// Create rotation service for manual key rotation admin endpoints.
	rotationService, err := cryptoutilTemplateBarrier.NewRotationService(
		template.JWKGen(),
		barrierRepo,
		unsealKeysService,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create rotation service: %w", err)
	}

	// Register rotation endpoints on admin server.
	// Routes: POST /admin/v1/barrier/rotate/{root,intermediate,content}
	cryptoutilTemplateBarrier.RegisterRotationRoutes(adminServer.App(), rotationService)

	// Create status service for barrier keys status endpoint.
	statusService, err := cryptoutilTemplateBarrier.NewStatusService(barrierRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to create status service: %w", err)
	}

	// Register status endpoint on admin server.
	// Route: GET /admin/v1/barrier/keys/status
	cryptoutilTemplateBarrier.RegisterStatusRoutes(adminServer.App(), statusService)

	// Create application with both servers.
	app, err := cryptoutilTemplateServer.NewApplication(ctx, publicServer, adminServer)
	if err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	return &CipherIMServer{
		app:              app,
		db:               db,
		telemetryService: template.Telemetry(),
		jwkGenService:    template.JWKGen(),
		barrierService:   barrierService,
		userRepo:         userRepo,
		messageRepo:      messageRepo,
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

// SetReady marks the server as ready to accept traffic.
//
// Applications should call SetReady(true) after initializing all dependencies
// but before starting the server. This enables the /admin/v1/readyz endpoint.
func (s *CipherIMServer) SetReady(ready bool) {
	s.app.SetReady(ready)
}
