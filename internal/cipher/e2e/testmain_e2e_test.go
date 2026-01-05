// Copyright (c) 2025 Justin Cranford
//
//

package e2e_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	cryptoutilCipherServer "cryptoutil/internal/cipher/server"
	"cryptoutil/internal/cipher/server/config"
	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// Shared test resources (initialized once per package).
var (
	sharedHTTPClient   *http.Client
	testCipherIMServer *cryptoutilCipherServer.CipherIMServer
	baseURL            string
	adminURL           string
)

// TestMain initializes cipher-im server with SQLite in-memory for fast E2E tests.
// Service-template handles database, telemetry, and all infrastructure automatically.
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Configure SQLite in-memory for fast E2E tests.
	settings := cryptoutilConfig.RequireNewForTest("cipher-im-e2e-test")
	settings.DatabaseURL = "file::memory:?cache=shared" // SQLite in-memory.
	settings.DatabaseContainer = "disabled"             // No container for E2E.

	cfg := &config.AppConfig{
		ServerSettings: *settings,
		JWTSecret:      uuid.Must(uuid.NewUUID()).String(),
	}

	// Create server with automatic infrastructure (SQLite, telemetry, etc.).
	var err error

	testCipherIMServer, err = cryptoutilCipherServer.NewFromConfig(ctx, cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create server: %v", err))
	}

	// Start server in background (Start() blocks until shutdown).
	errChan := make(chan error, 1)

	go func() {
		if startErr := testCipherIMServer.Start(ctx); startErr != nil {
			errChan <- startErr
		}
	}()

	// Wait for both servers to bind to ports.
	const (
		maxWaitAttempts = 50
		waitInterval    = 100 * time.Millisecond
	)

	var (
		publicPort int
		adminPort  int
	)

	for i := 0; i < maxWaitAttempts; i++ {
		publicPort = testCipherIMServer.PublicPort()

		adminPortValue, _ := testCipherIMServer.AdminPort()
		adminPort = adminPortValue

		if publicPort > 0 && adminPort > 0 {
			break
		}

		select {
		case err := <-errChan:
			panic(fmt.Sprintf("server start error: %v", err))
		case <-time.After(waitInterval):
		}
	}

	if publicPort == 0 {
		panic("public server did not bind to port")
	}

	if adminPort == 0 {
		panic("admin server did not bind to port")
	}

	// Setup HTTP client for tests.
	sharedHTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment only.
			},
		},
		Timeout: cryptoutilMagic.CipherDefaultTimeout,
	}

	// Get server URLs (ports already obtained from wait loop).
	baseURL = fmt.Sprintf("https://127.0.0.1:%d", publicPort)
	adminURL = fmt.Sprintf("https://127.0.0.1:%d", adminPort)

	// Run all E2E tests.
	exitCode := m.Run()

	// Automatic cleanup.
	_ = testCipherIMServer.Shutdown(context.Background())

	os.Exit(exitCode)
}
