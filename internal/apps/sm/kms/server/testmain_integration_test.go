//go:build integration

// Copyright (c) 2025 Justin Cranford
//
// TestMain for SM-KMS server integration tests.
package server

import (
	"context"
	"crypto/tls"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	http "net/http"
	"os"
	"testing"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceTestingE2eHelpers "cryptoutil/internal/apps/template/service/testing/e2e_helpers"
)

var (
	testIntegrationServer    *KMSServer
	testIntegrationClient    *http.Client
	testIntegrationPublicURL string
	testIntegrationAdminURL  string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Create test configuration using template helper.
	cfg := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("kms-server-integration")

	// Create KMS server.
	var err error

	testIntegrationServer, err = NewKMSServer(ctx, cfg)
	if err != nil {
		panic(fmt.Sprintf("TestMain: failed to create KMS server: %v", err))
	}

	// Use generic template helper for goroutine start + dual port polling + panic-on-failure.
	// KMSServer.Start() has no ctx parameter, so closure wraps accordingly.
	cryptoutilAppsTemplateServiceTestingE2eHelpers.MustStartAndWaitForDualPorts(testIntegrationServer, func() error {
		return testIntegrationServer.Start()
	})

	// Store base URLs for tests.
	testIntegrationPublicURL, testIntegrationAdminURL = cryptoutilAppsTemplateServiceTestingE2eHelpers.DualPortBaseURLs(testIntegrationServer)

	// Create HTTP client that accepts self-signed certificates.
	testIntegrationClient = &http.Client{
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
	testIntegrationServer.Shutdown()

	os.Exit(exitCode)
}
