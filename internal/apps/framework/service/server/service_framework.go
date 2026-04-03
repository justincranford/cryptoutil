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

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilAppsFrameworkServiceServerApplication "cryptoutil/internal/apps/framework/service/server/application"
	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps/framework/service/server/repository"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// ServiceFramework encapsulates reusable service infrastructure.
// Provides common initialization for telemetry, crypto, and dual HTTPS servers.
// Note: Barrier service is handled by ServerBuilder, not ServiceFramework.
type ServiceFramework struct {
	config    *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings
	db        *gorm.DB
	dbType    cryptoutilAppsFrameworkServiceServerRepository.DatabaseType
	telemetry *cryptoutilSharedTelemetry.TelemetryService
	jwkGen    *cryptoutilSharedCryptoJose.JWKGenService
}

// ServiceFrameworkOption is a functional option for configuring ServiceFramework.
type ServiceFrameworkOption func(*ServiceFramework) error

// NewServiceFramework creates a new ServiceFramework with common infrastructure.
// Initializes telemetry, JWK generation service, and optionally barrier service.
// Does NOT run migrations or create HTTP servers (caller-specific).
func NewServiceFramework(
	ctx context.Context,
	config *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings,
	db *gorm.DB,
	dbType cryptoutilAppsFrameworkServiceServerRepository.DatabaseType,
	options ...ServiceFrameworkOption,
) (*ServiceFramework, error) {
	return newServiceFrameworkInternal(
		ctx,
		config,
		db,
		dbType,
		func(ctx context.Context, config *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (*cryptoutilSharedTelemetry.TelemetryService, error) {
			return cryptoutilSharedTelemetry.NewTelemetryService(ctx, config.ToTelemetrySettings())
		},
		func(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, devMode bool) (*cryptoutilSharedCryptoJose.JWKGenService, error) {
			return cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetryService, devMode)
		},
		options...,
	)
}

func newServiceFrameworkInternal(
	ctx context.Context,
	config *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings,
	db *gorm.DB,
	dbType cryptoutilAppsFrameworkServiceServerRepository.DatabaseType,
	newTelemetryServiceFn func(ctx context.Context, config *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (*cryptoutilSharedTelemetry.TelemetryService, error),
	newJWKGenServiceFn func(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, devMode bool) (*cryptoutilSharedCryptoJose.JWKGenService, error),
	options ...ServiceFrameworkOption,
) (*ServiceFramework, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	} else if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	} else if db == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}

	// Validate database type.
	switch dbType {
	case cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypePostgreSQL, cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypeSQLite:
		// Valid database type.
	default:
		return nil, fmt.Errorf("invalid database type: %s", dbType)
	}

	// Initialize telemetry service.
	telemetryService, err := newTelemetryServiceFn(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	// Initialize JWK Generation Service for cryptographic operations.
	// Uses in-memory key pools with telemetry for monitoring.
	jwkGenService, err := newJWKGenServiceFn(ctx, telemetryService, false)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize JWK generation service: %w", err)
	}

	st := &ServiceFramework{
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
func (st *ServiceFramework) Config() *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings {
	return st.config
}

// DB returns the GORM database instance.
func (st *ServiceFramework) DB() *gorm.DB {
	return st.db
}

// SQLDB returns the underlying sql.DB instance.
func (st *ServiceFramework) SQLDB() (*sql.DB, error) {
	//nolint:wrapcheck // Pass-through to GORM, wrapping not needed.
	return st.db.DB()
}

// DBType returns the database type.
func (st *ServiceFramework) DBType() cryptoutilAppsFrameworkServiceServerRepository.DatabaseType {
	return st.dbType
}

// Telemetry returns the telemetry service.
func (st *ServiceFramework) Telemetry() *cryptoutilSharedTelemetry.TelemetryService {
	return st.telemetry
}

// JWKGen returns the JWK generation service.
func (st *ServiceFramework) JWKGen() *cryptoutilSharedCryptoJose.JWKGenService {
	return st.jwkGen
}

// Shutdown gracefully shuts down all service components.
func (st *ServiceFramework) Shutdown() {
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
func StartApplicationCore(ctx context.Context, settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (*cryptoutilAppsFrameworkServiceServerApplication.Core, error) {
	//nolint:wrapcheck // Pass-through to application layer.
	return cryptoutilAppsFrameworkServiceServerApplication.StartCore(ctx, settings)
}
