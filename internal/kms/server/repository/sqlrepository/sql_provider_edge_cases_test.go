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

// TestNewSQLRepository_NilTelemetryService tests nil telemetry service error.
func TestNewSQLRepository_NilTelemetryService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7() //nolint:errcheck // UUID generation error virtually impossible
	testName := "TestNewSQLRepository_NilTelemetryService_" + uuidVal.String()
	testSettings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest(testName)

	sqlRepo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, nil, testSettings)

	testify.Error(t, err)
	testify.Nil(t, sqlRepo)
	testify.Contains(t, err.Error(), "telemetryService must be non-nil")
}

// TestNewSQLRepository_NilSettings tests nil settings error.
func TestNewSQLRepository_NilSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7() //nolint:errcheck // UUID generation error virtually impossible
	testName := "TestNewSQLRepository_NilSettings_" + uuidVal.String()
	testSettings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	sqlRepo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, nil)

	testify.Error(t, err)
	testify.Nil(t, sqlRepo)
	testify.Contains(t, err.Error(), "settings must be non-nil")
}

// TestNewSQLRepository_ContainerModeInvalid tests invalid container mode.
func TestNewSQLRepository_ContainerModeInvalid(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7() //nolint:errcheck // UUID generation error virtually impossible
	testName := "TestNewSQLRepository_ContainerModeInvalid_" + uuidVal.String()
	testSettings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	// Set invalid container mode.
	testSettings.DatabaseContainer = "invalid-mode"

	sqlRepo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, testSettings)

	testify.Error(t, err)
	testify.Nil(t, sqlRepo)
	testify.Contains(t, err.Error(), "failed to determine container mode")
}

// TestHealthCheck tests database health check.
func TestHealthCheck(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7() //nolint:errcheck // UUID generation error virtually impossible
	testName := "TestHealthCheck_" + uuidVal.String()
	testSettings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	sqlRepo := cryptoutilSQLRepository.RequireNewForTest(ctx, telemetryService, testSettings)
	defer sqlRepo.Shutdown()

	// Test health check passes.
	status, err := sqlRepo.HealthCheck(ctx)
	testify.NoError(t, err, "Health check should pass for valid database")
	testify.NotNil(t, status)
	testify.Contains(t, status, "status")
	testify.Equal(t, "ok", status["status"])
}

// TestHealthCheck_AfterShutdown tests health check after shutdown.
func TestHealthCheck_AfterShutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7() //nolint:errcheck // UUID generation error virtually impossible
	testName := "TestHealthCheck_AfterShutdown_" + uuidVal.String()
	testSettings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	sqlRepo := cryptoutilSQLRepository.RequireNewForTest(ctx, telemetryService, testSettings)

	// Shutdown the repository.
	sqlRepo.Shutdown()

	// Health check should fail.
	status, err := sqlRepo.HealthCheck(ctx)
	testify.Error(t, err, "Health check should fail after shutdown")
	testify.NotNil(t, status)
	testify.Contains(t, status, "status")
	testify.Equal(t, "error", status["status"])
}
