// Copyright (c) 2025 Justin Cranford
//
//

package server_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
)

// mockPublicServer is a mock implementation of IPublicServer for testing.
type mockPublicServer struct {
	mu             sync.RWMutex
	started        bool
	startErr       error
	shutdownErr    error
	port           int
	baseURL        string
	startDelay     time.Duration
	blockUntilStop bool
	stopChan       chan struct{}
}

func newMockPublicServer(port int, baseURL string) *mockPublicServer {
	return &mockPublicServer{
		port:     port,
		baseURL:  baseURL,
		stopChan: make(chan struct{}),
	}
}

func (m *mockPublicServer) Start(ctx context.Context) error {
	m.mu.Lock()
	m.started = true
	m.mu.Unlock()

	if m.startDelay > 0 {
		time.Sleep(m.startDelay)
	}

	if m.startErr != nil {
		return m.startErr
	}

	if m.blockUntilStop {
		select {
		case <-ctx.Done():
			return fmt.Errorf("public server context cancelled: %w", ctx.Err())
		case <-m.stopChan:
			return nil
		}
	}

	return nil
}

func (m *mockPublicServer) Shutdown(_ context.Context) error {
	close(m.stopChan)

	return m.shutdownErr
}

func (m *mockPublicServer) ActualPort() int {
	return m.port
}

func (m *mockPublicServer) PublicBaseURL() string {
	return m.baseURL
}

// mockAdminServer is a mock implementation of IAdminServer for testing.
type mockAdminServer struct {
	mu             sync.RWMutex
	started        bool
	ready          bool
	startErr       error
	shutdownErr    error
	port           int
	baseURL        string
	startDelay     time.Duration
	blockUntilStop bool
	stopChan       chan struct{}
}

func newMockAdminServer(port int, baseURL string) *mockAdminServer {
	return &mockAdminServer{
		port:     port,
		baseURL:  baseURL,
		stopChan: make(chan struct{}),
	}
}

func (m *mockAdminServer) Start(ctx context.Context) error {
	m.mu.Lock()
	m.started = true
	m.mu.Unlock()

	if m.startDelay > 0 {
		time.Sleep(m.startDelay)
	}

	if m.startErr != nil {
		return m.startErr
	}

	if m.blockUntilStop {
		select {
		case <-ctx.Done():
			return fmt.Errorf("admin server context cancelled: %w", ctx.Err())
		case <-m.stopChan:
			return nil
		}
	}

	return nil
}

func (m *mockAdminServer) Shutdown(_ context.Context) error {
	close(m.stopChan)

	return m.shutdownErr
}

func (m *mockAdminServer) ActualPort() int {
	return m.port
}

func (m *mockAdminServer) SetReady(ready bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ready = ready
}

func (m *mockAdminServer) AdminBaseURL() string {
	return m.baseURL
}

func (m *mockAdminServer) isReady() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.ready
}

// TestNewApplication_HappyPath tests successful application creation.
func TestNewApplication_HappyPath(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080, "https://localhost:8080")
	adminServer := newMockAdminServer(9090, "https://localhost:9090")

	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)

	require.NoError(t, err)
	require.NotNil(t, app)
	require.False(t, app.IsShutdown())
}

// TestNewApplication_NilContext tests application creation with nil context.
func TestNewApplication_NilContext(t *testing.T) {
	t.Parallel()

	publicServer := newMockPublicServer(8080, "https://localhost:8080")
	adminServer := newMockAdminServer(9090, "https://localhost:9090")

	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(nil, publicServer, adminServer) //nolint:staticcheck // SA1012 - Testing nil context behavior intentionally

	require.Error(t, err)
	require.Nil(t, app)
	require.Contains(t, err.Error(), "context cannot be nil")
}

// TestNewApplication_NilPublicServer tests application creation with nil public server.
func TestNewApplication_NilPublicServer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	adminServer := newMockAdminServer(9090, "https://localhost:9090")

	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, nil, adminServer)

	require.Error(t, err)
	require.Nil(t, app)
	require.Contains(t, err.Error(), "publicServer cannot be nil")
}

// TestNewApplication_NilAdminServer tests application creation with nil admin server.
func TestNewApplication_NilAdminServer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080, "https://localhost:8080")

	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, nil)

	require.Error(t, err)
	require.Nil(t, app)
	require.Contains(t, err.Error(), "adminServer cannot be nil")
}

// TestApplication_Start_NilContext tests Start with nil context.
func TestApplication_Start_NilContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080, "https://localhost:8080")
	adminServer := newMockAdminServer(9090, "https://localhost:9090")

	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	err = app.Start(nil) //nolint:staticcheck // SA1012 - Testing nil context behavior intentionally

	require.Error(t, err)
	require.Contains(t, err.Error(), "context cannot be nil")
}

