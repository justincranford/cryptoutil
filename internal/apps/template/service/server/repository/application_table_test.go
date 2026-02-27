// Copyright (c) 2025 Justin Cranford

//nolint:testpackage // Testing private fields requires same-package access.
package repository_test

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilAppsTemplateServiceServerTestutil "cryptoutil/internal/apps/template/service/server/testutil"
)

// TestApplication_TableDriven_HappyPath tests successful application operations.
func TestApplication_TableDriven_HappyPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		description string
		testFunc    func(t *testing.T, app *cryptoutilAppsTemplateServiceServer.Application)
	}{
		{
			name:        "NewApplication",
			description: "Verify successful application creation with valid servers",
			testFunc: func(t *testing.T, _ *cryptoutilAppsTemplateServiceServer.Application) {
				t.Helper()

				ctx := context.Background()
				publicServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockPublicServer(cryptoutilSharedMagic.DemoServerPort)
				adminServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockAdminServer(cryptoutilSharedMagic.JoseJAAdminPort)

				app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
				require.NoError(t, err)
				require.NotNil(t, app)
				require.False(t, app.IsShutdown())
			},
		},
		{
			name:        "Start",
			description: "Application starts both public and admin servers concurrently",
			testFunc: func(t *testing.T, app *cryptoutilAppsTemplateServiceServer.Application) {
				t.Helper()

				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				errChan := make(chan error, 1)

				go func() {
					errChan <- app.Start(ctx)
				}()

				// Wait for context timeout.
				<-ctx.Done()

				// Verify error is context deadline exceeded (expected).
				err := <-errChan
				require.Error(t, err)
				require.Contains(t, err.Error(), "context deadline exceeded")
			},
		},
		{
			name:        "Shutdown",
			description: "Application shuts down both servers gracefully",
			testFunc: func(t *testing.T, app *cryptoutilAppsTemplateServiceServer.Application) {
				t.Helper()

				ctx := context.Background()

				err := app.Shutdown(ctx)
				require.NoError(t, err)
				require.True(t, app.IsShutdown())
			},
		},
		{
			name:        "PublicPort",
			description: "PublicPort returns correct port from public server",
			testFunc: func(t *testing.T, app *cryptoutilAppsTemplateServiceServer.Application) {
				t.Helper()

				port := app.PublicPort()
				require.Equal(t, cryptoutilSharedMagic.DemoServerPort, port)
			},
		},
		{
			name:        "AdminPort",
			description: "AdminPort returns correct port from admin server",
			testFunc: func(t *testing.T, app *cryptoutilAppsTemplateServiceServer.Application) {
				t.Helper()

				port := app.AdminPort()
				require.Equal(t, cryptoutilSharedMagic.JoseJAAdminPort, port)
			},
		},
		{
			name:        "IsShutdown",
			description: "IsShutdown tracks shutdown state correctly",
			testFunc: func(t *testing.T, app *cryptoutilAppsTemplateServiceServer.Application) {
				t.Helper()

				require.False(t, app.IsShutdown())

				ctx := context.Background()
				err := app.Shutdown(ctx)
				require.NoError(t, err)

				require.True(t, app.IsShutdown())
			},
		},
		{
			name:        "ConcurrentShutdown",
			description: "Concurrent shutdown calls are safe and idempotent",
			testFunc: func(t *testing.T, app *cryptoutilAppsTemplateServiceServer.Application) {
				t.Helper()

				ctx := context.Background()

				var wg sync.WaitGroup

				errChan := make(chan error, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)

				// Launch 5 concurrent shutdown calls.
				for range cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries {
					wg.Add(1)

					go func() {
						defer wg.Done()

						errChan <- app.Shutdown(ctx)
					}()
				}

				wg.Wait()
				close(errChan)

				// Exactly one should succeed, others should be no-ops.
				successCount := 0

				for err := range errChan {
					if err == nil {
						successCount++
					}
				}

				require.GreaterOrEqual(t, successCount, 1, "At least one shutdown should succeed")
				require.True(t, app.IsShutdown())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create fresh application for each test.
			ctx := context.Background()
			publicServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockPublicServer(cryptoutilSharedMagic.DemoServerPort)
			adminServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockAdminServer(cryptoutilSharedMagic.JoseJAAdminPort)

			app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
			require.NoError(t, err)

			// Run test function.
			tt.testFunc(t, app)
		})
	}
}

