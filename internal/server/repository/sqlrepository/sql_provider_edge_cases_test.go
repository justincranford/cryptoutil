package sqlrepository_test

import (
	"context"
	"cryptoutil/internal/server/repository/sqlrepository"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
)

// TestNewSQLRepository_NilTelemetryService tests nil telemetry service error.
func TestNewSQLRepository_NilTelemetryService(t *testing.T) {
	ctx := context.Background()
	uuidVal, _ := uuid.NewV7()
	testName := "TestNewSQLRepository_NilTelemetryService_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)

	sqlRepo, err := sqlrepository.NewSQLRepository(ctx, nil, testSettings)

	require.Error(t, err)
	require.Nil(t, sqlRepo)
	require.Contains(t, err.Error(), "telemetryService must be non-nil")
}

// TestNewSQLRepository_NilSettings tests nil settings error.
func TestNewSQLRepository_NilSettings(t *testing.T) {
	ctx := context.Background()
	uuidVal, _ := uuid.NewV7()
	testName := "TestNewSQLRepository_NilSettings_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	sqlRepo, err := sqlrepository.NewSQLRepository(ctx, telemetryService, nil)

	require.Error(t, err)
	require.Nil(t, sqlRepo)
	require.Contains(t, err.Error(), "settings must be non-nil")
}

// TestNewSQLRepository_ContainerModeInvalid tests invalid container mode.
func TestNewSQLRepository_ContainerModeInvalid(t *testing.T) {
	ctx := context.Background()
	uuidVal, _ := uuid.NewV7()
	testName := "TestNewSQLRepository_ContainerModeInvalid_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	// Set invalid container mode.
	testSettings.DatabaseContainer = "invalid-mode"

	sqlRepo, err := sqlrepository.NewSQLRepository(ctx, telemetryService, testSettings)

	require.Error(t, err)
	require.Nil(t, sqlRepo)
	require.Contains(t, err.Error(), "failed to determine container mode")
}

// TestHealthCheck tests database health check.
func TestHealthCheck(t *testing.T) {
	ctx := context.Background()
	uuidVal, _ := uuid.NewV7()
	testName := "TestHealthCheck_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	sqlRepo := sqlrepository.RequireNewForTest(ctx, telemetryService, testSettings)
	defer sqlRepo.Shutdown()

	// Test health check passes.
	status, err := sqlRepo.HealthCheck(ctx)
	require.NoError(t, err, "Health check should pass for valid database")
	require.NotNil(t, status)
	require.Contains(t, status, "status")
	require.Equal(t, "ok", status["status"])
}

// TestHealthCheck_AfterShutdown tests health check after shutdown.
func TestHealthCheck_AfterShutdown(t *testing.T) {
	ctx := context.Background()
	uuidVal, _ := uuid.NewV7()
	testName := "TestHealthCheck_AfterShutdown_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	sqlRepo := sqlrepository.RequireNewForTest(ctx, telemetryService, testSettings)

	// Shutdown the repository.
	sqlRepo.Shutdown()

	// Health check should fail.
	status, err := sqlRepo.HealthCheck(ctx)
	require.Error(t, err, "Health check should fail after shutdown")
	require.NotNil(t, status)
	require.Contains(t, status, "status")
	require.Equal(t, "error", status["status"])
}
