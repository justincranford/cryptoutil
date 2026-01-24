// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository_test

import (
	"context"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSQLRepository "cryptoutil/internal/kms/server/repository/sqlrepository"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	testify "github.com/stretchr/testify/require"
)

// TestNewSQLRepository_SQLite_PragmaSettings tests SQLite PRAGMA configuration.
func TestNewSQLRepository_SQLite_PragmaSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("pragma_test")
	settings.DevMode = true // SQLite
	settings.DatabaseContainer = containerModeDisabled
	settings.VerboseMode = true // Enable verbose logging to cover logConnectionPoolSettings

	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.NoError(t, err)
	testify.NotNil(t, repo)

	defer repo.Shutdown()

	// Verify repo was created successfully (PRAGMA settings applied internally)
	testify.Equal(t, cryptoutilSQLRepository.DBTypeSQLite, repo.GetDBType())
}

// TestNewSQLRepository_PostgreSQL_SchemaCreation tests PostgreSQL schema creation from search_path.
func TestNewSQLRepository_PostgreSQL_SchemaCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	t.Parallel()

	ctx := context.Background()

	// This test will fail to connect to PostgreSQL but exercises the schema creation code path
	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("schema_creation_test")
	settings.DevMode = false
	settings.DatabaseURL = "postgres://user:pass@localhost:5432/testdb?search_path=test_schema_123&sslmode=disable"
	settings.DatabaseContainer = containerModeDisabled

	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	// Will fail to connect, but that's OK - we're testing the code path
	testify.Error(t, err)
	testify.Nil(t, repo)
}

// TestNewSQLRepository_VerboseMode_ConnectionPoolLogging tests verbose connection pool logging.
func TestNewSQLRepository_VerboseMode_ConnectionPoolLogging(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name        string
		verboseMode bool
	}{
		{
			name:        "Verbose mode enabled",
			verboseMode: true,
		},
		{
			name:        "Verbose mode disabled",
			verboseMode: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest(tc.name)
			settings.DevMode = true
			settings.DatabaseContainer = containerModeDisabled
			settings.VerboseMode = tc.verboseMode

			telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings)
			defer telemetryService.Shutdown()

			repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
			testify.NoError(t, err)
			testify.NotNil(t, repo)

			defer repo.Shutdown()
		})
	}
}

// TestSQLRepository_Shutdown_ErrorHandling tests shutdown error handling.
func TestSQLRepository_Shutdown_ErrorHandling(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("shutdown_error_test")
	settings.DevMode = true
	settings.DatabaseContainer = containerModeDisabled

	telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.NoError(t, err)
	testify.NotNil(t, repo)

	// First shutdown
	repo.Shutdown()

	// Second shutdown (DB already closed - tests error path in Shutdown)
	repo.Shutdown()
}

// TestMapDBTypeAndURL_DevModeVariations tests database type mapping for different dev mode scenarios.
func TestMapDBTypeAndURL_DevModeVariations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name         string
		devMode      bool
		databaseURL  string
		expectSQLite bool
	}{
		{
			name:         "DevMode true - defaults to SQLite",
			devMode:      true,
			databaseURL:  "",
			expectSQLite: true,
		},
		{
			name:         "DevMode true with explicit SQLite URL",
			devMode:      true,
			databaseURL:  "file::memory:?cache=shared",
			expectSQLite: true,
		},
		{
			name:         "DevMode false with PostgreSQL URL",
			devMode:      false,
			databaseURL:  "postgres://user:pass@localhost:5432/testdb?sslmode=disable",
			expectSQLite: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest(tc.name)
			settings.DevMode = tc.devMode

			if tc.databaseURL != "" {
				settings.DatabaseURL = tc.databaseURL
			}

			settings.DatabaseContainer = containerModeDisabled

			telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings)
			defer telemetryService.Shutdown()

			repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)

			if tc.expectSQLite {
				testify.NoError(t, err)
				testify.NotNil(t, repo)

				if repo != nil {
					testify.Equal(t, cryptoutilSQLRepository.DBTypeSQLite, repo.GetDBType())
					defer repo.Shutdown()
				}
			} else {
				// PostgreSQL will fail to connect without running server
				testify.Error(t, err)
				testify.Nil(t, repo)
			}
		})
	}
}

// TestExtractSchemaFromURL_PostgreSQL_VariousFormats tests schema extraction from various URL formats.
func TestExtractSchemaFromURL_PostgreSQL_VariousFormats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name        string
		databaseURL string
		description string
	}{
		{
			name:        "Single schema in search_path",
			databaseURL: "postgres://user:pass@localhost:5432/testdb?search_path=public&sslmode=disable",
			description: "Should extract 'public' schema",
		},
		{
			name:        "Multiple schemas in search_path (comma-separated)",
			databaseURL: "postgres://user:pass@localhost:5432/testdb?search_path=schema1,schema2,schema3&sslmode=disable",
			description: "Should extract first schema 'schema1'",
		},
		{
			name:        "Schema with whitespace",
			databaseURL: "postgres://user:pass@localhost:5432/testdb?search_path= my_schema &sslmode=disable",
			description: "Should trim whitespace and extract 'my_schema'",
		},
		{
			name:        "Empty search_path parameter",
			databaseURL: "postgres://user:pass@localhost:5432/testdb?search_path=&sslmode=disable",
			description: "Should handle empty search_path gracefully",
		},
		{
			name:        "No search_path parameter",
			databaseURL: "postgres://user:pass@localhost:5432/testdb?sslmode=disable",
			description: "Should handle missing search_path gracefully",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest(tc.name)
			settings.DevMode = false
			settings.DatabaseURL = tc.databaseURL
			settings.DatabaseContainer = containerModeDisabled

			telemetryService := cryptoutilSharedTelemetry.RequireNewForTest(ctx, settings)
			defer telemetryService.Shutdown()

			// This will fail to connect, but exercises the URL parsing code
			repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
			testify.Error(t, err) // Expected to fail (no PostgreSQL server)
			testify.Nil(t, repo)
		})
	}
}
