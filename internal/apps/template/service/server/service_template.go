// Copyright (c) 2025 Justin Cranford
//
//

// Package server provides reusable service template infrastructure.
package server

import (
	"context"
	"database/sql"
	"fmt"

	"gorm.io/gorm"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTemplateServerApplication "cryptoutil/internal/apps/template/service/server/application"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilBarrierService "cryptoutil/internal/shared/barrier"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

// ServiceTemplate encapsulates reusable service infrastructure.
// Provides common initialization for telemetry, crypto, barrier, and dual HTTPS servers.
type ServiceTemplate struct {
	config    *cryptoutilConfig.ServiceTemplateServerSettings
	db        *gorm.DB
	dbType    cryptoutilAppsTemplateServiceServerRepository.DatabaseType
	telemetry *cryptoutilTelemetry.TelemetryService
	jwkGen    *cryptoutilJose.JWKGenService
	barrier   *cryptoutilBarrierService.BarrierService // Optional (nil for demo services).
}

// ServiceTemplateOption is a functional option for configuring ServiceTemplate.
type ServiceTemplateOption func(*ServiceTemplate) error

// WithBarrier configures an optional barrier service for key encryption at rest.
func WithBarrier(barrier *cryptoutilBarrierService.BarrierService) ServiceTemplateOption {
	return func(st *ServiceTemplate) error {
		st.barrier = barrier

		return nil
	}
}

// NewServiceTemplate creates a new ServiceTemplate with common infrastructure.
// Initializes telemetry, JWK generation service, and optionally barrier service.
// Does NOT run migrations or create HTTP servers (caller-specific).
func NewServiceTemplate(
	ctx context.Context,
	config *cryptoutilConfig.ServiceTemplateServerSettings,
	db *gorm.DB,
	dbType cryptoutilAppsTemplateServiceServerRepository.DatabaseType,
	options ...ServiceTemplateOption,
) (*ServiceTemplate, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	} else if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}

	// Validate database type.
	switch dbType {
	case cryptoutilAppsTemplateServiceServerRepository.DatabaseTypePostgreSQL, cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite:
		// Valid database type.
	default:
		return nil, fmt.Errorf("invalid database type: %s", dbType)
	}

	// Initialize telemetry service.
	telemetryService, err := cryptoutilTelemetry.NewTelemetryService(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	// Initialize JWK Generation Service for cryptographic operations.
	// Uses in-memory key pools with telemetry for monitoring.
	jwkGenService, err := cryptoutilJose.NewJWKGenService(ctx, telemetryService, false)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize JWK generation service: %w", err)
	}

	st := &ServiceTemplate{
		config:    config,
		db:        db,
		dbType:    dbType,
		telemetry: telemetryService,
		jwkGen:    jwkGenService,
		barrier:   nil, // Optional, set via WithBarrier option.
	}

	// Apply functional options.
	for _, opt := range options {
		if err := opt(st); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	return st, nil
}

// Config returns the server configuration.
func (st *ServiceTemplate) Config() *cryptoutilConfig.ServiceTemplateServerSettings {
	return st.config
}

// DB returns the GORM database instance.
func (st *ServiceTemplate) DB() *gorm.DB {
	return st.db
}

// SQLDB returns the underlying sql.DB instance.
func (st *ServiceTemplate) SQLDB() (*sql.DB, error) {
	//nolint:wrapcheck // Pass-through to GORM, wrapping not needed.
	return st.db.DB()
}

// DBType returns the database type.
func (st *ServiceTemplate) DBType() cryptoutilAppsTemplateServiceServerRepository.DatabaseType {
	return st.dbType
}

// Telemetry returns the telemetry service.
func (st *ServiceTemplate) Telemetry() *cryptoutilTelemetry.TelemetryService {
	return st.telemetry
}

// JWKGen returns the JWK generation service.
func (st *ServiceTemplate) JWKGen() *cryptoutilJose.JWKGenService {
	return st.jwkGen
}

// Barrier returns the optional barrier service (may be nil).
func (st *ServiceTemplate) Barrier() *cryptoutilBarrierService.BarrierService {
	return st.barrier
}

// Shutdown gracefully shuts down all service components.
func (st *ServiceTemplate) Shutdown() {
	if st.telemetry != nil {
		st.telemetry.Shutdown()
	}

	if st.jwkGen != nil {
		st.jwkGen.Shutdown()
	}

	if st.barrier != nil {
		st.barrier.Shutdown()
	}
}

// StartApplicationCore is a convenience wrapper for application.StartApplicationCore.
// Creates ApplicationCore with automatic database provisioning.
// Returns ApplicationCore with initialized telemetry, JWK gen, unseal, and database.
func StartApplicationCore(ctx context.Context, settings *cryptoutilConfig.ServiceTemplateServerSettings) (*cryptoutilTemplateServerApplication.ApplicationCore, error) {
	//nolint:wrapcheck // Pass-through to application layer.
	return cryptoutilTemplateServerApplication.StartApplicationCore(ctx, settings)
}
