// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps-framework/service/server"
	cryptoutilAppsIdentityAuthzServerConfig "cryptoutil/internal/apps/identity-authz/server/config"
)

type authzServerErrorPathPublicServer struct {
	cryptoutilAppsFrameworkServiceServer.IPublicServer
	shutdownErr error
}

func (s authzServerErrorPathPublicServer) Start(context.Context) error {
	return nil
}

func (s authzServerErrorPathPublicServer) Shutdown(context.Context) error {
	return s.shutdownErr
}

type authzServerErrorPathAdminServer struct {
	cryptoutilAppsFrameworkServiceServer.IAdminServer
}

func (authzServerErrorPathAdminServer) Start(context.Context) error {
	return nil
}

func (authzServerErrorPathAdminServer) Shutdown(context.Context) error {
	return nil
}

func TestAuthzServer_NewFromConfigBuildFailure(t *testing.T) {
	t.Parallel()

	cfg := cryptoutilAppsIdentityAuthzServerConfig.DefaultTestConfig()
	cfg.TLSPrivateProvisionMode = cryptoutilAppsFrameworkServiceConfig.TLSProvisionModeStatic
	cfg.TLSStaticCertPEM = []byte("invalid certificate")
	cfg.TLSStaticKeyPEM = []byte("invalid key")

	_, err := NewFromConfig(context.Background(), cfg)
	require.ErrorContains(t, err, "failed to build identity-authz service")
}

func TestAuthzServer_ShutdownError(t *testing.T) {
	t.Parallel()

	application, err := cryptoutilAppsFrameworkServiceServer.NewApplication(
		context.Background(),
		authzServerErrorPathPublicServer{shutdownErr: errors.New("injected shutdown failure")},
		authzServerErrorPathAdminServer{},
	)
	require.NoError(t, err)

	server := &AuthzServer{app: application}
	require.ErrorContains(t, server.Shutdown(context.Background()), "failed to shutdown application")
}

func TestAuthzServer_ShutdownInvokesCallbacks(t *testing.T) {
	t.Parallel()

	application, err := cryptoutilAppsFrameworkServiceServer.NewApplication(
		context.Background(),
		authzServerErrorPathPublicServer{},
		authzServerErrorPathAdminServer{},
	)
	require.NoError(t, err)

	coreShutdownCalled := false
	containerShutdownCalled := false
	server := &AuthzServer{
		app: application,
		shutdownCore: func() {
			coreShutdownCalled = true
		},
		shutdownContainer: func() {
			containerShutdownCalled = true
		},
	}

	require.NoError(t, server.Shutdown(context.Background()))
	require.True(t, coreShutdownCalled)
	require.True(t, containerShutdownCalled)
}

func TestAuthzServer_StartContextCanceled(t *testing.T) {
	t.Parallel()

	application, err := cryptoutilAppsFrameworkServiceServer.NewApplication(
		context.Background(),
		authzServerErrorPathPublicServer{},
		authzServerErrorPathAdminServer{},
	)
	require.NoError(t, err)

	server := &AuthzServer{app: application}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = server.Start(ctx)
	require.ErrorContains(t, err, "failed to start application")
	require.ErrorContains(t, err, "application startup cancelled")
}
