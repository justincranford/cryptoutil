// Copyright (c) 2025 Justin Cranford
//
// TestMain for pki-ca server integration tests.
package server

import (
	"context"
	"crypto/tls"
	"fmt"
	http "net/http"
	"os"
	"testing"
	"time"

	cryptoutilE2EHelpers "cryptoutil/internal/apps/framework/service/testing/e2e_helpers"
	cryptoutilAppsCaServerConfig "cryptoutil/internal/apps/pki-ca/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	testServer        *CAServer
	testHTTPClient    *http.Client
	testPublicBaseURL string
	testAdminBaseURL  string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Create test configuration.
	cfg := cryptoutilAppsCaServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	// Create server.
	var err error

	testServer, err = NewFromConfig(ctx, cfg)
	if err != nil {
		panic(fmt.Sprintf("TestMain: failed to create server: %v", err))
	}

	// Start server in background and wait for both ports to bind.
	cryptoutilE2EHelpers.MustStartAndWaitForDualPorts(testServer, func() error {
		return testServer.Start(ctx)
	})

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
			DisableKeepAlives: true,
		},
		Timeout: cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Second,
	}

	// Run all tests.
	exitCode := m.Run()

	// Cleanup: Shutdown server.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.JoseJADefaultMaxMaterials*time.Second)
	defer cancel()

	_ = testServer.Shutdown(shutdownCtx)

	os.Exit(exitCode)
}
