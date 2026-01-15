// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package im

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"cryptoutil/internal/apps/cipher/im/server"
	"cryptoutil/internal/apps/cipher/im/server/config"
	cipherTesting "cryptoutil/internal/apps/cipher/im/testing"
	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTestutil "cryptoutil/internal/apps/template/service/testutil"
	cryptoutilTLS "cryptoutil/internal/shared/crypto/tls"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

var (
	testCipherIMService *server.CipherIMServer
	sharedHTTPClient    *http.Client
	publicBaseURL       string
	adminBaseURL        string
	// Shared mock servers from template testutil.
	testMockServerOK    = cryptoutilTestutil.NewMockServerOK()
	testMockServerError = cryptoutilTestutil.NewMockServerError()
)

func TestMain(m *testing.M) {
	// Create in-memory SQLite configuration for testing.
	serviceTemplateServerSettings := cryptoutilConfig.RequireNewForTest("cipher-im-test")
	serviceTemplateServerSettings.DatabaseURL = sqliteInMemoryURL

	sharedAppConfig := &config.CipherImServerSettings{
		ServiceTemplateServerSettings: serviceTemplateServerSettings,
	}

	// Start service once for all tests in this package (following e2e pattern).
	testCipherIMService = cipherTesting.StartCipherIMService(sharedAppConfig)

	// Defer shutdown.
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := testCipherIMService.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to shutdown test server: %v\n", err)
		}
	}()

	// Create shared HTTP client for all tests (accepts self-signed certs).
	sharedHTTPClient = cryptoutilTLS.NewClientForTest()

	// Get base URLs for tests.
	publicBaseURL = testCipherIMService.PublicBaseURL()
	adminBaseURL = testCipherIMService.AdminBaseURL()

	// Create shared mock server that returns 200 OK.
	tesShared mock servers already initialized from template testutil.
	defer testMockServerOK.Close()
	defer testMockServerError
	// Run all tests.
	exitCode := m.Run()

	os.Exit(exitCode)
}
