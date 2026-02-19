// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package im

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTestutil "cryptoutil/internal/shared/testutil"
)

// TestIM_SubcommandHelpFlags tests help flag for all subcommands in table-driven format.
func TestIM_SubcommandHelpFlags(t *testing.T) {
	tests := []struct {
		subcommand string
		helpTexts  []string
	}{
		{
			subcommand: "client",
			helpTexts:  []string{"cipher im client", "Run client operations"},
		},
		{
			subcommand: "init",
			helpTexts:  []string{"cipher im init", "Initialize database schema"},
		},
		{
			subcommand: "health",
			helpTexts:  []string{"cipher im health", "Check service health"},
		},
		{
			subcommand: "livez",
			helpTexts:  []string{"cipher im livez", "Check service liveness"},
		},
		{
			subcommand: "readyz",
			helpTexts:  []string{"cipher im readyz", "Check service readiness"},
		},
		{
			subcommand: "shutdown",
			helpTexts:  []string{"cipher im shutdown", "Trigger graceful shutdown"},
		},
	}

	for _, tt := range tests {
		// Capture range variable.
		t.Run(tt.subcommand, func(t *testing.T) {
			t.Parallel()

			// Test --help flag.

			var stdout, stderr bytes.Buffer

			exitCode := Im([]string{tt.subcommand, "--help"}, nil, &stdout, &stderr)
			require.Equal(t, 0, exitCode, "%s --help should succeed", tt.subcommand)

			for _, expected := range tt.helpTexts {
				require.Contains(t, stdout.String()+stderr.String(), expected, "Help output should contain: %s", expected)
			}

			// Test -h flag.
			stdout.Reset()
			stderr.Reset()
			exitCode = Im([]string{tt.subcommand, "-h"}, nil, &stdout, &stderr)
			require.Equal(t, 0, exitCode, "%s -h should succeed", tt.subcommand)

			for _, expected := range tt.helpTexts {
				require.Contains(t, stdout.String()+stderr.String(), expected, "Help output (-h) should contain: %s", expected)
			}

			// Test with help as positional argument.
			stdout.Reset()
			stderr.Reset()
			exitCode = Im([]string{tt.subcommand, "help"}, nil, &stdout, &stderr)
			require.Equal(t, 0, exitCode, "%s help should succeed", tt.subcommand)

			for _, expected := range tt.helpTexts {
				require.Contains(t, stdout.String()+stderr.String(), expected, "Help output (positional) should contain: %s", expected)
			}
		})
	}
}

// TestPrintIMVersion tests the version command.
func TestPrintIMVersion(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Im([]string{"version"}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode, "version command should succeed")

	output := stdout.String() + stderr.String()

	require.Contains(t, output, "cipher-im service", "Version output should contain cipher-im service")
}

// TestIM_SubcommandNotImplemented tests subcommands that are not yet implemented.
func TestIM_SubcommandNotImplemented(t *testing.T) {
	tests := []struct {
		subcommand      string
		expectedMessage string
	}{
		{
			subcommand:      "client",
			expectedMessage: "Client subcommand not yet implemented",
		},
		{
			subcommand:      "init",
			expectedMessage: "Init subcommand not yet implemented",
		},
	}

	for _, tt := range tests {
		// Capture range variable.
		t.Run(tt.subcommand, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := Im([]string{tt.subcommand}, nil, &stdout, &stderr)
			require.Equal(t, 1, exitCode, "%s subcommand should fail (not implemented)", tt.subcommand)

			output := stdout.String() + stderr.String()
			require.Contains(t, output, tt.expectedMessage)
		})
	}
}

// TestIM_SubcommandLiveServer tests health check subcommands with shared test server.
func TestIM_SubcommandLiveServer(t *testing.T) {
	tests := []struct {
		subcommand       string
		url              string
		expectedExitCode int
		expectedOutputs  []string
		customCheck      func(t *testing.T, output string) // For special cases like readyz
	}{
		{
			subcommand:       "health",
			url:              publicBaseURL + "/service/api/v1",
			expectedExitCode: 0,
			expectedOutputs:  []string{"Service is healthy", "HTTP 200"},
		},
		{
			subcommand:       "livez",
			url:              adminBaseURL,
			expectedExitCode: 0,
			expectedOutputs:  []string{"Service is alive", "HTTP 200"},
		},
		{
			subcommand: "readyz",
			url:        adminBaseURL,
			customCheck: func(t *testing.T, output string) {
				// Readyz may return 0 (ready) or 1 (not ready) depending on service state
				// Check that we got a valid response (either ready or not ready)
				validResponse := strings.Contains(output, "Service is ready") || strings.Contains(output, "Service is not ready")
				require.True(t, validResponse, "Output should indicate readiness status")
			},
		},
	}

	for _, tt := range tests {
		// Capture range variable.
		t.Run(tt.subcommand, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := Im([]string{tt.subcommand, "--url", tt.url}, nil, &stdout, &stderr)

			if tt.customCheck != nil {
				// For readyz: check exit code is 0 or 1, and custom output check
				require.Contains(t, []int{0, 1}, exitCode, "%s should return 0 or 1", tt.subcommand)

				output := stdout.String() + stderr.String()
				tt.customCheck(t, output)
			} else {
				// For health and livez: exact exit code and output checks
				require.Equal(t, tt.expectedExitCode, exitCode, "%s should succeed", tt.subcommand)

				output := stdout.String() + stderr.String()
				for _, expected := range tt.expectedOutputs {
					require.Contains(t, output, expected, "Output should contain: %s", expected)
				}
			}
		})
	}
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

			Im([]string{tt.subcommand, "--url", tt.url}, nil, &stdout, &stderr)
			output := stdout.String() + stderr.String()
			// Check that output contains at least one of the expected strings.
			require.True(t, cryptoutilSharedTestutil.ContainsAny(output, tt.contains),
				"output should contain one of %v: %s", tt.contains, output)
		})
	}
}

