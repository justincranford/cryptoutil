// Copyright (c) 2025 Justin Cranford

package learn_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cryptoutilLearnCmd "cryptoutil/internal/cmd/learn"

	"github.com/stretchr/testify/require"
)

// TestIM_HealthSubcommand_MultipleURLFlags tests health check with multiple --url flags (first wins).
func TestIM_HealthSubcommand_MultipleURLFlags(t *testing.T) {
	t.Parallel()

	// Create test server.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "Healthy")
	}))
	defer server.Close()

	// Pass multiple --url flags (first one should win, second ignored).
	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{
			"health",
			"--url", server.URL + "/health",
			"--url", "https://invalid-second-url:9999",
		})
		require.Equal(t, 0, exitCode, "Should use first --url flag")
	})

	require.Contains(t, stdout, "✅ Service is healthy")
	require.Empty(t, stderr)
}

// TestIM_LivezSubcommand_URLFlagWithoutValue tests livez with --url flag but missing value.
func TestIM_LivezSubcommand_URLFlagWithoutValue(t *testing.T) {
	t.Parallel()

	// Pass --url flag without value (should use default URL).
	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{"livez", "--url"})
		require.Equal(t, 1, exitCode, "Should fail with connection error to default")
	})

	require.Empty(t, stdout)
	require.Contains(t, stderr, "❌ Liveness check failed")
	require.True(t,
		containsAny(stderr, []string{
			"connection refused",
			"actively refused",
			"dial tcp",
		}),
		"Should contain connection error: %s", stderr)
}

// TestIM_ReadyzSubcommand_ExtraArgumentsIgnored tests readyz with extra arguments.
func TestIM_ReadyzSubcommand_ExtraArgumentsIgnored(t *testing.T) {
	t.Parallel()

	// Create test server.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "Ready")
	}))
	defer server.Close()

	// Pass extra arguments after --url (should be ignored).
	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{
			"readyz",
			"--url", server.URL + "/admin/v1/readyz",
			"extra", "ignored", "args",
		})
		require.Equal(t, 0, exitCode, "Extra args should be ignored")
	})

	require.Contains(t, stdout, "✅ Service is ready")
	require.Empty(t, stderr)
}

// TestIM_ShutdownSubcommand_URLWithoutQueryParameters tests shutdown URL handling.
func TestIM_ShutdownSubcommand_URLWithoutQueryParameters(t *testing.T) {
	t.Parallel()

	// Create test server.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, "Shutting down")
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{
			"shutdown",
			"--url", server.URL + "/admin/v1/shutdown",
		})
		require.Equal(t, 0, exitCode, "Shutdown should succeed")
	})

	require.Contains(t, stdout, "✅ Shutdown initiated")
	require.Empty(t, stderr)
}

// TestIM_HealthSubcommand_URLWithFragment tests health check with URL fragment (fragment should be ignored by HTTP).
func TestIM_HealthSubcommand_URLWithFragment(t *testing.T) {
	t.Parallel()

	// Create test server.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// HTTP strips fragment before sending request.
		require.Equal(t, "/health", r.URL.Path, "URL path should be /health")
		require.Empty(t, r.URL.Fragment, "Fragment should be stripped")

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "Healthy")
	}))
	defer server.Close()

	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{
			"health",
			"--url", server.URL + "/health#section",
		})
		require.Equal(t, 0, exitCode, "Health check with fragment should succeed")
	})

	require.Contains(t, stdout, "✅ Service is healthy")
	require.Empty(t, stderr)
}

// TestIM_LivezSubcommand_URLWithUserInfo tests livez with URL containing user info (basic auth style).
func TestIM_LivezSubcommand_URLWithUserInfo(t *testing.T) {
	t.Parallel()

	// Create test server that expects basic auth in URL.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for Authorization header (httptest.NewTLSServer strips user info).
		// User info in URL doesn't auto-set Authorization header in Go http.Client.
		// This tests URL parsing handles user:pass@ format without error.
		_ = r.Header.Get("Authorization") // Explicitly ignore auth header.

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "Alive")
	}))
	defer server.Close()

	// Extract host from server URL and add user info.
	urlParts := strings.Split(server.URL, "//")
	urlWithUserInfo := urlParts[0] + "//user:pass@" + urlParts[1] + "/livez"

	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{
			"livez",
			"--url", urlWithUserInfo,
		})
		require.Equal(t, 0, exitCode, "Livez with user info in URL should succeed")
	})

	require.Contains(t, stdout, "✅ Service is alive")
	require.Empty(t, stderr)
}

// TestIM_ReadyzSubcommand_CaseInsensitiveHTTPStatus tests readyz response with different status code messages.
func TestIM_ReadyzSubcommand_CaseInsensitiveHTTPStatus(t *testing.T) {
	t.Parallel()

	// Create test server returning non-standard status code.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot) // 418 I'm a teapot.
		_, _ = fmt.Fprint(w, "Not a coffee machine")
	}))
	defer server.Close()

	stdout, stderr := captureOutput(t, func() {
		exitCode := cryptoutilLearnCmd.IM([]string{
			"readyz",
			"--url", server.URL + "/admin/v1/readyz",
		})
		require.Equal(t, 1, exitCode, "Non-200 status should fail")
	})

	require.Empty(t, stdout)
	require.Contains(t, stderr, "❌ Service is not ready")
	require.Contains(t, stderr, "418")
}
