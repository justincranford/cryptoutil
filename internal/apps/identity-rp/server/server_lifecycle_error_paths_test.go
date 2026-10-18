// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps-framework/service/server"
	cryptoutilAppsFrameworkTestingStubs "cryptoutil/internal/apps-framework/service/testing/stubs"
)

func TestRPServer_LifecycleErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		publicStartError error
		publicStopError  error
		adminStopError   error
		wantError        string
	}{
		{
			name:             testRPStartFailureCaseName,
			publicStartError: errors.New("injected public start failure"),
			wantError:        "failed to start application",
		},
		{
			name:            "public shutdown failure",
			publicStopError: errors.New("injected public shutdown failure"),
			wantError:       testRPShutdownError,
		},
		{
			name:           "admin shutdown failure",
			adminStopError: errors.New("injected admin shutdown failure"),
			wantError:      testRPShutdownError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			publicServer := &cryptoutilAppsFrameworkTestingStubs.StubPublicServer{
				StartErr:    tc.publicStartError,
				ShutdownErr: tc.publicStopError,
			}
			adminServer := &cryptoutilAppsFrameworkTestingStubs.StubAdminServer{ShutdownErr: tc.adminStopError}
			app, err := cryptoutilAppsFrameworkServiceServer.NewApplication(
				context.Background(),
				publicServer,
				adminServer,
			)
			require.NoError(t, err)

			server := &RPServer{app: app}

			if tc.name != testRPStartFailureCaseName {
				shutdownCoreCalled := false
				shutdownContainerCalled := false
				server.shutdownCore = func() {
					shutdownCoreCalled = true
				}
				server.shutdownContainer = func() {
					shutdownContainerCalled = true
				}

				require.ErrorContains(t, server.Shutdown(context.Background()), tc.wantError)
				require.False(t, shutdownCoreCalled)
				require.False(t, shutdownContainerCalled)

				return
			}

			require.ErrorContains(t, server.Start(context.Background()), tc.wantError)
		})
	}

	t.Run("successful shutdown callbacks", func(t *testing.T) {
		t.Parallel()

		app, err := cryptoutilAppsFrameworkServiceServer.NewApplication(
			context.Background(),
			&cryptoutilAppsFrameworkTestingStubs.StubPublicServer{},
			&cryptoutilAppsFrameworkTestingStubs.StubAdminServer{},
		)
		require.NoError(t, err)

		shutdownCoreCalled := false
		shutdownContainerCalled := false
		server := &RPServer{
			app: app,
			shutdownCore: func() {
				shutdownCoreCalled = true
			},
			shutdownContainer: func() {
				shutdownContainerCalled = true
			},
		}

		require.NoError(t, server.Shutdown(context.Background()))
		require.True(t, shutdownCoreCalled)
		require.True(t, shutdownContainerCalled)
	})
}
