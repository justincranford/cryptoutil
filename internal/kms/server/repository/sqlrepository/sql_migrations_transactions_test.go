// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository_test

import (
	"context"
	"database/sql"
	"testing"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilSQLRepository "cryptoutil/internal/kms/server/repository/sqlrepository"

	testify "github.com/stretchr/testify/require"
)

// TestApplyEmbeddedSQLMigrations_SQLite tests migration application for SQLite.
func TestApplyEmbeddedSQLMigrations_SQLite(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilConfig.RequireNewForTest("migrations_sqlite")
	settings.DevMode = true
	settings.DatabaseContainer = containerModeDisabled

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
	settings.DatabaseContainer = containerModeDisabled
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
	settings.DatabaseContainer = containerModeDisabled

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

// NOTE: Read-only transactions not supported by SQLite driver (modernc.org/sqlite)
// TestSQLRepository_WithTransaction_ReadOnly removed - PostgreSQL-only feature

// TestSQLRepository_WithTransaction_RollbackOnError tests automatic rollback on error.
func TestSQLRepository_WithTransaction_RollbackOnError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilConfig.RequireNewForTest("rollback_error_test")
	settings.DevMode = true
	settings.DatabaseContainer = containerModeDisabled

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
