// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package im

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"

	"cryptoutil/internal/apps/cipher/im/server"
	"cryptoutil/internal/apps/cipher/im/server/config"
	cipherTesting "cryptoutil/internal/apps/cipher/im/testing"
	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTLS "cryptoutil/internal/shared/crypto/tls"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

var (
	testCipherIMService *server.CipherIMServer
	sharedHTTPClient    *http.Client
	publicBaseURL       string
	adminBaseURL        string
	// Shared mock servers for testing different response scenarios.
	testMockServerOK     *httptest.Server
	testMockServerError  *httptest.Server
	testMockServerCustom *httptest.Server
)

func TestMain(m *testing.M) {
	// Create in-memory SQLite configuration for testing.
	settings := cryptoutilConfig.RequireNewForTest("cipher-im-test")
	settings.DatabaseURL = sqliteInMemoryURL

	sharedAppConfig := &config.CipherImServerSettings{
		ServiceTemplateServerSettings: *settings,
		JWTSecret:                     googleUuid.Must(googleUuid.NewUUID()).String(),
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
	testMockServerOK = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "OK")
	}))
	defer testMockServerOK.Close()

	// Create shared mock server that returns errors.
	testMockServerError = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprint(w, "Service Unavailable")
	}))
	defer testMockServerError.Close()

	// Create shared mock server for custom responses (controlled by path).
	testMockServerCustom = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case cryptoutilMagic.DefaultPrivateAdminAPIContextPath + "/health":
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, "All systems operational")
		case cryptoutilMagic.DefaultPrivateAdminAPIContextPath + cryptoutilMagic.PrivateAdminLivezRequestPath:
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, "Process is alive and running")
		case cryptoutilMagic.DefaultPrivateAdminAPIContextPath + cryptoutilMagic.PrivateAdminReadyzRequestPath:
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, "Ready")
		case cryptoutilMagic.DefaultPrivateAdminAPIContextPath + cryptoutilMagic.PrivateAdminShutdownRequestPath:
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprint(w, "Shutting down")
			} else {
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer testMockServerCustom.Close()

	// Run all tests.
	exitCode := m.Run()

	os.Exit(exitCode)
}
