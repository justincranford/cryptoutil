// Copyright (c) 2025 Justin Cranford

package server_test

import (
	"context"
	"crypto/tls"
	"fmt"
	http "net/http"
	"os"
	"testing"
	"time"

	cryptoutilAppsIdentityRpServer "cryptoutil/internal/apps/identity/rp/server"
	cryptoutilAppsIdentityRpServerConfig "cryptoutil/internal/apps/identity/rp/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilPoll "cryptoutil/internal/shared/util/poll"

	"github.com/stretchr/testify/require"
)

// Timeout constants for test operations.
const (
	readyTimeout    = 10 * time.Second
	checkInterval   = 100 * time.Millisecond
	shutdownTimeout = 5 * time.Second
)

// testServer is the shared server instance for all tests.
var testServer *cryptoutilAppsIdentityRpServer.RPServer //nolint:gochecknoglobals

// testHTTPClient is a pre-configured HTTP client for testing.
var testHTTPClient *http.Client //nolint:gochecknoglobals

// testPublicBaseURL is the public server base URL.
var testPublicBaseURL string //nolint:gochecknoglobals

// testAdminBaseURL is the admin server base URL.
var testAdminBaseURL string //nolint:gochecknoglobals

// TestMain sets up the test server once for all tests.
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Create test configuration.
	cfg := cryptoutilAppsIdentityRpServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	// Create server.
	var err error

	testServer, err = cryptoutilAppsIdentityRpServer.NewFromConfig(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create RP server: %v\n", err)
		os.Exit(1)
	}

	// Start server in background.
	go func() {
		if startErr := testServer.Start(ctx); startErr != nil {
			fmt.Fprintf(os.Stderr, "Server start error: %v\n", startErr)
		}
	}()

	// Wait for server to be ready using polling pattern.
	if readyErr := waitForReady(ctx, testServer); readyErr != nil {
		fmt.Fprintf(os.Stderr, "Server not ready: %v\n", readyErr)
		os.Exit(1)
	}

	// Capture base URLs.
	testPublicBaseURL = testServer.PublicBaseURL()
	testAdminBaseURL = testServer.AdminBaseURL()

	// Create HTTP client with TLS config.
	testHTTPClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment
				MinVersion:         tls.VersionTLS12,
			},
		},
	}

	// Mark server as ready.
	testServer.SetReady(true)

	// Run tests.
	exitCode := m.Run()

	// Shutdown server.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if shutdownErr := testServer.Shutdown(shutdownCtx); shutdownErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to shutdown server: %v\n", shutdownErr)
	}

	os.Exit(exitCode)
}

// waitForReady waits for the server to be ready by polling port allocation.
func waitForReady(ctx context.Context, srv *cryptoutilAppsIdentityRpServer.RPServer) error {
	return cryptoutilSharedUtilPoll.Until(ctx, readyTimeout, checkInterval, func(_ context.Context) (bool, error) {
		return srv.PublicPort() > 0, nil
	})
}

// requireTestSetup verifies test infrastructure is available.
func requireTestSetup(t *testing.T) {
	t.Helper()

	require.NotNil(t, testServer, "test server should be initialized")
	require.NotNil(t, testHTTPClient, "test HTTP client should be initialized")
	require.NotEmpty(t, testPublicBaseURL, "public base URL should be set")
	require.NotEmpty(t, testAdminBaseURL, "admin base URL should be set")
}
