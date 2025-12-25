// Copyright (c) 2025 Justin Cranford
//
//

package server_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	cryptoutilTemplateServer "cryptoutil/internal/template/server"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPublicServer is a test double for PublicServer interface.
type mockPublicServer struct {
	startCalled    bool
	startErr       error
	shutdownCalled bool
	shutdownErr    error
	actualPort     int
	startBlock     chan struct{} // Used to control blocking behavior.
	mu             sync.Mutex
}

func newMockPublicServer(actualPort int) *mockPublicServer {
	return &mockPublicServer{
		actualPort: actualPort,
		startBlock: make(chan struct{}),
	}
}

func (m *mockPublicServer) Start(ctx context.Context) error {
	m.mu.Lock()
	m.startCalled = true
	m.mu.Unlock()

	if m.startErr != nil {
		return m.startErr
	}

	// Block until shutdown or context cancelled.
	select {
	case <-m.startBlock:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	}
}

func (m *mockPublicServer) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.shutdownCalled = true

	select {
	case <-m.startBlock:
		// Already closed, do nothing.
	default:
		close(m.startBlock) // Unblock Start().
	}

	return m.shutdownErr
}

func (m *mockPublicServer) ActualPort() int {
	return m.actualPort
}

// mockAdminServer is a test double for AdminServer interface.
type mockAdminServer struct {
	startCalled    bool
	startErr       error
	shutdownCalled bool
	shutdownErr    error
	actualPort     int
	startBlock     chan struct{} // Used to control blocking behavior.
	mu             sync.Mutex
}

func newMockAdminServer(actualPort int) *mockAdminServer {
	return &mockAdminServer{
		actualPort: actualPort,
		startBlock: make(chan struct{}),
	}
}

func (m *mockAdminServer) Start(ctx context.Context) error {
	m.mu.Lock()
	m.startCalled = true
	m.mu.Unlock()

	if m.startErr != nil {
		return m.startErr
	}

	// Block until shutdown or context cancelled.
	select {
	case <-m.startBlock:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	}
}

func (m *mockAdminServer) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.shutdownCalled = true

	select {
	case <-m.startBlock:
		// Already closed, do nothing.
	default:
		close(m.startBlock) // Unblock Start().
	}

	return m.shutdownErr
}

func (m *mockAdminServer) ActualPort() (int, error) {
	if m.actualPort == 0 {
		return 0, errors.New("admin server not initialized")
	}

	return m.actualPort, nil
}

// TestNewApplication_HappyPath tests successful application creation.
func TestNewApplication_HappyPath(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080)
	adminServer := newMockAdminServer(9090)

	app, err := cryptoutilTemplateServer.NewApplication(ctx, publicServer, adminServer)

	require.NoError(t, err)
	require.NotNil(t, app)
	assert.False(t, app.IsShutdown())
}

// TestNewApplication_NilContext tests error when context is nil.
func TestNewApplication_NilContext(t *testing.T) {
	t.Parallel()

	publicServer := newMockPublicServer(8080)
	adminServer := newMockAdminServer(9090)

	app, err := cryptoutilTemplateServer.NewApplication(nil, publicServer, adminServer) //nolint:staticcheck // Testing nil context handling.

	require.Error(t, err)
	assert.Nil(t, app)
	assert.Contains(t, err.Error(), "context cannot be nil")
}

// TestNewApplication_NilPublicServer tests error when publicServer is nil.
func TestNewApplication_NilPublicServer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	adminServer := newMockAdminServer(9090)

	app, err := cryptoutilTemplateServer.NewApplication(ctx, nil, adminServer)

	require.Error(t, err)
	assert.Nil(t, app)
	assert.Contains(t, err.Error(), "publicServer cannot be nil")
}

// TestNewApplication_NilAdminServer tests error when adminServer is nil.
func TestNewApplication_NilAdminServer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080)

	app, err := cryptoutilTemplateServer.NewApplication(ctx, publicServer, nil)

	require.Error(t, err)
	assert.Nil(t, app)
	assert.Contains(t, err.Error(), "adminServer cannot be nil")
}

// TestApplication_Start_HappyPath tests successful concurrent server startup.
func TestApplication_Start_HappyPath(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	publicServer := newMockPublicServer(8080)
	adminServer := newMockAdminServer(9090)

	app, err := cryptoutilTemplateServer.NewApplication(context.Background(), publicServer, adminServer)
	require.NoError(t, err)

	// Start in background.
	errChan := make(chan error, 1)

	go func() {
		errChan <- app.Start(ctx)
	}()

	// Give servers time to start.
	time.Sleep(100 * time.Millisecond)

	// Verify servers started.
	assert.True(t, publicServer.startCalled)
	assert.True(t, adminServer.startCalled)

	// Trigger shutdown via context cancellation.
	cancel()

	// Wait for Start() to return.
	err = <-errChan

	require.Error(t, err)
	assert.Contains(t, err.Error(), "application startup cancelled")
	assert.True(t, publicServer.shutdownCalled)
	assert.True(t, adminServer.shutdownCalled)
}

// TestApplication_Start_PublicServerFailure tests error handling when public server fails to start.
func TestApplication_Start_PublicServerFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080)
	publicServer.startErr = errors.New("public server bind error")
	adminServer := newMockAdminServer(9090)

	app, err := cryptoutilTemplateServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	err = app.Start(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "public server failed")
	assert.Contains(t, err.Error(), "public server bind error")
	assert.True(t, app.IsShutdown())
}

// TestApplication_Start_AdminServerFailure tests error handling when admin server fails to start.
func TestApplication_Start_AdminServerFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080)
	adminServer := newMockAdminServer(9090)
	adminServer.startErr = errors.New("admin server bind error")

	app, err := cryptoutilTemplateServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	err = app.Start(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin server failed")
	assert.Contains(t, err.Error(), "admin server bind error")
	assert.True(t, app.IsShutdown())
}

