// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package im

import (
	"context"
	"fmt"
	http "net/http"
	"os"
	"testing"
	"time"

	cryptoutilAppsSmImServer "cryptoutil/internal/apps/sm/im/server"
	cryptoutilAppsSmImServerConfig "cryptoutil/internal/apps/sm/im/server/config"
	cryptoutilAppsSmImTesting "cryptoutil/internal/apps/sm/im/testing"
	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceTestutil "cryptoutil/internal/apps/template/service/testutil"
	cryptoutilSharedCryptoTls "cryptoutil/internal/shared/crypto/tls"
)

var (
	testSmIMService *cryptoutilAppsSmImServer.SmIMServer
	sharedHTTPClient    *http.Client
	publicBaseURL       string
	adminBaseURL        string
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
	serviceTemplateServerSettings.DatabaseURL = sqliteInMemoryURL

	sharedAppConfig := &cryptoutilAppsSmImServerConfig.SmIMServerSettings{
		ServiceTemplateServerSettings: serviceTemplateServerSettings,
	}

	// Start service once for all tests in this package (following e2e pattern).
	testSmIMService = cryptoutilAppsSmImTesting.StartSmIMService(sharedAppConfig)

	// Defer shutdown.
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := testSmIMService.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to shutdown test server: %v\n", err)
		}
	}()

	// Create shared HTTP client for all tests (accepts self-signed certs).
	sharedHTTPClient = cryptoutilSharedCryptoTls.NewClientForTest()

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
