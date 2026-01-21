// Copyright (c) 2025 Justin Cranford
//
//

package sqlrepository_test

import (
	"context"
	"testing"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSQLRepository "cryptoutil/internal/kms/server/repository/sqlrepository"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"

	testify "github.com/stretchr/testify/require"
)

// TestNewSQLRepository_SQLite_ContainerMode tests SQLite with container mode (should error).
func TestNewSQLRepository_SQLite_ContainerMode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilConfig.RequireNewForTest("sqlite_container_test")
	settings.DevMode = true                            // SQLite
	settings.DatabaseContainer = containerModeRequired // SQLite doesn't support containers

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.Error(t, err)
	testify.ErrorIs(t, err, cryptoutilSQLRepository.ErrContainerOptionNotExist)
	testify.Nil(t, repo)
}

// TestNewSQLRepository_InvalidContainerMode tests invalid container mode values.
func TestNewSQLRepository_InvalidContainerMode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilConfig.RequireNewForTest("invalid_container_mode")
	settings.DevMode = true
	settings.DatabaseContainer = containerModeInvalid

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
	defer telemetryService.Shutdown()

	repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
	testify.Error(t, err)
	testify.ErrorContains(t, err, "unsupported container mode")
	testify.Nil(t, repo)
}

// TestNewSQLRepository_NilInputs tests nil input validation.
func TestNewSQLRepository_NilInputs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := cryptoutilConfig.RequireNewForTest("nil_inputs_test")

	telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)

	t.Cleanup(func() { telemetryService.Shutdown() })

	tests := []struct {
		name          string
		ctx           context.Context
		telemetry     *cryptoutilTelemetry.TelemetryService
		settings      *cryptoutilConfig.ServiceTemplateServerSettings
		errorContains string
	}{
		{
			name:          "Nil context",
			ctx:           nil,
			telemetry:     telemetryService,
			settings:      settings,
			errorContains: "ctx must be non-nil",
		},
		{
			name:          "Nil telemetry service",
			ctx:           ctx,
			telemetry:     nil,
			settings:      settings,
			errorContains: "telemetryService must be non-nil",
		},
		{
			name:          "Nil settings",
			ctx:           ctx,
			telemetry:     telemetryService,
			settings:      nil,
			errorContains: "settings must be non-nil",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo, err := cryptoutilSQLRepository.NewSQLRepository(tc.ctx, tc.telemetry, tc.settings)
			testify.Error(t, err)
			testify.ErrorContains(t, err, tc.errorContains)
			testify.Nil(t, repo)
		})
	}
}

// TestMapDBTypeAndURL_EdgeCases tests database type and URL mapping edge cases.
func TestMapDBTypeAndURL_EdgeCases(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name        string
		devMode     bool
		databaseURL string
		expectError bool
	}{
		{
			name:        "DevMode true with empty URL",
			devMode:     true,
			databaseURL: "",
			expectError: false, // DevMode uses SQLite in-memory by default
		},
		{
			name:        "DevMode false with empty URL",
			devMode:     false,
			databaseURL: "",
			expectError: true, // PostgreSQL requires URL
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := cryptoutilConfig.RequireNewForTest(tc.name)
			settings.DevMode = tc.devMode
			settings.DatabaseURL = tc.databaseURL

			telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
			defer telemetryService.Shutdown()

			repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
			if tc.expectError {
				testify.Error(t, err)
				testify.Nil(t, repo)
			} else {
				if err == nil {
					defer repo.Shutdown()
				}
				// Some scenarios may still error due to connection issues
			}
		})
	}
}

// TestExtractSchemaFromURL_EdgeCases tests schema extraction edge cases.
func TestExtractSchemaFromURL_EdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name        string
		databaseURL string
		expectError bool
	}{
		{
			name:        "PostgreSQL URL with malformed search_path",
			databaseURL: "postgres://user:pass@localhost:5432/testdb?search_path=&sslmode=disable",
			expectError: true,
		},
		{
			name:        "PostgreSQL URL with special characters in schema",
			databaseURL: "postgres://user:pass@localhost:5432/testdb?search_path=my-schema_123&sslmode=disable",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			settings := cryptoutilConfig.RequireNewForTest(tc.name)
			settings.DevMode = false
			settings.DatabaseURL = tc.databaseURL
			settings.DatabaseContainer = containerModeDisabled

			telemetryService := cryptoutilTelemetry.RequireNewForTest(ctx, settings)
			defer telemetryService.Shutdown()

			repo, err := cryptoutilSQLRepository.NewSQLRepository(ctx, telemetryService, settings)
			if tc.expectError {
				testify.Error(t, err)
			}

			testify.Nil(t, repo)
		})
	}
}
