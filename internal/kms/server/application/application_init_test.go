// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"os"
	"path/filepath"
	"testing"

	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilMagic "cryptoutil/internal/shared/magic"

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
		settings *cryptoutilConfig.ServerSettings
	}{
		{
			name: "ValidConfig_InMemoryDB_UnsealModeSysInfo",
			settings: &cryptoutilConfig.ServerSettings{
				LogLevel:          "ERROR",
				VerboseMode:       false,
				DevMode:           true,
				DatabaseURL:       "sqlite://file::memory:?cache=shared",
				UnsealMode:        cryptoutilMagic.DefaultUnsealModeSysInfo,
				OTLPEnabled:       false,
				OTLPService:       "application_test",
				OTLPEndpoint:      "grpc://localhost:4317",
				TLSPublicDNSNames: []string{"localhost", "127.0.0.1"},
				TLSPublicIPAddresses: []string{
					cryptoutilMagic.IPv4Loopback,
					cryptoutilMagic.IPv6Loopback,
				},
				TLSPrivateDNSNames: []string{"localhost", "127.0.0.1"},
				TLSPrivateIPAddresses: []string{
					cryptoutilMagic.IPv4Loopback,
					cryptoutilMagic.IPv6Loopback,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create subdirectory for this specific test case
			testCaseDir := filepath.Join(tempDir, tt.name)
			err := os.MkdirAll(testCaseDir, cryptoutilMagic.FilePermOwnerReadWriteExecuteGroupReadExecute)
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
		settings    *cryptoutilConfig.ServerSettings
		expectedErr string
	}{
		{
			name: "InvalidPublicIPAddress",
			settings: &cryptoutilConfig.ServerSettings{
				LogLevel:              "ERROR",
				VerboseMode:           false,
				DevMode:               true,
				DatabaseURL:           "sqlite://file::memory:?cache=shared",
				UnsealMode:            cryptoutilMagic.DefaultUnsealModeSysInfo,
				OTLPEnabled:           false,
				OTLPService:           "application_test",
				OTLPEndpoint:          "grpc://localhost:4317",
				TLSPublicDNSNames:     []string{"localhost"},
				TLSPublicIPAddresses:  []string{"invalid-ip"},
				TLSPrivateDNSNames:    []string{"localhost"},
				TLSPrivateIPAddresses: []string{cryptoutilMagic.IPv4Loopback},
			},
			expectedErr: "failed to parse public TLS server IP addresses",
		},
		{
			name: "InvalidPrivateIPAddress",
			settings: &cryptoutilConfig.ServerSettings{
				LogLevel:              "ERROR",
				VerboseMode:           false,
				DevMode:               true,
				DatabaseURL:           "sqlite://file::memory:?cache=shared",
				UnsealMode:            cryptoutilMagic.DefaultUnsealModeSysInfo,
				OTLPEnabled:           false,
				OTLPService:           "application_test",
				OTLPEndpoint:          "grpc://localhost:4317",
				TLSPublicDNSNames:     []string{"localhost"},
				TLSPublicIPAddresses:  []string{cryptoutilMagic.IPv4Loopback},
				TLSPrivateDNSNames:    []string{"localhost"},
				TLSPrivateIPAddresses: []string{"999.999.999.999"},
			},
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
