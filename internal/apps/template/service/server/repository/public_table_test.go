// Copyright (c) 2025 Justin Cranford

//nolint:testpackage // Testing private fields requires same-package access.
package repository_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	http "net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilAppsTemplateServiceServerListener "cryptoutil/internal/apps/template/service/server/listener"
	cryptoutilAppsTemplateServiceServerTestutil "cryptoutil/internal/apps/template/service/server/testutil"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestPublicHTTPServer_TableDriven_HappyPath tests successful public server operations.
func TestPublicHTTPServer_TableDriven_HappyPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		description string
		testFunc    func(t *testing.T, server cryptoutilAppsTemplateServiceServer.IPublicServer)
	}{
		{
			name:        "NewPublicHTTPServer",
			description: "Verify successful public server creation",
			testFunc: func(t *testing.T, _ cryptoutilAppsTemplateServiceServer.IPublicServer) {
				t.Helper()

				tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PublicTLS()

				server, err := cryptoutilAppsTemplateServiceServerListener.NewPublicHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
				require.NoError(t, err)
				require.NotNil(t, server)
			},
		},
		{
			name:        "Start",
			description: "Verify public server starts and listens on dynamic port",
			testFunc: func(t *testing.T, server cryptoutilAppsTemplateServiceServer.IPublicServer) {
				t.Helper()

				var wg sync.WaitGroup
				wg.Add(1)

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				go func() {
					defer wg.Done()

					_ = server.Start(ctx)
				}()

				// Wait for server to be ready.
				time.Sleep(200 * time.Millisecond)

				port := server.ActualPort()
				require.Greater(t, port, 0, "Expected dynamic port allocation")

				// Shutdown.
				shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer shutdownCancel()

				err := server.Shutdown(shutdownCtx)
				require.NoError(t, err)

				wg.Wait()
			},
		},
		{
			name:        "ServiceHealth_Healthy",
			description: "Service health endpoint returns 200 when server is healthy",
			testFunc: func(t *testing.T, server cryptoutilAppsTemplateServiceServer.IPublicServer) {
				t.Helper()

				var wg sync.WaitGroup
				wg.Add(1)

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				go func() {
					defer wg.Done()

					_ = server.Start(ctx)
				}()

				// Wait for server to be ready.
				time.Sleep(200 * time.Millisecond)

				port := server.ActualPort()
				require.Greater(t, port, 0)

				// Query service health endpoint.
				client := &http.Client{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: true, //nolint:gosec // Test uses self-signed cert.
						},
					},
					Timeout: 5 * time.Second,
				}

				url := fmt.Sprintf("https://%s:%d/service/api/v1/health", cryptoutilSharedMagic.IPv4Loopback, port)
				req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
				require.NoError(t, err)
				resp, err := client.Do(req)
				require.NoError(t, err)

				defer func() { _ = resp.Body.Close() }()

				require.Equal(t, http.StatusOK, resp.StatusCode)

				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				require.Contains(t, string(body), "healthy")

				// Shutdown.
				shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer shutdownCancel()

				err = server.Shutdown(shutdownCtx)
				require.NoError(t, err)

				wg.Wait()
			},
		},
		{
			name:        "BrowserHealth_Healthy",
			description: "Browser health endpoint returns 200 when server is healthy",
			testFunc: func(t *testing.T, server cryptoutilAppsTemplateServiceServer.IPublicServer) {
				t.Helper()

				var wg sync.WaitGroup
				wg.Add(1)

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				go func() {
					defer wg.Done()

					_ = server.Start(ctx)
				}()

				// Wait for server to be ready.
				time.Sleep(200 * time.Millisecond)

				port := server.ActualPort()
				require.Greater(t, port, 0)

				// Query browser health endpoint.
				client := &http.Client{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: true, //nolint:gosec // Test uses self-signed cert.
						},
					},
					Timeout: 5 * time.Second,
				}

				url := fmt.Sprintf("https://%s:%d/browser/api/v1/health", cryptoutilSharedMagic.IPv4Loopback, port)
				req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
				require.NoError(t, err)
				resp, err := client.Do(req)
				require.NoError(t, err)

				defer func() { _ = resp.Body.Close() }()

				require.Equal(t, http.StatusOK, resp.StatusCode)

				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				require.Contains(t, string(body), "healthy")

				// Shutdown.
				shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer shutdownCancel()

				err = server.Shutdown(shutdownCtx)
				require.NoError(t, err)

				wg.Wait()
			},
		},
		{
			name:        "Shutdown_Graceful",
			description: "Shutdown gracefully stops server and waits for connections to drain",
			testFunc: func(t *testing.T, server cryptoutilAppsTemplateServiceServer.IPublicServer) {
				t.Helper()

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				var wg sync.WaitGroup
				wg.Add(1)

				go func() {
					defer wg.Done()

					_ = server.Start(ctx)
				}()

				// Wait for server to be ready.
				time.Sleep(200 * time.Millisecond)

				// Shutdown.
				shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer shutdownCancel()

				err := server.Shutdown(shutdownCtx)
				require.NoError(t, err)

				wg.Wait()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create fresh server for each test.
			tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PublicTLS()
			server, err := cryptoutilAppsTemplateServiceServerListener.NewPublicHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
			require.NoError(t, err)

			// Run test function.
			tt.testFunc(t, server)
		})
	}
}

