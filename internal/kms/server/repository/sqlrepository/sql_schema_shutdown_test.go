// Copyright (c) 2025 Justin Cranford

package sqlrepository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// TestLogSchema_PostgreSQL tests LogSchema with PostgreSQL backend (logPostgresSchema coverage).
func TestLogSchema_PostgreSQL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("sql_schema_postgres")

	telemetry := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings)
	defer telemetry.Shutdown()

	// Create PostgreSQL repository.
	repo := RequireNewForTest(ctx, telemetry, settings)
	defer repo.Shutdown()

	// Call LogSchema - should exercise logPostgresSchema path.
	err := LogSchema(repo)
	require.NoError(t, err, "LogSchema should succeed for PostgreSQL")
}

// TestLogSchema_SQLite tests LogSchema with SQLite backend (logSQLiteSchema coverage).
func TestLogSchema_SQLite(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("sql_schema_sqlite")

	telemetry := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings)
	defer telemetry.Shutdown()

	// Create SQLite repository.
	repo := RequireNewForTest(ctx, telemetry, settings)
	defer repo.Shutdown()

	// Call LogSchema - should exercise logSQLiteSchema path.
	err := LogSchema(repo)
	require.NoError(t, err, "LogSchema should succeed for SQLite")
}

// TestShutdown_SQLite tests Shutdown function with SQLite backend.
func TestShutdown_SQLite(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("sql_shutdown_sqlite")

	telemetry := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings)
	defer telemetry.Shutdown()

	// Create SQLite repository.
	repo := RequireNewForTest(ctx, telemetry, settings)

	// Call Shutdown - should close DB connection.
	repo.Shutdown()

	// Verify DB is closed by attempting operation (should fail).
	_, err := repo.HealthCheck(ctx)
	require.Error(t, err, "HealthCheck should fail after Shutdown")
	require.Contains(t, err.Error(), "sql: database is closed", "Error should indicate database is closed")
}

// TestShutdown_PostgreSQL tests Shutdown function with PostgreSQL backend.
func TestShutdown_PostgreSQL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("sql_shutdown_postgres")

	telemetry := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings)
	defer telemetry.Shutdown()

	// Create PostgreSQL repository.
	repo := RequireNewForTest(ctx, telemetry, settings)

	// Call Shutdown - should close DB connection and container.
	repo.Shutdown()

	// Verify DB is closed by attempting operation (should fail).
	_, err := repo.HealthCheck(ctx)
	require.Error(t, err, "HealthCheck should fail after Shutdown")
	require.Contains(t, err.Error(), "sql: database is closed", "Error should indicate database is closed")
}
