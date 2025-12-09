// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository_test

import (
	"context"
	"testing"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilSQLRepository "cryptoutil/internal/kms/server/repository/sqlrepository"

	testify "github.com/stretchr/testify/require"
)

// TestNewSQLRepository_PostgreSQL_ContainerModes tests PostgreSQL with different container modes.
func TestNewSQLRepository_PostgreSQL_ContainerModes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL container test in short mode")
	}

	t.Parallel()

	tests := []struct {
		name          string
		containerMode string
		databaseURL   string
		expectError   bool
		errorContains string
	}{
		{
			name:          "PostgreSQL with disabled container mode and valid URL",
			containerMode: "disabled",
			databaseURL:   getTestPostgresURL(),
			expectError:   false, // In CI, PostgreSQL service container is running, so connection succeeds
			errorContains: "",
		},
		{
			name:          "PostgreSQL with preferred container mode (will start container)",
			containerMode: "preferred",
			databaseURL:   getTestPostgresURL(),
			expectError:   false, // Container will start successfully when Docker available
			errorContains: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			settings := cryptoutilConfig.RequireNewForTest(tc.name)
			settings.DevMode = false // Use PostgreSQL
			settings.DatabaseURL = tc.databaseURL
			settings.DatabaseContainer = tc.containerMode

			telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
			defer telemetryService.Shutdown()

			repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
			if tc.expectError {
				testify.Error(t, err)

				if tc.errorContains != "" {
					testify.ErrorContains(t, err, tc.errorContains)
				}
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

// TestNewSQLRepository_PostgreSQL_InvalidURL tests PostgreSQL with invalid database URLs.
func TestNewSQLRepository_PostgreSQL_InvalidURL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	t.Parallel()

	tests := []struct {
		name        string
		databaseURL string
		errorCheck  func(t *testing.T, err error)
	}{
		{
			name:        "Empty PostgreSQL URL",
			databaseURL: "",
			errorCheck: func(t *testing.T, err error) {
				t.Helper()
				testify.Error(t, err)
				testify.ErrorContains(t, err, "unsupported database URL format")
			},
		},
		{
			name:        "Invalid PostgreSQL URL format",
			databaseURL: "not-a-valid-url",
			errorCheck: func(t *testing.T, err error) {
				t.Helper()
				testify.Error(t, err)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			settings := cryptoutilConfig.RequireNewForTest(tc.name)
			settings.DevMode = false
			settings.DatabaseURL = tc.databaseURL
			settings.DatabaseContainer = containerModeDisabled

			telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
			defer telemetryService.Shutdown()

			repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
			tc.errorCheck(t, err)
			testify.Nil(t, repo)
		})
	}
}

// TestExtractSchemaFromURL_PostgreSQL tests schema extraction from PostgreSQL URLs.
func TestExtractSchemaFromURL_PostgreSQL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	t.Parallel()

	// This tests the extractSchemaFromURL function indirectly through NewSQLRepository.
	tests := []struct {
		name        string
		databaseURL string
		expectError bool
	}{
		{
			name:        "PostgreSQL URL with search_path single schema",
			databaseURL: "postgres://user:pass@localhost:5432/testdb?search_path=test_schema&sslmode=disable",
			expectError: true, // Will fail to connect but exercises the code path
		},
		{
			name:        "PostgreSQL URL with search_path multiple schemas",
			databaseURL: "postgres://user:pass@localhost:5432/testdb?search_path=schema1,schema2,schema3&sslmode=disable",
			expectError: true,
		},
		{
			name:        "PostgreSQL URL without search_path",
			databaseURL: getTestPostgresURL(),
			expectError: false, // In CI, PostgreSQL service container is running with valid schema
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			settings := cryptoutilConfig.RequireNewForTest(tc.name)
			settings.DevMode = false
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
					defer repo.Shutdown()
				}
			}
		})
	}
}
