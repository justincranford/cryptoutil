// Copyright (c) 2025 Justin Cranford
//

// Package server implements the pki-ca HTTPS server using the service template.
package server

import (
	"context"
	"fmt"

	cryptoutilAppsPkiCaDomain "cryptoutil/internal/apps/pki/ca/domain"
	cryptoutilAppsPkiCaRepository "cryptoutil/internal/apps/pki/ca/repository"
	cryptoutilAppsCaServerConfig "cryptoutil/internal/apps/pki/ca/server/config"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilAppsTemplateServiceServerBuilder "cryptoutil/internal/apps/template/service/server/builder"

	"gorm.io/gorm"
)

// PKICAServer wraps the service template Application for the pki-ca service.
type PKICAServer struct {
	app *cryptoutilAppsTemplateServiceServer.Application
	db  *gorm.DB
}

// NewFromConfig creates a new PKI CA server from configuration.
func NewFromConfig(ctx context.Context, cfg *cryptoutilAppsCaServerConfig.PKICAServerSettings) (*PKICAServer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("failed to create pki-ca server: %w", fmt.Errorf("context is nil"))
	}

	if cfg == nil {
		return nil, fmt.Errorf("failed to create pki-ca server: %w", fmt.Errorf("config is nil"))
	}

	builder := cryptoutilAppsTemplateServiceServerBuilder.NewServerBuilder(ctx, cfg.ServiceTemplateServerSettings)
	builder.WithDomainMigrations(cryptoutilAppsPkiCaRepository.MigrationsFS, "migrations")
	builder.WithPublicRouteRegistration(func(
		_ *cryptoutilAppsTemplateServiceServer.PublicServerBase,
		res *cryptoutilAppsTemplateServiceServerBuilder.ServiceResources,
	) error {
		// Auto-migrate domain models for GORM compatibility.
		if err := res.DB.AutoMigrate(&cryptoutilAppsPkiCaDomain.CAItem{}); err != nil {
			return fmt.Errorf("failed to auto-migrate pki-ca domain models: %w", err)
		}

		return nil
	})

	resources, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build pki-ca server: %w", err)
	}

	return &PKICAServer{app: resources.Application, db: resources.DB}, nil
}

// Start starts the pki-ca server (blocking).
func (s *PKICAServer) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("failed to start pki-ca server: %w", fmt.Errorf("context is nil"))
	}

	if err := s.app.Start(ctx); err != nil {
		return fmt.Errorf("failed to start pki-ca server: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the pki-ca server.
func (s *PKICAServer) Shutdown(ctx context.Context) error {
	if err := s.app.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown pki-ca server: %w", err)
	}

	return nil
}

// PublicPort returns the actual port the public server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *PKICAServer) PublicPort() int {
	return s.app.PublicPort()
}

// AdminPort returns the actual port the admin server is listening on (for tests).
// Useful when configured with port 0 for dynamic allocation.
func (s *PKICAServer) AdminPort() int {
	return s.app.AdminPort()
}

// SetReady marks the server as ready (enables /admin/v1/readyz to return 200 OK).
func (s *PKICAServer) SetReady(ready bool) {
	s.app.SetReady(ready)
}

// PublicBaseURL returns the public server base URL (for tests).
func (s *PKICAServer) PublicBaseURL() string {
	return s.app.PublicBaseURL()
}

// AdminBaseURL returns the admin server base URL (for tests).
func (s *PKICAServer) AdminBaseURL() string {
	return s.app.AdminBaseURL()
}

// PublicServerActualPort returns the actual port the public server is listening on.
// Useful when configured with port 0 for dynamic allocation.
func (s *PKICAServer) PublicServerActualPort() int {
	return s.app.PublicPort()
}

// AdminServerActualPort returns the actual port the admin server is listening on.
// Useful when configured with port 0 for dynamic allocation.
func (s *PKICAServer) AdminServerActualPort() int {
	return s.app.AdminPort()
}

// DB returns the GORM database instance for domain operations.
func (s *PKICAServer) DB() *gorm.DB {
	return s.db
}
