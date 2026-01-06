// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package im

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"cryptoutil/internal/apps/cipher/im/server"
	"cryptoutil/internal/apps/cipher/im/server/config"
	cipherTesting "cryptoutil/internal/apps/cipher/im/testing"
	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilTLS "cryptoutil/internal/shared/crypto/tls"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTestutil "cryptoutil/internal/shared/testutil"
)

var (
	testCipherIMServer *server.CipherIMServer
	sharedHTTPClient   *http.Client
	publicBaseURL      string
	adminBaseURL       string
	// Shared mock servers for testing different response scenarios.
	testMockServerOK     *httptest.Server
	testMockServerError  *httptest.Server
	testMockServerCustom *httptest.Server
)

func TestMain(m *testing.M) {
	// Create in-memory SQLite configuration for testing.
	settings := cryptoutilConfig.RequireNewForTest("cipher-im-test")
	settings.DatabaseURL = sqliteInMemoryURL

	sharedAppConfig := &config.AppConfig{
		ServerSettings: *settings,
		JWTSecret:      googleUuid.Must(googleUuid.NewUUID()).String(),
	}

	// Start server once for all tests in this package (following e2e pattern).
	testCipherIMServer = cipherTesting.StartCipherIMServer(sharedAppConfig)

	// Defer shutdown.
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := testCipherIMServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to shutdown test server: %v\n", err)
		}
	}()

	// Create shared HTTP client for all tests (accepts self-signed certs).
	sharedHTTPClient = cryptoutilTLS.NewClientForTest()

	// Get base URLs for tests.
	publicBaseURL = testCipherIMServer.PublicBaseURL()
	adminBaseURL = testCipherIMServer.AdminBaseURL()

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

// TestIM_SubcommandHelpFlags tests help flag for all subcommands in table-driven format.
func TestIM_SubcommandHelpFlags(t *testing.T) {
	tests := []struct {
		subcommand string
		flagValue  string
		helpTexts  []string
	}{
		{
			subcommand: "client",
			flagValue:  "--help",
			helpTexts:  []string{"cipher im client", "Run client operations"},
		},
		{
			subcommand: "init",
			flagValue:  "--help",
			helpTexts:  []string{"cipher im init", "Initialize database schema"},
		},
		{
			subcommand: "health",
			flagValue:  "--help",
			helpTexts:  []string{"cipher im health", "Check service health"},
		},
		{
			subcommand: "livez",
			flagValue:  "--help",
			helpTexts:  []string{"cipher im livez", "Check service liveness"},
		},
		{
			subcommand: "readyz",
			flagValue:  "--help",
			helpTexts:  []string{"cipher im readyz", "Check service readiness"},
		},
		{
			subcommand: "shutdown",
			flagValue:  "--help",
			helpTexts:  []string{"cipher im shutdown", "Trigger graceful shutdown"},
		},
	}

	for _, tt := range tests {
		// Capture range variable.
		t.Run(tt.subcommand, func(t *testing.T) {
			t.Parallel()

			// Test --help flag.
			var stdout, stderr bytes.Buffer
			exitCode := internalIM([]string{tt.subcommand, tt.flagValue}, &stdout, &stderr)
			require.Equal(t, 0, exitCode, "%s %s should succeed", tt.subcommand, tt.flagValue)

			helpOutput := stdout.String() + stderr.String()
			for _, expected := range tt.helpTexts {
				require.Contains(t, helpOutput, expected, "Help output should contain: %s", expected)
			}

			// Test -h flag.
			stdout.Reset()
			stderr.Reset()
			exitCode = internalIM([]string{tt.subcommand, "-h"}, &stdout, &stderr)
			require.Equal(t, 0, exitCode, "%s -h should succeed", tt.subcommand)

			shortHelpOutput := stdout.String() + stderr.String()
			for _, expected := range tt.helpTexts {
				require.Contains(t, shortHelpOutput, expected, "Help output (-h) should contain: %s", expected)
			}

			// Test with help as positional argument.
			stdout.Reset()
			stderr.Reset()
			exitCode = internalIM([]string{tt.subcommand, "help"}, &stdout, &stderr)
			require.Equal(t, 0, exitCode, "%s help should succeed", tt.subcommand)

			positionalOutput := stdout.String() + stderr.String()
			for _, expected := range tt.helpTexts {
				require.Contains(t, positionalOutput, expected, "Help output (positional) should contain: %s", expected)
			}
		})
	}
}

// TestPrintIMVersion tests the version command.
func TestPrintIMVersion(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	exitCode := internalIM([]string{"version"}, &stdout, &stderr)
	require.Equal(t, 0, exitCode, "version command should succeed")

	output := stdout.String() + stderr.String()

	require.Contains(t, output, "cipher-im service", "Version output should contain cipher-im service")
}

