// Copyright (c) 2025-2026 Justin Cranford.
// Package stubs provides exported IPublicServer and IAdminServer stub implementations
// for use in PS-ID unit tests. Centralises the stub definitions so each PS-ID does
// not need to re-define identical local types.
//
// Usage in a PS-ID server test:
//
//	import cryptoutilTestingStubs "cryptoutil/internal/apps-framework/service/testing/stubs"
//
//	app := cryptoutilTestingStubs.NewTestApplication(t)
//	srv := &MyServer{resources: &builder.ServiceResources{Application: app}}
package stubs

import (
	"context"
	"crypto/x509"
	"testing"

	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps-framework/service/server"
)

const (
	// StubPublicPortValue is the port returned by StubPublicServer.ActualPort().
	StubPublicPortValue = 8443
	// StubAdminPortValue is the port returned by StubAdminServer.ActualPort().
	StubAdminPortValue = 9090

	stubPublicBaseURL = "https://localhost:8443"
	stubAdminBaseURL  = "https://localhost:9090"
)

// StubPublicServer is a minimal IPublicServer for test infrastructure setup.
// Set StartErr to simulate startup failures.
type StubPublicServer struct {
	StartErr error
}

func (s *StubPublicServer) Start(_ context.Context) error    { return s.StartErr }
func (s *StubPublicServer) Shutdown(_ context.Context) error { return nil }
func (s *StubPublicServer) ActualPort() int                  { return StubPublicPortValue }
func (s *StubPublicServer) PublicBaseURL() string            { return stubPublicBaseURL }

// StubAdminServer is a minimal IAdminServer for test infrastructure setup.
type StubAdminServer struct{}

func (s *StubAdminServer) Start(_ context.Context) error      { return nil }
func (s *StubAdminServer) Shutdown(_ context.Context) error   { return nil }
func (s *StubAdminServer) ActualPort() int                    { return StubAdminPortValue }
func (s *StubAdminServer) SetReady(_ bool)                    {}
func (s *StubAdminServer) AdminBaseURL() string               { return stubAdminBaseURL }
func (s *StubAdminServer) AdminTLSRootCAPool() *x509.CertPool { return nil }

// NewTestApplication creates a test Application backed by stub servers.
// Calls t.Fatalf if Application construction fails.
func NewTestApplication(t testing.TB) *cryptoutilAppsFrameworkServiceServer.Application {
	t.Helper()

	app, err := cryptoutilAppsFrameworkServiceServer.NewApplication(
		context.Background(), &StubPublicServer{}, &StubAdminServer{},
	)
	if err != nil {
		t.Fatalf("stubs: failed to create test application: %v", err)
	}

	return app
}
