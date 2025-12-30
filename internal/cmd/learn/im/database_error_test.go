// Copyright (c) 2025 Justin Cranford

package im

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require")

const (
	adminHealthPath = "/admin/v1/healthz"
	adminLivezPath  = "/admin/v1/livez"
	adminReadyzPath = "/admin/v1/readyz"
)

// TestIM_HealthSubcommand_SlowResponse tests health check with slow server response.
func TestIM_HealthSubcommand_SlowResponse(t *testing.T) {
	t.Parallel()

	// Create slow server.
	lc := &net.ListenConfig{}
	listener, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	require.True(t, ok, "listener address should be TCP")

	actualPort := tcpAddr.Port

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Delay before responding.
			time.Sleep(3 * time.Second)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		}),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		_ = server.Serve(listener)
	}()

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = server.Shutdown(ctx)
	}()

	// Wait for server to start.
	time.Sleep(100 * time.Millisecond)

	// Test health check completes despite slow response.
	output := captureOutput(t, func() {
		exitCode := IM([]string{
			"health",
			"--url", fmt.Sprintf("http://127.0.0.1:%d%s", actualPort, adminHealthPath),
		})
		require.Equal(t, 0, exitCode, "Health check should succeed for slow but valid response")
	})

	require.Contains(t, output, "✅ Service is healthy")
}

// TestIM_LivezSubcommand_EmptyResponse tests livez check with empty body.
func TestIM_LivezSubcommand_EmptyResponse(t *testing.T) {
	t.Parallel()

	// Create server that returns empty body.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == adminLivezPath {
			w.WriteHeader(http.StatusOK)
			// Empty body - no write.
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Test livez check with empty response.
	output := captureOutput(t, func() {
		exitCode := IM([]string{"livez", "--url", server.URL + adminLivezPath})
		require.Equal(t, 0, exitCode, "Livez should succeed with 200 OK even if body empty")
	})

	require.Contains(t, output, "✅ Service is alive")
}

// TestIM_ReadyzSubcommand_404NotFound tests readyz check with 404 response.
func TestIM_ReadyzSubcommand_404NotFound(t *testing.T) {
	t.Parallel()

	// Create server that returns 404 for readyz endpoint.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	// Test readyz check with 404 response.
	output := captureOutput(t, func() {
		exitCode := IM([]string{"readyz", "--url", server.URL + adminReadyzPath})
		require.Equal(t, 1, exitCode, "Readyz should fail with non-200 status")
	})
	require.Contains(t, output, "❌ Service is not ready")
	require.Contains(t, output, "404")
}

// TestIM_ShutdownSubcommand_500InternalServerError tests shutdown with 500 error.
func TestIM_ShutdownSubcommand_500InternalServerError(t *testing.T) {
	t.Parallel()

	// Create server that returns 500 for shutdown endpoint.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	// Test shutdown with 500 response.
	output := captureOutput(t, func() {
		exitCode := IM([]string{"shutdown", "--url", server.URL + "/admin/v1/shutdown"})
		require.Equal(t, 1, exitCode, "Shutdown should fail with 500 status")
	})
	require.Contains(t, output, "❌ Shutdown request failed")
	require.Contains(t, output, "500")
}
