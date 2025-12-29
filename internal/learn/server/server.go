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
	cryptoutilBarrierService "cryptoutil/internal/shared/barrier"
	tlsGenerator "cryptoutil/internal/shared/config/tls_generator"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilTemplateServer "cryptoutil/internal/template/server"
)

// LearnIMServer represents the learn-im service application.
type LearnIMServer struct {
	app *cryptoutilTemplateServer.Application
	db  *gorm.DB

	// Services.
	telemetryService *cryptoutilTelemetry.TelemetryService
	jwkGenService    *cryptoutilJose.JWKGenService
	barrierService   *cryptoutilBarrierService.BarrierService

	// Repositories.
	userRepo    *repository.UserRepository
	messageRepo *repository.MessageRepository
}

// New creates a new learn-im server using the template.
// Takes AppConfig (which embeds ServerSettings), database instance, and database type.
func New(ctx context.Context, cfg *AppConfig, db *gorm.DB, dbType repository.DatabaseType) (*LearnIMServer, error) {
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

	// Initialize telemetry service using ServerSettings from AppConfig.
	telemetryService, err := cryptoutilTelemetry.NewTelemetryService(ctx, &cfg.ServerSettings)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	// Initialize JWK Generation Service for message encryption.
	// Uses in-memory key pools with telemetry for monitoring.
	jwkGenService, err := cryptoutilJose.NewJWKGenService(ctx, telemetryService, false)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize JWK generation service: %w", err)
	}

	// NOTE: Phase 5b will initialize Barrier Service for key encryption at rest.
	// For Phase 5a, message encryption JWKs are generated in-memory without barrier encryption.
	// Phase 5b will add barrier service to encrypt JWKs before storing in messages_jwks table.

	// Initialize repositories.
	userRepo := repository.NewUserRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(db)

	// Create TLS config for public server using auto-generated certificates.
	publicTLSCfg, err := tlsGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{"localhost", "learn-im-server"},
		[]string{"127.0.0.1", "::1"},
		cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate public TLS config: %w", err)
	}

	// Create public server with handlers.
	// Use BindPublicPort from embedded ServerSettings.
	publicServer, err := NewPublicServer(ctx, int(cfg.BindPublicPort), userRepo, messageRepo, messageRecipientJWKRepo, jwkGenService, cfg.JWTSecret, publicTLSCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create public server: %w", err)
	}

	// Create admin server TLS config using auto-generated certificates.
	adminTLSCfg, err := tlsGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{"localhost"},
		[]string{"127.0.0.1", "::1"},
		cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate admin TLS config: %w", err)
	}

	// Create admin server using ServerSettings from AppConfig.
	adminServer, err := cryptoutilTemplateServer.NewAdminHTTPServer(ctx, &cfg.ServerSettings, adminTLSCfg)
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
		telemetryService: telemetryService,
		jwkGenService:    jwkGenService,
		barrierService:   nil, // NOTE: Phase 5b will initialize barrier service for encrypted key storage.
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
