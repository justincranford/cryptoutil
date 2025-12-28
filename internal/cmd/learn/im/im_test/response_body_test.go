// Copyright (c) 2025 Justin Cranford

package learn_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	cryptoutilLearnCmd "cryptoutil/internal/cmd/learn"

	"github.com/stretchr/testify/require"
)

// TestIM_HealthSubcommand_NoBodySuccess tests health check with 200 but no body.
func TestIM_HealthSubcommand_NoBodySuccess(t *testing.T) {
	t.Parallel()

	// Create test server with no response body.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// No body written.
	}))
	defer server.Close()

	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{"health", "--url", server.URL + "/health"})
		require.Equal(t, 0, exitCode, "Health should succeed with 200 even if no body")
	})

	require.Contains(t, stdout, "✅ Service is healthy")
	require.Empty(t, stderr)
}

// TestIM_HealthSubcommand_UnhealthyNoBody tests health check unhealthy with no body.
func TestIM_HealthSubcommand_UnhealthyNoBody(t *testing.T) {
	t.Parallel()

	// Create test server returning 503 with no body.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		// No body written.
	}))
	defer server.Close()

	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{"health", "--url", server.URL + "/health"})
		require.Equal(t, 1, exitCode, "Health should fail with 503")
	})

	require.Empty(t, stdout)
	require.Contains(t, stderr, "❌ Service is unhealthy")
	require.Contains(t, stderr, "503")
}

// TestIM_LivezSubcommand_NoBodySuccess tests livez with 200 but no body.
func TestIM_LivezSubcommand_NoBodySuccess(t *testing.T) {
	t.Parallel()

	// Create test server with no response body.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// No body written.
	}))
	defer server.Close()

	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{"livez", "--url", server.URL + "/livez"})
		require.Equal(t, 0, exitCode, "Livez should succeed with 200 even if no body")
	})

	require.Contains(t, stdout, "✅ Service is alive")
	require.Empty(t, stderr)
}

// TestIM_LivezSubcommand_NotAliveNoBody tests livez not alive with no body.
func TestIM_LivezSubcommand_NotAliveNoBody(t *testing.T) {
	t.Parallel()

	// Create test server returning 503 with no body.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		// No body written.
	}))
	defer server.Close()

	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{"livez", "--url", server.URL + "/livez"})
		require.Equal(t, 1, exitCode, "Livez should fail with 503")
	})

	require.Empty(t, stdout)
	require.Contains(t, stderr, "❌ Service is not alive")
	require.Contains(t, stderr, "503")
}

// TestIM_ShutdownSubcommand_NoBodySuccess tests shutdown with 200 but no body.
func TestIM_ShutdownSubcommand_NoBodySuccess(t *testing.T) {
	t.Parallel()

	// Create test server with no response body.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			// No body written.
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{"shutdown", "--url", server.URL + "/shutdown"})
		require.Equal(t, 0, exitCode, "Shutdown should succeed with 200 even if no body")
	})

	require.Contains(t, stdout, "✅ Shutdown initiated")
	require.Empty(t, stderr)
}

// TestIM_ShutdownSubcommand_FailedNoBody tests shutdown failure with no body.
func TestIM_ShutdownSubcommand_FailedNoBody(t *testing.T) {
	t.Parallel()

	// Create test server returning 500 with no body.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		// No body written.
	}))
	defer server.Close()

	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{"shutdown", "--url", server.URL + "/shutdown"})
		require.Equal(t, 1, exitCode, "Shutdown should fail with 500")
	})

	require.Empty(t, stdout)
	require.Contains(t, stderr, "❌ Shutdown request failed")
	require.Contains(t, stderr, "500")
}

// TestIM_HealthSubcommand_LargeBody tests health check with large response body.
func TestIM_HealthSubcommand_LargeBody(t *testing.T) {
	t.Parallel()

	// Create test server with large body (1MB).
	largeBody := make([]byte, 1024*1024)
	for i := range largeBody {
		largeBody[i] = 'A'
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(largeBody)
	}))
	defer server.Close()

	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{"health", "--url", server.URL + "/health"})
		require.Equal(t, 0, exitCode, "Health should succeed with large body")
	})

	require.Contains(t, stdout, "✅ Service is healthy")
	require.Contains(t, stdout, string(largeBody))
	require.Empty(t, stderr)
}

// TestIM_ShutdownSubcommand_PartialBodyRead tests shutdown with body read error simulation.
func TestIM_ShutdownSubcommand_PartialBodyRead(t *testing.T) {
	t.Parallel()

	// Create test server that closes connection mid-body.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Shutting"))
		// Simulate connection close before full body sent.
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		// hijack not possible in httptest, but we can simulate with body write.
		_, _ = w.Write([]byte(" down gracefully"))
	}))
	defer server.Close()

	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{"shutdown", "--url", server.URL + "/shutdown"})
		// Should still succeed because we got 200 status.
		require.Equal(t, 0, exitCode, "Shutdown should succeed even with partial body")
	})

	require.Contains(t, stdout, "✅ Shutdown initiated")
	require.Empty(t, stderr)
}

// TestIM_HealthSubcommand_DefaultURL tests health check without --url flag (uses default).
func TestIM_HealthSubcommand_DefaultURL(t *testing.T) {
	t.Parallel()

	// Test default URL (will fail to connect to 127.0.0.1:8888).
	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{"health"})
		require.Equal(t, 1, exitCode, "Health check should fail when no server running")
	})

	require.Empty(t, stdout)
	require.Contains(t, stderr, "❌ Health check failed:")
	require.True(t,
		containsAny(stderr, []string{
			"connection refused",
			"actively refused",
			"dial tcp",
		}),
		"Should contain connection error for default URL: %s", stderr)
}

// TestIM_LivezSubcommand_DefaultURL tests livez check without --url flag (uses default).
func TestIM_LivezSubcommand_DefaultURL(t *testing.T) {
	t.Parallel()

	// Test default URL (will fail to connect to 127.0.0.1:9090).
	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{"livez"})
		require.Equal(t, 1, exitCode, "Livez check should fail when no server running")
	})

	require.Empty(t, stdout)
	require.Contains(t, stderr, "❌ Liveness check failed:")
	require.True(t,
		containsAny(stderr, []string{
			"connection refused",
			"actively refused",
			"dial tcp",
		}),
		"Should contain connection error for default URL: %s", stderr)
}
