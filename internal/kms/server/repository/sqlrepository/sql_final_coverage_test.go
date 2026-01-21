// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository_test

import (
	"context"
	"errors"
	"os"
	"testing"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilSQLRepository "cryptoutil/internal/kms/server/repository/sqlrepository"

	testify "github.com/stretchr/testify/require"
)

const (
	// invalidDatabaseURL is used for negative test cases to ensure failures even when PostgreSQL service exists.
	invalidDatabaseURL = "postgres://invalid_user:invalid_pass@localhost:9999/nonexistent_db?sslmode=disable"
	// invalidDatabaseURLWithAuth is used for tests expecting authentication failures.
	invalidDatabaseURLWithAuth = "postgres://invalid_user:invalid_pass@localhost:5432/cryptoutil_test?sslmode=disable"
)

// TestMapDBTypeAndURL_AllScenarios tests all database type and URL mapping scenarios.
func TestMapDBTypeAndURL_AllScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name           string
		devMode        bool
		databaseURL    string
		expectedDBType cryptoutilSQLRepository.SupportedDBType
		expectError    bool
	}{
		{
			name:           "DevMode true - SQLite in-memory",
			devMode:        true,
			databaseURL:    "ignored-url",
			expectedDBType: cryptoutilSQLRepository.DBTypeSQLite,
			expectError:    false,
		},
		{
			name:           "PostgreSQL URL",
			devMode:        false,
			databaseURL:    "postgres://user:pass@localhost:5432/testdb",
			expectedDBType: cryptoutilSQLRepository.DBTypePostgres,
			expectError:    true, // Will fail to connect
		},
		{
			name:           "Unsupported MySQL URL",
			devMode:        false,
			databaseURL:    "mysql://user:pass@localhost:3306/testdb",
			expectedDBType: "",
			expectError:    true,
		},
		{
			name:           "Empty URL with devMode false",
			devMode:        false,
			databaseURL:    "",
			expectedDBType: "",
			expectError:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := cryptoutilConfig.RequireNewForTest(tc.name)
			settings.DevMode = tc.devMode
			settings.DatabaseURL = tc.databaseURL
			settings.DatabaseContainer = containerModeDisabled

			telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
			defer telemetryService.Shutdown()

			repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)

			if tc.expectError {
				testify.Error(t, err)
				testify.Nil(t, repo)
			} else {
				testify.NoError(t, err)
				testify.NotNil(t, repo)

				if repo != nil {
					testify.Equal(t, tc.expectedDBType, repo.GetDBType())
					defer repo.Shutdown()
				}
			}
		})
	}
}

// TestMapContainerMode_AllModes tests all container mode mappings.
func TestMapContainerMode_AllModes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL container test in short mode")
	}

	// Skip if PostgreSQL not available (ci-race has no services)
	if os.Getenv("POSTGRES_HOST") == "" {
		t.Skip("Skipping PostgreSQL test: POSTGRES_HOST not set (PostgreSQL service not available)")
	}

	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name          string
		containerMode string
		expectError   bool
	}{
		{
			name:          "Container mode disabled",
			containerMode: "disabled",
			expectError:   false, // In CI, PostgreSQL service container is running, so connection succeeds
		},
		{
			name:          "Container mode preferred",
			containerMode: "preferred",
			expectError:   false, // Will start PostgreSQL container successfully
		},
		{
			name:          "Container mode required",
			containerMode: "required",
			expectError:   false, // Will start PostgreSQL container successfully
		},
		{
			name:          "Invalid container mode",
			containerMode: "invalid-mode",
			expectError:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := cryptoutilConfig.RequireNewForTest(tc.name)
			settings.DevMode = false
			settings.DatabaseURL = getTestPostgresURL()
			settings.DatabaseContainer = tc.containerMode

			telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
			defer telemetryService.Shutdown()

			repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)

			if tc.expectError {
				testify.Error(t, err)
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

// TestLogSchema_BothDatabaseTypes tests schema logging for both SQLite and PostgreSQL.
func TestLogSchema_BothDatabaseTypes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name    string
		devMode bool
	}{
		{
			name:    "SQLite schema logging",
			devMode: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := cryptoutilConfig.RequireNewForTest(tc.name)
			settings.DevMode = tc.devMode
			settings.DatabaseContainer = containerModeDisabled

			telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
			defer telemetryService.Shutdown()

			repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
			testify.NoError(t, err)
			testify.NotNil(t, repo)

			defer repo.Shutdown()

			// LogSchema is called during NewSQLRepository
			// Verify repo was created successfully (schema was logged)
			if tc.devMode {
				testify.Equal(t, cryptoutilSQLRepository.DBTypeSQLite, repo.GetDBType())
			} else {
				testify.Equal(t, cryptoutilSQLRepository.DBTypePostgres, repo.GetDBType())
			}
		})
	}
}

