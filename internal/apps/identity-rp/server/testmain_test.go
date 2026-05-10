// Copyright (c) 2025-2026 Justin Cranford.
package server_test

import (
	"context"
	"crypto/tls"
	http "net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilTestOrcIntegration "cryptoutil/internal/apps-framework/service/test_orch_integration"
	cryptoutilAppsIdentityRpServer "cryptoutil/internal/apps/identity-rp/server"
	cryptoutilAppsIdentityRpServerConfig "cryptoutil/internal/apps/identity-rp/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// testServer is the shared server instance for all tests.
var testServer *cryptoutilAppsIdentityRpServer.RPServer //nolint:gochecknoglobals

// testPublicHTTPClient is a pre-configured HTTP client for public endpoints.
var testPublicHTTPClient *http.Client //nolint:gochecknoglobals

// testAdminHTTPClient is a pre-configured HTTP client for admin endpoints.
var testAdminHTTPClient *http.Client //nolint:gochecknoglobals

// testPublicBaseURL is the public server base URL.
var testPublicBaseURL string //nolint:gochecknoglobals

// testAdminBaseURL is the admin server base URL.
var testAdminBaseURL string //nolint:gochecknoglobals

// testIntegrationServer is the orchestration handle for cleanup.
var testIntegrationServer *cryptoutilTestOrcIntegration.IntegrationServer //nolint:gochecknoglobals

// TestMain sets up the test server once for all tests.
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Create test configuration.
	cfg := cryptoutilAppsIdentityRpServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	// Create server.
	var err error

	testServer, err = cryptoutilAppsIdentityRpServer.NewFromConfig(ctx, cfg)
	if err != nil {
		panic("TestMain: failed to create server: " + err.Error())
	}

	// Start server and wait for both ports to bind.
	testIntegrationServer, err = cryptoutilTestOrcIntegration.StartIntegrationServerForTestMain(ctx, testServer, nil)
	if err != nil {
		panic("TestMain: failed to start server: " + err.Error())
	}

	// Capture base URLs.
	testPublicBaseURL = testServer.PublicBaseURL()
	testAdminBaseURL = testServer.AdminBaseURL()

	// Create HTTP clients with TLS config.
	testPublicHTTPClient = &http.Client{
		Timeout: cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
				RootCAs:    testServer.TLSRootCAPool(),
			},
			DisableKeepAlives: true,
		},
	}
	testAdminHTTPClient = &http.Client{
		Timeout: cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
				RootCAs:    testServer.AdminTLSRootCAPool(),
			},
			DisableKeepAlives: true,
		},
	}

	// Run tests.
	exitCode := m.Run()

	// Shutdown server.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultDataServerShutdownTimeout)
	defer cancel()

	_ = testIntegrationServer.Shutdown(shutdownCtx)

	os.Exit(exitCode)
}

// requireTestSetup verifies test infrastructure is available.
func requireTestSetup(t *testing.T) {
	t.Helper()

	require.NotNil(t, testServer, "test server should be initialized")
	require.NotNil(t, testPublicHTTPClient, "public HTTP client should be initialized")
	require.NotNil(t, testAdminHTTPClient, "admin HTTP client should be initialized")
	require.NotEmpty(t, testPublicBaseURL, "public base URL should be set")
	require.NotEmpty(t, testAdminBaseURL, "admin base URL should be set")
}
