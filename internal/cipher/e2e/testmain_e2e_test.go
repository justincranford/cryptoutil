// Copyright (c) 2025 Justin Cranford
//
//

package e2e_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	cryptoutilCipherServer "cryptoutil/internal/cipher/server"
	cipherTesting "cryptoutil/internal/cipher/testing"
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

	// Setup: Create shared heavyweight resources ONCE using helper.
	resources, err := cipherTesting.SetupTestServer(ctx, true) // Use in-memory DB for E2E tests.
	if err != nil {
		panic("TestMain: failed to setup test server: " + err.Error())
	}
	defer resources.Shutdown(context.Background())

	// Assign to package-level variables for E2E tests.
	testCipherIMServer = resources.CipherIMServer
	baseURL = resources.BaseURL
	adminURL = resources.AdminURL
	sharedHTTPClient = resources.HTTPClient

	// Run all E2E tests.
	exitCode := m.Run()

	os.Exit(exitCode)
}
