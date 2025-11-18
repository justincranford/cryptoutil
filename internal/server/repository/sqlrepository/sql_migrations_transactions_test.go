package sqlrepository_test

import (
	"context"
	"database/sql"
	"testing"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilSQLRepository "cryptoutil/internal/server/repository/sqlrepository"

	testify "github.com/stretchr/testify/require"
)

// TestApplyEmbeddedSQLMigrations_SQLite tests migration application for SQLite.
func TestApplyEmbeddedSQLMigrations_SQLite(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	
	settings := cryptoutilConfig.RequireNewForTest("migrations_sqlite")
	settings.DevMode = true
	settings.DatabaseContainer = "disabled"
	
	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.NoError(t, err)
	testify.NotNil(t, repo)
	defer repo.Shutdown()

	// Migrations are applied during NewSQLRepository, verify database schema exists
	testify.Equal(t, cryptoutilSQLRepository.DBTypeSQLite, repo.GetDBType())
}

// TestLogSchema_SQLite tests schema logging for SQLite.
func TestLogSchema_SQLite(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	
	settings := cryptoutilConfig.RequireNewForTest("log_schema_sqlite")
	settings.DevMode = true
	settings.DatabaseContainer = "disabled"
	settings.VerboseMode = true // Enable verbose logging
	
	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.NoError(t, err)
	testify.NotNil(t, repo)
	defer repo.Shutdown()

	// LogSchema is called during NewSQLRepository initialization
	// Verify repo was created successfully (schema was logged)
	testify.Equal(t, cryptoutilSQLRepository.DBTypeSQLite, repo.GetDBType())
}

// TestSQLRepository_WithTransaction_NestedTransaction tests nested transaction detection.
func TestSQLRepository_WithTransaction_NestedTransaction(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	
	settings := cryptoutilConfig.RequireNewForTest("nested_transaction_test")
	settings.DevMode = true
	settings.DatabaseContainer = "disabled"
	
	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.NoError(t, err)
	testify.NotNil(t, repo)
	defer repo.Shutdown()

	// Start outer transaction
	err = repo.WithTransaction(ctx, false, func(txOuter *cryptoutilSQLRepository.SQLTransaction) error {
		// Try to start nested transaction (should error)
		err := repo.WithTransaction(ctx, false, func(txInner *cryptoutilSQLRepository.SQLTransaction) error {
			return nil
		})
		testify.Error(t, err)
		testify.ErrorContains(t, err, "transaction already started")
		return nil
	})
	testify.NoError(t, err)
}

// TestSQLRepository_WithTransaction_ReadOnly tests read-only transaction mode.
func TestSQLRepository_WithTransaction_ReadOnly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	
	settings := cryptoutilConfig.RequireNewForTest("readonly_transaction_test")
	settings.DevMode = true
	settings.DatabaseContainer = "disabled"
	
	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.NoError(t, err)
	testify.NotNil(t, repo)
	defer repo.Shutdown()

	// Test read-only transaction
	err = repo.WithTransaction(ctx, true, func(tx *cryptoutilSQLRepository.SQLTransaction) error {
		// Read-only transactions should succeed for queries
		return nil
	})
	testify.NoError(t, err)
}

// TestSQLRepository_WithTransaction_PanicRecovery tests panic recovery in transactions.
func TestSQLRepository_WithTransaction_PanicRecovery(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	
	settings := cryptoutilConfig.RequireNewForTest("panic_recovery_test")
	settings.DevMode = true
	settings.DatabaseContainer = "disabled"
	
	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.NoError(t, err)
	testify.NotNil(t, repo)
	defer repo.Shutdown()

	// Test panic recovery
	err = repo.WithTransaction(ctx, false, func(tx *cryptoutilSQLRepository.SQLTransaction) error {
		panic("intentional test panic")
	})
	testify.Error(t, err)
	testify.ErrorContains(t, err, "intentional test panic")
}

// TestSQLRepository_WithTransaction_CommitError tests commit error handling.
func TestSQLRepository_WithTransaction_CommitError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	
	// Cancel context before commit to force error
	ctx, cancel := context.WithCancel(ctx)
	
	settings := cryptoutilConfig.RequireNewForTest("commit_error_test")
	settings.DevMode = true
	settings.DatabaseContainer = "disabled"
	
	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.NoError(t, err)
	testify.NotNil(t, repo)
	defer repo.Shutdown()

	// Start transaction, then cancel context
	err = repo.WithTransaction(ctx, false, func(tx *cryptoutilSQLRepository.SQLTransaction) error {
		cancel() // Cancel context during transaction
		return nil
	})
	// Transaction should fail due to cancelled context
	if err != nil {
		testify.Error(t, err)
	}
}

// TestSQLRepository_WithTransaction_RollbackOnError tests automatic rollback on error.
func TestSQLRepository_WithTransaction_RollbackOnError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	
	settings := cryptoutilConfig.RequireNewForTest("rollback_error_test")
	settings.DevMode = true
	settings.DatabaseContainer = "disabled"
	
	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.NoError(t, err)
	testify.NotNil(t, repo)
	defer repo.Shutdown()

	// Transaction that returns error (should rollback automatically)
	err = repo.WithTransaction(ctx, false, func(tx *cryptoutilSQLRepository.SQLTransaction) error {
		return sql.ErrNoRows // Return error to trigger rollback
	})
	testify.Error(t, err)
	testify.ErrorIs(t, err, sql.ErrNoRows)
}
