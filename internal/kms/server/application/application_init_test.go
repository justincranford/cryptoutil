// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// TestServerInit_HappyPath tests ServerInit with valid configuration.
func TestServerInit_HappyPath(t *testing.T) {
	t.Parallel()

	// Create temporary directory for PEM file outputs
	tempDir, err := os.MkdirTemp("", "serverinit_test_*")
	require.NoError(t, err)

	defer func() { _ = os.RemoveAll(tempDir) }()

	// Change to temp directory so ServerInit writes PEM files there
	originalWD, err := os.Getwd()
	require.NoError(t, err)

	defer func() { _ = os.Chdir(originalWD) }()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	tests := []struct {
		name     string
		settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
	}{
		{
			name:     "ValidConfig_InMemoryDB_UnsealModeSysInfo",
			settings: cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create subdirectory for this specific test case
			testCaseDir := filepath.Join(tempDir, tt.name)
			err := os.MkdirAll(testCaseDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute)
			require.NoError(t, err)

			// Change to test case directory
			err = os.Chdir(testCaseDir)
			require.NoError(t, err)

			// Execute ServerInit - should complete without error
			err = ServerInit(tt.settings)
			require.NoError(t, err)

			// Verify expected PEM files were created
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

// TestServerInit_InvalidIPAddresses tests ServerInit with invalid IP address configurations.
func TestServerInit_InvalidIPAddresses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		settings    *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
		expectedErr string
	}{
		{
			name: "InvalidPublicIPAddress",
			settings: func() *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings {
				s := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
				s.TLSPublicIPAddresses = []string{"invalid-ip"}

				return s
			}(),
			expectedErr: "failed to parse public TLS server IP addresses",
		},
		{
			name: "InvalidPrivateIPAddress",
			settings: func() *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings {
				s := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
				s.TLSPrivateIPAddresses = []string{"999.999.999.999"}

				return s
			}(),
			expectedErr: "failed to parse private TLS server IP addresses",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ServerInit(tt.settings)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}
