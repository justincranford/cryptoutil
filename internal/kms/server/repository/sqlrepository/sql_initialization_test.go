// Copyright (c) 2025 Justin Cranford

package sqlrepository

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// TestNewSQLRepository_NilContext tests NewSQLRepository with nil context.
func TestNewSQLRepository_NilContext(t *testing.T) {
	t.Parallel()

	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("sql_init_nil_ctx")

	telemetry := cryptoutilSharedTelemetry.RequireNewForTest(context.Background(), settings)
	defer telemetry.Shutdown()

	// Call NewSQLRepository with nil context.
	repo, err := NewSQLRepository(nil, telemetry, settings) //nolint:staticcheck // Testing nil context error handling
	require.Error(t, err, "NewSQLRepository should fail with nil context")
	require.Nil(t, repo, "Repository should be nil on error")
	require.Contains(t, err.Error(), "ctx must be non-nil", "Error should indicate nil context")
}

// TestNewSQLRepository_NilTelemetry tests NewSQLRepository with nil telemetry service.
func TestNewSQLRepository_NilTelemetry(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("sql_init_nil_telemetry")

	// Call NewSQLRepository with nil telemetry.
	repo, err := NewSQLRepository(ctx, nil, settings)
	require.Error(t, err, "NewSQLRepository should fail with nil telemetry")
	require.Nil(t, repo, "Repository should be nil on error")
	require.Contains(t, err.Error(), "telemetryService must be non-nil", "Error should indicate nil telemetry")
}

// TestNewSQLRepository_NilSettings tests NewSQLRepository with nil settings.
func TestNewSQLRepository_NilSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("sql_init_nil_settings")

	telemetry := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings)
	defer telemetry.Shutdown()

	// Call NewSQLRepository with nil settings.
	repo, err := NewSQLRepository(ctx, telemetry, nil)
	require.Error(t, err, "NewSQLRepository should fail with nil settings")
	require.Nil(t, repo, "Repository should be nil on error")
	require.Contains(t, err.Error(), "settings must be non-nil", "Error should indicate nil settings")
}

// TestNewTransaction_NilRepository tests newTransaction with nil repository.
func TestNewTransaction_NilRepository(t *testing.T) {
	t.Parallel()

	var repo *SQLRepository

	// Call newTransaction on nil repository.
	tx, err := repo.newTransaction()
	require.Error(t, err, "newTransaction should fail with nil repository")
	require.Nil(t, tx, "Transaction should be nil on error")
	require.Contains(t, err.Error(), "SQL repository cannot be nil", "Error should indicate nil repository")
}

// TestWithTransaction_ReadOnlySQLite tests WithTransaction attempting read-only transaction on SQLite.
func TestWithTransaction_ReadOnlySQLite(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("sql_transaction_readonly_sqlite")

	telemetry := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings)
	defer telemetry.Shutdown()

	// Create SQLite repository (dev mode uses SQLite).
	repo := RequireNewForTest(ctx, telemetry, settings)
	defer repo.Shutdown()

	// Attempt read-only transaction on SQLite (should fail - SQLite doesn't support read-only transactions).
	err := repo.WithTransaction(ctx, true, func(_ *SQLTransaction) error {
		return nil
	})
	require.Error(t, err, "WithTransaction should fail for read-only on SQLite")
	require.Contains(t, err.Error(), "doesn't support read-only transactions", "Error should indicate no read-only support")
}

// TestWithTransaction_FunctionError tests WithTransaction with function returning error.
func TestWithTransaction_FunctionError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("sql_transaction_func_error")

	telemetry := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings)
	defer telemetry.Shutdown()

	// Create SQLite repository.
	repo := RequireNewForTest(ctx, telemetry, settings)
	defer repo.Shutdown()

	testErr := fmt.Errorf("test transaction function error")

	// Call WithTransaction with function that returns error.
	err := repo.WithTransaction(ctx, false, func(_ *SQLTransaction) error {
		return testErr
	})
	require.Error(t, err, "WithTransaction should propagate function error")
	require.Contains(t, err.Error(), "failed to execute transaction", "Error should indicate transaction execution failure")
	require.ErrorIs(t, err, testErr, "Error should wrap original error")
}

// TestWithTransaction_PanicRecovery tests WithTransaction panic recovery and rollback.
func TestWithTransaction_PanicRecovery(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("sql_transaction_panic")

	telemetry := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings)
	defer telemetry.Shutdown()

	// Create SQLite repository.
	repo := RequireNewForTest(ctx, telemetry, settings)
	defer repo.Shutdown()

	// Use defer to recover from panic.
	defer func() {
		if r := recover(); r != nil {
			require.NotNil(t, r, "Panic should be recovered")
			require.Equal(t, "simulated panic in transaction", r, "Panic value should match")
		}
	}()

	// Call WithTransaction with function that panics.
	_ = repo.WithTransaction(ctx, false, func(_ *SQLTransaction) error { //nolint:errcheck // Panic test - error handling not applicable
		panic("simulated panic in transaction")
	})

	// Should not reach here because panic is re-thrown.
	t.Fatal("Should not reach here - panic should be re-thrown")
}
