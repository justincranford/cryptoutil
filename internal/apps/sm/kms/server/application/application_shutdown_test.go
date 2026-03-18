//go:build integration
// +build integration

// Copyright (c) 2025 Justin Cranford
//
// NOTE: These tests require a PostgreSQL database and are skipped in CI without the integration tag.
//

package application

import (
	"context"
	"testing"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"

	"github.com/stretchr/testify/require"
)

func TestServerApplicationBasic_Shutdown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) *ServerApplicationBasic
		expectPanic bool
	}{
		{
			name: "Shutdown_AllComponents_Success",
			setupFunc: func(t *testing.T) *ServerApplicationBasic {
				t.Helper()

				settings := cryptoutilAppsFrameworkServiceConfig.RequireNewForTest("shutdown_test_basic")
				ctx := context.Background()
				app, err := StartServerApplicationBasic(ctx, settings)
				require.NoError(t, err, "failed to start server application basic")
				require.NotNil(t, app, "server application basic should not be nil")

				return app
			},
			expectPanic: false,
		},
		{
			name: "Shutdown_NilComponents_NoP anic",
			setupFunc: func(t *testing.T) *ServerApplicationBasic {
				t.Helper()

				return &ServerApplicationBasic{}
			},
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := tt.setupFunc(t)

			shutdownFunc := app.Shutdown()
			require.NotNil(t, shutdownFunc, "shutdown function should not be nil")

			if tt.expectPanic {
				require.Panics(t, shutdownFunc, "expected panic during shutdown")
			} else {
				require.NotPanics(t, shutdownFunc, "shutdown should not panic")
			}

			t.Logf("âœ“ Shutdown test passed: %s", tt.name)
		})
	}
}
