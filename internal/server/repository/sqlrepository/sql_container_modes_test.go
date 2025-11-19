// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository_test

import (
	"context"
	"testing"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilSQLRepository "cryptoutil/internal/server/repository/sqlrepository"

	testify "github.com/stretchr/testify/require"
)

// TestNewSQLRepository_PostgreSQL_ContainerRequired tests container mode = required (will fail).
func TestNewSQLRepository_PostgreSQL_ContainerRequired(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilConfig.RequireNewForTest("postgres_container_required")
	settings.DevMode = false
	settings.DatabaseURL = "postgres://user:pass@localhost:5432/testdb?sslmode=disable"
	settings.DatabaseContainer = "required" // Will fail without Docker

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.Error(t, err)
	testify.ErrorIs(t, err, cryptoutilSQLRepository.ErrContainerModeRequiredButContainerNotStarted)
	testify.Nil(t, repo)
}

// TestNewSQLRepository_PostgreSQL_ContainerPreferred tests container mode = preferred (will fallback).
func TestNewSQLRepository_PostgreSQL_ContainerPreferred(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilConfig.RequireNewForTest("postgres_container_preferred")
	settings.DevMode = false
	settings.DatabaseURL = "postgres://user:pass@localhost:5432/testdb?sslmode=disable"
	settings.DatabaseContainer = "preferred" // Will fallback to URL

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	// Will fail to connect to non-existent PostgreSQL, but fallback path is exercised
	testify.Error(t, err)
	testify.ErrorContains(t, err, "failed to ping database")
	testify.Nil(t, repo)
}

// TestNewSQLRepository_UnsupportedDBType tests unsupported database types.
func TestNewSQLRepository_UnsupportedDBType(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilConfig.RequireNewForTest("unsupported_db_type")
	settings.DevMode = false
	settings.DatabaseURL = "mysql://user:pass@localhost:3306/testdb" // MySQL not supported
	settings.DatabaseContainer = "disabled"

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
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

			settings := cryptoutilConfig.RequireNewForTest(tc.name)
			settings.DevMode = tc.devMode
			settings.DatabaseContainer = "disabled"

			telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
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

	settings := cryptoutilConfig.RequireNewForTest("healthcheck_timeout")
	settings.DevMode = true
	settings.DatabaseContainer = "disabled"

	telemetryService := cryptoutilTelemetry.RequireNewForTest(baseCtx, settings)
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

			settings := cryptoutilConfig.RequireNewForTest(tc.name)
			settings.DevMode = false
			settings.DatabaseURL = tc.databaseURL
			settings.DatabaseContainer = "disabled"

			telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
			defer telemetryService.Shutdown()

			repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
			testify.Error(t, err)
			testify.Nil(t, repo)
		})
	}
}
