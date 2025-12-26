// Copyright (c) 2025 Justin Cranford
//
//

package server_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
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
	ready          bool
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

func (m *mockAdminServer) SetReady(ready bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ready = ready
}
