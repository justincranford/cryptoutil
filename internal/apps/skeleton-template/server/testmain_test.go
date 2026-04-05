// Copyright (c) 2025 Justin Cranford
//
// TestMain for skeleton-template server integration tests.
package server

import (
	"context"
	"crypto/tls"
	"fmt"
	http "net/http"
	"os"
	"testing"
	"time"

	cryptoutilAppsFrameworkServiceTestingE2eHelpers "cryptoutil/internal/apps/framework/service/testing/e2e_helpers"
	cryptoutilTestingHealthclient "cryptoutil/internal/apps/framework/service/testing/healthclient"
	cryptoutilAppsSkeletonTemplateServerConfig "cryptoutil/internal/apps/skeleton-template/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	testServer           *SkeletonTemplateServer
	testPublicHTTPClient *http.Client
	testAdminHTTPClient  *http.Client
	testHealthClient     *cryptoutilTestingHealthclient.HealthClient
	testPublicBaseURL    string
	testAdminBaseURL     string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Create test configuration.
	cfg := cryptoutilAppsSkeletonTemplateServerConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	// Create server.
	var err error

	testServer, err = NewFromConfig(ctx, cfg)
	if err != nil {
		panic(fmt.Sprintf("TestMain: failed to create server: %v", err))
	}

	// Use generic template helper for goroutine start + dual port polling + panic-on-failure.
	cryptoutilAppsFrameworkServiceTestingE2eHelpers.MustStartAndWaitForDualPorts(testServer, func() error {
		return testServer.Start(ctx)
	})

	// Mark server as ready.
	testServer.SetReady(true)

	// Store base URLs for tests.
	testPublicBaseURL, testAdminBaseURL = cryptoutilAppsFrameworkServiceTestingE2eHelpers.DualPortBaseURLs(testServer)
	testHealthClient = cryptoutilTestingHealthclient.NewHealthClient(testPublicBaseURL, testAdminBaseURL)

	// Create public and admin HTTPS clients.
	testPublicHTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
				RootCAs:    testServer.TLSRootCAPool(),
			},
			DisableKeepAlives: true,
		},
		Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
	}

	testAdminHTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
				RootCAs:    testServer.AdminTLSRootCAPool(),
			},
			DisableKeepAlives: true,
		},
		Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
	}

	// Run all tests.
	exitCode := m.Run()

	// Cleanup: Shutdown server.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultDataServerShutdownTimeout)
	defer cancel()

	_ = testServer.Shutdown(shutdownCtx)

	os.Exit(exitCode)
}
