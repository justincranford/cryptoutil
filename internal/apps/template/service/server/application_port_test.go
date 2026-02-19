// Copyright (c) 2025 Justin Cranford
//
//

package server_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
)

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

// TestApplication_PublicServerBase_MockServer tests PublicServerBase when public server is a mock (not PublicServerBase).
func TestApplication_PublicServerBase_MockServer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080, "https://localhost:8080")
	adminServer := newMockAdminServer(9090, "https://localhost:9090")

	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	// When public server is NOT *PublicServerBase, should return nil.
	base := app.PublicServerBase()
	require.Nil(t, base, "Expected nil when public server is not *PublicServerBase")

	// Verify the type assertion actually failed.
	// The mock is definitely not a PublicServerBase from package server.
	// This test covers application.go:294 (return nil path).
}

// TestApplication_PublicServerBase_RealServer tests PublicServerBase when public server is a real PublicServerBase.
func TestApplication_PublicServerBase_RealServer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create a real PublicServerBase.
	publicServer, err := cryptoutilAppsTemplateServiceServer.NewPublicServerBase(&cryptoutilAppsTemplateServiceServer.PublicServerConfig{
		BindAddress: "127.0.0.1",
		Port:        0, // Dynamic allocation.
		TLSMaterial: createTestTLSMaterial(t),
	})
	require.NoError(t, err)

	adminServer := newMockAdminServer(9090, "https://localhost:9090")

	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	// When public server IS *PublicServerBase, should return it.
	base := app.PublicServerBase()
	require.NotNil(t, base, "Expected non-nil when public server is *PublicServerBase")
	require.Same(t, publicServer, base, "Expected same instance")

	// This test covers application.go:290 (return base path).
}

// TestApplication_IsShutdown tests IsShutdown method.
func TestApplication_IsShutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	publicServer := newMockPublicServer(8080, "https://localhost:8080")
	adminServer := newMockAdminServer(9090, "https://localhost:9090")

	app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
	require.NoError(t, err)

	// Initially not shutdown.
	require.False(t, app.IsShutdown())

	// After shutdown.
	err = app.Shutdown(ctx)
	require.NoError(t, err)
	require.True(t, app.IsShutdown())
}