// TestIM_ClientSubcommand_NotImplemented tests client subcommand is not implemented.
func TestIM_ClientSubcommand_NotImplemented(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	exitCode := internalIM([]string{"client"}, &stdout, &stderr)
	require.Equal(t, 1, exitCode, "client subcommand should fail (not implemented)")

	output := stdout.String() + stderr.String()
	require.Contains(t, output, "Client subcommand not yet implemented")
}

// TestIM_InitSubcommand_NotImplemented tests init subcommand is not implemented.
func TestIM_InitSubcommand_NotImplemented(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	exitCode := internalIM([]string{"init"}, &stdout, &stderr)
	require.Equal(t, 1, exitCode, "init subcommand should fail (not implemented)")

	output := stdout.String() + stderr.String()
	require.Contains(t, output, "Init subcommand not yet implemented")
}

// TestIM_HealthSubcommand_LiveServer tests health check with shared test server.
func TestIM_HealthSubcommand_LiveServer(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	args := []string{"health", "--url", publicBaseURL + "/service/api/v1"}
	exitCode := internalIM(args, &stdout, &stderr)
	require.Equal(t, 0, exitCode)

	output := stdout.String() + stderr.String()
	require.Contains(t, output, "Service is healthy")
	require.Contains(t, output, "HTTP 200")
}

// TestIM_LivezSubcommand_LiveServer tests livez check with shared test server.
func TestIM_LivezSubcommand_LiveServer(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	args := []string{"livez", "--url", adminBaseURL}
	exitCode := internalIM(args, &stdout, &stderr)
	require.Equal(t, 0, exitCode)

	output := stdout.String() + stderr.String()
	require.Contains(t, output, "Service is alive")
	require.Contains(t, output, "HTTP 200")
}

// TestIM_ReadyzSubcommand_LiveServer tests readyz check with shared test server.
func TestIM_ReadyzSubcommand_LiveServer(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	args := []string{"readyz", "--url", adminBaseURL}
	exitCode := internalIM(args, &stdout, &stderr)
	// Readyz may return 0 (ready) or 1 (not ready) depending on service state
	// Both are valid responses
	require.Contains(t, []int{0, 1}, exitCode, "readyz should return 0 or 1")

	output := stdout.String() + stderr.String()
	// Check that we got a valid response (either ready or not ready)
	validResponse := strings.Contains(output, "Service is ready") || strings.Contains(output, "Service is not ready")
	require.True(t, validResponse, "Output should indicate readiness status")
}

// TestIM_SubcommandErrors tests error handling for all health check subcommands.
func TestIM_SubcommandErrors(t *testing.T) {
	tests := []struct {
		name       string
		subcommand string
		url        string
		contains   []string
	}{
		{
			name:       "health_invalid_url",
			subcommand: "health",
			url:        "://invalid-url",
			contains:   []string{"Health check failed:"},
		},
		{
			name:       "health_connection_refused",
			subcommand: "health",
			url:        "https://127.0.0.1:9999",
			contains:   []string{"Health check failed:", "connection refused", "actively refused"},
		},
		{
			name:       "livez_invalid_url",
			subcommand: "livez",
			url:        "://invalid-url",
			contains:   []string{"Liveness check failed:"},
		},
		{
			name:       "livez_connection_refused",
			subcommand: "livez",
			url:        "https://127.0.0.1:9997",
			contains:   []string{"Liveness check failed:", "connection refused", "actively refused"},
		},
		{
			name:       "readyz_invalid_url",
			subcommand: "readyz",
			url:        "://invalid-url",
			contains:   []string{"Readiness check failed:"},
		},
		{
			name:       "readyz_connection_refused",
			subcommand: "readyz",
			url:        "https://127.0.0.1:9995",
			contains:   []string{"Readiness check failed:", "connection refused", "actively refused"},
		},
		{
			name:       "shutdown_invalid_url",
			subcommand: "shutdown",
			url:        "://invalid-url",
			contains:   []string{"Shutdown request failed:"},
		},
		{
			name:       "shutdown_connection_refused",
			subcommand: "shutdown",
			url:        "https://127.0.0.1:9994",
			contains:   []string{"Shutdown request failed:", "connection refused", "actively refused"},
		},
	}

	for _, tt := range tests {
		// Capture range variable.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			exitCode := internalIM([]string{tt.subcommand, "--url", tt.url}, &stdout, &stderr)
			require.Equal(t, 1, exitCode, "%s should fail", tt.subcommand)

			output := stdout.String() + stderr.String()
			// Check that output contains at least one of the expected strings.
			require.True(t, cryptoutilTestutil.ContainsAny(output, tt.contains),
				"output should contain one of %v: %s", tt.contains, output)
		})
	}
}

