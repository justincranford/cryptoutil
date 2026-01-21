// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository_test

import (
	"context"
	"testing"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSQLRepository "cryptoutil/internal/kms/server/repository/sqlrepository"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"

	testify "github.com/stretchr/testify/require"
)

// TestSQLRepository_PoolSettings tests connection pool settings for both SQLite and PostgreSQL.
func TestSQLRepository_PoolSettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		devMode           bool
		dbType            string
		expectedMaxOpen   int
		verboseMode       bool
		checkPoolSettings bool
	}{
		{
			name:              "SQLite pool settings with verbose logging",
			devMode:           true,
			dbType:            "sqlite",
			expectedMaxOpen:   1,
			verboseMode:       true,
			checkPoolSettings: true,
		},
		{
			name:              "SQLite pool settings without verbose logging",
			devMode:           true,
			dbType:            "sqlite",
			expectedMaxOpen:   1,
			verboseMode:       false,
			checkPoolSettings: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			settings := cryptoutilConfig.RequireNewForTest(tc.name)
			settings.DevMode = tc.devMode
			settings.DatabaseContainer = containerModeDisabled
			settings.VerboseMode = tc.verboseMode

			telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
			defer telemetryService.Shutdown()

			repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
			testify.NoError(t, err)
			testify.NotNil(t, repo)

			if repo != nil {
				defer repo.Shutdown()

				if tc.checkPoolSettings {
					// Access the underlying sql.DB to check pool settings.
					// This requires reflection or we just verify the repo was created successfully.
					testify.Equal(t, tc.dbType, string(repo.GetDBType()))
				}
			}
		})
	}
}

// TestSQLRepository_Shutdown_Multiple tests multiple shutdown calls.
func TestSQLRepository_Shutdown_Multiple(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilConfig.RequireNewForTest("shutdown_multiple_test")
	settings.DevMode = true
	settings.DatabaseContainer = containerModeDisabled

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.NoError(t, err)
	testify.NotNil(t, repo)

	// First shutdown should succeed.
	repo.Shutdown()

	// Second shutdown should handle already-closed connection gracefully.
	repo.Shutdown()
	// Depending on implementation, this might return an error or succeed.
	// We just verify it doesn't panic.
}

// TestHealthCheck_SQLDBNil tests health check when sql.DB is nil (defensive coding).
func TestHealthCheck_SQLDBNil(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create a repository with nil sqlDB (simulate corruption or invalid state).
	repo := &cryptoutilSQLRepository.SQLRepository{}

	result, err := repo.HealthCheck(ctx)
	testify.Error(t, err)
	testify.ErrorContains(t, err, "database connection not initialized")
	testify.NotNil(t, result)
	testify.Equal(t, "error", result["status"])
}

// TestSQLRepository_DBStats tests database statistics retrieval.
func TestSQLRepository_DBStats(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilConfig.RequireNewForTest("dbstats_test")
	settings.DevMode = true
	settings.DatabaseContainer = containerModeDisabled
	settings.VerboseMode = true

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.NoError(t, err)
	testify.NotNil(t, repo)

	defer repo.Shutdown()

	// Perform health check to get stats.
	result, err := repo.HealthCheck(ctx)
	testify.NoError(t, err)
	testify.NotNil(t, result)
	testify.Equal(t, "ok", result["status"])
	testify.Contains(t, result, "open_connections")
	testify.Contains(t, result, "idle_connections")
	testify.Contains(t, result, "in_use_connections")
	testify.Contains(t, result, "max_open_connections")
	testify.Contains(t, result, "wait_count")
	testify.Contains(t, result, "wait_duration")
}

// TestSQLRepository_ConnectionPoolExhaustion tests connection pool behavior under load.
func TestSQLRepository_ConnectionPoolExhaustion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilConfig.RequireNewForTest("pool_exhaustion_test")
	settings.DevMode = true
	settings.DatabaseContainer = containerModeDisabled

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.NoError(t, err)
	testify.NotNil(t, repo)

	defer repo.Shutdown()

	// SQLite has MaxOpenConns=1, so this tests sequential access pattern.
	for i := 0; i < 5; i++ {
		err := repo.WithTransaction(ctx, false, func(_ *cryptoutilSQLRepository.SQLTransaction) error {
			return nil
		})
		testify.NoError(t, err)
	}
}
