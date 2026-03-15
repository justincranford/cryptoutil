// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package im

import (
	"context"
	"crypto/tls"
	"fmt"
	http "net/http"
	"os"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilAppsSmImServer "cryptoutil/internal/apps/sm/im/server"
	cryptoutilAppsSmImServerConfig "cryptoutil/internal/apps/sm/im/server/config"
	cryptoutilAppsSmImTesting "cryptoutil/internal/apps/sm/im/testing"
	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceTestutil "cryptoutil/internal/apps/template/service/testutil"
)

var (
	testSmIMService  *cryptoutilAppsSmImServer.SmIMServer
	sharedHTTPClient *http.Client
	publicBaseURL    string
	adminBaseURL     string
)

// Shared mock servers from template testutil.
var (
	testMockServerOK     = cryptoutilAppsTemplateServiceTestutil.NewMockServerOK()
	testMockServerError  = cryptoutilAppsTemplateServiceTestutil.NewMockServerError()
	testMockServerCustom = cryptoutilAppsTemplateServiceTestutil.NewMockServerCustom()
)

func TestMain(m *testing.M) {
	// Create in-memory SQLite configuration for testing.
	serviceTemplateServerSettings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("sm-im-test")
	serviceTemplateServerSettings.DatabaseURL = cryptoutilSharedMagic.SQLiteInMemoryDSN

	sharedAppConfig := &cryptoutilAppsSmImServerConfig.SmIMServerSettings{
		ServiceTemplateServerSettings: serviceTemplateServerSettings,
	}

	// Start service once for all tests in this package (following e2e pattern).
	testSmIMService = cryptoutilAppsSmImTesting.StartSmIMService(sharedAppConfig)

	// Defer shutdown.
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
		defer cancel()

		if err := testSmIMService.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to shutdown test server: %v\n", err)
		}
	}()

	// Create shared HTTP client using proper TLS certificate validation.
	sharedHTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
				RootCAs:    testSmIMService.TLSRootCAPool(),
			},
			DisableKeepAlives: true,
		},
		Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
	}

	// Get base URLs for tests.
	publicBaseURL = testSmIMService.PublicBaseURL()
	adminBaseURL = testSmIMService.AdminBaseURL()

	// Shared mock servers already initialized from template testutil.
	defer testMockServerOK.Close()
	defer testMockServerError.Close()
	defer testMockServerCustom.Close()

	// Run all tests.
	exitCode := m.Run()

	os.Exit(exitCode)
}
