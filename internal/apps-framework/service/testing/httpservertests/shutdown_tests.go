// Copyright (c) 2025-2026 Justin Cranford.
// Package httpservertests provides reusable test cases for HTTP server implementations.
// Deprecated: migrate to package-specific tests using the newer shared helper packages instead.
// These generic tests verify shutdown behavior, double-shutdown handling, and health checks
// during shutdown scenarios. Product services should use these tests instead of duplicating
// shutdown test logic across multiple implementations.
package httpservertests

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Test-specific timing constants.
const (
	testPortReleaseWait = 500 * time.Millisecond
)

func waitForServerExit(t *testing.T, startErr <-chan error) error {
	t.Helper()

	timer := time.NewTimer(cryptoutilSharedMagic.TestDefaultServerShutdownTimeout)
	defer timer.Stop()

	select {
	case err := <-startErr:
		return err
	case <-timer.C:
		t.Fatal("server Start did not return before the shutdown deadline")

		return nil
	}
}

// HTTPServer represents a minimal HTTP server interface for shutdown testing.
type HTTPServer interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
	ActualPort() int
}

func waitForServerStart(t *testing.T, server HTTPServer) {
	t.Helper()

	require.Eventually(t, func() bool {
		port := server.ActualPort()
		if port <= 0 {
			return false
		}

		dialer := net.Dialer{Timeout: cryptoutilSharedMagic.TestTimeoutHTTPRetryInterval}

		conn, err := dialer.DialContext(context.Background(), "tcp", net.JoinHostPort(cryptoutilSharedMagic.IPv4Loopback, strconv.Itoa(port)))
		if err != nil {
			return false
		}

		return conn.Close() == nil
	}, cryptoutilSharedMagic.TestTimeoutCryptoutilReady, cryptoutilSharedMagic.TestTimeoutHTTPRetryInterval, "server did not accept a TCP connection")
}

// TestShutdownGraceful verifies that server shuts down gracefully.
func TestShutdownGraceful(t *testing.T, createServer func(t *testing.T) HTTPServer) {
	t.Helper()

	server := createServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background.
	startErr := make(chan error, 1)

	go func() {
		startErr <- server.Start(ctx)
	}()

	waitForServerStart(t, server)

	// Shutdown server.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TestDefaultServerShutdownTimeout)
	defer shutdownCancel()

	err := server.Shutdown(shutdownCtx)
	require.NoError(t, err, "Shutdown should succeed gracefully")

	// Verify Start() returned (may be nil or context-cancelled error).
	if err := waitForServerExit(t, startErr); err != nil {
		assert.Contains(t, err.Error(), "server stopped", "Start should indicate server stopped")
	}

	// Wait for port to be fully released.
	time.Sleep(testPortReleaseWait)
}

// TestShutdownNilContext verifies that Shutdown accepts nil context (background fallback).
func TestShutdownNilContext(t *testing.T, createServer func(t *testing.T) HTTPServer) {
	t.Helper()

	server := createServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background.
	startErr := make(chan error, 1)

	go func() {
		startErr <- server.Start(ctx)
	}()

	waitForServerStart(t, server)

	// Shutdown with nil context (should use background context internally).
	err := server.Shutdown(nil) //nolint:staticcheck // Testing nil context handling.
	require.NoError(t, err, "Shutdown should accept nil context")

	require.NoError(t, waitForServerExit(t, startErr), "Start should return cleanly after shutdown")

	// Wait for port to be fully released.
	time.Sleep(testPortReleaseWait)
}

// TestShutdownDoubleCall verifies that calling Shutdown twice returns error.
func TestShutdownDoubleCall(t *testing.T, createServer func(t *testing.T) HTTPServer) {
	t.Helper()

	server := createServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background.
	startErr := make(chan error, 1)

	go func() {
		startErr <- server.Start(ctx)
	}()

	waitForServerStart(t, server)

	// First shutdown - should succeed.
	shutdownCtx1, shutdownCancel1 := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TestDefaultServerShutdownTimeout)
	defer shutdownCancel1()

	err := server.Shutdown(shutdownCtx1)
	require.NoError(t, err, "First shutdown should succeed")

	// Second shutdown - should return error.
	shutdownCtx2, shutdownCancel2 := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TestDefaultServerShutdownTimeout)
	defer shutdownCancel2()

	err = server.Shutdown(shutdownCtx2)
	require.Error(t, err, "Second shutdown should return error")
	assert.Contains(t, err.Error(), "already shutdown", "Error should indicate already shutdown")

	require.NoError(t, waitForServerExit(t, startErr), "Start should return cleanly after shutdown")

	// Wait for port to be fully released.
	time.Sleep(testPortReleaseWait)
}

// TestStartNilContext verifies that Start rejects nil context.
func TestStartNilContext(t *testing.T, createServer func(t *testing.T) HTTPServer) {
	t.Helper()

	server := createServer(t)

	err := server.Start(nil) //nolint:staticcheck // Testing nil context handling.
	require.Error(t, err, "Start should reject nil context")
	assert.Contains(t, err.Error(), "context cannot be nil", "Error should indicate nil context")
}
