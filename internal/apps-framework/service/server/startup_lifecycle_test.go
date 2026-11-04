// Copyright (c) 2025-2026 Justin Cranford.
package server_test

import (
	"context"
	"crypto/x509"
	"errors"
	"testing"
	"time"

	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps-framework/service/server"
	cryptoutilAppsFrameworkServiceServerReadiness "cryptoutil/internal/apps-framework/service/server/readiness"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

type startupTestPublicServer struct {
	startGate chan struct{}
	started   chan struct{}
	entered   chan struct{}
}

func (server *startupTestPublicServer) Start(ctx context.Context) error {
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

func (*startupTestPublicServer) Shutdown(context.Context) error  { return nil }
func (*startupTestPublicServer) ActualPort() int                 { return 0 }
func (*startupTestPublicServer) PublicBaseURL() string           { return "" }
func (server *startupTestPublicServer) Started() <-chan struct{} { return server.started }

type startupTestAdminServer struct {
	startGate       chan struct{}
	started         chan struct{}
	entered         chan struct{}
	startupComplete chan struct{}
}

func (server *startupTestAdminServer) Start(ctx context.Context) error {
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

func (*startupTestAdminServer) Shutdown(context.Context) error { return nil }
func (*startupTestAdminServer) ActualPort() int                { return 0 }
func (*startupTestAdminServer) SetReady(ready bool)            {}
func (*startupTestAdminServer) AdminBaseURL() string           { return "" }
func (*startupTestAdminServer) AdminTLSRootCAPool() *x509.CertPool {
	return nil
}
func (server *startupTestAdminServer) Started() <-chan struct{} { return server.started }
func (server *startupTestAdminServer) MarkStartupComplete() {
	close(server.startupComplete)
}
func (server *startupTestAdminServer) BeginDrain() {}

func TestApplicationStartWaitsForBothListeners(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	publicServer := &startupTestPublicServer{
		startGate: make(chan struct{}),
		started:   make(chan struct{}),
		entered:   make(chan struct{}),
	}
	adminServer := &startupTestAdminServer{
		startGate:       make(chan struct{}),
		started:         make(chan struct{}),
		entered:         make(chan struct{}),
		startupComplete: make(chan struct{}),
	}
	application, err := cryptoutilAppsFrameworkServiceServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	startResult := make(chan error, 1)
	go func() {
		startResult <- application.Start(ctx)
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

	cancel()
	require.ErrorContains(t, <-startResult, "application startup cancelled")
}

func TestApplicationStartFailsBeforeStartupWhenDependencyObservationFails(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	publicServer := &startupTestPublicServer{
		startGate: make(chan struct{}),
		started:   make(chan struct{}),
		entered:   make(chan struct{}),
	}
	adminServer := &startupTestAdminServer{
		startGate:       make(chan struct{}),
		started:         make(chan struct{}),
		entered:         make(chan struct{}),
		startupComplete: make(chan struct{}),
	}
	application, err := cryptoutilAppsFrameworkServiceServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	monitor, err := cryptoutilAppsFrameworkServiceServerReadiness.New(cryptoutilAppsFrameworkServiceServerReadiness.Config{
		Checks: []cryptoutilAppsFrameworkServiceServerReadiness.Check{{
			Name:       cryptoutilSharedMagic.RealmStorageTypeDatabase,
			ReasonCode: "database_unavailable",
			Run: func(context.Context) error {
				return errors.New("private database connection details")
			},
		}},
		Jitter: func(period time.Duration) time.Duration { return period },
	})
	require.NoError(t, err)
	application.SetReadinessMonitor(monitor)

	startResult := make(chan error, 1)
	go func() {
		startResult <- application.Start(ctx)
	}()

	<-publicServer.entered
	<-adminServer.entered
	close(publicServer.startGate)
	close(adminServer.startGate)

	require.ErrorIs(t, monitor.WaitInitial(ctx), cryptoutilAppsFrameworkServiceServerReadiness.ErrInitialChecksFailed)
	cancel()

	err = <-startResult
	require.ErrorContains(t, err, "database_unavailable")
	require.NotContains(t, err.Error(), "private database connection details")

	select {
	case <-adminServer.startupComplete:
		t.Fatal("startup latch opened after failed dependency observations")
	default:
	}
}
