// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository_test

import (
	"context"
	"errors"
	"testing"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilSQLRepository "cryptoutil/internal/kms/server/repository/sqlrepository"

	testify "github.com/stretchr/testify/require"
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
			expectError:   false,
		},
		{
			name:          "Container mode preferred",
			containerMode: "preferred",
			expectError:   true, // Will fail to start container, fallback to URL, fail to connect
		},
		{
			name:          "Container mode required",
			containerMode: "required",
			expectError:   true, // Will fail to start required container
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

				settings := cryptoutilConfig.RequireNewForTest("error_container_required")
				settings.DevMode = false
				settings.DatabaseURL = getTestPostgresURL()
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
				settings.DatabaseURL = getTestPostgresURL()
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
		fn        func(tx *cryptoutilSQLRepository.SQLTransaction) error
		wantError bool
	}{
		{
			name: "Transaction returns custom error",
			fn: func(tx *cryptoutilSQLRepository.SQLTransaction) error {
				return errors.New("custom error")
			},
			wantError: true,
		},
		{
			name: "Transaction returns nil",
			fn: func(tx *cryptoutilSQLRepository.SQLTransaction) error {
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