// TestApplication_Start_NilContext tests error when Start called with nil context.
func TestApplication_Start_NilContext(t *testing.T) {
	t.Parallel()

	publicServer := newMockPublicServer(8080)
	adminServer := newMockAdminServer(9090)

	app, err := cryptoutilTemplateServer.NewApplication(context.Background(), publicServer, adminServer)
	require.NoError(t, err)

	err = app.Start(nil) //nolint:staticcheck // Testing nil context handling.

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context cannot be nil")
}

// TestApplication_Shutdown_HappyPath tests graceful shutdown of both servers.
func TestApplication_Shutdown_HappyPath(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080)
	adminServer := newMockAdminServer(9090)

	app, err := cryptoutilTemplateServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	err = app.Shutdown(ctx)

	require.NoError(t, err)
	assert.True(t, publicServer.shutdownCalled)
	assert.True(t, adminServer.shutdownCalled)
	assert.True(t, app.IsShutdown())
}

// TestApplication_Shutdown_PublicServerError tests error handling when public server fails to shutdown.
func TestApplication_Shutdown_PublicServerError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080)
	publicServer.shutdownErr = errors.New("public server shutdown error")
	adminServer := newMockAdminServer(9090)

	app, err := cryptoutilTemplateServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	err = app.Shutdown(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to shutdown public server")
	assert.True(t, publicServer.shutdownCalled)
	assert.True(t, adminServer.shutdownCalled)
}

// TestApplication_Shutdown_AdminServerError tests error handling when admin server fails to shutdown.
func TestApplication_Shutdown_AdminServerError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080)
	adminServer := newMockAdminServer(9090)
	adminServer.shutdownErr = errors.New("admin server shutdown error")

	app, err := cryptoutilTemplateServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	err = app.Shutdown(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to shutdown admin server")
	assert.True(t, publicServer.shutdownCalled)
	assert.True(t, adminServer.shutdownCalled)
}

// TestApplication_Shutdown_BothServersError tests error reporting when both servers fail to shutdown.
func TestApplication_Shutdown_BothServersError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080)
	publicServer.shutdownErr = errors.New("public shutdown error")
	adminServer := newMockAdminServer(9090)
	adminServer.shutdownErr = errors.New("admin shutdown error")

	app, err := cryptoutilTemplateServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	err = app.Shutdown(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "multiple shutdown errors")
	assert.Contains(t, err.Error(), "public shutdown error")
	assert.Contains(t, err.Error(), "admin shutdown error")
}

// TestApplication_Shutdown_NilContext tests Shutdown uses context.Background when context is nil.
func TestApplication_Shutdown_NilContext(t *testing.T) {
	t.Parallel()

	publicServer := newMockPublicServer(8080)
	adminServer := newMockAdminServer(9090)

	app, err := cryptoutilTemplateServer.NewApplication(context.Background(), publicServer, adminServer)
	require.NoError(t, err)

	err = app.Shutdown(context.Background())

	require.NoError(t, err)
	assert.True(t, publicServer.shutdownCalled)
	assert.True(t, adminServer.shutdownCalled)
}

// TestApplication_PublicPort tests retrieval of public server actual port.
func TestApplication_PublicPort(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080)
	adminServer := newMockAdminServer(9090)

	app, err := cryptoutilTemplateServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	actualPort := app.PublicPort()

	assert.Equal(t, 8080, actualPort)
}

// TestApplication_AdminPort tests retrieval of admin server actual port.
func TestApplication_AdminPort(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080)
	adminServer := newMockAdminServer(9090)

	app, err := cryptoutilTemplateServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	actualPort, err := app.AdminPort()

	require.NoError(t, err)
	assert.Equal(t, 9090, actualPort)
}

// TestApplication_AdminPort_NotInitialized tests error when admin server not initialized.
func TestApplication_AdminPort_NotInitialized(t *testing.T) {
	t.Parallel()

	// Create application with nil admin server (will fail NewApplication).
	ctx := context.Background()
	publicServer := newMockPublicServer(8080)
	adminServer := newMockAdminServer(0) // Port 0 simulates uninitialized.

	app, err := cryptoutilTemplateServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	_, err = app.AdminPort()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin server not initialized")
}

// TestApplication_IsShutdown tests shutdown state tracking.
func TestApplication_IsShutdown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		performShutdown  bool
		expectedShutdown bool
	}{
		{
			name:             "not shutdown initially",
			performShutdown:  false,
			expectedShutdown: false,
		},
		{
			name:             "shutdown after Shutdown() called",
			performShutdown:  true,
			expectedShutdown: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			publicServer := newMockPublicServer(8080)
			adminServer := newMockAdminServer(9090)

			app, err := cryptoutilTemplateServer.NewApplication(ctx, publicServer, adminServer)
			require.NoError(t, err)

			if tt.performShutdown {
				err := app.Shutdown(ctx)
				require.NoError(t, err)
			}

			isShutdown := app.IsShutdown()
			assert.Equal(t, tt.expectedShutdown, isShutdown)
		})
	}
}

// TestApplication_ConcurrentShutdown tests thread-safety of shutdown state.
func TestApplication_ConcurrentShutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080)
	adminServer := newMockAdminServer(9090)

	app, err := cryptoutilTemplateServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	// Shutdown from multiple goroutines concurrently.
	const concurrency = 10

	var wg sync.WaitGroup

	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()

			_ = app.Shutdown(ctx)
		}()
	}

	wg.Wait()

	assert.True(t, app.IsShutdown())
}
