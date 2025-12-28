// Copyright (c) 2025 Justin Cranford

package learn_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	cryptoutilLearnCmd "cryptoutil/internal/cmd/learn"

	"github.com/stretchr/testify/require"
)

// TestIM_HealthSubcommand_UnhealthyWithBody tests health check unhealthy with error body.
func TestIM_HealthSubcommand_UnhealthyWithBody(t *testing.T) {
	t.Parallel()

	// Create test server returning 503 with error body.
	errorMessage := "Database connection timeout"

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprint(w, errorMessage)
	}))
	defer server.Close()

	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{"health", "--url", server.URL + "/health"})
		require.Equal(t, 1, exitCode, "Health should fail with 503")
	})

	require.Empty(t, stdout)
	require.Contains(t, stderr, "❌ Service is unhealthy")
	require.Contains(t, stderr, "503")
	require.Contains(t, stderr, errorMessage, "Should include error message from response body")
}

// TestIM_LivezSubcommand_NotAliveWithBody tests livez not alive with error body.
func TestIM_LivezSubcommand_NotAliveWithBody(t *testing.T) {
	t.Parallel()

	// Create test server returning 503 with error body.
	errorMessage := "Service initialization failed"

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprint(w, errorMessage)
	}))
	defer server.Close()

	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{"livez", "--url", server.URL + "/livez"})
		require.Equal(t, 1, exitCode, "Livez should fail with 503")
	})

	require.Empty(t, stdout)
	require.Contains(t, stderr, "❌ Service is not alive")
	require.Contains(t, stderr, "503")
	require.Contains(t, stderr, errorMessage, "Should include error message from response body")
}

// TestIM_HealthSubcommand_SuccessWithBody tests health check success with response body.
func TestIM_HealthSubcommand_SuccessWithBody(t *testing.T) {
	t.Parallel()

	// Create test server returning 200 with body.
	responseBody := "All systems operational"

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, responseBody)
	}))
	defer server.Close()

	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{"health", "--url", server.URL + "/health"})
		require.Equal(t, 0, exitCode, "Health should succeed with 200")
	})

	require.Contains(t, stdout, "✅ Service is healthy")
	require.Contains(t, stdout, "200")
	require.Contains(t, stdout, responseBody, "Should include response body in stdout")
	require.Empty(t, stderr)
}

// TestIM_LivezSubcommand_AliveWithBody tests livez alive with response body.
func TestIM_LivezSubcommand_AliveWithBody(t *testing.T) {
	t.Parallel()

	// Create test server returning 200 with body.
	responseBody := "Process is alive and running"

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, responseBody)
	}))
	defer server.Close()

	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{"livez", "--url", server.URL + "/livez"})
		require.Equal(t, 0, exitCode, "Livez should succeed with 200")
	})

	require.Contains(t, stdout, "✅ Service is alive")
	require.Contains(t, stdout, "200")
	require.Contains(t, stdout, responseBody, "Should include response body in stdout")
	require.Empty(t, stderr)
}