// TestApplication_Start_PublicServerFails tests Start when public server fails.
func TestApplication_Start_PublicServerFails(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080, "https://localhost:8080")
	publicServer.startErr = fmt.Errorf("public server startup failed")

	adminServer := newMockAdminServer(9090, "https://localhost:9090")
	adminServer.blockUntilStop = true

	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	startCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err = app.Start(startCtx)

	require.Error(t, err)
	require.Contains(t, err.Error(), "public server failed")
	require.True(t, app.IsShutdown())
}

// TestApplication_Start_AdminServerFails tests Start when admin server fails.
func TestApplication_Start_AdminServerFails(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080, "https://localhost:8080")
	publicServer.blockUntilStop = true

	adminServer := newMockAdminServer(9090, "https://localhost:9090")
	adminServer.startErr = fmt.Errorf("admin server startup failed")

	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	startCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err = app.Start(startCtx)

	require.Error(t, err)
	require.Contains(t, err.Error(), "admin server failed")
	require.True(t, app.IsShutdown())
}

// TestApplication_Start_ContextCancelled tests Start when context is cancelled.
func TestApplication_Start_ContextCancelled(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080, "https://localhost:8080")
	publicServer.blockUntilStop = true

	adminServer := newMockAdminServer(9090, "https://localhost:9090")
	adminServer.blockUntilStop = true

	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	startCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	err = app.Start(startCtx)

	require.Error(t, err)
	require.Contains(t, err.Error(), "application startup cancelled")
	require.True(t, app.IsShutdown())
}

// TestApplication_Shutdown_NilContext tests Shutdown with nil context (uses Background).
func TestApplication_Shutdown_NilContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080, "https://localhost:8080")
	adminServer := newMockAdminServer(9090, "https://localhost:9090")

	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	err = app.Shutdown(nil) //nolint:staticcheck // SA1012 - Testing nil context behavior intentionally (Shutdown accepts nil)

	require.NoError(t, err)
	require.True(t, app.IsShutdown())
}

// TestApplication_Shutdown_PublicServerFails tests Shutdown when public server fails.
func TestApplication_Shutdown_PublicServerFails(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080, "https://localhost:8080")
	publicServer.shutdownErr = fmt.Errorf("public server shutdown failed")

	adminServer := newMockAdminServer(9090, "https://localhost:9090")

	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	err = app.Shutdown(ctx)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to shutdown public server")
	require.True(t, app.IsShutdown())
}

// TestApplication_Shutdown_AdminServerFails tests Shutdown when admin server fails.
func TestApplication_Shutdown_AdminServerFails(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080, "https://localhost:8080")
	adminServer := newMockAdminServer(9090, "https://localhost:9090")
	adminServer.shutdownErr = fmt.Errorf("admin server shutdown failed")

	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	err = app.Shutdown(ctx)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to shutdown admin server")
	require.True(t, app.IsShutdown())
}

// TestApplication_Shutdown_BothServersFail tests Shutdown when both servers fail.
func TestApplication_Shutdown_BothServersFail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080, "https://localhost:8080")
	publicServer.shutdownErr = fmt.Errorf("public server shutdown failed")

	adminServer := newMockAdminServer(9090, "https://localhost:9090")
	adminServer.shutdownErr = fmt.Errorf("admin server shutdown failed")

	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	err = app.Shutdown(ctx)

	require.Error(t, err)
	require.Contains(t, err.Error(), "multiple shutdown errors")
	require.Contains(t, err.Error(), "public=")
	require.Contains(t, err.Error(), "admin=")
	require.True(t, app.IsShutdown())
}

// TestApplication_PublicPort tests PublicPort method.
func TestApplication_PublicPort(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080, "https://localhost:8080")
	adminServer := newMockAdminServer(9090, "https://localhost:9090")

	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	port := app.PublicPort()
	require.Equal(t, 8080, port)
}

// TestApplication_AdminPort tests AdminPort method.
func TestApplication_AdminPort(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080, "https://localhost:8080")
	adminServer := newMockAdminServer(9090, "https://localhost:9090")

	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	port := app.AdminPort()
	require.Equal(t, 9090, port)
}

// TestApplication_PublicBaseURL tests PublicBaseURL method.
func TestApplication_PublicBaseURL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080, "https://localhost:8080")
	adminServer := newMockAdminServer(9090, "https://localhost:9090")

	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	baseURL := app.PublicBaseURL()
	require.Equal(t, "https://localhost:8080", baseURL)
}

// TestApplication_AdminBaseURL tests AdminBaseURL method.
func TestApplication_AdminBaseURL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080, "https://localhost:8080")
	adminServer := newMockAdminServer(9090, "https://localhost:9090")

	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	baseURL := app.AdminBaseURL()
	require.Equal(t, "https://localhost:9090", baseURL)
}

// TestApplication_SetReady tests SetReady method.
func TestApplication_SetReady(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080, "https://localhost:8080")
	adminServer := newMockAdminServer(9090, "https://localhost:9090")

	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	require.False(t, adminServer.isReady())

	app.SetReady(true)
	require.True(t, adminServer.isReady())

	app.SetReady(false)
	require.False(t, adminServer.isReady())
}
