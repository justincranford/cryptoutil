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

// TestNewSQLRepository_InvalidDatabaseURL tests error handling for malformed database URLs.
func TestNewSQLRepository_InvalidDatabaseURL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7() //nolint:errcheck // UUID generation error virtually impossible
	testName := "TestNewSQLRepository_InvalidDatabaseURL_" + uuidVal.String()
	testSettings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	// Disable dev mode and set invalid database URL.
	testSettings.DevMode = false
	testSettings.DatabaseURL = "not-a-valid-url"

	sqlRepo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, testSettings)

	testify.Error(t, err)
	testify.Nil(t, sqlRepo)
	testify.Contains(t, err.Error(), "failed to determine database type")
}

// TestNewSQLRepository_EmptyDatabaseURL tests error handling for empty database URL.
func TestNewSQLRepository_EmptyDatabaseURL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7() //nolint:errcheck // UUID generation error virtually impossible
	testName := "TestNewSQLRepository_EmptyDatabaseURL_" + uuidVal.String()
	testSettings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	// Disable dev mode and set empty database URL.
	testSettings.DevMode = false
	testSettings.DatabaseURL = ""

	sqlRepo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, testSettings)

	testify.Error(t, err)
	testify.Nil(t, sqlRepo)
	testify.Contains(t, err.Error(), "unsupported database URL format")
}

// TestNewSQLRepository_ContainerModePreferred tests container mode "preferred" behavior.
func TestNewSQLRepository_ContainerModePreferred(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7() //nolint:errcheck // UUID generation error virtually impossible
	testName := "TestNewSQLRepository_ContainerModePreferred_" + uuidVal.String()
	testSettings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	// Set container mode to preferred (will fail since container not available for SQLite).
	testSettings.DatabaseContainer = containerModePreferred

	sqlRepo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, testSettings)

	// Container mode not supported for SQLite even in "preferred" mode.
	testify.Error(t, err)
	testify.Nil(t, sqlRepo)
	testify.Contains(t, err.Error(), "container option not available for sqlite")
}

// TestNewSQLRepository_ContainerModeRequired tests container mode "required" behavior.
func TestNewSQLRepository_ContainerModeRequired(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7() //nolint:errcheck // UUID generation error virtually impossible
	testName := "TestNewSQLRepository_ContainerModeRequired_" + uuidVal.String()
	testSettings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	// Set container mode to required (will fail since container not available for SQLite).
	testSettings.DatabaseContainer = containerModeRequired

	sqlRepo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, testSettings)

	// Should fail because container mode not available for SQLite.
	testify.Error(t, err)
	testify.Nil(t, sqlRepo)
	testify.Contains(t, err.Error(), "container option not available for sqlite")
}

// TestNewSQLRepository_VerboseMode tests verbose mode logging.
func TestNewSQLRepository_VerboseMode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7() //nolint:errcheck // UUID generation error virtually impossible
	testName := "TestNewSQLRepository_VerboseMode_" + uuidVal.String()
	testSettings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	// Enable verbose mode to trigger detailed logging.
	testSettings.VerboseMode = true

	sqlRepo := cryptoutilSQLRepository.RequireNewForTest(ctx, telemetryService, testSettings)
	defer sqlRepo.Shutdown()

	// Verbose mode should trigger schema logging during initialization.
	testify.NotNil(t, sqlRepo)
}
