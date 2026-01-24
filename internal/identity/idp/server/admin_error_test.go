// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceTestingHttpservertests "cryptoutil/internal/apps/template/service/testing/httpservertests"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
)

// TestAdminServer_Start_NilContext tests Start with nil context.
func TestAdminServer_Start_NilContext(t *testing.T) {
	t.Parallel()

	createServer := func(t *testing.T) cryptoutilAppsTemplateServiceTestingHttpservertests.HTTPServer {
		t.Helper()

		cfg := cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin_start_nil_ctx")
		ctx := context.Background()
		server, err := NewAdminHTTPServer(ctx, cfg)
		require.NoError(t, err)

		return server
	}

	cryptoutilAppsTemplateServiceTestingHttpservertests.TestStartNilContext(t, createServer)
}

// TestAdminServer_LoadTLSConfig_InvalidCertFile tests loadTLSConfig with invalid certificate file.
func TestAdminServer_LoadTLSConfig_InvalidCertFile(t *testing.T) {
	t.Parallel()

	cfg := cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin_invalid_cert")

	// Create temporary directory for test files.
	tmpDir := t.TempDir()

	// Create invalid cert file (not a valid PEM).
	certFile := filepath.Join(tmpDir, "invalid_cert.pem")
	require.NoError(t, os.WriteFile(certFile, []byte("invalid certificate data"), 0o600))

	// Create valid-looking key file (will fail at LoadX509KeyPair due to cert issue).
	keyFile := filepath.Join(tmpDir, "invalid_key.pem")
	require.NoError(t, os.WriteFile(keyFile, []byte("invalid key data"), 0o600))

	// Configure server to use invalid files.
	cfg.IDP.TLSCertFile = certFile
	cfg.IDP.TLSKeyFile = keyFile

	ctx := context.Background()

	server, err := NewAdminHTTPServer(ctx, cfg)
	require.NoError(t, err)

	// loadTLSConfig should fail with invalid certificate file.
	_, err = server.loadTLSConfig(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to load TLS certificate and key")
}

// TestAdminServer_LoadTLSConfig_NonexistentFiles tests loadTLSConfig with nonexistent files.
func TestAdminServer_LoadTLSConfig_NonexistentFiles(t *testing.T) {
	t.Parallel()

	cfg := cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin_nonexistent_files")

	// Configure server to use nonexistent files.
	cfg.IDP.TLSCertFile = "/nonexistent/cert.pem"
	cfg.IDP.TLSKeyFile = "/nonexistent/key.pem"

	ctx := context.Background()

	server, err := NewAdminHTTPServer(ctx, cfg)
	require.NoError(t, err)

	// loadTLSConfig should fail with file not found error.
	_, err = server.loadTLSConfig(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to load TLS certificate and key")
}

// TestAdminServer_LoadTLSConfig_NilContext tests loadTLSConfig with nil context.
func TestAdminServer_LoadTLSConfig_NilContext(t *testing.T) {
	t.Parallel()

	cfg := cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin_loadtls_nil_ctx")
	ctx := context.Background()

	server, err := NewAdminHTTPServer(ctx, cfg)
	require.NoError(t, err)

	// loadTLSConfig with nil context should fail.
	_, err = server.loadTLSConfig(nil) //nolint:staticcheck // Testing nil context validation requires passing nil.
	require.Error(t, err)
	require.Contains(t, err.Error(), "context cannot be nil")
}

// TestAdminServer_LoadTLSConfig_SelfSigned tests loadTLSConfig generates self-signed cert when no files provided.
func TestAdminServer_LoadTLSConfig_SelfSigned(t *testing.T) {
	t.Parallel()

	cfg := cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin_selfsigned")

	// Do not set TLSCertFile or TLSKeyFile - should trigger self-signed generation.
	cfg.IDP.TLSCertFile = ""
	cfg.IDP.TLSKeyFile = ""

	ctx := context.Background()

	server, err := NewAdminHTTPServer(ctx, cfg)
	require.NoError(t, err)

	// loadTLSConfig should generate self-signed certificate.
	tlsConfig, err := server.loadTLSConfig(ctx)
	require.NoError(t, err)
	require.NotNil(t, tlsConfig)
	require.NotEmpty(t, tlsConfig.Certificates)
	require.Equal(t, uint16(0x0304), tlsConfig.MinVersion) // TLS 1.3 = 0x0304.
}

// TestAdminServer_Start_InvalidBindAddress tests Start with invalid bind address.
func TestAdminServer_Start_InvalidBindAddress(t *testing.T) {
	t.Parallel()

	cfg := cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin_invalid_bind")

	// Set invalid bind address (should fail to listen).
	cfg.IDP.AdminBindAddress = "999.999.999.999" // Invalid IP address.
	cfg.IDP.AdminPort = 0

	ctx := context.Background()

	server, err := NewAdminHTTPServer(ctx, cfg)
	require.NoError(t, err)

	// Start should fail with invalid bind address.
	err = server.Start(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create TLS listener")
}

// TestAdminServer_ActualPort_BeforeStart tests ActualPort before server starts.
func TestAdminServer_ActualPort_BeforeStart(t *testing.T) {
	t.Parallel()

	cfg := cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin_port_nostart")
	ctx := context.Background()

	server, err := NewAdminHTTPServer(ctx, cfg)
	require.NoError(t, err)

	// ActualPort before Start should return 0.
	port := server.ActualPort()
	require.Zero(t, port)
}
