// Copyright (c) 2025 Justin Cranford
//
// TestMain for JOSE-JA server integration tests.
package server

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"crypto/tls"
	"fmt"
	http "net/http"
	"os"
	"testing"
	"time"

	cryptoutilAppsJoseJaServerConfig "cryptoutil/internal/apps/jose/ja/server/config"
	cryptoutilAppsTemplateServiceTestingE2eHelpers "cryptoutil/internal/apps/template/service/testing/e2e_helpers"
)

var (
	testServer        *JoseJAServer
	testHTTPClient    *http.Client
	testPublicBaseURL string
	testAdminBaseURL  string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Create test configuration.
	cfg := cryptoutilAppsJoseJaServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	// Create server.
	var err error

	testServer, err = NewFromConfig(ctx, cfg)
	if err != nil {
		panic(fmt.Sprintf("TestMain: failed to create server: %v", err))
	}

	// Use generic template helper for goroutine start + dual port polling + panic-on-failure.
	cryptoutilAppsTemplateServiceTestingE2eHelpers.MustStartAndWaitForDualPorts(testServer, func() error {
		return testServer.Start(ctx)
	})

	// Mark server as ready.
	testServer.SetReady(true)

	// Store base URLs for tests.
	testPublicBaseURL, testAdminBaseURL = cryptoutilAppsTemplateServiceTestingE2eHelpers.DualPortBaseURLs(testServer)

	// Create HTTP client that accepts self-signed certificates.
	testHTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // G402: Test client for self-signed certs.
			},
		},
		Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
	}

	// Run all tests.
	exitCode := m.Run()

	// Cleanup: Shutdown server.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.JoseJADefaultMaxMaterials*time.Second)
	defer cancel()

	_ = testServer.Shutdown(shutdownCtx)

	os.Exit(exitCode)
}