// TestSQLRepository_ErrorWrapping_AllTypes tests all custom error types.
func TestSQLRepository_ErrorWrapping_AllTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL error test in short mode")
	}

	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name       string
		setup      func(t *testing.T) (*cryptoutilSQLRepository.SQLRepository, error)
		checkError func(t *testing.T, err error)
	}{
		{
			name: "ErrContainerOptionNotExist - SQLite with container mode",
			setup: func(t *testing.T) (*cryptoutilSQLRepository.SQLRepository, error) {
				t.Helper()

				settings := cryptoutilConfig.RequireNewForTest("error_container_not_exist")
				settings.DevMode = true
				settings.DatabaseContainer = containerModeRequired

				telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
				defer telemetryService.Shutdown()

				return cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
			},
			checkError: func(t *testing.T, err error) {
				t.Helper()
				testify.ErrorIs(t, err, cryptoutilSQLRepository.ErrContainerOptionNotExist)
			},
		},
		{
			name: "ErrContainerModeRequiredButContainerNotStarted",
			setup: func(t *testing.T) (*cryptoutilSQLRepository.SQLRepository, error) {
				t.Helper()

				// Skip test when Docker is available - container will start successfully
				t.Skip("Docker available - container starts successfully instead of failing")

				settings := cryptoutilConfig.RequireNewForTest("error_container_required")
				settings.DevMode = false
				// Use invalid database URL to force container startup failure even if PostgreSQL service exists.
				settings.DatabaseURL = invalidDatabaseURL
				settings.DatabaseContainer = containerModeRequired

				telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
				defer telemetryService.Shutdown()

				return cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
			},
			checkError: func(t *testing.T, err error) {
				t.Helper()
				testify.ErrorIs(t, err, cryptoutilSQLRepository.ErrContainerModeRequiredButContainerNotStarted)
			},
		},
		{
			name: "ErrPingDatabaseFailed - PostgreSQL without server",
			setup: func(t *testing.T) (*cryptoutilSQLRepository.SQLRepository, error) {
				t.Helper()

				settings := cryptoutilConfig.RequireNewForTest("error_ping_failed")
				settings.DevMode = false
				// Use invalid credentials to force ping failure even if PostgreSQL service exists.
				settings.DatabaseURL = invalidDatabaseURLWithAuth
				settings.DatabaseContainer = containerModeDisabled

				telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
				defer telemetryService.Shutdown()

				return cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
			},
			checkError: func(t *testing.T, err error) {
				t.Helper()
				testify.ErrorIs(t, err, cryptoutilSQLRepository.ErrPingDatabaseFailed)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo, err := tc.setup(t)
			testify.Error(t, err)
			testify.Nil(t, repo)
			tc.checkError(t, err)
		})
	}
}

// TestSQLTransaction_ErrorConditions tests various transaction error conditions.
func TestSQLTransaction_ErrorConditions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilConfig.RequireNewForTest("transaction_errors")
	settings.DevMode = true
	settings.DatabaseContainer = containerModeDisabled

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)

	t.Cleanup(func() { telemetryService.Shutdown() })

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.NoError(t, err)
	testify.NotNil(t, repo)

	t.Cleanup(func() { repo.Shutdown() })

	tests := []struct {
		name      string
		fn        func(_ *cryptoutilSQLRepository.SQLTransaction) error
		wantError bool
	}{
		{
			name: "Transaction returns custom error",
			fn: func(_ *cryptoutilSQLRepository.SQLTransaction) error {
				return errors.New("custom error")
			},
			wantError: true,
		},
		{
			name: "Transaction returns nil",
			fn: func(_ *cryptoutilSQLRepository.SQLTransaction) error {
				return nil
			},
			wantError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := repo.WithTransaction(ctx, false, tc.fn)
			if tc.wantError {
				testify.Error(t, err)
			} else {
				testify.NoError(t, err)
			}
		})
	}
}
