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

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerApplication "cryptoutil/internal/apps/template/service/server/application"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// ServiceTemplate encapsulates reusable service infrastructure.
// Provides common initialization for telemetry, crypto, and dual HTTPS servers.
// Note: Barrier service is handled by ServerBuilder, not ServiceTemplate.
type ServiceTemplate struct {
	config    *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
	db        *gorm.DB
	dbType    cryptoutilAppsTemplateServiceServerRepository.DatabaseType
	telemetry *cryptoutilSharedTelemetry.TelemetryService
	jwkGen    *cryptoutilSharedCryptoJose.JWKGenService
}

// ServiceTemplateOption is a functional option for configuring ServiceTemplate.
type ServiceTemplateOption func(*ServiceTemplate) error

// NewServiceTemplate creates a new ServiceTemplate with common infrastructure.
// Initializes telemetry, JWK generation service, and optionally barrier service.
// Does NOT run migrations or create HTTP servers (caller-specific).
func NewServiceTemplate(
	ctx context.Context,
	config *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings,
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
	telemetryService, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	// Initialize JWK Generation Service for cryptographic operations.
	// Uses in-memory key pools with telemetry for monitoring.
	jwkGenService, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetryService, false)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize JWK generation service: %w", err)
	}

	st := &ServiceTemplate{
		config:    config,
		db:        db,
		dbType:    dbType,
		telemetry: telemetryService,
		jwkGen:    jwkGenService,
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
func (st *ServiceTemplate) Config() *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings {
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
func (st *ServiceTemplate) Telemetry() *cryptoutilSharedTelemetry.TelemetryService {
	return st.telemetry
}

// JWKGen returns the JWK generation service.
func (st *ServiceTemplate) JWKGen() *cryptoutilSharedCryptoJose.JWKGenService {
	return st.jwkGen
}

// Shutdown gracefully shuts down all service components.
func (st *ServiceTemplate) Shutdown() {
	if st.telemetry != nil {
		st.telemetry.Shutdown()
	}

	if st.jwkGen != nil {
		st.jwkGen.Shutdown()
	}
}

// StartApplicationCore is a convenience wrapper for application.StartApplicationCore.
// Creates ApplicationCore with automatic database provisioning.
// Returns ApplicationCore with initialized telemetry, JWK gen, unseal, and database.
func StartApplicationCore(ctx context.Context, settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) (*cryptoutilAppsTemplateServiceServerApplication.Core, error) {
	//nolint:wrapcheck // Pass-through to application layer.
	return cryptoutilAppsTemplateServiceServerApplication.StartCore(ctx, settings)
}
