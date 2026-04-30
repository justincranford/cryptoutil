// Copyright (c) 2025-2026 Justin Cranford.
package stubs_test

import (
	"context"
	"testing"

	cryptoutilTestingStubs "cryptoutil/internal/apps-framework/service/testing/stubs"

	"github.com/stretchr/testify/require"
)

func TestStubPublicServer_DefaultBehavior(t *testing.T) {
	t.Parallel()

	s := &cryptoutilTestingStubs.StubPublicServer{}

	require.NoError(t, s.Start(context.Background()))
	require.NoError(t, s.Shutdown(context.Background()))
	require.Equal(t, cryptoutilTestingStubs.StubPublicPortValue, s.ActualPort())
	require.Equal(t, "https://localhost:8443", s.PublicBaseURL())
}

func TestStubPublicServer_StartErr(t *testing.T) {
	t.Parallel()

	expected := context.DeadlineExceeded
	s := &cryptoutilTestingStubs.StubPublicServer{StartErr: expected}

	require.ErrorIs(t, s.Start(context.Background()), expected)
}

func TestStubAdminServer_DefaultBehavior(t *testing.T) {
	t.Parallel()

	s := &cryptoutilTestingStubs.StubAdminServer{}

	require.NoError(t, s.Start(context.Background()))
	require.NoError(t, s.Shutdown(context.Background()))
	require.Equal(t, cryptoutilTestingStubs.StubAdminPortValue, s.ActualPort())
	require.Equal(t, "https://localhost:9090", s.AdminBaseURL())
	require.Nil(t, s.AdminTLSRootCAPool())

	require.NotPanics(t, func() { s.SetReady(true) })
	require.NotPanics(t, func() { s.SetReady(false) })
}

func TestNewTestApplication(t *testing.T) {
	t.Parallel()

	app := cryptoutilTestingStubs.NewTestApplication(t)
	require.NotNil(t, app)

	require.Equal(t, cryptoutilTestingStubs.StubPublicPortValue, app.PublicPort())
	require.Equal(t, cryptoutilTestingStubs.StubAdminPortValue, app.AdminPort())
	require.Equal(t, "https://localhost:8443", app.PublicBaseURL())
	require.Equal(t, "https://localhost:9090", app.AdminBaseURL())
}
