// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps-framework/service/server"
	cryptoutilAppsIdentityIdpServerConfig "cryptoutil/internal/apps/identity-idp/server/config"
)

type idpServerErrorPathPublicServer struct {
	cryptoutilAppsFrameworkServiceServer.IPublicServer
	shutdownErr error
}

func (s idpServerErrorPathPublicServer) Start(context.Context) error {
	return nil
}

func (s idpServerErrorPathPublicServer) Shutdown(context.Context) error {
	return s.shutdownErr
}

type idpServerErrorPathAdminServer struct {
	cryptoutilAppsFrameworkServiceServer.IAdminServer
}

func (idpServerErrorPathAdminServer) Start(context.Context) error {
	return nil
}

func (idpServerErrorPathAdminServer) Shutdown(context.Context) error {
	return nil
}

func TestIDPServer_NewFromConfigBuildFailure(t *testing.T) {
	t.Parallel()

	cfg := cryptoutilAppsIdentityIdpServerConfig.DefaultTestConfig()
	cfg.TLSPrivateProvisionMode = cryptoutilAppsFrameworkServiceConfig.TLSProvisionModeStatic
	cfg.TLSStaticCertPEM = []byte("invalid certificate")
	cfg.TLSStaticKeyPEM = []byte("invalid key")

	_, err := NewFromConfig(context.Background(), cfg)
	require.ErrorContains(t, err, "failed to build identity-idp service")
}

func TestIDPServer_ShutdownError(t *testing.T) {
	t.Parallel()

	application, err := cryptoutilAppsFrameworkServiceServer.NewApplication(
		context.Background(),
		idpServerErrorPathPublicServer{shutdownErr: errors.New("injected shutdown failure")},
		idpServerErrorPathAdminServer{},
	)
	require.NoError(t, err)

	server := &IDPServer{app: application}
	require.ErrorContains(t, server.Shutdown(context.Background()), "failed to shutdown application")
}

func TestIDPServer_ShutdownInvokesCallbacks(t *testing.T) {
	t.Parallel()

	application, err := cryptoutilAppsFrameworkServiceServer.NewApplication(
		context.Background(),
		idpServerErrorPathPublicServer{},
		idpServerErrorPathAdminServer{},
	)
	require.NoError(t, err)

	coreShutdownCalled := false
	containerShutdownCalled := false
	server := &IDPServer{
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

func TestIDPServer_StartContextCanceled(t *testing.T) {
	t.Parallel()

	application, err := cryptoutilAppsFrameworkServiceServer.NewApplication(
		context.Background(),
		idpServerErrorPathPublicServer{},
		idpServerErrorPathAdminServer{},
	)
	require.NoError(t, err)

	server := &IDPServer{app: application}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = server.Start(ctx)
	require.ErrorContains(t, err, "failed to start application")
	require.ErrorContains(t, err, "application startup cancelled")
}
