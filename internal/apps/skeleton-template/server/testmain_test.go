// Copyright (c) 2025-2026 Justin Cranford.
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

	cryptoutilTestHelpApi "cryptoutil/internal/apps-framework/service/test_help_api"
	cryptoutilTestOrcIntegration "cryptoutil/internal/apps-framework/service/test_orch_integration"
	cryptoutilAppsSkeletonTemplateServerConfig "cryptoutil/internal/apps/skeleton-template/server/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	testServer            *SkeletonTemplateServer
	testIntegrationServer *cryptoutilTestOrcIntegration.IntegrationServer
	testPublicHTTPClient  *http.Client
	testAdminHTTPClient   *http.Client
	testHealthClient      *cryptoutilTestHelpApi.HealthClient
	testPublicBaseURL     string
	testAdminBaseURL      string
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

	// Start server and wait for both ports to bind.
	testIntegrationServer, err = cryptoutilTestOrcIntegration.StartIntegrationServerForTestMain(ctx, testServer, nil)
	if err != nil {
		panic(fmt.Sprintf("TestMain: failed to start server: %v", err))
	}

	// Store base URLs for tests.
	testPublicBaseURL = testServer.PublicBaseURL()
	testAdminBaseURL = testServer.AdminBaseURL()
	testHealthClient = cryptoutilTestHelpApi.NewHealthClient(testPublicBaseURL, testAdminBaseURL)

	// Create public and admin HTTPS clients.
	testPublicHTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
				RootCAs:    testServer.TLSRootCAPool(),
			},
			DisableKeepAlives: true,
		},
		Timeout: cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Second,
	}

	testAdminHTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
				RootCAs:    testServer.AdminTLSRootCAPool(),
			},
			DisableKeepAlives: true,
		},
		Timeout: cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Second,
	}

	// Run all tests.
	exitCode := m.Run()

	// Cleanup: Shutdown server.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultDataServerShutdownTimeout)
	defer cancel()

	_ = testIntegrationServer.Shutdown(shutdownCtx)

	os.Exit(exitCode)
}
