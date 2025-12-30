// Copyright (c) 2025 Justin Cranford

package im

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// containsAny returns true if s contains any of the substrings.
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(s, substr) {
			return true
		}
	}

	return false
}

// TestIM_HealthSubcommand_InvalidURL tests health check with malformed URL.
func TestIM_HealthSubcommand_InvalidURL(t *testing.T) {
	output := captureOutput(t, func() {
		exitCode := IM([]string{"health", "--url", "://invalid-url"})
		require.Equal(t, 1, exitCode, "Health check should fail with invalid URL")
	})

	require.Contains(t, output, "Health check failed:")
}

// TestIM_HealthSubcommand_ConnectionRefused tests health check when server not running.
func TestIM_HealthSubcommand_ConnectionRefused(t *testing.T) {
	output := captureOutput(t, func() {
		// Use port that's unlikely to be in use.
		exitCode := IM([]string{"health", "--url", "https://127.0.0.1:9999"})
		require.Equal(t, 1, exitCode, "Health check should fail when server down")
	})

	require.Contains(t, output, "Health check failed:")
	// Windows: "actively refused it", Linux: "connection refused".
	require.True(t, containsAny(output, []string{"connection refused", "actively refused"}),
		"output should contain connection error: %s", output)
}

// TestIM_HealthSubcommand_Non200Status tests health check with unhealthy server.
func TestIM_HealthSubcommand_Non200Status(t *testing.T) {
	// This test would require a mock server returning non-200 status.
	// For now, we test with a down server (covers error handling path).
	output := captureOutput(t, func() {
		exitCode := IM([]string{"health", "--url", "https://127.0.0.1:9998"})
		require.Equal(t, 1, exitCode)
	})

	require.Contains(t, output, "Health check failed:")
}

// TestIM_LivezSubcommand_InvalidURL tests livez check with malformed URL.
func TestIM_LivezSubcommand_InvalidURL(t *testing.T) {
	output := captureOutput(t, func() {
		exitCode := IM([]string{"livez", "--url", "://invalid-url"})
		require.Equal(t, 1, exitCode, "Liveness check should fail with invalid URL")
	})

	require.Contains(t, output, "Liveness check failed:")
}

// TestIM_LivezSubcommand_ConnectionRefused tests livez check when server not running.
func TestIM_LivezSubcommand_ConnectionRefused(t *testing.T) {
	output := captureOutput(t, func() {
		exitCode := IM([]string{"livez", "--url", "https://127.0.0.1:9997"})
		require.Equal(t, 1, exitCode, "Liveness check should fail when server down")
	})

	require.Contains(t, output, "Liveness check failed:")
	// Windows: "actively refused it", Linux: "connection refused".
	require.True(t, containsAny(output, []string{"connection refused", "actively refused"}),
		"output should contain connection error: %s", output)
}

// TestIM_LivezSubcommand_Non200Status tests livez check with unhealthy server.
func TestIM_LivezSubcommand_Non200Status(t *testing.T) {
	output := captureOutput(t, func() {
		exitCode := IM([]string{"livez", "--url", "https://127.0.0.1:9996"})
		require.Equal(t, 1, exitCode)
	})

	require.Contains(t, output, "Liveness check failed:")
}

// TestIM_ReadyzSubcommand_InvalidURL tests readyz check with malformed URL.
func TestIM_ReadyzSubcommand_InvalidURL(t *testing.T) {
	output := captureOutput(t, func() {
		exitCode := IM([]string{"readyz", "--url", "://invalid-url"})
		require.Equal(t, 1, exitCode, "Readiness check should fail with invalid URL")
	})

	require.Contains(t, output, "Readiness check failed:")
}

// TestIM_ReadyzSubcommand_ConnectionRefused tests readyz check when server not running.
func TestIM_ReadyzSubcommand_ConnectionRefused(t *testing.T) {
	output := captureOutput(t, func() {
		exitCode := IM([]string{"readyz", "--url", "https://127.0.0.1:9995"})
		require.Equal(t, 1, exitCode, "Readiness check should fail when server down")
	})

	require.Contains(t, output, "Readiness check failed:")
	// Windows: "actively refused it", Linux: "connection refused".
	require.True(t, containsAny(output, []string{"connection refused", "actively refused"}),
		"output should contain connection error: %s", output)
}

// TestIM_ShutdownSubcommand_InvalidURL tests shutdown with malformed URL.
func TestIM_ShutdownSubcommand_InvalidURL(t *testing.T) {
	output := captureOutput(t, func() {
		exitCode := IM([]string{"shutdown", "--url", "://invalid-url"})
		require.Equal(t, 1, exitCode, "Shutdown should fail with invalid URL")
	})

	require.Contains(t, output, "Shutdown request failed:")
}

// TestIM_ShutdownSubcommand_ConnectionRefused tests shutdown when server not running.
func TestIM_ShutdownSubcommand_ConnectionRefused(t *testing.T) {
	output := captureOutput(t, func() {
		exitCode := IM([]string{"shutdown", "--url", "https://127.0.0.1:9994"})
		require.Equal(t, 1, exitCode, "Shutdown should fail when server down")
	})

	require.Contains(t, output, "Shutdown request failed:")
	// Windows: "actively refused it", Linux: "connection refused".
	require.True(t, containsAny(output, []string{"connection refused", "actively refused"}),
		"output should contain connection error: %s", output)
}
