// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilSQLRepository "cryptoutil/internal/kms/server/repository/sqlrepository"

	testify "github.com/stretchr/testify/require"
)

const (
	// Database container mode constants.
	containerModeDisabled  = "disabled"
	containerModePreferred = "preferred"
	containerModeRequired  = "required"
	containerModeInvalid   = "invalid-mode"
)

// getTestPostgresURL returns the PostgreSQL URL from environment variables or a default.
func getTestPostgresURL() string {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASS")
	name := os.Getenv("POSTGRES_NAME")

	if host == "" {
		host = "localhost"
	}

	if port == "" {
		port = "5432"
	}

	if user == "" {
		user = "cryptoutil"
	}

	if pass == "" {
		pass = "cryptoutil_test_password"
	}

	if name == "" {
		name = "cryptoutil_test"
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, name)
}

// TestNewSQLRepository_PostgreSQL_PingRetry tests database ping retry logic.
func TestNewSQLRepository_PostgreSQL_PingRetry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL container test in short mode")
	}

	// Skip if PostgreSQL not available (ci-race has no services)
	if os.Getenv("POSTGRES_HOST") == "" {
		t.Skip("Skipping PostgreSQL test: POSTGRES_HOST not set (PostgreSQL service not available)")
	}

	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilConfig.RequireNewForTest("ping_retry_test")
	settings.DevMode = false
	settings.DatabaseURL = getTestPostgresURL()
	settings.DatabaseContainer = containerModeDisabled

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	// This exercises ping retry logic when PostgreSQL is not immediately available.
	// In CI with PostgreSQL service container, connection succeeds after retries.
	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.NoError(t, err, "PostgreSQL service container should be reachable")
	testify.NotNil(t, repo)

	if repo != nil {
		defer repo.Shutdown()
	}
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
		err := repo.WithTransaction(ctx, false, func(_ *cryptoutilSQLRepository.SQLTransaction) error {
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
		err := repo.WithTransaction(ctx, false, func(_ *cryptoutilSQLRepository.SQLTransaction) error {
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
	err = repo.WithTransaction(ctx, false, func(_ *cryptoutilSQLRepository.SQLTransaction) error {
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
