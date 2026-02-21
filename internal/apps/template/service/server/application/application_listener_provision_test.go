// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestProvisionDatabase_ErrorPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name          string
		databaseURL   string
		containerMode string
		expectError   bool
		errorContains string
	}{
		{
			name:          "Unsupported database scheme",
			databaseURL:   "mysql://user:pass@localhost:3306/db",
			containerMode: "disabled",
			expectError:   true,
			errorContains: "unsupported database URL scheme",
		},
		{
			name:          "Invalid SQLite file path",
			databaseURL:   "file:///nonexistent/path/to/invalid.db",
			containerMode: "disabled",
			expectError:   true,
			errorContains: "failed to open database",
		},
		{
			name:          "file::memory: format",
			databaseURL:   "file::memory:?cache=shared",
			containerMode: "disabled",
			expectError:   false,
		},
		{
			name:          "file:NAME?mode=memory format",
			databaseURL:   "file:provision_test_mode?mode=memory&cache=shared",
			containerMode: "disabled",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				LogLevel:          "info",
				OTLPEndpoint:      "grpc://localhost:4317",
				OTLPService:       "test-error-paths",
				OTLPVersion:       "1.0.0",
				OTLPEnvironment:   "test",
				UnsealMode:        "sysinfo",
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

				if tt.errorContains != "" {
					require.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, db)
			}
		})
	}
}

func TestOpenSQLite_FileMode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create temporary database file.
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"

	db, err := openSQLite(ctx, dbPath, false)
	require.NoError(t, err)
	require.NotNil(t, db)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NotNil(t, sqlDB)

	err = sqlDB.Close()
	require.NoError(t, err)
}

func TestStartCoreWithServices_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:                    true,
		VerboseMode:                false,
		DatabaseURL:                cryptoutilSharedMagic.SQLiteInMemoryDSN,
		OTLPService:                "template-test-cws",
		OTLPEnabled:                false,
		OTLPEndpoint:               "grpc://127.0.0.1:4317",
		LogLevel:                   "INFO",
		BrowserSessionAlgorithm:    "JWS",
		BrowserSessionJWSAlgorithm: "RS256",
		BrowserSessionJWEAlgorithm: "RSA-OAEP",
		BrowserSessionExpiration:   15 * time.Minute,
		ServiceSessionAlgorithm:    "JWS",
		ServiceSessionJWSAlgorithm: "RS256",
		ServiceSessionJWEAlgorithm: "RSA-OAEP",
		ServiceSessionExpiration:   1 * time.Hour,
		SessionIdleTimeout:         30 * time.Minute,
		SessionCleanupInterval:     1 * time.Hour,
	}

	// FIRST create just Core to get DB.
	core, err := StartCore(ctx, settings)
	require.NoError(t, err)

	require.NotNil(t, core)
	defer core.Shutdown()

	// THEN run migrations (required for BarrierService).
	err = core.DB.AutoMigrate(
		&cryptoutilAppsTemplateServiceServerBarrier.RootKey{},
		&cryptoutilAppsTemplateServiceServerBarrier.IntermediateKey{},
		&cryptoutilAppsTemplateServiceServerBarrier.ContentKey{},
		&cryptoutilAppsTemplateServiceServerRepository.BrowserSessionJWK{},
		&cryptoutilAppsTemplateServiceServerRepository.ServiceSessionJWK{},
		&cryptoutilAppsTemplateServiceServerRepository.BrowserSession{},
		&cryptoutilAppsTemplateServiceServerRepository.ServiceSession{},
	)
	require.NoError(t, err)

	// FINALLY initialize services on migrated Core.
	coreWithSvcs, err := InitializeServicesOnCore(ctx, core, settings)
	require.NoError(t, err)
	require.NotNil(t, coreWithSvcs)

	// Verify all services initialized.
	require.NotNil(t, coreWithSvcs.Repository)
	require.NotNil(t, coreWithSvcs.BarrierService)
	require.NotNil(t, coreWithSvcs.RealmRepository)
	require.NotNil(t, coreWithSvcs.RealmService)
	require.NotNil(t, coreWithSvcs.SessionManager)
	require.NotNil(t, coreWithSvcs.TenantRepository)
	require.NotNil(t, coreWithSvcs.UserRepository)
	require.NotNil(t, coreWithSvcs.JoinRequestRepository)
	require.NotNil(t, coreWithSvcs.RegistrationService)
	require.NotNil(t, coreWithSvcs.RotationService)
	require.NotNil(t, coreWithSvcs.StatusService)
}

