// Copyright (c) 2025 Justin Cranford
//
//

// Package server implements the learn-im HTTPS server using the service template.
package server

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"cryptoutil/internal/learn/repository"
	"cryptoutil/internal/learn/server/config"
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

// LearnIMServer represents the learn-im service application.
type LearnIMServer struct {
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

// New creates a new learn-im server using the template.
// Takes AppConfig (which embeds ServerSettings), database instance, and database type.
func New(ctx context.Context, cfg *config.AppConfig, db *gorm.DB, dbType repository.DatabaseType) (*LearnIMServer, error) {
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
	// Create a simple in-memory unseal keys service for demo purposes.
	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceFromSettings(ctx, template.Telemetry(), &cfg.ServerSettings)
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
		[]string{cryptoutilMagic.HostnameLocalhost, "learn-im-server"},
		[]string{"127.0.0.1", "::1"},
		cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate public TLS config: %w", err)
	}

	// Create public server with handlers.
	// Use BindPublicPort from embedded ServerSettings.
	publicServer, err := NewPublicServer(ctx, int(cfg.BindPublicPort), userRepo, messageRepo, messageRecipientJWKRepo, template.JWKGen(), cfg.JWTSecret, publicTLSCfg)
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

	// Create application with both servers.
	app, err := cryptoutilTemplateServer.NewApplication(ctx, publicServer, adminServer)
	if err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	return &LearnIMServer{
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
func (s *LearnIMServer) Start(ctx context.Context) error {
	//nolint:wrapcheck // Pass-through to template, wrapping not needed.
	return s.app.Start(ctx)
}

// Shutdown gracefully shuts down both servers.
func (s *LearnIMServer) Shutdown(ctx context.Context) error {
	//nolint:wrapcheck // Pass-through to template, wrapping not needed.
	return s.app.Shutdown(ctx)
}

// PublicPort returns the actual public server port.
func (s *LearnIMServer) PublicPort() int {
	return s.app.PublicPort()
}

// AdminPort returns the actual admin server port.
func (s *LearnIMServer) AdminPort() (int, error) {
	//nolint:wrapcheck // Pass-through to template, wrapping not needed.
	return s.app.AdminPort()
}
