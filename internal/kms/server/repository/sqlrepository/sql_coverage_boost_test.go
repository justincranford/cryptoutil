// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository_test

import (
	"context"
	"testing"

	cryptoutilSQLRepository "cryptoutil/internal/kms/server/repository/sqlrepository"

	googleUuid "github.com/google/uuid"
	testify "github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// TestCreateGormDB_SQLite tests CreateGormDB with SQLite.
func TestCreateGormDB_SQLite(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create test configuration and telemetry (creates SQLite by default).
	uuidVal, _ := googleUuid.NewV7() //nolint:errcheck // UUID generation error virtually impossible
	testName := "TestCreateGormDB_SQLite_" + uuidVal.String()
	testSettings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	// Create SQL repository (CreateGormDB is called internally).
	sqlRepo := cryptoutilSQLRepository.RequireNewForTest(ctx, telemetryService, testSettings)
	defer sqlRepo.Shutdown()

	// Verify repository was created successfully with SQLite.
	testify.NotNil(t, sqlRepo)
	testify.Equal(t, cryptoutilSQLRepository.DBTypeSQLite, sqlRepo.GetDBType())
}

// TestCreateGormDB_Postgres tests CreateGormDB with PostgreSQL.
func TestCreateGormDB_Postgres(t *testing.T) {
	t.Parallel()

	t.Skip("Requires PostgreSQL instance - tested in integration tests")
}

// TestCreateGormDB_UnsupportedType tests CreateGormDB with unsupported database type.
func TestCreateGormDB_UnsupportedType(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create test configuration and telemetry.
	uuidVal, _ := googleUuid.NewV7() //nolint:errcheck // UUID generation error virtually impossible
	testName := "TestCreateGormDB_UnsupportedType_" + uuidVal.String()
	testSettings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	// Disable dev mode (otherwise it forces SQLite).
	testSettings.DevMode = false

	// Override database URL with unsupported type (mysql).
	testSettings.DatabaseURL = "mysql://test:test@tcp(localhost:3306)/testdb"

	// Try to create SQL repository with unsupported type.
	sqlRepo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, testSettings)

	// Should fail with unsupported database type error.
	testify.Error(t, err)
	testify.Nil(t, sqlRepo)
	testify.Contains(t, err.Error(), "unsupported database URL format")
}

// TestSQLRepository_Shutdown_Isolated tests shutdown in isolation.
func TestSQLRepository_Shutdown_Isolated(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create test configuration and telemetry.
	uuidVal, _ := googleUuid.NewV7() //nolint:errcheck // UUID generation error virtually impossible
	testName := "TestSQLRepository_Shutdown_Isolated_" + uuidVal.String()
	testSettings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	// Create SQL repository.
	sqlRepo := cryptoutilSQLRepository.RequireNewForTest(ctx, telemetryService, testSettings)

	// Test shutdown (doesn't panic, closes connections).
	sqlRepo.Shutdown()

	// Verify shutdown was successful (health check should fail after shutdown).
	_, err := sqlRepo.HealthCheck(ctx)
	testify.Error(t, err, "Health check should fail after shutdown")
}

// TestLogPostgresSchema tests PostgreSQL schema logging.
func TestLogPostgresSchema(t *testing.T) {
	t.Parallel()

	t.Skip("Requires PostgreSQL instance - tested in integration tests")
}

// TestLogSchema_Postgres tests LogSchema with PostgreSQL (verbose mode).
func TestLogSchema_Postgres(t *testing.T) {
	t.Parallel()

	t.Skip("Requires PostgreSQL instance - tested in integration tests")
}