func TestStartCoreWithServices_StartCoreFails(t *testing.T) {
	t.Parallel()

	coreWithSvcs, err := StartCoreWithServices(nil, nil) //nolint:staticcheck // Testing nil context error handling
	require.Error(t, err)
	require.Nil(t, coreWithSvcs)
	require.Contains(t, err.Error(), "failed to start application core")
}

// TestStartCoreWithServices_InitializeServicesFails tests StartCoreWithServices when InitializeServicesOnCore fails.
// This tests the error path where StartCore succeeds but service initialization fails.
func TestStartCoreWithServices_InitializeServicesFails(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Use temporary file database for test isolation.
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Convert to proper file URI (file:///abs/path on all platforms).
	slashPath := filepath.ToSlash(dbPath)
	if !strings.HasPrefix(slashPath, "/") {
		slashPath = "/" + slashPath
	}

	dbName := fmt.Sprintf("file://%s?mode=rwc&cache=shared", slashPath)

	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:                    true,
		VerboseMode:                false,
		DatabaseURL:                dbName,
		OTLPService:                "template-service-test",
		OTLPEnabled:                false,
		OTLPEndpoint:               "grpc://127.0.0.1:4317",
		LogLevel:                   "INFO",
		BrowserSessionAlgorithm:    "JWS",
		BrowserSessionJWSAlgorithm: "RS256",
		BrowserSessionJWEAlgorithm: "RSA-OAEP",
		BrowserSessionExpiration:   15 * time.Minute,
		ServiceSessionAlgorithm:    "JWS",
		ServiceSessionJWSAlgorithm: "RS256",
		ServiceSessionJWEAlgorithm: "RSA-OAEP",
		ServiceSessionExpiration:   1 * time.Hour,
		SessionIdleTimeout:         30 * time.Minute,
		SessionCleanupInterval:     1 * time.Hour,
	}

	// StartCoreWithServices without running migrations.
	// This will cause BarrierService initialization to fail when it queries barrier_root_keys table.
	coreWithSvcs, err := StartCoreWithServices(ctx, settings)
	require.Error(t, err)
	require.Nil(t, coreWithSvcs)
	require.Contains(t, err.Error(), "barrier service")
}

// TestStartBasic_UnsealKeysServiceFailure tests StartBasic when unseal keys service initialization fails.
// This triggers the error path at lines 47-52 in application_basic.go.
func TestStartBasic_UnsealKeysServiceFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Use invalid unseal mode to trigger unseal keys service failure.
	// "invalid-mode" is not "sysinfo", not a number, and not "M-of-N" format.
	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:         false, // Must be false to use UnsealMode
		VerboseMode:     false,
		LogLevel:        "info", // Required for telemetry to succeed
		OTLPEnabled:     false,  // Disable OTLP to avoid endpoint issues
		OTLPEndpoint:    "grpc://localhost:4317",
		OTLPService:     "test-service",
		OTLPVersion:     "1.0.0",
		OTLPEnvironment: "test",
		UnsealMode:      "invalid-mode", // Invalid mode triggers unseal service error
		DatabaseURL:     cryptoutilSharedMagic.SQLiteInMemoryDSN,
	}

	// StartBasic should fail because unseal keys service initialization fails.
	basic, err := StartBasic(ctx, settings)
	require.Error(t, err)
	require.Nil(t, basic)
	require.Contains(t, err.Error(), "failed to create unseal repository")
}
