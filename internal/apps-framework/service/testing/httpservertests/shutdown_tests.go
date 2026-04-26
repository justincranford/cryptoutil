// Copyright (c) 2025 Justin Cranford

// Package httpservertests provides reusable test cases for HTTP server implementations.
// These generic tests verify shutdown behavior, double-shutdown handling, and health checks
// during shutdown scenarios. Product services should use these tests instead of duplicating
// shutdown test logic across multiple implementations.
package httpservertests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test-specific timing constants.
const (
	testShutdownTimeout   = 5 * time.Second
	testServerStartupWait = 1 * time.Second
	testPortReleaseWait   = 500 * time.Millisecond
)

// HTTPServer represents a minimal HTTP server interface for shutdown testing.
type HTTPServer interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// TestShutdownGraceful verifies that server shuts down gracefully.
func TestShutdownGraceful(t *testing.T, createServer func(t *testing.T) HTTPServer) {
	t.Helper()

	server := createServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background.
	var wg sync.WaitGroup

	wg.Add(1)

	var startErr error

	go func() {
		defer wg.Done()

		startErr = server.Start(ctx)
	}()

	// Wait for server to start.
	time.Sleep(testServerStartupWait)

	// Shutdown server.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), testShutdownTimeout)
	defer shutdownCancel()

	err := server.Shutdown(shutdownCtx)
	require.NoError(t, err, "Shutdown should succeed gracefully")

	wg.Wait()

	// Verify Start() returned (may be nil or context-cancelled error).
	if startErr != nil {
		assert.Contains(t, startErr.Error(), "server stopped", "Start should indicate server stopped")
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
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		_ = server.Start(ctx)
	}()

	// Wait for server to start.
	time.Sleep(testServerStartupWait)

	// Shutdown with nil context (should use background context internally).
	err := server.Shutdown(nil) //nolint:staticcheck // Testing nil context handling.
	require.NoError(t, err, "Shutdown should accept nil context")

	wg.Wait()

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
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		_ = server.Start(ctx)
	}()

	// Wait for server to start.
	time.Sleep(testServerStartupWait)

	// First shutdown - should succeed.
	shutdownCtx1, shutdownCancel1 := context.WithTimeout(context.Background(), testShutdownTimeout)
	defer shutdownCancel1()

	err := server.Shutdown(shutdownCtx1)
	require.NoError(t, err, "First shutdown should succeed")

	// Second shutdown - should return error.
	shutdownCtx2, shutdownCancel2 := context.WithTimeout(context.Background(), testShutdownTimeout)
	defer shutdownCancel2()

	err = server.Shutdown(shutdownCtx2)
	require.Error(t, err, "Second shutdown should return error")
	assert.Contains(t, err.Error(), "already shutdown", "Error should indicate already shutdown")

	// Cleanup.
	wg.Wait()

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