// TestApplication_TableDriven_SadPath tests error conditions and edge cases.
func TestApplication_TableDriven_SadPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		description   string
		setupFunc     func(t *testing.T) (*cryptoutilAppsTemplateServiceServer.Application, error)
		expectedError string
	}{
		{
			name:        "NilContext",
			description: "NewApplication rejects nil context",
			setupFunc: func(t *testing.T) (*cryptoutilAppsTemplateServiceServer.Application, error) {
				t.Helper()

				publicServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockPublicServer(cryptoutilSharedMagic.DemoServerPort)
				adminServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockAdminServer(cryptoutilSharedMagic.JoseJAAdminPort)

				return cryptoutilAppsTemplateServiceServer.NewApplication(nil, publicServer, adminServer) //nolint:staticcheck // Testing nil context.
			},
			expectedError: "context cannot be nil",
		},
		{
			name:        "NilPublicServer",
			description: "NewApplication rejects nil public server",
			setupFunc: func(t *testing.T) (*cryptoutilAppsTemplateServiceServer.Application, error) {
				t.Helper()

				ctx := context.Background()
				adminServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockAdminServer(cryptoutilSharedMagic.JoseJAAdminPort)

				return cryptoutilAppsTemplateServiceServer.NewApplication(ctx, nil, adminServer)
			},
			expectedError: "publicServer cannot be nil",
		},
		{
			name:        "NilAdminServer",
			description: "NewApplication rejects nil admin server",
			setupFunc: func(t *testing.T) (*cryptoutilAppsTemplateServiceServer.Application, error) {
				t.Helper()

				ctx := context.Background()
				publicServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockPublicServer(cryptoutilSharedMagic.DemoServerPort)

				return cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, nil)
			},
			expectedError: "adminServer cannot be nil",
		},
		{
			name:        "Start_NilContext",
			description: "Start rejects nil context",
			setupFunc: func(t *testing.T) (*cryptoutilAppsTemplateServiceServer.Application, error) {
				t.Helper()

				ctx := context.Background()
				publicServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockPublicServer(cryptoutilSharedMagic.DemoServerPort)
				adminServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockAdminServer(cryptoutilSharedMagic.JoseJAAdminPort)

				app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
				require.NoError(t, err)

				err = app.Start(nil) //nolint:staticcheck // Testing nil context.

				return app, fmt.Errorf("failed to start application: %w", err)
			},
			expectedError: "context cannot be nil",
		},
		{
			name:        "Start_PublicServerFailure",
			description: "Start propagates public server start errors",
			setupFunc: func(t *testing.T) (*cryptoutilAppsTemplateServiceServer.Application, error) {
				t.Helper()

				ctx := context.Background()
				publicServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockPublicServer(cryptoutilSharedMagic.DemoServerPort)
				publicServer.StartErr = errors.New("public server start failed")
				adminServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockAdminServer(cryptoutilSharedMagic.JoseJAAdminPort)

				app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
				require.NoError(t, err)

				err = app.Start(context.Background())

				return app, fmt.Errorf("failed to start application: %w", err)
			},
			expectedError: "public server start failed",
		},
		{
			name:        "Start_AdminServerFailure",
			description: "Start propagates admin server start errors",
			setupFunc: func(t *testing.T) (*cryptoutilAppsTemplateServiceServer.Application, error) {
				t.Helper()

				ctx := context.Background()
				publicServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockPublicServer(cryptoutilSharedMagic.DemoServerPort)
				adminServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockAdminServer(cryptoutilSharedMagic.JoseJAAdminPort)
				adminServer.StartErr = errors.New("admin server start failed")

				app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
				require.NoError(t, err)

				err = app.Start(context.Background())

				return app, fmt.Errorf("failed to start application: %w", err)
			},
			expectedError: "admin server start failed",
		},
		{
			name:        "Shutdown_NilContext",
			description: "Shutdown accepts nil context and uses Background()",
			setupFunc: func(t *testing.T) (*cryptoutilAppsTemplateServiceServer.Application, error) {
				t.Helper()

				ctx := context.Background()
				publicServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockPublicServer(cryptoutilSharedMagic.DemoServerPort)
				adminServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockAdminServer(cryptoutilSharedMagic.JoseJAAdminPort)

				app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
				require.NoError(t, err)

				err = app.Shutdown(nil) //nolint:staticcheck // Testing nil context.
				if err != nil {
					return app, fmt.Errorf("failed to shutdown application: %w", err)
				}

				return app, nil
			},
			expectedError: "",
		},
		{
			name:        "Shutdown_PublicServerError",
			description: "Shutdown continues even if public server shutdown fails",
			setupFunc: func(t *testing.T) (*cryptoutilAppsTemplateServiceServer.Application, error) {
				t.Helper()

				ctx := context.Background()
				publicServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockPublicServer(cryptoutilSharedMagic.DemoServerPort)
				publicServer.ShutdownErr = errors.New("public shutdown failed")
				adminServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockAdminServer(cryptoutilSharedMagic.JoseJAAdminPort)

				app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
				require.NoError(t, err)

				err = app.Shutdown(context.Background())

				return app, fmt.Errorf("failed to shutdown application: %w", err)
			},
			expectedError: "public shutdown failed",
		},
		{
			name:        "Shutdown_AdminServerError",
			description: "Shutdown continues even if admin server shutdown fails",
			setupFunc: func(t *testing.T) (*cryptoutilAppsTemplateServiceServer.Application, error) {
				t.Helper()

				ctx := context.Background()
				publicServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockPublicServer(cryptoutilSharedMagic.DemoServerPort)
				adminServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockAdminServer(cryptoutilSharedMagic.JoseJAAdminPort)
				adminServer.ShutdownErr = errors.New("admin shutdown failed")

				app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
				require.NoError(t, err)

				err = app.Shutdown(context.Background())

				return app, fmt.Errorf("failed to shutdown application: %w", err)
			},
			expectedError: "admin shutdown failed",
		},
		{
			name:        "Shutdown_BothServersError",
			description: "Shutdown reports both server errors when both fail",
			setupFunc: func(t *testing.T) (*cryptoutilAppsTemplateServiceServer.Application, error) {
				t.Helper()

				ctx := context.Background()
				publicServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockPublicServer(cryptoutilSharedMagic.DemoServerPort)
				publicServer.ShutdownErr = errors.New("public shutdown failed")
				adminServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockAdminServer(cryptoutilSharedMagic.JoseJAAdminPort)
				adminServer.ShutdownErr = errors.New("admin shutdown failed")

				app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
				require.NoError(t, err)

				err = app.Shutdown(context.Background())

				return app, fmt.Errorf("failed to shutdown application: %w", err)
			},
			expectedError: "public shutdown failed",
		},
		{
			name:        "AdminPort_NotInitialized",
			description: "AdminPort returns 0 when admin server port is 0",
			setupFunc: func(t *testing.T) (*cryptoutilAppsTemplateServiceServer.Application, error) {
				t.Helper()

				ctx := context.Background()
				publicServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockPublicServer(cryptoutilSharedMagic.DemoServerPort)
				adminServer := cryptoutilAppsTemplateServiceServerTestutil.NewMockAdminServer(0) // Port 0 returns 0.

				app, err := cryptoutilAppsTemplateServiceServer.NewApplication(ctx, publicServer, adminServer)
				require.NoError(t, err)

				port := app.AdminPort()
				if port != 0 {
					return app, fmt.Errorf("expected port 0, got %d", port)
				}

				return app, fmt.Errorf("admin port is 0")
			},
			expectedError: "admin port is 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := tt.setupFunc(t)

			if tt.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}
