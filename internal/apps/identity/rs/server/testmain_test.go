// Copyright (c) 2025 Justin Cranford

package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	cryptoutilAppsIdentityRsServerConfig "cryptoutil/internal/apps/identity/rs/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	testServer  *RSServer
	testBaseURL string
	testErr     error
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Create test configuration with dynamic port allocation.
	cfg := cryptoutilAppsIdentityRsServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

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
func waitForReady(ctx context.Context, server *RSServer) error {
	deadline := time.Now().UTC().Add(readyTimeout)

	for time.Now().UTC().Before(deadline) {
		// Check if public port is allocated.
		if server.PublicPort() > 0 {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for server: %w", ctx.Err())
		default:
			time.Sleep(checkInterval)
		}
	}

	return context.DeadlineExceeded
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
