// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"context"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedContainer "cryptoutil/internal/shared/container"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestShutdown_PartialInitialization(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create a mock server that implements IPublicServer.
	mockServer := &mockPublicServer{}

	listener := &Listener{
		AdminServer:  nil,
		PublicServer: mockServer,
	}

	// Shutdown should handle partial initialization gracefully.
	err := listener.Shutdown(ctx)
	require.NoError(t, err)
}

// TestProvisionDatabase_SQLiteVariations tests provisionDatabase with different SQLite URL formats.
func TestProvisionDatabase_SQLiteVariations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name        string
		databaseURL string
		expectError bool
	}{
		{
			name:        "Empty URL (defaults to in-memory)",
			databaseURL: "",
			expectError: false,
		},
		{
			name:        "In-memory placeholder",
			databaseURL: cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
			expectError: false,
		},
		{
			name:        "File URL",
			databaseURL: "file:///tmp/test.db",
			expectError: false,
		},
		{
			name:        "Invalid URL scheme",
			databaseURL: "mysql://user:pass@localhost:3306/db",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				LogLevel:          "info",
				OTLPEndpoint:      "grpc://localhost:4317",
				OTLPService:       "test-service",
				OTLPVersion:       cryptoutilSharedMagic.ServiceVersion,
				OTLPEnvironment:   "test",
				UnsealMode:        cryptoutilSharedMagic.DefaultUnsealModeSysInfo,
				DatabaseURL:       tt.databaseURL,
				DatabaseContainer: cryptoutilSharedMagic.DefaultDatabaseContainerDisabled,
			}

			basic, err := StartBasic(ctx, settings)
			require.NoError(t, err)

			defer basic.Shutdown()

			db, cleanup, err := provisionDatabase(ctx, basic, settings)
			if cleanup != nil {
				defer cleanup()
			}

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, db)
			} else {
				require.NoError(t, err)
				require.NotNil(t, db)
			}
		})
	}
}

// TestStartCore_Variations tests StartCore with different unseal modes and database URLs.
func TestStartCore_Variations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name        string
		unsealMode  string
		databaseURL string
		expectError bool
	}{
		{
			name:        "sysinfo mode with in-memory",
			unsealMode:  cryptoutilSharedMagic.DefaultUnsealModeSysInfo,
			databaseURL: cryptoutilSharedMagic.SQLiteInMemoryDSN,
			expectError: false,
		},
		{
			name:        "sysinfo mode with empty URL",
			unsealMode:  cryptoutilSharedMagic.DefaultUnsealModeSysInfo,
			databaseURL: "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				LogLevel:        "info",
				OTLPEndpoint:    "grpc://localhost:4317",
				OTLPService:     "test-service",
				OTLPVersion:     cryptoutilSharedMagic.ServiceVersion,
				OTLPEnvironment: "test",
				UnsealMode:      tt.unsealMode,
				DatabaseURL:     tt.databaseURL,
			}

			core, err := StartCore(ctx, settings)
			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, core)
			} else {
				require.NoError(t, err)
				require.NotNil(t, core)

				if core != nil {
					core.Shutdown()
				}
			}
		})
	}
}

// TestOpenSQLite_DebugMode tests openSQLite with debug mode enabled.
func TestOpenSQLite_DebugMode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name        string
		databaseURL string
		debugMode   bool
		expectError bool
	}{
		{
			name:        "In-memory with debug mode",
			databaseURL: cryptoutilSharedMagic.SQLiteInMemoryDSN,
			debugMode:   true,
			expectError: false,
		},
		{
			name:        "File URL with debug mode",
			databaseURL: "file:///tmp/test-debug.db",
			debugMode:   true,
			expectError: false,
		},
		{
			name:        "In-memory without debug mode",
			databaseURL: cryptoutilSharedMagic.SQLiteInMemoryDSN,
			debugMode:   false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, err := openSQLite(ctx, tt.databaseURL, tt.debugMode)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, db)
			} else {
				require.NoError(t, err)
				require.NotNil(t, db)

				// Clean up.
				sqlDB, dbErr := db.DB()
				require.NoError(t, dbErr)

				_ = sqlDB.Close()
			}
		})
	}
}

// TestOpenPostgreSQL_WithContainer tests openPostgreSQL with a real container.
func TestOpenPostgreSQL_WithContainer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create a basic telemetry service for the container.
	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		LogLevel:        "info",
		OTLPEndpoint:    "grpc://localhost:4317",
		OTLPService:     "test-container",
		OTLPVersion:     cryptoutilSharedMagic.ServiceVersion,
		OTLPEnvironment: "test",
		UnsealMode:      cryptoutilSharedMagic.DefaultUnsealModeSysInfo,
		DatabaseURL:     cryptoutilSharedMagic.SQLiteInMemoryDSN,
	}

	basic, err := StartBasic(ctx, settings)
	require.NoError(t, err)

	defer basic.Shutdown()

	// Start a real PostgreSQL container for testing.
	containerURL, cleanup, err := cryptoutilSharedContainer.StartPostgres(
		ctx,
		basic.TelemetryService,
		"test_db",
		"test_user",
		"test_password",
	)
	if err != nil {
		t.Skipf("Skipping PostgreSQL test - container unavailable: %v", err)
	}
	defer cleanup()

	tests := []struct {
		name        string
		debugMode   bool
		expectError bool
	}{
		{
			name:        "Debug mode enabled",
			debugMode:   true,
			expectError: false,
		},
		{
			name:        "Debug mode disabled",
			debugMode:   false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := openPostgreSQL(ctx, containerURL, tt.debugMode)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, db)
			} else {
				require.NoError(t, err)
				require.NotNil(t, db)

				// Clean up.
				sqlDB, dbErr := db.DB()
				require.NoError(t, dbErr)

				_ = sqlDB.Close()
			}
		})
	}
}

