// Copyright (c) 2025 Justin Cranford

package learn_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	cryptoutilLearnCmd "cryptoutil/internal/cmd/learn"

	"github.com/stretchr/testify/require"
)

// TestIM_HealthSubcommand_URLWithHealthSuffix tests health check preserves /health suffix.
func TestIM_HealthSubcommand_URLWithHealthSuffix(t *testing.T) {
	t.Parallel()

	// Create test server.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/admin/v1/health" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Healthy"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Test health check with URL that already has /health suffix.
	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{
			"health",
			"--url", server.URL + "/admin/v1/health",
		})
		require.Equal(t, 0, exitCode, "Health check should succeed with explicit /health suffix")
	})

	require.Contains(t, stdout, "✅ Service is healthy")
	require.Empty(t, stderr)
}

// TestIM_LivezSubcommand_URLWithLivezSuffix tests livez check preserves /livez suffix.
func TestIM_LivezSubcommand_URLWithLivezSuffix(t *testing.T) {
	t.Parallel()

	// Create test server.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/admin/v1/livez" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Alive"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Test livez check with URL that already has /livez suffix.
	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{
			"livez",
			"--url", server.URL + "/admin/v1/livez",
		})
		require.Equal(t, 0, exitCode, "Livez check should succeed with explicit /livez suffix")
	})

	require.Contains(t, stdout, "✅ Service is alive")
	require.Empty(t, stderr)
}

// TestIM_ReadyzSubcommand_URLWithReadyzSuffix tests readyz check preserves /readyz suffix.
func TestIM_ReadyzSubcommand_URLWithReadyzSuffix(t *testing.T) {
	t.Parallel()

	// Create test server.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/admin/v1/readyz" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Ready"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Test readyz check with URL that already has /readyz suffix.
	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{
			"readyz",
			"--url", server.URL + "/admin/v1/readyz",
		})
		require.Equal(t, 0, exitCode, "Readyz check should succeed with explicit /readyz suffix")
	})

	require.Contains(t, stdout, "✅ Service is ready")
	require.Empty(t, stderr)
}

// TestIM_ShutdownSubcommand_URLWithShutdownSuffix tests shutdown preserves /shutdown suffix.
func TestIM_ShutdownSubcommand_URLWithShutdownSuffix(t *testing.T) {
	t.Parallel()

	// Create test server.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/admin/v1/shutdown" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Shutting down"))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Test shutdown with URL that already has /shutdown suffix.
	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{
			"shutdown",
			"--url", server.URL + "/admin/v1/shutdown",
		})
		require.Equal(t, 0, exitCode, "Shutdown should succeed with explicit /shutdown suffix")
	})

	require.Contains(t, stdout, "✅ Shutdown initiated")
	require.Empty(t, stderr)
}
