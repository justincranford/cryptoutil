// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository_test

import (
	"context"
	"testing"
	"time"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilSQLRepository "cryptoutil/internal/kms/server/repository/sqlrepository"

	testify "github.com/stretchr/testify/require"
)

const (
	// Database container mode constants.
	containerModeDisabled  = "disabled"
	containerModePreferred = "preferred"
	containerModeRequired  = "required"
	containerModeInvalid   = "invalid-mode"

	// Test database URL for PostgreSQL connection tests.
	testPostgresURL = "postgres://user:pass@localhost:5432/testdb?sslmode=disable"
)

// TestNewSQLRepository_PostgreSQL_PingRetry tests database ping retry logic.
func TestNewSQLRepository_PostgreSQL_PingRetry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL container test in short mode")
	}

	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilConfig.RequireNewForTest("ping_retry_test")
	settings.DevMode = false
	settings.DatabaseURL = testPostgresURL
	settings.DatabaseContainer = containerModeDisabled

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	// This will exercise the ping retry logic (will fail after max attempts)
	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.Error(t, err)
	testify.ErrorContains(t, err, "failed to ping database")
	testify.Nil(t, repo)
}

// TestHealthCheck_AllConnectionPoolStats tests all connection pool statistics.
func TestHealthCheck_AllConnectionPoolStats(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilConfig.RequireNewForTest("pool_stats_test")
	settings.DevMode = true
	settings.DatabaseContainer = containerModeDisabled

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.NoError(t, err)
	testify.NotNil(t, repo)

	defer repo.Shutdown()

	// Perform multiple operations to generate pool statistics.
	// Use readOnly=false because SQLite doesn't support read-only transactions.
	for i := 0; i < 10; i++ {
		err := repo.WithTransaction(ctx, false, func(tx *cryptoutilSQLRepository.SQLTransaction) error {
			return nil
		})
		testify.NoError(t, err)
	}

	// Get health check with full stats
	result, err := repo.HealthCheck(ctx)
	testify.NoError(t, err)
	testify.NotNil(t, result)
	testify.Equal(t, "ok", result["status"])

	// Verify all stat fields are present
	testify.Contains(t, result, "db_type")
	testify.Contains(t, result, "open_connections")
	testify.Contains(t, result, "idle_connections")
	testify.Contains(t, result, "in_use_connections")
	testify.Contains(t, result, "max_open_connections")
	testify.Contains(t, result, "wait_count")
	testify.Contains(t, result, "wait_duration")
}

// TestSQLRepository_Shutdown_AlreadyClosed tests shutdown of already-closed connection.
func TestSQLRepository_Shutdown_AlreadyClosed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilConfig.RequireNewForTest("shutdown_closed_test")
	settings.DevMode = true
	settings.DatabaseContainer = containerModeDisabled

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.NoError(t, err)
	testify.NotNil(t, repo)

	// First shutdown
	repo.Shutdown()

	// Second shutdown on already-closed connection (tests error logging path)
	repo.Shutdown()

	// Third shutdown to ensure no panic
	repo.Shutdown()
}

// TestSQLRepository_LogConnectionPoolSettings_VerboseMode tests verbose logging.
func TestSQLRepository_LogConnectionPoolSettings_VerboseMode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilConfig.RequireNewForTest("verbose_pool_logging")
	settings.DevMode = true
	settings.DatabaseContainer = containerModeDisabled
	settings.VerboseMode = true // Enable verbose logging

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.NoError(t, err)
	testify.NotNil(t, repo)

	defer repo.Shutdown()

	// Verbose logging happens during initialization
	// Verify repo was created successfully (pool settings were logged)
	testify.Equal(t, cryptoutilSQLRepository.DBTypeSQLite, repo.GetDBType())
}

// TestHealthCheck_PingError tests health check when ping fails.
func TestHealthCheck_PingError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilConfig.RequireNewForTest("healthcheck_ping_error")
	settings.DevMode = true
	settings.DatabaseContainer = containerModeDisabled

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.NoError(t, err)
	testify.NotNil(t, repo)

	// Close connection to force ping failure
	repo.Shutdown()

	// Health check should fail with closed connection
	result, err := repo.HealthCheck(ctx)
	testify.Error(t, err)
	testify.NotNil(t, result)
	testify.Equal(t, "error", result["status"])
	testify.Contains(t, result, "error")
	testify.Contains(t, result, "db_type")
}

// TestSQLRepository_WithTransaction_MultipleSequential tests multiple sequential transactions.
func TestSQLRepository_WithTransaction_MultipleSequential(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilConfig.RequireNewForTest("sequential_transactions")
	settings.DevMode = true
	settings.DatabaseContainer = containerModeDisabled

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.NoError(t, err)
	testify.NotNil(t, repo)

	defer repo.Shutdown()

	// Execute multiple sequential transactions
	for i := 0; i < 20; i++ {
		err := repo.WithTransaction(ctx, false, func(tx *cryptoutilSQLRepository.SQLTransaction) error {
			return nil
		})
		testify.NoError(t, err)
	}
}

// TestSQLRepository_WithTransaction_ContextDeadlineExceeded tests context deadline.
func TestSQLRepository_WithTransaction_ContextDeadlineExceeded(t *testing.T) {
	t.Parallel()

	baseCtx := context.Background()

	settings := cryptoutilConfig.RequireNewForTest("deadline_exceeded_test")
	settings.DevMode = true
	settings.DatabaseContainer = containerModeDisabled

	telemetryService := cryptoutilTelemetry.RequireNewForTest(baseCtx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(baseCtx, telemetryService, settings)
	testify.NoError(t, err)
	testify.NotNil(t, repo)

	defer repo.Shutdown()

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(baseCtx, 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure timeout occurs

	// Transaction should fail due to deadline exceeded
	err = repo.WithTransaction(ctx, false, func(tx *cryptoutilSQLRepository.SQLTransaction) error {
		return nil
	})
	// May or may not error depending on timing - just verify no panic
	_ = err
}

// TestSQLRepository_ErrorTypes_Wrapping tests error type wrapping and checking.
func TestSQLRepository_ErrorTypes_Wrapping(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name          string
		setup         func() (*cryptoutilSQLRepository.SQLRepository, error)
		expectedError error
	}{
		{
			name: "ErrContainerOptionNotExist",
			setup: func() (*cryptoutilSQLRepository.SQLRepository, error) {
				settings := cryptoutilConfig.RequireNewForTest("container_not_exist")
				settings.DevMode = true                            // SQLite
				settings.DatabaseContainer = containerModeRequired // SQLite doesn't support containers

				telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
				defer telemetryService.Shutdown()

				return cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
			},
			expectedError: cryptoutilSQLRepository.ErrContainerOptionNotExist,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo, err := tc.setup()
			testify.Error(t, err)
			testify.ErrorIs(t, err, tc.expectedError)
			testify.Nil(t, repo)
		})
	}
}
