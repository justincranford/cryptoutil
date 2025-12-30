// Copyright (c) 2025 Justin Cranford

package im

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require")

// TestIM_HealthSubcommand_UnhealthyWithBody tests health check unhealthy with error body.
func TestIM_HealthSubcommand_UnhealthyWithBody(t *testing.T) {

	// Create test server returning 503 with error body.
	errorMessage := "Database connection timeout"

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprint(w, errorMessage)
	}))
	defer server.Close()

	output := captureOutput(t, func() {
		exitCode := IM([]string{"health", "--url", server.URL + "/health"})
		require.Equal(t, 1, exitCode, "Health should fail with 503")
	})
	require.Contains(t, output, "Service is unhealthy")
	require.Contains(t, output, "503")
	require.Contains(t, output, errorMessage, "Should include error message from response body")
}

// TestIM_LivezSubcommand_NotAliveWithBody tests livez not alive with error body.
func TestIM_LivezSubcommand_NotAliveWithBody(t *testing.T) {

	// Create test server returning 503 with error body.
	errorMessage := "Service initialization failed"

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprint(w, errorMessage)
	}))
	defer server.Close()

	output := captureOutput(t, func() {
		exitCode := IM([]string{"livez", "--url", server.URL + "/livez"})
		require.Equal(t, 1, exitCode, "Livez should fail with 503")
	})
	require.Contains(t, output, "Service is not alive")
	require.Contains(t, output, "503")
	require.Contains(t, output, errorMessage, "Should include error message from response body")
}

// TestIM_HealthSubcommand_SuccessWithBody tests health check success with response body.
func TestIM_HealthSubcommand_SuccessWithBody(t *testing.T) {

	// Create test server returning 200 with body.
	responseBody := "All systems operational"

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, responseBody)
	}))
	defer server.Close()

	output := captureOutput(t, func() {
		exitCode := IM([]string{"health", "--url", server.URL + "/health"})
		require.Equal(t, 0, exitCode, "Health should succeed with 200")
	})

	require.Contains(t, output, "Service is healthy")
	require.Contains(t, output, "200")
	require.Contains(t, output, responseBody, "Should include response body in output")
}

// TestIM_LivezSubcommand_AliveWithBody tests livez alive with response body.
func TestIM_LivezSubcommand_AliveWithBody(t *testing.T) {

	// Create test server returning 200 with body.
	responseBody := "Process is alive and running"

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, responseBody)
	}))
	defer server.Close()

	output := captureOutput(t, func() {
		exitCode := IM([]string{"livez", "--url", server.URL + "/livez"})
		require.Equal(t, 0, exitCode, "Livez should succeed with 200")
	})

	require.Contains(t, output, "Service is alive")
	require.Contains(t, output, "200")
	require.Contains(t, output, responseBody, "Should include response body in output")
}
