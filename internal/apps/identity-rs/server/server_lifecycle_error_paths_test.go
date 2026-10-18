// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps-framework/service/server"
	cryptoutilAppsFrameworkTestingStubs "cryptoutil/internal/apps-framework/service/testing/stubs"
	cryptoutilAppsIdentityRsServerConfig "cryptoutil/internal/apps/identity-rs/server/config"
)

func TestRSServer_NewFromConfigBuilderFailure(t *testing.T) {
	t.Parallel()

	cfg := cryptoutilAppsIdentityRsServerConfig.DefaultTestConfig()
	cfg.ServiceFrameworkServerSettings = nil

	_, err := NewFromConfig(context.Background(), cfg)

	require.ErrorContains(t, err, "failed to build identity-rs service")
}

func TestRSServer_StartRejectsNilContext(t *testing.T) {
	t.Parallel()

	app, err := cryptoutilAppsFrameworkServiceServer.NewApplication(
		context.Background(),
		&cryptoutilAppsFrameworkTestingStubs.StubPublicServer{},
		&cryptoutilAppsFrameworkTestingStubs.StubAdminServer{},
	)
	require.NoError(t, err)

	var nilContext context.Context
	require.ErrorContains(t, (&RSServer{app: app}).Start(nilContext), "failed to start application")
}

func TestRSServer_ShutdownErrorHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		publicError   error
		adminError    error
		wantError     string
		wantCallbacks bool
	}{
		{
			name:        "public shutdown failure",
			publicError: errors.New("injected public shutdown failure"),
			wantError:   testRSShutdownError,
		},
		{
			name:       "admin shutdown failure",
			adminError: errors.New("injected admin shutdown failure"),
			wantError:  testRSShutdownError,
		},
		{name: "both shutdown callbacks run", wantCallbacks: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app, err := cryptoutilAppsFrameworkServiceServer.NewApplication(
				context.Background(),
				&cryptoutilAppsFrameworkTestingStubs.StubPublicServer{ShutdownErr: tc.publicError},
				&cryptoutilAppsFrameworkTestingStubs.StubAdminServer{ShutdownErr: tc.adminError},
			)
			require.NoError(t, err)

			shutdownCoreCalled := false
			shutdownContainerCalled := false
			server := &RSServer{
				app: app,
				shutdownCore: func() {
					shutdownCoreCalled = true
				},
				shutdownContainer: func() {
					shutdownContainerCalled = true
				},
			}

			err = server.Shutdown(context.Background())
			if tc.wantCallbacks {
				require.NoError(t, err)
				require.True(t, shutdownCoreCalled)
				require.True(t, shutdownContainerCalled)

				return
			}

			require.ErrorContains(t, err, tc.wantError)
			require.False(t, shutdownCoreCalled)
			require.False(t, shutdownContainerCalled)
		})
	}
}
