// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"context"
	"fmt"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSQLRepository "cryptoutil/internal/kms/server/repository/sqlrepository"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	"gorm.io/gorm"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver for database/sql
	_ "modernc.org/sqlite"             // SQLite driver for database/sql
)

// OrmRepository provides GORM-based database operations for the KMS server.
type OrmRepository struct {
	telemetryService *cryptoutilSharedTelemetry.TelemetryService
	sqlRepository    *cryptoutilSQLRepository.SQLRepository
	jwkGenService    *cryptoutilSharedCryptoJose.JWKGenService
	verboseMode      bool
	gormDB           *gorm.DB
}

// NewOrmRepository creates a new OrmRepository with the provided dependencies.
func NewOrmRepository(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, sqlRepository *cryptoutilSQLRepository.SQLRepository, jwkGenService *cryptoutilSharedCryptoJose.JWKGenService, settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) (*OrmRepository, error) {
	if ctx == nil {
		return nil, fmt.Errorf("ctx must be non-nil")
	} else if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	} else if sqlRepository == nil {
		return nil, fmt.Errorf("sqlRepository must be non-nil")
	} else if jwkGenService == nil {
		return nil, fmt.Errorf("jwkGenService must be non-nil")
	}

	gormDB, err := cryptoutilSQLRepository.CreateGormDB(sqlRepository)
	if err != nil {
		return nil, fmt.Errorf("failed to connect with gormDB: %w", err)
	}

	return &OrmRepository{telemetryService: telemetryService, sqlRepository: sqlRepository, jwkGenService: jwkGenService, gormDB: gormDB, verboseMode: settings.VerboseMode}, nil
}

// Shutdown releases resources held by the OrmRepository.
func (r *OrmRepository) Shutdown() {
	// no-op
}

// NewOrmRepositoryFromGORM creates a new OrmRepository with GORM directly (template pattern).
// This constructor allows KMS to integrate with ServerBuilder which provides GORM instances.
func NewOrmRepositoryFromGORM(_ context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, gormDB *gorm.DB, jwkGenService *cryptoutilSharedCryptoJose.JWKGenService, verboseMode bool) (*OrmRepository, error) {
	if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	} else if gormDB == nil {
		return nil, fmt.Errorf("gormDB must be non-nil")
	} else if jwkGenService == nil {
		return nil, fmt.Errorf("jwkGenService must be non-nil")
	}

	return &OrmRepository{telemetryService: telemetryService, sqlRepository: nil, jwkGenService: jwkGenService, gormDB: gormDB, verboseMode: verboseMode}, nil
}

// GormDB returns the underlying GORM database connection.
func (r *OrmRepository) GormDB() *gorm.DB {
	return r.gormDB
}

// HealthCheck performs a database connectivity check and returns detailed status.
// Uses GORM's underlying sql.DB for connection pool statistics.
func (r *OrmRepository) HealthCheck(ctx context.Context) (map[string]any, error) {
	if r.gormDB == nil {
		return map[string]any{
			"status": "error",
			"error":  "database connection not initialized",
		}, fmt.Errorf("database connection not initialized")
	}

	// Get underlying sql.DB for health check.
	sqlDB, err := r.gormDB.DB()
	if err != nil {
		return map[string]any{
			"status": "error",
			"error":  fmt.Sprintf("failed to get sql.DB from GORM: %v", err),
		}, fmt.Errorf("failed to get sql.DB from GORM: %w", err)
	}

	// Ping with timeout.
	err = sqlDB.PingContext(ctx)
	if err != nil {
		return map[string]any{
			"status": "error",
			"error":  fmt.Sprintf("database ping failed: %v", err),
		}, fmt.Errorf("database ping failed: %w", err)
	}

	// Get connection pool stats.
	stats := sqlDB.Stats()

	// Get database type from GORM dialector.
	dbType := r.gormDB.Dialector.Name()

	return map[string]any{
		"status":               "ok",
		"db_type":              dbType,
		"open_connections":     stats.OpenConnections,
		"idle_connections":     stats.Idle,
		"in_use_connections":   stats.InUse,
		"max_open_connections": stats.MaxOpenConnections,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
	}, nil
}
