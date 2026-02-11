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

	cryptoutilAppsCipherImServer "cryptoutil/internal/apps/cipher/im/server"
	cryptoutilAppsCipherImServerConfig "cryptoutil/internal/apps/cipher/im/server/config"
	cryptoutilAppsCipherImTesting "cryptoutil/internal/apps/cipher/im/testing"
	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceTestutil "cryptoutil/internal/apps/template/service/testutil"
	cryptoutilSharedCryptoTls "cryptoutil/internal/shared/crypto/tls"
)

var (
	testCipherIMService *cryptoutilAppsCipherImServer.CipherIMServer
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
	serviceTemplateServerSettings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("cipher-im-test")
	serviceTemplateServerSettings.DatabaseURL = sqliteInMemoryURL

	sharedAppConfig := &cryptoutilAppsCipherImServerConfig.CipherImServerSettings{
		ServiceTemplateServerSettings: serviceTemplateServerSettings,
	}

	// Start service once for all tests in this package (following e2e pattern).
	testCipherIMService = cryptoutilAppsCipherImTesting.StartCipherIMService(sharedAppConfig)

	// Defer shutdown.
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := testCipherIMService.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to shutdown test server: %v\n", err)
		}
	}()

	// Create shared HTTP client for all tests (accepts self-signed certs).
	sharedHTTPClient = cryptoutilSharedCryptoTls.NewClientForTest()

	// Get base URLs for tests.
	publicBaseURL = testCipherIMService.PublicBaseURL()
	adminBaseURL = testCipherIMService.AdminBaseURL()

	// Shared mock servers already initialized from template testutil.
	defer testMockServerOK.Close()
	defer testMockServerError.Close()
	defer testMockServerCustom.Close()

	// Run all tests.
	exitCode := m.Run()

	os.Exit(exitCode)
}
