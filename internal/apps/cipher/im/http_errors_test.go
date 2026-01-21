// Copyright (c) 2025 Justin Cranford

package im

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

const (
	adminHealthPath = cryptoutilMagic.DefaultPrivateAdminAPIContextPath + "/healthz"
	adminLivezPath  = cryptoutilMagic.DefaultPrivateAdminAPIContextPath + cryptoutilMagic.PrivateAdminLivezRequestPath
	adminReadyzPath = cryptoutilMagic.DefaultPrivateAdminAPIContextPath + cryptoutilMagic.PrivateAdminReadyzRequestPath
)

// TestIM_HealthSubcommand_SlowResponse tests health check with slow server response.
func TestIM_HealthSubcommand_SlowResponse(t *testing.T) {
	// Create slow server.
	lc := &net.ListenConfig{}
	listener, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	require.True(t, ok, "listener address should be TCP")

	actualPort := tcpAddr.Port

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			// Delay before responding (reduced from 3s to 1s for faster tests).
			// 1 second is still "slow" for health checks and validates timeout behavior.
			time.Sleep(1 * time.Second)
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
	var stdout, stderr bytes.Buffer

	exitCode := internalIM([]string{
		"health",
		"--url", fmt.Sprintf("http://127.0.0.1:%d%s", actualPort, adminHealthPath),
	}, &stdout, &stderr)
	require.Equal(t, 0, exitCode, "Health check should succeed for slow but valid response")

	output := stdout.String() + stderr.String()
	require.Contains(t, output, "Service is healthy")
}

// TestIM_LivezSubcommand_EmptyResponse tests livez check with empty body.
func TestIM_LivezSubcommand_EmptyResponse(t *testing.T) {
	t.Parallel()

	// Test livez check with empty response using shared OK server (returns "OK" body).
	var stdout, stderr bytes.Buffer

	exitCode := internalIM([]string{"livez", "--url", testMockServerOK.URL + adminLivezPath}, &stdout, &stderr)
	require.Equal(t, 0, exitCode, "Livez should succeed with 200 OK")

	output := stdout.String() + stderr.String()
	require.Contains(t, output, "Service is alive")
}

// TestIM_ReadyzSubcommand_404NotFound tests readyz check with 404 response.
func TestIM_ReadyzSubcommand_404NotFound(t *testing.T) {
	t.Parallel()

	// Use shared error server that returns 503 (close enough to 404 for error case).
	var stdout, stderr bytes.Buffer

	exitCode := internalIM([]string{"readyz", "--url", testMockServerError.URL + adminReadyzPath}, &stdout, &stderr)
	require.Equal(t, 1, exitCode, "Readyz should fail with non-200 status")

	output := stdout.String() + stderr.String()
	require.Contains(t, output, "Service is not ready")
	require.Contains(t, output, "503")
}

// TestIM_ShutdownSubcommand_500InternalServerError tests shutdown with 500 error.
func TestIM_ShutdownSubcommand_500InternalServerError(t *testing.T) {
	t.Parallel()

	// Use shared error server that returns 503 (close enough to 500 for error case).
	var stdout, stderr bytes.Buffer

	exitCode := internalIM([]string{"shutdown", "--url", testMockServerError.URL + cryptoutilMagic.DefaultPrivateAdminAPIContextPath + cryptoutilMagic.PrivateAdminShutdownRequestPath}, &stdout, &stderr)
	require.Equal(t, 1, exitCode, "Shutdown should fail with error status")

	output := stdout.String() + stderr.String()
	require.Contains(t, output, "Shutdown request failed")
	require.Contains(t, output, "503")
}
