// Copyright (c) 2025 Justin Cranford
//

package server_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/apps/cipher/im/server"
	"cryptoutil/internal/apps/cipher/im/server/config"
)

// TestServer_AccessorMethods tests all server accessor methods (delegation to Application).
// These methods exist for test infrastructure - they should just return non-nil values.
func TestServer_AccessorMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		accessorFn func() any
	}{
		{
			name:       "DB() returns non-nil GORM database",
			accessorFn: func() any { return testCipherIMServer.DB() },
		},
		{
			name:       "App() returns non-nil Application",
			accessorFn: func() any { return testCipherIMServer.App() },
		},
		{
			name:       "JWKGen() returns non-nil JWK generation service",
			accessorFn: func() any { return testCipherIMServer.JWKGen() },
		},
		{
			name:       "Telemetry() returns non-nil telemetry service",
			accessorFn: func() any { return testCipherIMServer.Telemetry() },
		},
		{
			name:       "PublicPort() returns valid port number",
			accessorFn: func() any { return testCipherIMServer.PublicPort() },
		},
		{
			name:       "AdminPort() returns valid port number",
			accessorFn: func() any { return testCipherIMServer.AdminPort() },
		},
		{
			name:       "PublicBaseURL() returns non-empty URL",
			accessorFn: func() any { return testCipherIMServer.PublicBaseURL() },
		},
		{
			name:       "AdminBaseURL() returns non-empty URL",
			accessorFn: func() any { return testCipherIMServer.AdminBaseURL() },
		},
		{
			name:       "PublicServerActualPort() returns valid port number",
			accessorFn: func() any { return testCipherIMServer.PublicServerActualPort() },
		},
		{
			name:       "AdminServerActualPort() returns valid port number",
			accessorFn: func() any { return testCipherIMServer.AdminServerActualPort() },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.accessorFn()
			require.NotNil(t, result, "Accessor method should return non-nil value")

			// Additional type-specific validations.
			switch v := result.(type) {
			case int:
				require.Greater(t, v, 0, "Port number should be > 0")
			case string:
				require.NotEmpty(t, v, "URL should not be empty")
			}
		})
	}
}

// TestServer_SetReady tests the SetReady method (marks server ready for health checks).
func TestServer_SetReady(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		ready bool
	}{
		{
			name:  "SetReady(true) marks server as ready",
			ready: true,
		},
		{
			name:  "SetReady(false) marks server as not ready",
			ready: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Cannot run in parallel because SetReady mutates shared testCipherIMServer state.
			// Test that SetReady doesn't panic (actual health check testing done in E2E tests).
			require.NotPanics(t, func() {
				testCipherIMServer.SetReady(tt.ready)
			}, "SetReady should not panic")

			// Reset to ready state for other tests.
			testCipherIMServer.SetReady(true)
		})
	}
}

// TestServer_Shutdown tests graceful shutdown behavior.
func TestServer_Shutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create test configuration.
	cfg := config.DefaultTestConfig()

	// Create new server instance (separate from TestMain's testCipherIMServer).
	testServer, err := server.NewFromConfig(ctx, cfg)
	require.NoError(t, err, "Failed to create test server")
	require.NotNil(t, testServer, "Server should not be nil")

	// Start server in background.
	startCtx, startCancel := context.WithCancel(ctx)
	defer startCancel()

	startErrCh := make(chan error, 1)

	go func() {
		startErrCh <- testServer.Start(startCtx)
	}()

	// Wait for server to become ready (check public port is assigned).
	require.Eventually(t, func() bool {
		return testServer.PublicPort() > 0
	}, 5*time.Second, 100*time.Millisecond, "Server should start within 5 seconds")

	// Test graceful shutdown.
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancel()

	err = testServer.Shutdown(shutdownCtx)
	require.NoError(t, err, "Shutdown should succeed without errors")

	// Verify Start() exits after shutdown.
	select {
	case startErr := <-startErrCh:
		// Start() may return error after graceful shutdown (context canceled) - this is acceptable.
		// The important part is that Start() exits (doesn't block forever).
		if startErr != nil {
			t.Logf("Start() returned error after shutdown (expected): %v", startErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Start() did not exit within 2 seconds after shutdown")
	}
}

// TestServer_Start_WithInvalidContext tests Start() error handling.
func TestServer_Start_WithInvalidContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create test configuration.
	cfg := config.DefaultTestConfig()

	// Create new server instance.
	testServer, err := server.NewFromConfig(ctx, cfg)
	require.NoError(t, err, "Failed to create test server")
	require.NotNil(t, testServer, "Server should not be nil")

	// Test with already-cancelled context (should fail immediately).
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel() // Cancel immediately.

	err = testServer.Start(cancelledCtx)
	require.Error(t, err, "Start() should error with cancelled context")
	require.Contains(t, err.Error(), "context canceled", "Error should mention context cancellation")

	// Cleanup: Shutdown server (may already be stopped, ignore error).
	_ = testServer.Shutdown(context.Background())
}

// TestNewFromConfig_WithNilContext tests NewFromConfig error handling for nil context.
func TestNewFromConfig_WithNilContext(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultTestConfig()

	testServer, err := server.NewFromConfig(nil, cfg)
	require.Error(t, err, "NewFromConfig should error with nil context")
	require.Nil(t, testServer, "Server should be nil on error")
	require.Contains(t, err.Error(), "context cannot be nil", "Error should mention nil context")
}

// TestNewFromConfig_WithNilConfig tests NewFromConfig error handling for nil config.
func TestNewFromConfig_WithNilConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testServer, err := server.NewFromConfig(ctx, nil)
	require.Error(t, err, "NewFromConfig should error with nil config")
	require.Nil(t, testServer, "Server should be nil on error")
	require.Contains(t, err.Error(), "config cannot be nil", "Error should mention nil config")
}

// TestNewFromConfig_SuccessfulCreation tests successful server creation from config.
func TestNewFromConfig_SuccessfulCreation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := config.DefaultTestConfig()

	testServer, err := server.NewFromConfig(ctx, cfg)
	require.NoError(t, err, "NewFromConfig should succeed with valid config")
	require.NotNil(t, testServer, "Server should not be nil")

	// Verify server has initialized all required services.
	require.NotNil(t, testServer.DB(), "DB should be initialized")
	require.NotNil(t, testServer.App(), "App should be initialized")
	require.NotNil(t, testServer.JWKGen(), "JWKGen should be initialized")
	require.NotNil(t, testServer.Telemetry(), "Telemetry should be initialized")

	// Cleanup: Shutdown server.
	err = testServer.Shutdown(context.Background())
	require.NoError(t, err, "Shutdown should succeed")
}
