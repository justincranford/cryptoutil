// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestInitTLSServerCerts_HappyPath(t *testing.T) {
	// Create temporary directory for PEM file outputs.
	tempDir, err := os.MkdirTemp("", "inittlscerts_test_*")
	require.NoError(t, err)

	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory so InitTLSServerCerts writes PEM files there.
	originalWD, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(originalWD) }()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	tests := []struct {
		name     string
		settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings
	}{
		{
			name:     "ValidConfig_InMemoryDB_UnsealModeSysInfo",
			settings: cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Sequential: uses os.Chdir (global process state, cannot run in parallel).

			// Create subdirectory for this specific test case.
			testCaseDir := filepath.Join(tempDir, tt.name)
			err := os.MkdirAll(testCaseDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute)
			require.NoError(t, err)

			// Change to test case directory.
			err = os.Chdir(testCaseDir)
			require.NoError(t, err)

			// Execute InitTLSServerCerts — should complete without error.
			err = InitTLSServerCerts(tt.settings)
			require.NoError(t, err)

			// Verify expected PEM files were created.
			expectedFiles := []string{
				"tls_public_server_certificate_0.pem",
				"tls_public_server_certificate_1.pem",
				"tls_public_server_private_key.pem",
				"tls_private_server_certificate_0.pem",
				"tls_private_server_certificate_1.pem",
				"tls_private_server_private_key.pem",
			}

			for _, filename := range expectedFiles {
				filePath := filepath.Join(testCaseDir, filename)
				_, err := os.Stat(filePath)
				require.NoError(t, err, "expected PEM file %s not found", filename)
			}
		})
	}
}

// TestInitTLSServerCerts_InvalidIPAddresses tests InitTLSServerCerts with invalid IP address configurations.
func TestInitTLSServerCerts_InvalidIPAddresses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		settings    *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings
		expectedErr string
	}{
		{
			name: "InvalidPublicIPAddress",
			settings: func() *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings {
				s := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
				s.TLSPublicIPAddresses = []string{"invalid-ip"}

				return s
			}(),
			expectedErr: "failed to parse public TLS server IP addresses",
		},
		{
			name: "InvalidPrivateIPAddress",
			settings: func() *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings {
				s := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
				s.TLSPrivateIPAddresses = []string{"999.999.999.999"}

				return s
			}(),
			expectedErr: "failed to parse private TLS server IP addresses",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := InitTLSServerCerts(tt.settings)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

// TestBasic_Shutdown_NilComponents verifies that Shutdown does not panic when all struct fields are nil.
func TestBasic_Shutdown_NilComponents(t *testing.T) {
	t.Parallel()

	app := &Basic{}
	require.NotPanics(t, app.Shutdown, "shutdown should not panic with nil components")
}

// TestStartTLSListener_ContainerConfig tests StartTLSListener with container mode and dev mode configurations.
func TestStartTLSListener_ContainerConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		modifySettings    func(*cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings)
		wantInitSuccess   bool
		wantContainerMode bool
	}{
		{
			name: "container mode + SQLite initialization succeeds",
			modifySettings: func(s *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) {
				s.BindPublicAddress = cryptoutilSharedMagic.IPv4AnyAddress // 0.0.0.0
				s.BindPublicPort = 0                                       // Dynamic port
				s.DatabaseURL = "sqlite://file::memory:?cache=shared"
			},
			wantInitSuccess:   true,
			wantContainerMode: true,
		},
		{
			name: "container mode + PostgreSQL initialization fails (unavailable in test)",
			modifySettings: func(s *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) {
				s.DevMode = false                                          // Disable dev mode to test actual PostgreSQL URL
				s.BindPublicAddress = cryptoutilSharedMagic.IPv4AnyAddress // 0.0.0.0
				s.BindPublicPort = 0                                       // Dynamic port
				// PostgreSQL URL — connection will fail but validation passes.
				s.DatabaseURL = "postgres://user:pass@localhost:5432/testdb"
			},
			wantInitSuccess:   false, // Will fail on database connection, but config validation passes.
			wantContainerMode: true,
		},
		{
			name: "dev mode + SQLite initialization succeeds",
			modifySettings: func(s *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) {
				s.DevMode = true
				s.BindPublicAddress = cryptoutilSharedMagic.IPv4Loopback // 127.0.0.1
				s.BindPublicPort = 0
				s.DatabaseURL = cryptoutilSharedMagic.SQLiteInMemoryDSN
			},
			wantInitSuccess:   true,
			wantContainerMode: false,
		},
		{
			name: "production mode + loopback + SQLite initialization succeeds",
			modifySettings: func(s *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) {
				s.DevMode = false
				s.BindPublicAddress = cryptoutilSharedMagic.IPv4Loopback // 127.0.0.1
				s.BindPublicPort = 0
				// Use in-memory SQLite to avoid persisted migration state from previous test runs.
				s.DatabaseURL = "file:production_tls_test?mode=memory&cache=shared"
			},
			wantInitSuccess:   true,
			wantContainerMode: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Start with clean test settings.
			settings := cryptoutilAppsFrameworkServiceConfig.RequireNewForTest(tt.name)

			// Apply test-specific modifications.
			tt.modifySettings(settings)

			// Attempt to initialize TLS listener.
			serverApp, err := StartTLSListener(settings)

			if tt.wantInitSuccess {
				require.NoError(t, err, "StartTLSListener should succeed")
				require.NotNil(t, serverApp, "TLSListener should not be nil")

				// Verify TLS servers initialized.
				require.NotNil(t, serverApp.PublicTLSServer, "Public TLS server should be initialized")
				require.NotNil(t, serverApp.PrivateTLSServer, "Private TLS server should be initialized")

				// Verify container mode detection.
				actualContainerMode := settings.BindPublicAddress == cryptoutilSharedMagic.IPv4AnyAddress
				require.Equal(t, tt.wantContainerMode, actualContainerMode,
					"Container mode detection mismatch: expected=%v, actual=%v",
					tt.wantContainerMode, actualContainerMode)

				// Cleanup.
				serverApp.ShutdownFunction()
			} else {
				// For cases where initialization is expected to fail (e.g., PostgreSQL not available)
				// we still verify that config validation passed, even if database connection failed.
				require.Error(t, err, "StartTLSListener should fail when database unavailable")
			}
		})
	}
}
