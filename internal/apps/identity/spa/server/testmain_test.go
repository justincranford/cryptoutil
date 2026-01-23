// Copyright (c) 2025 Justin Cranford

package server

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	cryptoutilSPAConfig "cryptoutil/internal/apps/identity/spa/server/config"
)

var (
	testServer  *SPAServer
	testBaseURL string
	testErr     error
)

func TestMain(m *testing.M) {
	// Create test configuration.
	cfg := cryptoutilSPAConfig.DefaultTestConfig()

	// Create server.
	ctx := context.Background()

	testServer, testErr = NewFromConfig(ctx, cfg)
	if testErr != nil {
		fmt.Printf("Failed to create test server: %v\n", testErr)
		os.Exit(1)
	}

	// Start server in background.
	go func() {
		if err := testServer.Start(ctx); err != nil {
			fmt.Printf("Server start error: %v\n", err)
		}
	}()

	// Wait for server to be ready.
	if !waitForReady(testServer, serverReadyTimeout) {
		fmt.Println("Server failed to become ready")
		os.Exit(1)
	}

	// Set base URL after server starts (uses dynamic port).
	testBaseURL = testServer.PublicBaseURL()

	// Run tests.
	exitCode := m.Run()

	// Cleanup.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := testServer.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Shutdown error: %v\n", err)
	}

	os.Exit(exitCode)
}

// waitForReady waits for the server to become ready.
func waitForReady(server *SPAServer, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if server.PublicPort() > 0 && server.AdminPort() > 0 {
			return true
		}

		time.Sleep(checkInterval)
	}

	return false
}

// requireTestSetup checks that the test server is properly initialized.
func requireTestSetup(t *testing.T) {
	t.Helper()

	if testErr != nil {
		t.Fatalf("Test server setup failed: %v", testErr)
	}

	if testServer == nil {
		t.Fatal("Test server is nil")
	}
}

// Timeout constants for test operations.
const (
	serverReadyTimeout = 30 * time.Second
	shutdownTimeout    = 10 * time.Second
	checkInterval      = 100 * time.Millisecond
)
