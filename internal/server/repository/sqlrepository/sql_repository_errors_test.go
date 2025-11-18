package sqlrepository_test

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	testify "github.com/stretchr/testify/require"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	"cryptoutil/internal/server/repository/sqlrepository"
)

// TestNewSQLRepository_InvalidDatabaseURL tests error handling for malformed database URLs.
func TestNewSQLRepository_InvalidDatabaseURL(t *testing.T) {
	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7()
	testName := "TestNewSQLRepository_InvalidDatabaseURL_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	// Disable dev mode and set invalid database URL.
	testSettings.DevMode = false
	testSettings.DatabaseURL = "not-a-valid-url"

	sqlRepo, err := sqlrepository.NewSQLRepository(ctx, telemetryService, testSettings)

	testify.Error(t, err)
	testify.Nil(t, sqlRepo)
	testify.Contains(t, err.Error(), "failed to determine database type")
}

// TestNewSQLRepository_EmptyDatabaseURL tests error handling for empty database URL.
func TestNewSQLRepository_EmptyDatabaseURL(t *testing.T) {
	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7()
	testName := "TestNewSQLRepository_EmptyDatabaseURL_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	// Disable dev mode and set empty database URL.
	testSettings.DevMode = false
	testSettings.DatabaseURL = ""

	sqlRepo, err := sqlrepository.NewSQLRepository(ctx, telemetryService, testSettings)

	testify.Error(t, err)
	testify.Nil(t, sqlRepo)
	testify.Contains(t, err.Error(), "unsupported database URL format")
}

// TestNewSQLRepository_ContainerModePreferred tests container mode "preferred" behavior.
func TestNewSQLRepository_ContainerModePreferred(t *testing.T) {
	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7()
	testName := "TestNewSQLRepository_ContainerModePreferred_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	// Set container mode to preferred (will fail since container not available for SQLite).
	testSettings.DatabaseContainer = "preferred"

	sqlRepo, err := sqlrepository.NewSQLRepository(ctx, telemetryService, testSettings)

	// Container mode not supported for SQLite even in "preferred" mode.
	testify.Error(t, err)
	testify.Nil(t, sqlRepo)
	testify.Contains(t, err.Error(), "container option not available for sqlite")
}

// TestNewSQLRepository_ContainerModeRequired tests container mode "required" behavior.
func TestNewSQLRepository_ContainerModeRequired(t *testing.T) {
	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7()
	testName := "TestNewSQLRepository_ContainerModeRequired_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	// Set container mode to required (will fail since container not available for SQLite).
	testSettings.DatabaseContainer = "required"

	sqlRepo, err := sqlrepository.NewSQLRepository(ctx, telemetryService, testSettings)

	// Should fail because container mode not available for SQLite.
	testify.Error(t, err)
	testify.Nil(t, sqlRepo)
	testify.Contains(t, err.Error(), "container option not available for sqlite")
}

// TestNewSQLRepository_VerboseMode tests verbose mode logging.
func TestNewSQLRepository_VerboseMode(t *testing.T) {
	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7()
	testName := "TestNewSQLRepository_VerboseMode_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	// Enable verbose mode to trigger detailed logging.
	testSettings.VerboseMode = true

	sqlRepo := sqlrepository.RequireNewForTest(ctx, telemetryService, testSettings)
	defer sqlRepo.Shutdown()

	// Verbose mode should trigger schema logging during initialization.
	testify.NotNil(t, sqlRepo)
}
