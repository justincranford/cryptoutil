// Copyright (c) 2025 Justin Cranford

package testutil

import (
	"context"
	"fmt"
	"sync"
)

// MockPublicServer is a test double for PublicServer interface.
type MockPublicServer struct {
	StartCalled    bool
	StartErr       error
	ShutdownCalled bool
	ShutdownErr    error
	actualPort     int
	startBlock     chan struct{} // Used to control blocking behavior.
	mu             sync.Mutex
}

// NewMockPublicServer creates a new mock public server for testing.
func NewMockPublicServer(actualPort int) *MockPublicServer {
	return &MockPublicServer{
		actualPort: actualPort,
		startBlock: make(chan struct{}),
	}
}

// Start simulates starting the public server.
func (m *MockPublicServer) Start(ctx context.Context) error {
	m.mu.Lock()
	m.StartCalled = true
	m.mu.Unlock()

	if m.StartErr != nil {
		return m.StartErr
	}

	// Block until shutdown or context cancelled.
	select {
	case <-m.startBlock:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	}
}

// Shutdown simulates shutting down the public server.
func (m *MockPublicServer) Shutdown(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ShutdownCalled = true

	select {
	case <-m.startBlock:
		// Already closed, do nothing.
	default:
		close(m.startBlock) // Unblock Start().
	}

	return m.ShutdownErr
}

// ActualPort returns the actual port the server is listening on.
func (m *MockPublicServer) ActualPort() int {
	return m.actualPort
}

// PublicBaseURL returns the base URL for the public server.
func (m *MockPublicServer) PublicBaseURL() string {
	return fmt.Sprintf("https://127.0.0.1:%d", m.actualPort)
}

// MockAdminServer is a test double for AdminServer interface.
type MockAdminServer struct {
	StartCalled    bool
	StartErr       error
	ShutdownCalled bool
	ShutdownErr    error
	actualPort     int
	startBlock     chan struct{} // Used to control blocking behavior.
	ready          bool
	mu             sync.Mutex
}

// NewMockAdminServer creates a new mock admin server for testing.
func NewMockAdminServer(actualPort int) *MockAdminServer {
	return &MockAdminServer{
		actualPort: actualPort,
		startBlock: make(chan struct{}),
	}
}

// Start simulates starting the admin server.
func (m *MockAdminServer) Start(ctx context.Context) error {
	m.mu.Lock()
	m.StartCalled = true
	m.mu.Unlock()

	if m.StartErr != nil {
		return m.StartErr
	}

	// Block until shutdown or context cancelled.
	select {
	case <-m.startBlock:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	}
}

// Shutdown simulates shutting down the admin server.
func (m *MockAdminServer) Shutdown(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ShutdownCalled = true

	select {
	case <-m.startBlock:
		// Already closed, do nothing.
	default:
		close(m.startBlock) // Unblock Start().
	}

	return m.ShutdownErr
}

// ActualPort returns the actual port the admin server is listening on.
func (m *MockAdminServer) ActualPort() int {
	return m.actualPort
}

// AdminBaseURL returns the base URL for the admin server.
func (m *MockAdminServer) AdminBaseURL() string {
	return fmt.Sprintf("https://127.0.0.1:%d", m.actualPort)
}

// SetReady sets the readiness status of the admin server.
func (m *MockAdminServer) SetReady(ready bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ready = ready
}