// TestPublicHTTPServer_TableDriven_SadPath tests error conditions and edge cases.
func TestPublicHTTPServer_TableDriven_SadPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		description   string
		setupFunc     func(t *testing.T) (cryptoutilAppsTemplateServiceServer.IPublicServer, error)
		expectedError string
	}{
		{
			name:        "NilContext",
			description: "NewPublicHTTPServer rejects nil context",
			setupFunc: func(t *testing.T) (cryptoutilAppsTemplateServiceServer.IPublicServer, error) {
				t.Helper()

				tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PublicTLS()

				return cryptoutilAppsTemplateServiceServerListener.NewPublicHTTPServer(nil, cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg) //nolint:staticcheck // Testing nil context.
			},
			expectedError: "context cannot be nil",
		},
		{
			name:        "Start_NilContext",
			description: "Start rejects nil context",
			setupFunc: func(t *testing.T) (cryptoutilAppsTemplateServiceServer.IPublicServer, error) {
				t.Helper()

				tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PublicTLS()

				server, err := cryptoutilAppsTemplateServiceServerListener.NewPublicHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
				require.NoError(t, err)

				err = server.Start(nil) //nolint:staticcheck // Testing nil context.

				return server, fmt.Errorf("failed to start server: %w", err)
			},
			expectedError: "context cannot be nil",
		},
		{
			name:        "Shutdown_NilContext",
			description: "Shutdown accepts nil context and uses Background()",
			setupFunc: func(t *testing.T) (cryptoutilAppsTemplateServiceServer.IPublicServer, error) {
				t.Helper()

				tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PublicTLS()

				server, err := cryptoutilAppsTemplateServiceServerListener.NewPublicHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
				require.NoError(t, err)

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				var wg sync.WaitGroup
				wg.Add(1)

				go func() {
					defer wg.Done()

					_ = server.Start(ctx)
				}()

				// Wait for server to start.
				time.Sleep(200 * time.Millisecond)

				// Shutdown with nil context (should NOT error).
				err = server.Shutdown(nil) //nolint:staticcheck // Testing nil context.

				wg.Wait()

				if err != nil {
					return server, fmt.Errorf("failed to shutdown server: %w", err)
				}

				return server, nil
			},
			expectedError: "",
		},
		{
			name:        "ActualPort_BeforeStart",
			description: "ActualPort returns 0 before server starts",
			setupFunc: func(t *testing.T) (cryptoutilAppsTemplateServiceServer.IPublicServer, error) {
				t.Helper()

				tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PublicTLS()

				server, err := cryptoutilAppsTemplateServiceServerListener.NewPublicHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
				require.NoError(t, err)

				port := server.ActualPort()
				if port != 0 {
					return server, fmt.Errorf("expected port 0, got %d", port)
				}

				return server, nil
			},
			expectedError: "",
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