// TestIM_SubcommandResponseBodies tests response body handling for all subcommands.
func TestIM_SubcommandResponseBodies(t *testing.T) {
	tests := []struct {
		name         string
		subcommand   string
		serverURL    string
		path         string
		expectExit   int
		expectOutput []string
	}{
		// Health subcommand tests.
		{
			name:         "health_success_with_body",
			subcommand:   "health",
			serverURL:    testMockServerCustom.URL,
			path:         cryptoutilMagic.DefaultPrivateAdminAPIContextPath + "/health",
			expectExit:   0,
			expectOutput: []string{"Service is healthy", "200", "All systems operational"},
		},
		{
			name:         "health_success_no_body",
			subcommand:   "health",
			serverURL:    testMockServerOK.URL,
			path:         cryptoutilMagic.DefaultPrivateAdminAPIContextPath + "/health",
			expectExit:   0,
			expectOutput: []string{"Service is healthy"},
		},
		{
			name:         "health_unhealthy_with_body",
			subcommand:   "health",
			serverURL:    testMockServerError.URL,
			path:         cryptoutilMagic.DefaultPrivateAdminAPIContextPath + "/health",
			expectExit:   1,
			expectOutput: []string{"Service is unhealthy", "503", "Service Unavailable"},
		},
		// Livez subcommand tests.
		{
			name:         "livez_alive_with_body",
			subcommand:   "livez",
			serverURL:    testMockServerCustom.URL,
			path:         cryptoutilMagic.DefaultPrivateAdminAPIContextPath + cryptoutilMagic.PrivateAdminLivezRequestPath,
			expectExit:   0,
			expectOutput: []string{"Service is alive", "200", "Process is alive and running"},
		},
		{
			name:         "livez_alive_no_body",
			subcommand:   "livez",
			serverURL:    testMockServerOK.URL,
			path:         cryptoutilMagic.DefaultPrivateAdminAPIContextPath + cryptoutilMagic.PrivateAdminLivezRequestPath,
			expectExit:   0,
			expectOutput: []string{"Service is alive"},
		},
		{
			name:         "livez_not_alive_with_body",
			subcommand:   "livez",
			serverURL:    testMockServerError.URL,
			path:         cryptoutilMagic.DefaultPrivateAdminAPIContextPath + cryptoutilMagic.PrivateAdminLivezRequestPath,
			expectExit:   1,
			expectOutput: []string{"Service is not alive", "503", "Service Unavailable"},
		},
		// Shutdown subcommand tests.
		{
			name:         "shutdown_success_no_body",
			subcommand:   "shutdown",
			serverURL:    testMockServerOK.URL,
			path:         cryptoutilMagic.DefaultPrivateAdminAPIContextPath + cryptoutilMagic.PrivateAdminShutdownRequestPath,
			expectExit:   0,
			expectOutput: []string{"Shutdown initiated"},
		},
		{
			name:         "shutdown_failed_no_body",
			subcommand:   "shutdown",
			serverURL:    testMockServerError.URL,
			path:         cryptoutilMagic.DefaultPrivateAdminAPIContextPath + cryptoutilMagic.PrivateAdminShutdownRequestPath,
			expectExit:   1,
			expectOutput: []string{"Shutdown request failed", "503"},
		},
	}

	for _, tt := range tests {
		// Capture range variable.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			exitCode := internalIM([]string{tt.subcommand, "--url", tt.serverURL + tt.path}, &stdout, &stderr)
			require.Equal(t, tt.expectExit, exitCode, "%s should exit with code %d", tt.subcommand, tt.expectExit)

			output := stdout.String() + stderr.String()
			for _, expected := range tt.expectOutput {
				require.Contains(t, output, expected, "Output should contain: %s", expected)
			}
		})
	}
}

// TestIM_URLHandling tests URL parsing and suffix handling for all subcommands.
func TestIM_URLHandling(t *testing.T) {
	tests := []struct {
		name       string
		subcommand string
		urlSuffix  string
	}{
		{
			name:       "health_with_health_suffix",
			subcommand: "health",
			urlSuffix:  cryptoutilMagic.DefaultPrivateAdminAPIContextPath + "/health",
		},
		{
			name:       "livez_with_livez_suffix",
			subcommand: "livez",
			urlSuffix:  cryptoutilMagic.DefaultPrivateAdminAPIContextPath + cryptoutilMagic.PrivateAdminLivezRequestPath,
		},
		{
			name:       "readyz_with_readyz_suffix",
			subcommand: "readyz",
			urlSuffix:  cryptoutilMagic.DefaultPrivateAdminAPIContextPath + cryptoutilMagic.PrivateAdminReadyzRequestPath,
		},
		{
			name:       "shutdown_with_shutdown_suffix",
			subcommand: "shutdown",
			urlSuffix:  cryptoutilMagic.DefaultPrivateAdminAPIContextPath + cryptoutilMagic.PrivateAdminShutdownRequestPath,
		},
	}

	for _, tt := range tests {
		// Capture range variable.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer
			exitCode := internalIM([]string{tt.subcommand, "--url", testMockServerCustom.URL + tt.urlSuffix}, &stdout, &stderr)
			require.Equal(t, 0, exitCode, "%s should succeed with explicit suffix", tt.subcommand)

			output := stdout.String() + stderr.String()
			require.NotContains(t, output, "failed")
		})
	}
}
