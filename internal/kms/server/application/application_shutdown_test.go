// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"context"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"

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

				settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("shutdown_test_basic")
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

			t.Logf("✓ Shutdown test passed: %s", tt.name)
		})
	}
}

func TestServerApplicationCore_Shutdown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupFunc   func(t *testing.T) *ServerApplicationCore
		expectPanic bool
	}{
		{
			name: "Shutdown_AllComponents_Success",
			setupFunc: func(t *testing.T) *ServerApplicationCore {
				t.Helper()

				settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("shutdown_test_core")
				ctx := context.Background()

				// Start core (which starts basic internally)
				core, err := StartServerApplicationCore(ctx, settings)
				require.NoError(t, err, "failed to start server application core")
				require.NotNil(t, core, "server application core should not be nil")

				return core
			},
			expectPanic: false,
		},
		{
			name: "Shutdown_NilComponents_NoPanic",
			setupFunc: func(t *testing.T) *ServerApplicationCore {
				t.Helper()

				return &ServerApplicationCore{}
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

			t.Logf("✓ Shutdown test passed: %s", tt.name)
		})
	}
}

func TestSendServerListenerShutdownRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		settingsFunc  func(t *testing.T) *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
		expectError   bool
		errorContains string
	}{
		{
			name: "Shutdown_InvalidURL_Error",
			settingsFunc: func(t *testing.T) *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings {
				t.Helper()

				settings := cryptoutilAppsTemplateServiceConfig.RequireNewForTest("shutdown_request_invalid")
				settings.BindPrivateAddress = "invalid-url-that-does-not-exist"
				settings.BindPrivatePort = 9999

				return settings
			},
			expectError:   true,
			errorContains: "failed to send shutdown request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			settings := tt.settingsFunc(t)
			err := SendServerListenerShutdownRequest(settings)

			if tt.expectError {
				require.Error(t, err, "expected error for invalid shutdown request")
				require.Contains(t, err.Error(), tt.errorContains, "error message should match expected")
			} else {
				require.NoError(t, err, "shutdown request should succeed")
			}

			t.Logf("✓ Shutdown request test passed: %s", tt.name)
		})
	}
}
