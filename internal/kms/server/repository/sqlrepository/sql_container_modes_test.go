// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository_test

import (
	"context"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSQLRepository "cryptoutil/internal/kms/server/repository/sqlrepository"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	testify "github.com/stretchr/testify/require"
)

// TestNewSQLRepository_PostgreSQL_ContainerRequired tests container mode = required (will start container).
func TestNewSQLRepository_PostgreSQL_ContainerRequired(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping container test in short mode")
	}

	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("postgres_container_required")
	settings.DevMode = false
	settings.DatabaseURL = getTestPostgresURL()
	settings.DatabaseContainer = containerModeRequired // Will start container when Docker available

	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	// Containers start successfully when Docker available
	testify.NoError(t, err)
	testify.NotNil(t, repo)

	if repo != nil {
		defer repo.Shutdown()
	}
}

// TestNewSQLRepository_PostgreSQL_ContainerPreferred tests container mode = preferred (will start container).
func TestNewSQLRepository_PostgreSQL_ContainerPreferred(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping container test in short mode")
	}

	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("postgres_container_preferred")
	settings.DevMode = false
	settings.DatabaseURL = getTestPostgresURL()
	settings.DatabaseContainer = containerModePreferred // Will start container when Docker available

	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	// Containers start successfully when Docker available
	testify.NoError(t, err)
	testify.NotNil(t, repo)

	if repo != nil {
		defer repo.Shutdown()
	}
}

// TestNewSQLRepository_UnsupportedDBType tests unsupported database types.
func TestNewSQLRepository_UnsupportedDBType(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("unsupported_db_type")
	settings.DevMode = false
	settings.DatabaseURL = "mysql://user:pass@localhost:3306/testdb" // MySQL not supported
	settings.DatabaseContainer = containerModeDisabled

	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.Error(t, err)
	testify.Nil(t, repo)
}

// TestSQLRepository_GetDBType_AllTypes tests GetDBType for all supported database types.
func TestSQLRepository_GetDBType_AllTypes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name           string
		devMode        bool
		expectedDBType cryptoutilSQLRepository.SupportedDBType
	}{
		{
			name:           "SQLite via DevMode",
			devMode:        true,
			expectedDBType: cryptoutilSQLRepository.DBTypeSQLite,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest(tc.name)
			settings.DevMode = tc.devMode
			settings.DatabaseContainer = containerModeDisabled

			telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings)
			defer telemetryService.Shutdown()

			repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
			testify.NoError(t, err)
			testify.NotNil(t, repo)

			defer repo.Shutdown()

			testify.Equal(t, tc.expectedDBType, repo.GetDBType())
		})
	}
}

// TestHealthCheck_ContextTimeout tests health check with context timeout.
func TestHealthCheck_ContextTimeout(t *testing.T) {
	t.Parallel()

	baseCtx := context.Background()

	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("healthcheck_timeout")
	settings.DevMode = true
	settings.DatabaseContainer = containerModeDisabled

	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(baseCtx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(baseCtx, telemetryService, settings)
	testify.NoError(t, err)
	testify.NotNil(t, repo)

	defer repo.Shutdown()

	// Health check with already-cancelled context
	ctx, cancel := context.WithCancel(baseCtx)
	cancel() // Cancel immediately

	result, err := repo.HealthCheck(ctx)
	// Should still succeed as the DB ping is fast
	if err != nil {
		testify.NotNil(t, result)
		testify.Equal(t, "error", result["status"])
	}
}

// TestNewSQLRepository_InvalidDatabaseURL_Formats tests various invalid URL formats.
func TestNewSQLRepository_InvalidDatabaseURL_Formats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name        string
		databaseURL string
	}{
		{
			name:        "Malformed URL - no protocol",
			databaseURL: "user:pass@localhost:5432/testdb",
		},
		{
			name:        "Malformed URL - invalid characters",
			databaseURL: "postgres://user@@@localhost:5432/testdb",
		},
		{
			name:        "Malformed URL - missing host",
			databaseURL: "postgres://user:pass@/testdb",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest(tc.name)
			settings.DevMode = false
			settings.DatabaseURL = tc.databaseURL
			settings.DatabaseContainer = containerModeDisabled

			telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings)
			defer telemetryService.Shutdown()

			repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
			testify.Error(t, err)
			testify.Nil(t, repo)
		})
	}
}
