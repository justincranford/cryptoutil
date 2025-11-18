package sqlrepository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	"cryptoutil/internal/server/repository/sqlrepository"
)

// TestCreateGormDB_SQLite tests CreateGormDB with SQLite.
func TestCreateGormDB_SQLite(t *testing.T) {
	ctx := context.Background()

	// Create test configuration and telemetry (creates SQLite by default).
	uuidVal, _ := uuid.NewV7()
	testName := "TestCreateGormDB_SQLite_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	// Create SQL repository (CreateGormDB is called internally).
	sqlRepo := sqlrepository.RequireNewForTest(ctx, telemetryService, testSettings)
	defer sqlRepo.Shutdown()

	// Verify repository was created successfully with SQLite.
	require.NotNil(t, sqlRepo)
	require.Equal(t, sqlrepository.DBTypeSQLite, sqlRepo.GetDBType())
}

// TestCreateGormDB_Postgres tests CreateGormDB with PostgreSQL.
func TestCreateGormDB_Postgres(t *testing.T) {
	t.Skip("Requires PostgreSQL instance - tested in integration tests")
}

// TestCreateGormDB_UnsupportedType tests CreateGormDB with unsupported database type.
func TestCreateGormDB_UnsupportedType(t *testing.T) {
	ctx := context.Background()

	// Create test configuration and telemetry.
	uuidVal, _ := uuid.NewV7()
	testName := "TestCreateGormDB_UnsupportedType_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	// Disable dev mode (otherwise it forces SQLite).
	testSettings.DevMode = false

	// Override database URL with unsupported type (mysql).
	testSettings.DatabaseURL = "mysql://test:test@tcp(localhost:3306)/testdb"

	// Try to create SQL repository with unsupported type.
	sqlRepo, err := sqlrepository.NewSQLRepository(ctx, telemetryService, testSettings)

		// Should fail with unsupported database type error.
	require.Error(t, err)
	require.Nil(t, sqlRepo)
	require.Contains(t, err.Error(), "unsupported database URL format")
}

// TestSQLRepository_Shutdown_Isolated tests shutdown in isolation.
func TestSQLRepository_Shutdown_Isolated(t *testing.T) {
	ctx := context.Background()

	// Create test configuration and telemetry.
	uuidVal, _ := uuid.NewV7()
	testName := "TestSQLRepository_Shutdown_Isolated_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	// Create SQL repository.
	sqlRepo := sqlrepository.RequireNewForTest(ctx, telemetryService, testSettings)

	// Test shutdown (doesn't panic, closes connections).
	sqlRepo.Shutdown()

	// Verify shutdown was successful (health check should fail after shutdown).
	_, err := sqlRepo.HealthCheck(ctx)
	require.Error(t, err, "Health check should fail after shutdown")
}

// TestLogPostgresSchema tests PostgreSQL schema logging.
func TestLogPostgresSchema(t *testing.T) {
	t.Skip("Requires PostgreSQL instance - tested in integration tests")
}

// TestLogSchema_SQLite tests LogSchema with SQLite (verbose mode).
func TestLogSchema_SQLite(t *testing.T) {
	ctx := context.Background()

	// Create test configuration and telemetry.
	uuidVal, _ := uuid.NewV7()
	testName := "TestLogSchema_SQLite_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	// Enable verbose mode to trigger schema logging.
	testSettings.VerboseMode = true

	// Create SQL repository (LogSchema is called internally when verbose mode is enabled).
	sqlRepo := sqlrepository.RequireNewForTest(ctx, telemetryService, testSettings)
	defer sqlRepo.Shutdown()

	// Verify repository was created successfully (LogSchema didn't panic).
	require.NotNil(t, sqlRepo)
}

// TestLogSchema_Postgres tests LogSchema with PostgreSQL (verbose mode).
func TestLogSchema_Postgres(t *testing.T) {
	t.Skip("Requires PostgreSQL instance - tested in integration tests")
}
