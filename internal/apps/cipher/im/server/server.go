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
	"cryptoutil/internal/apps/cipher/im/server/config"
	tlsGenerator "cryptoutil/internal/template/config/tls_generator"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilTemplateServer "cryptoutil/internal/template/server"
	cryptoutilTemplateBarrier "cryptoutil/internal/template/server/barrier"
	cryptoutilTemplateServerListener "cryptoutil/internal/template/server/listener"
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

// SetReady marks the server as ready to accept traffic.
//
// Applications should call SetReady(true) after initializing all dependencies
// but before starting the server. This enables the /admin/v1/readyz endpoint.
func (s *CipherIMServer) SetReady(ready bool) {
	s.app.SetReady(ready)
}
