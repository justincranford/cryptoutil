// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"context"
	"fmt"

	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/apps/template/service/telemetry"

	"gorm.io/gorm"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver for database/sql
	_ "modernc.org/sqlite"             // SQLite driver for database/sql
)

// OrmRepository provides GORM-based database operations for the KMS server.
type OrmRepository struct {
	telemetryService *cryptoutilSharedTelemetry.TelemetryService
	jwkGenService    *cryptoutilSharedCryptoJose.JWKGenService
	verboseMode      bool
	gormDB           *gorm.DB
}

// Shutdown releases resources held by the OrmRepository.
func (r *OrmRepository) Shutdown() {
	// no-op
}

// NewOrmRepository creates a new OrmRepository with GORM directly (template pattern).
// KMS integrates with ServerBuilder which provides GORM instances.
func NewOrmRepository(_ context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, gormDB *gorm.DB, jwkGenService *cryptoutilSharedCryptoJose.JWKGenService, verboseMode bool) (*OrmRepository, error) {
	if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	} else if gormDB == nil {
		return nil, fmt.Errorf("gormDB must be non-nil")
	} else if jwkGenService == nil {
		return nil, fmt.Errorf("jwkGenService must be non-nil")
	}

	return &OrmRepository{telemetryService: telemetryService, jwkGenService: jwkGenService, gormDB: gormDB, verboseMode: verboseMode}, nil
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
	dbType := r.gormDB.Name()

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