// TestIM_SubcommandResponseBodies tests response body handling for all subcommands.
func TestIM_SubcommandResponseBodies(t *testing.T) {
	tests := []struct {
		name         string
		subcommand   string
		url          string
		expectExit   int
		expectOutput []string
	}{
		// Health subcommand tests.
		{
			name:         "health_success_no_body",
			subcommand:   "health",
			url:          testMockServerOK.URL + cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + "/health",
			expectExit:   0,
			expectOutput: []string{"Service is healthy"},
		},
		{
			name:         "health_success_with_body",
			subcommand:   "health",
			url:          testMockServerCustom.URL + cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + "/health",
			expectExit:   0,
			expectOutput: []string{"Service is healthy", "200", "All systems operational"},
		},
		{
			name:         "health_unhealthy_with_body",
			subcommand:   "health",
			url:          testMockServerError.URL + cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + "/health",
			expectExit:   1,
			expectOutput: []string{"Service is unhealthy", "503", "Service Unavailable"},
		},
		// Livez subcommand tests.
		{
			name:         "livez_alive_no_body",
			subcommand:   "livez",
			url:          testMockServerOK.URL + cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminLivezRequestPath,
			expectExit:   0,
			expectOutput: []string{"Service is alive"},
		},
		{
			name:         "livez_alive_with_body",
			subcommand:   "livez",
			url:          testMockServerCustom.URL + cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminLivezRequestPath,
			expectExit:   0,
			expectOutput: []string{"Service is alive", "200", "Process is alive and running"},
		},
		{
			name:         "livez_not_alive_with_body",
			subcommand:   "livez",
			url:          testMockServerError.URL + cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminLivezRequestPath,
			expectExit:   1,
			expectOutput: []string{"Service is not alive", "503", "Service Unavailable"},
		},
		// Shutdown subcommand tests.
		{
			name:         "shutdown_success_no_body",
			subcommand:   "shutdown",
			url:          testMockServerOK.URL + cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminShutdownRequestPath,
			expectExit:   0,
			expectOutput: []string{"Shutdown initiated"},
		},
		{
			name:         "shutdown_failed_no_body",
			subcommand:   "shutdown",
			url:          testMockServerError.URL + cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminShutdownRequestPath,
			expectExit:   1,
			expectOutput: []string{"Shutdown request failed", "503"},
		},
	}

	for _, tt := range tests {
		// Capture range variable.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := Im([]string{tt.subcommand, "--url", tt.url}, nil, &stdout, &stderr)
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
		url        string
	}{
		{
			name:       "health_with_health_suffix",
			subcommand: "health",
			url:        testMockServerCustom.URL + cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + "/health",
		},
		{
			name:       "livez_with_livez_suffix",
			subcommand: "livez",
			url:        testMockServerCustom.URL + cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminLivezRequestPath,
		},
		{
			name:       "readyz_with_readyz_suffix",
			subcommand: "readyz",
			url:        testMockServerCustom.URL + cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath,
		},
		{
			name:       "shutdown_with_shutdown_suffix",
			subcommand: "shutdown",
			url:        testMockServerCustom.URL + cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminShutdownRequestPath,
		},
	}

	for _, tt := range tests {
		// Capture range variable.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := Im([]string{tt.subcommand, "--url", tt.url}, nil, &stdout, &stderr)
			require.Equal(t, 0, exitCode, "%s should succeed with explicit suffix", tt.subcommand)

			output := stdout.String() + stderr.String()
			require.NotContains(t, output, "failed")
		})
	}
}

// TestIM_URLEdgeCases tests various URL edge cases using table-driven tests.
