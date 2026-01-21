// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository_test

import (
	"context"
	"testing"

	"cryptoutil/internal/kms/server/repository/sqlrepository"

	googleUuid "github.com/google/uuid"
	testify "github.com/stretchr/testify/require"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

// TestNewSQLRepository_ErrorPaths tests various error conditions during repository creation.
func TestNewSQLRepository_ErrorPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name        string
		setupConfig func() *cryptoutilConfig.ServiceTemplateServerSettings
		expectError bool
		errorText   string
	}{
		{
			name: "Nil context",
			setupConfig: func() *cryptoutilConfig.ServiceTemplateServerSettings {
				settings := cryptoutilConfig.RequireNewForTest("nil_ctx_test")
				settings.DevMode = true
				settings.DatabaseContainer = containerModeDisabled

				return settings
			},
			expectError: true,
			errorText:   "ctx must be non-nil",
		},
		{
			name: "Nil telemetry service",
			setupConfig: func() *cryptoutilConfig.ServiceTemplateServerSettings {
				settings := cryptoutilConfig.RequireNewForTest("nil_telemetry_test")
				settings.DevMode = true
				settings.DatabaseContainer = containerModeDisabled

				return settings
			},
			expectError: true,
			errorText:   "telemetryService must be non-nil",
		},
		{
			name: "Nil settings",
			setupConfig: func() *cryptoutilConfig.ServiceTemplateServerSettings {
				return nil
			},
			expectError: true,
			errorText:   "settings must be non-nil",
		},
		{
			name: "Empty database URL (non-dev mode)",
			setupConfig: func() *cryptoutilConfig.ServiceTemplateServerSettings {
				settings := cryptoutilConfig.RequireNewForTest("empty_url_test")
				settings.DevMode = false // Production mode
				settings.DatabaseURL = ""
				settings.DatabaseContainer = containerModeDisabled

				return settings
			},
			expectError: true,
			errorText:   "unsupported database URL format",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			settings := tc.setupConfig()

			var repo *sqlrepository.SQLRepository

			var err error

			if settings == nil {
				// Test nil settings
				telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, cryptoutilConfig.RequireNewForTest("temp"))
				defer telemetryService.Shutdown()

				repo, err = sqlrepository.NewSQLRepository(ctx, telemetryService, nil)
			} else if tc.name == "Nil context" {
				// Test nil context
				telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
				defer telemetryService.Shutdown()

				repo, err = sqlrepository.NewSQLRepository(nil, telemetryService, settings) //nolint:staticcheck // Testing nil context error handling
			} else if tc.name == "Nil telemetry service" {
				// Test nil telemetry service
				repo, err = sqlrepository.NewSQLRepository(ctx, nil, settings)
			} else {
				// Normal test
				telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
				defer telemetryService.Shutdown()

				repo, err = sqlrepository.NewSQLRepository(ctx, telemetryService, settings)
			}

			if tc.expectError {
				testify.Error(t, err)
				testify.ErrorContains(t, err, tc.errorText)
				testify.Nil(t, repo)
			} else {
				testify.NoError(t, err)
				testify.NotNil(t, repo)

				if repo != nil {
					defer repo.Shutdown()
				}
			}
		})
	}
}

// TestSQLRepository_WithTransaction_NilContext tests transaction with nil context.
func TestSQLRepository_WithTransaction_NilContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7() //nolint:errcheck // UUID generation error virtually impossible
	testName := "TestWithTransaction_NilContext_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	sqlRepo := sqlrepository.RequireNewForTest(ctx, telemetryService, testSettings)
	defer sqlRepo.Shutdown()

	// Test with nil context.
	err := sqlRepo.WithTransaction(nil, false, func(_ *sqlrepository.SQLTransaction) error { //nolint:staticcheck // Testing nil context error handling
		return nil
	})

	testify.Error(t, err)
	testify.ErrorContains(t, err, "context cannot be nil")
}

// TestSQLRepository_WithTransaction_NilFunction tests transaction with nil function.
func TestSQLRepository_WithTransaction_NilFunction(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7() //nolint:errcheck // UUID generation error virtually impossible
	testName := "TestWithTransaction_NilFunction_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	sqlRepo := sqlrepository.RequireNewForTest(ctx, telemetryService, testSettings)
	defer sqlRepo.Shutdown()

	// Test with nil function.
	err := sqlRepo.WithTransaction(ctx, false, nil)

	testify.Error(t, err)
	testify.ErrorContains(t, err, "function cannot be nil")
}

// TestSQLTransaction_PublicMethods tests all public methods on SQLTransaction.
func TestSQLTransaction_PublicMethods(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7() //nolint:errcheck // UUID generation error virtually impossible
	testName := "TestSQLTransaction_PublicMethods_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	sqlRepo := sqlrepository.RequireNewForTest(ctx, telemetryService, testSettings)
	defer sqlRepo.Shutdown()

	// Test all public methods.
	err := sqlRepo.WithTransaction(ctx, false, func(tx *sqlrepository.SQLTransaction) error {
		// Test TransactionID().
		txID := tx.TransactionID()
		testify.NotNil(t, txID)

		// Test Context().
		txCtx := tx.Context()
		testify.NotNil(t, txCtx)
		testify.Equal(t, ctx, txCtx)

		// Test IsReadOnly().
		isReadOnly := tx.IsReadOnly()
		testify.False(t, isReadOnly)

		return nil
	})

	testify.NoError(t, err)

	// Test read-only transaction IsReadOnly() returns true (if supported).
	// Note: SQLite doesn't support read-only transactions, so this will fail.
	// This test is here to demonstrate the code path.
	err = sqlRepo.WithTransaction(ctx, true, func(tx *sqlrepository.SQLTransaction) error {
		isReadOnly := tx.IsReadOnly()
		testify.True(t, isReadOnly)

		return nil
	})
	// SQLite should fail with "database sqlite doesn't support read-only transactions".
	if err != nil {
		testify.ErrorContains(t, err, "doesn't support read-only transactions")
	}
}

// TestSQLRepository_Shutdown_MultipleCallsSafe tests that multiple shutdown calls are safe.
func TestSQLRepository_Shutdown_MultipleCallsSafe(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	uuidVal, _ := googleUuid.NewV7() //nolint:errcheck // UUID generation error virtually impossible
	testName := "TestShutdown_MultipleCalls_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	sqlRepo := sqlrepository.RequireNewForTest(ctx, telemetryService, testSettings)

	// Call shutdown multiple times (should be safe).
	sqlRepo.Shutdown()
	sqlRepo.Shutdown()
	sqlRepo.Shutdown()
	// No panic expected.
}

// TestSQLRepository_GetDBType_SQLiteOnly tests GetDBType for SQLite.
func TestSQLRepository_GetDBType_SQLiteOnly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Test SQLite.
	uuidVal, _ := googleUuid.NewV7() //nolint:errcheck // UUID generation error virtually impossible
	testName := "TestGetDBType_SQLite_" + uuidVal.String()
	testSettings := cryptoutilConfig.RequireNewForTest(testName)
	testSettings.DevMode = true
	testSettings.DatabaseContainer = containerModeDisabled

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, testSettings)
	defer telemetryService.Shutdown()

	sqlRepo := sqlrepository.RequireNewForTest(ctx, telemetryService, testSettings)
	defer sqlRepo.Shutdown()

	testify.Equal(t, sqlrepository.DBTypeSQLite, sqlRepo.GetDBType())
	// PostgreSQL test would require actual PostgreSQL instance, so skip it.
	// The code path is covered by sql_postgres_coverage_test.go (CI/CD only).
}
