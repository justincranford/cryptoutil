// Copyright (c) 2025-2026 Justin Cranford.
package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type startupAwarePublicServer struct {
	*mockPublicServer
	entered   chan struct{}
	startGate chan struct{}
	started   chan struct{}
}

func (server *startupAwarePublicServer) Start(ctx context.Context) error {
	close(server.entered)

	select {
	case <-server.startGate:
		close(server.started)
	case <-ctx.Done():
		return ctx.Err()
	}

	<-ctx.Done()

	return ctx.Err()
}

func (server *startupAwarePublicServer) Started() <-chan struct{} {
	return server.started
}

type startupAwareAdminServer struct {
	*mockAdminServer
	entered         chan struct{}
	startGate       chan struct{}
	started         chan struct{}
	startupComplete chan struct{}
}

func (server *startupAwareAdminServer) Start(ctx context.Context) error {
	close(server.entered)

	select {
	case <-server.startGate:
		close(server.started)
	case <-ctx.Done():
		return ctx.Err()
	}

	<-ctx.Done()

	return ctx.Err()
}

func (server *startupAwareAdminServer) Started() <-chan struct{} {
	return server.started
}

func (server *startupAwareAdminServer) MarkStartupComplete() {
	close(server.startupComplete)
}

func (server *startupAwareAdminServer) BeginDrain() {
	server.SetReady(false)
}

func TestListenerStartWaitsForBothListenersBeforeStartup(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	publicServer := &startupAwarePublicServer{
		mockPublicServer: &mockPublicServer{},
		entered:          make(chan struct{}),
		startGate:        make(chan struct{}),
		started:          make(chan struct{}),
	}
	adminServer := &startupAwareAdminServer{
		mockAdminServer: &mockAdminServer{},
		entered:         make(chan struct{}),
		startGate:       make(chan struct{}),
		started:         make(chan struct{}),
		startupComplete: make(chan struct{}),
	}
	listener := &Listener{PublicServer: publicServer, AdminServer: adminServer}

	startResult := make(chan error, 1)
	go func() {
		startResult <- listener.Start(ctx)
	}()

	<-publicServer.entered
	<-adminServer.entered
	close(publicServer.startGate)
	<-publicServer.started

	select {
	case <-adminServer.startupComplete:
		t.Fatal("startup latch opened before the admin listener started")
	default:
	}

	close(adminServer.startGate)
	<-adminServer.startupComplete
	require.True(t, adminServer.ready)

	cancel()
	require.Error(t, <-startResult)
	require.False(t, adminServer.ready)
}