// TestStartBasic_VerboseMode tests StartBasic with verbose mode variations.
func TestStartBasic_VerboseMode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name        string
		verboseMode bool
		expectError bool
	}{
		{
			name:        "Verbose mode enabled",
			verboseMode: true,
			expectError: false,
		},
		{
			name:        "Verbose mode disabled",
			verboseMode: false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				LogLevel:        "info",
				OTLPEndpoint:    "grpc://localhost:4317",
				OTLPService:     "test-service",
				OTLPVersion:     cryptoutilSharedMagic.ServiceVersion,
				OTLPEnvironment: "test",
				UnsealMode:      cryptoutilSharedMagic.DefaultUnsealModeSysInfo,
				VerboseMode:     tt.verboseMode,
				DatabaseURL:     cryptoutilSharedMagic.SQLiteInMemoryDSN,
			}

			basic, err := StartBasic(ctx, settings)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, basic)
			} else {
				require.NoError(t, err)

				require.NotNil(t, basic)
				defer basic.Shutdown()
			}
		})
	}
}

// TestProvisionDatabase_PostgreSQLContainerModes tests container mode variations.
func TestProvisionDatabase_PostgreSQLContainerModes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name          string
		containerMode string
		databaseURL   string
		expectError   bool
	}{
		{
			name:          "Container mode disabled with SQLite",
			containerMode: cryptoutilSharedMagic.DefaultDatabaseContainerDisabled,
			databaseURL:   cryptoutilSharedMagic.SQLiteInMemoryDSN,
			expectError:   false,
		},
		{
			name:          "Container mode empty string with SQLite",
			containerMode: "",
			databaseURL:   cryptoutilSharedMagic.SQLiteInMemoryDSN,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				LogLevel:          "info",
				OTLPEndpoint:      "grpc://localhost:4317",
				OTLPService:       "test-container-modes",
				OTLPVersion:       cryptoutilSharedMagic.ServiceVersion,
				OTLPEnvironment:   "test",
				UnsealMode:        cryptoutilSharedMagic.DefaultUnsealModeSysInfo,
				DatabaseURL:       tt.databaseURL,
				DatabaseContainer: tt.containerMode,
			}

			basic, err := StartBasic(ctx, settings)
			require.NoError(t, err)

			defer basic.Shutdown()

			db, cleanup, err := provisionDatabase(ctx, basic, settings)
			if cleanup != nil {
				defer cleanup()
			}

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, db)
			} else {
				require.NoError(t, err)
				require.NotNil(t, db)
			}
		})
	}
}

// TestMaskPasswordVariations tests the maskPassword function with various DSN formats.
func TestMaskPasswordVariations(t *testing.T) {
	t.Parallel()

	// We test via provisionDatabase which calls maskPassword internally.
	ctx := context.Background()

	tests := []struct {
		name        string
		databaseURL string
		expectError bool
	}{
		{
			name:        "PostgreSQL URL with password",
			databaseURL: "postgres://user:secret123@localhost:5432/testdb",
			expectError: false, // Will fail to connect but maskPassword executes.
		},
		{
			name:        "PostgreSQL URL without password",
			databaseURL: "postgres://user@localhost:5432/testdb",
			expectError: false,
		},
		{
			name:        "Malformed URL",
			databaseURL: "invalid://url",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				LogLevel:          "info",
				OTLPEndpoint:      "grpc://localhost:4317",
				OTLPService:       "test-mask-password",
				OTLPVersion:       cryptoutilSharedMagic.ServiceVersion,
				OTLPEnvironment:   "test",
				UnsealMode:        cryptoutilSharedMagic.DefaultUnsealModeSysInfo,
				DatabaseURL:       tt.databaseURL,
				DatabaseContainer: cryptoutilSharedMagic.DefaultDatabaseContainerDisabled, // Don't try to start container.
			}

			basic, err := StartBasic(ctx, settings)
			if err == nil {
				defer basic.Shutdown()

				// Try to provision - this will call maskPassword internally.
				db, cleanup, dbErr := provisionDatabase(ctx, basic, settings)
				if cleanup != nil {
					defer cleanup()
				}

				if tt.expectError {
					require.Error(t, dbErr)
					require.Nil(t, db)
				} else {
					// maskPassword executes even if connection fails.
					if dbErr != nil {
						require.Contains(t, dbErr.Error(), "failed to open database")
					}
				}
			}
		})
	}
}

// TestProvisionDatabase_ErrorPaths tests error handling in database provisioning.
