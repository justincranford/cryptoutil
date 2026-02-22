// Copyright (c) 2025 Justin Cranford

package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	cryptoutilAppsIdentityAuthzServerConfig "cryptoutil/internal/apps/identity/authz/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilPoll "cryptoutil/internal/shared/util/poll"
)

var (
	testServer  *AuthzServer
	testBaseURL string
	testErr     error
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Create test configuration with dynamic port allocation.
	cfg := cryptoutilAppsIdentityAuthzServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	testServer, testErr = NewFromConfig(ctx, cfg)
	if testErr != nil {
		log.Printf("Failed to create test server: %v", testErr)
		os.Exit(1)
	}

	// Start server in background.
	go func() {
		if startErr := testServer.Start(ctx); startErr != nil {
			log.Printf("Server start error: %v", startErr)
		}
	}()

	// Wait for server to be ready.
	if readyErr := waitForReady(ctx, testServer); readyErr != nil {
		log.Printf("Server not ready: %v", readyErr)
		os.Exit(1)
	}

	// Set base URL after server is running (dynamic port).
	testBaseURL = testServer.PublicBaseURL()

	// Mark server as ready.
	testServer.SetReady(true)

	// Run tests.
	code := m.Run()

	// Shutdown server.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if shutdownErr := testServer.Shutdown(shutdownCtx); shutdownErr != nil {
		log.Printf("Shutdown error: %v", shutdownErr)
	}

	os.Exit(code)
}

// Timeout constants for test operations.
const (
	readyTimeout    = 10 * time.Second
	checkInterval   = 100 * time.Millisecond
	shutdownTimeout = 5 * time.Second
)

// waitForReady waits for the server to be ready.
func waitForReady(ctx context.Context, server *AuthzServer) error {
	return cryptoutilSharedUtilPoll.Until(ctx, readyTimeout, checkInterval, func(_ context.Context) (bool, error) {
		return server.PublicPort() > 0, nil
	})
}

// requireTestSetup ensures test server is running before tests execute.
func requireTestSetup(t *testing.T) {
	t.Helper()

	if testErr != nil {
		t.Fatalf("Test setup failed: %v", testErr)
	}

	if testServer == nil {
		t.Fatal("Test server not initialized")
	}
}
