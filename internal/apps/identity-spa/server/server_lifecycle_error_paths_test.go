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

func TestSPAServer_LifecycleErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		publicStartError error
		publicStopError  error
		adminStopError   error
		wantError        string
	}{
		{
			name:             testSPAStartFailureCaseName,
			publicStartError: errors.New("injected public start failure"),
			wantError:        "failed to start SPA server",
		},
		{
			name:            "public shutdown failure",
			publicStopError: errors.New("injected public shutdown failure"),
			wantError:       testSPAShutdownErrors,
		},
		{
			name:           "admin shutdown failure",
			adminStopError: errors.New("injected admin shutdown failure"),
			wantError:      testSPAShutdownErrors,
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

			server := &SPAServer{app: app}

			if tc.name != testSPAStartFailureCaseName {
				shutdownCoreCalled := false
				server.shutdownCore = func() {
					shutdownCoreCalled = true
				}

				require.ErrorContains(t, server.Shutdown(context.Background()), tc.wantError)
				require.True(t, shutdownCoreCalled)

				return
			}

			require.ErrorContains(t, server.Start(context.Background()), tc.wantError)
		})
	}
}
