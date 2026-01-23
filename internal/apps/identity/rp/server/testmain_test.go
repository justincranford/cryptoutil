// Copyright (c) 2025 Justin Cranford

package server_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"cryptoutil/internal/apps/identity/rp/server"
	"cryptoutil/internal/apps/identity/rp/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// testServer is the shared server instance for all tests.
var testServer *server.RPServer //nolint:gochecknoglobals

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
	cfg := config.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	// Create server.
	var err error

	testServer, err = server.NewFromConfig(ctx, cfg)
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

	// Wait for server to be ready.
	time.Sleep(500 * time.Millisecond)

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
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if shutdownErr := testServer.Shutdown(shutdownCtx); shutdownErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to shutdown server: %v\n", shutdownErr)
	}

	os.Exit(exitCode)
}

// requireTestSetup verifies test infrastructure is available.
func requireTestSetup(t *testing.T) {
	t.Helper()

	require.NotNil(t, testServer, "test server should be initialized")
	require.NotNil(t, testHTTPClient, "test HTTP client should be initialized")
	require.NotEmpty(t, testPublicBaseURL, "public base URL should be set")
	require.NotEmpty(t, testAdminBaseURL, "admin base URL should be set")
}
