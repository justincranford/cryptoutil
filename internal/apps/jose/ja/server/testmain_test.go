// Copyright (c) 2025 Justin Cranford
//
// TestMain for JOSE-JA server integration tests.
package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"cryptoutil/internal/apps/jose/ja/server/config"
)

var (
	testServer *JoseJAServer
	testHTTPClient *http.Client
	testPublicBaseURL string
	testAdminBaseURL string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Create test configuration.
	cfg := config.NewTestConfig("127.0.0.1", 0, true)

	// Create server.
	var err error
	testServer, err = NewFromConfig(ctx, cfg)
	if err != nil {
		panic(fmt.Sprintf("TestMain: failed to create server: %v", err))
	}

	// Start server in background.
	errChan := make(chan error, 1)
	go func() {
		if startErr := testServer.Start(ctx); startErr != nil {
			errChan <- startErr
		}
	}()

	// Wait for server ports to be assigned.
	const (
		maxWaitAttempts = 50
		waitInterval    = 100 * time.Millisecond
	)

	var publicPort, adminPort int
	for i := 0; i < maxWaitAttempts; i++ {
		publicPort = testServer.PublicPort()
		adminPort = testServer.AdminPort()

		if publicPort > 0 && adminPort > 0 {
			break
		}

		select {
		case err := <-errChan:
			panic(fmt.Sprintf("TestMain: server failed to start: %v", err))
		case <-time.After(waitInterval):
		}
	}

	if publicPort == 0 || adminPort == 0 {
		panic("TestMain: server did not bind to ports")
	}

	// Mark server as ready.
	testServer.SetReady(true)

	// Store base URLs for tests.
	testPublicBaseURL = testServer.PublicBaseURL()
	testAdminBaseURL = testServer.AdminBaseURL()

	// Create HTTP client that accepts self-signed certificates.
	testHTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // G402: Test client for self-signed certs.
			},
		},
		Timeout: 5 * time.Second,
	}

	// Run all tests.
	exitCode := m.Run()

	// Cleanup: Shutdown server.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = testServer.Shutdown(shutdownCtx)

	os.Exit(exitCode)
}
