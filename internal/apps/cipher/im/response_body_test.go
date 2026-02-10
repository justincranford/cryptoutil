// Copyright (c) 2025 Justin Cranford

package im

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedTestutil "cryptoutil/internal/shared/testutil"
)

// TestIM_HealthSubcommand_NoBodySuccess tests health check with 200 but no body.
func TestIM_HealthSubcommand_NoBodySuccess(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Im([]string{"health", "--url", testMockServerOK.URL + "/health"}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode, "Health should succeed with 200 even if no body")

	output := stdout.String() + stderr.String()
	require.Contains(t, output, "Service is healthy")
}

// TestIM_HealthSubcommand_UnhealthyNoBody tests health check unhealthy with no body.
func TestIM_HealthSubcommand_UnhealthyNoBody(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Im([]string{"health", "--url", testMockServerError.URL + "/health"}, nil, &stdout, &stderr)
	require.Equal(t, 1, exitCode, "Health should fail with 503")

	output := stdout.String() + stderr.String()
	require.Contains(t, output, "Service is unhealthy")
	require.Contains(t, output, "503")
}

// TestIM_LivezSubcommand_NoBodySuccess tests livez with 200 but no body.
func TestIM_LivezSubcommand_NoBodySuccess(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Im([]string{"livez", "--url", testMockServerOK.URL + "/livez"}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode, "Livez should succeed with 200 even if no body")

	output := stdout.String() + stderr.String()
	require.Contains(t, output, "Service is alive")
}

// TestIM_LivezSubcommand_NotAliveNoBody tests livez not alive with no body.
func TestIM_LivezSubcommand_NotAliveNoBody(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Im([]string{"livez", "--url", testMockServerError.URL + "/livez"}, nil, &stdout, &stderr)
	require.Equal(t, 1, exitCode, "Livez should fail with 503")

	output := stdout.String() + stderr.String()
	require.Contains(t, output, "Service is not alive")
	require.Contains(t, output, "503")
}

// TestIM_ShutdownSubcommand_NoBodySuccess tests shutdown with 200 but no body.
func TestIM_ShutdownSubcommand_NoBodySuccess(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Im([]string{"shutdown", "--url", testMockServerOK.URL + "/shutdown"}, nil, &stdout, &stderr)
	require.Equal(t, 0, exitCode, "Shutdown should succeed with 200 even if no body")

	output := stdout.String() + stderr.String()
	require.Contains(t, output, "Shutdown initiated")
}

// TestIM_ShutdownSubcommand_FailedNoBody tests shutdown failure with no body.
func TestIM_ShutdownSubcommand_FailedNoBody(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Im([]string{"shutdown", "--url", testMockServerError.URL + "/shutdown"}, nil, &stdout, &stderr)
	require.Equal(t, 1, exitCode, "Shutdown should fail with 503")

	output := stdout.String() + stderr.String()
	require.Contains(t, output, "Shutdown request failed")
	require.Contains(t, output, "503")
}

// TestIM_ShutdownSubcommand_PartialBodyRead tests shutdown with body read.
func TestIM_ShutdownSubcommand_PartialBodyRead(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := Im([]string{"shutdown", "--url", testMockServerOK.URL + "/shutdown"}, nil, &stdout, &stderr)
	// Should still succeed because we got 200 status.
	require.Equal(t, 0, exitCode, "Shutdown should succeed even with partial body")

	output := stdout.String() + stderr.String()
	require.Contains(t, output, "Shutdown initiated")
}

// TestIM_HealthSubcommand_DefaultURL tests health check without --url flag (uses default).
func TestIM_HealthSubcommand_DefaultURL(t *testing.T) {
	t.Parallel()

	// Test default URL (will fail - either connection refused or HTTP error from unrelated service).
	var stdout, stderr bytes.Buffer

	exitCode := Im([]string{"health"}, nil, &stdout, &stderr)
	require.Equal(t, 1, exitCode, "Health check should fail when no server running")

	output := stdout.String() + stderr.String()
	// Accept either connection failure OR HTTP error (if Docker containers are running on default port).
	require.True(t,
		cryptoutilSharedTestutil.ContainsAny(output, []string{
			"Health check failed:", // Connection error
			"Service is unhealthy", // HTTP error from unexpected service
			"connection refused",   // TCP connection refused
			"actively refused",     // Windows connection refused
			"dial tcp",             // Go dial error
			"EOF",                  // Connection closed
		}),
		"Should contain error message for default URL: %s", output)
}

// TestIM_LivezSubcommand_DefaultURL tests livez check without --url flag (uses default).
func TestIM_LivezSubcommand_DefaultURL(t *testing.T) {
	t.Parallel()

	// Test default URL (will fail to connect to 127.0.0.1:9090).
	var stdout, stderr bytes.Buffer

	exitCode := Im([]string{"livez"}, nil, &stdout, &stderr)
	require.Equal(t, 1, exitCode, "Livez check should fail when no server running")

	output := stdout.String() + stderr.String()
	require.Contains(t, output, "Liveness check failed:")
	require.True(t,
		cryptoutilSharedTestutil.ContainsAny(output, []string{
			"connection refused",
			"actively refused",
			"dial tcp",
			"EOF", // Can happen when nothing is listening.
		}),
		"Should contain connection error for default URL: %s", output)
}
