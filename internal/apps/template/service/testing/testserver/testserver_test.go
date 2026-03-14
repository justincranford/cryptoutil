// Copyright (c) 2025 Justin Cranford
//

package testserver_test

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilTestingTestserver "cryptoutil/internal/apps/template/service/testing/testserver"
)

// mockServer implements ServiceServer for testing without starting real HTTPS servers.
type mockServer struct {
	mu          sync.Mutex
	publicPort  int
	adminPort   int
	ready       bool
	startCalled bool
	startDelay  time.Duration
	startErr    error
	shutdownErr error
}

func newMockServer() *mockServer {
	return &mockServer{}
}

// Start simulates a real server binding to ports: sets publicPort and adminPort under mutex.
func (m *mockServer) Start(_ context.Context) error {
	if m.startDelay > 0 {
		time.Sleep(m.startDelay)
	}

	if m.startErr != nil {
		return m.startErr
	}

	m.mu.Lock()
	m.startCalled = true
	m.publicPort = cryptoutilSharedMagic.DemoServerPort
	m.adminPort = cryptoutilSharedMagic.DemoAdminPort
	m.mu.Unlock()

	return nil
}

func (m *mockServer) Shutdown(_ context.Context) error {
	return m.shutdownErr
}

func (m *mockServer) DB() *gorm.DB {
	return nil
}

func (m *mockServer) App() *cryptoutilAppsTemplateServiceServer.Application {
	return nil
}

func (m *mockServer) PublicPort() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.publicPort
}

func (m *mockServer) AdminPort() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.adminPort
}

func (m *mockServer) SetReady(ready bool) {
	m.mu.Lock()
	m.ready = ready
	m.mu.Unlock()
}

func (m *mockServer) PublicBaseURL() string {
	return "https://127.0.0.1:8080"
}

func (m *mockServer) AdminBaseURL() string {
	return "https://127.0.0.1:9090"
}

func (m *mockServer) PublicServerActualPort() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.publicPort
}

func (m *mockServer) AdminServerActualPort() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.adminPort
}

func (m *mockServer) TLSRootCAPool() *x509.CertPool {
	return x509.NewCertPool()
}

func (m *mockServer) AdminTLSRootCAPool() *x509.CertPool {
	return x509.NewCertPool()
}

func (m *mockServer) isStartCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.startCalled
}

func (m *mockServer) isReady() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.ready
}

// captureTB wraps *testing.T to record Fatalf calls without exiting the test.
// Used to exercise error code paths in StartAndWait that call t.Fatalf.
type captureTB struct {
	*testing.T
	mu       sync.Mutex
	fatals   []string
	cleanups []func()
}

// Helper is a no-op override so captureTB call frames are not marked as test helpers.
func (c *captureTB) Helper() {}

// Fatalf records the fatal message without calling runtime.Goexit.
func (c *captureTB) Fatalf(format string, args ...any) {
	c.mu.Lock()
	c.fatals = append(c.fatals, fmt.Sprintf(format, args...))
	c.mu.Unlock()
}

// Cleanup captures cleanup functions instead of registering them with *testing.T.
func (c *captureTB) Cleanup(f func()) {
	c.mu.Lock()
	c.cleanups = append(c.cleanups, f)
	c.mu.Unlock()
}

func (c *captureTB) runCleanups() {
	c.mu.Lock()
	fns := make([]func(), len(c.cleanups))
	copy(fns, c.cleanups)
	c.mu.Unlock()

	for i := len(fns) - 1; i >= 0; i-- {
		fns[i]()
	}
}

func (c *captureTB) hasFatal() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.fatals) > 0
}

func TestStartAndWait_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	srv := newMockServer()

	result := cryptoutilTestingTestserver.StartAndWait(ctx, t, srv)

	require.NotNil(t, result)
	require.True(t, srv.isStartCalled(), "Start should have been called")
	require.True(t, srv.isReady(), "SetReady(true) should have been called")

	castedResult, ok := result.(*mockServer)
	require.True(t, ok, "returned server should be *mockServer")
	require.Same(t, srv, castedResult, "returned server should be the same instance")
}

func TestStartAndWait_RegistersCleanup(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	srv := newMockServer()

	t.Run("inner", func(t *testing.T) {
		t.Parallel()

		result := cryptoutilTestingTestserver.StartAndWait(ctx, t, srv)
		require.NotNil(t, result)
	})
}

func TestStartAndWait_ReturnsOriginalServer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	srv := newMockServer()

	returned := cryptoutilTestingTestserver.StartAndWait(ctx, t, srv)
	require.Equal(t, "https://127.0.0.1:8080", returned.PublicBaseURL())
	require.Equal(t, "https://127.0.0.1:9090", returned.AdminBaseURL())
}

func TestStartAndWait_PortsReady(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	srv := newMockServer()

	cryptoutilTestingTestserver.StartAndWait(ctx, t, srv)

	require.Equal(t, cryptoutilSharedMagic.DemoServerPort, srv.PublicPort())
	require.Equal(t, cryptoutilSharedMagic.DemoAdminPort, srv.AdminPort())
}

func TestStartAndWait_StartError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	srv := newMockServer()
	srv.startErr = errors.New("simulated start failure")

	ctb := &captureTB{T: t}
	_ = cryptoutilTestingTestserver.StartAndWait(ctx, ctb, srv)
	ctb.runCleanups()

	require.True(t, ctb.hasFatal(), "Fatalf should have been called for start error")
}

func TestStartAndWait_ShutdownError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("shutdown error is logged without failing test", func(t *testing.T) {
		t.Parallel()

		srv := newMockServer()
		srv.shutdownErr = errors.New("simulated shutdown failure")

		result := cryptoutilTestingTestserver.StartAndWait(ctx, t, srv)
		require.NotNil(t, result)
		// cleanup runs when inner test ends, exercising the t.Logf path for shutdownErr
	})
}
