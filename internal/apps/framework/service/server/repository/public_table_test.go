// Copyright (c) 2025 Justin Cranford

//nolint:testpackage // Testing private fields requires same-package access.
package repository_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	http "net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps/framework/service/server"
	cryptoutilAppsFrameworkServiceServerListener "cryptoutil/internal/apps/framework/service/server/listener"
	cryptoutilAppsFrameworkServiceServerTestutil "cryptoutil/internal/apps/framework/service/server/testutil"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestPublicHTTPServer_TableDriven_HappyPath tests successful public server operations.
func TestPublicHTTPServer_TableDriven_HappyPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		description string
		testFunc    func(t *testing.T, server cryptoutilAppsFrameworkServiceServer.IPublicServer)
	}{
		{
			name:        "NewPublicHTTPServer",
			description: "Verify successful public server creation",
			testFunc: func(t *testing.T, _ cryptoutilAppsFrameworkServiceServer.IPublicServer) {
				t.Helper()

				tlsCfg := cryptoutilAppsFrameworkServiceServerTestutil.PublicTLS()

				server, err := cryptoutilAppsFrameworkServiceServerListener.NewPublicHTTPServer(context.Background(), cryptoutilAppsFrameworkServiceServerTestutil.ServiceFrameworkServerSettings(), tlsCfg)
				require.NoError(t, err)
				require.NotNil(t, server)
			},
		},
		{
			name:        "Start",
			description: "Verify public server starts and listens on dynamic port",
			testFunc: func(t *testing.T, server cryptoutilAppsFrameworkServiceServer.IPublicServer) {
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
				shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
				defer shutdownCancel()

				err := server.Shutdown(shutdownCtx)
				require.NoError(t, err)

				wg.Wait()
			},
		},
		{
			name:        "ServiceHealth_Healthy",
			description: "Service health endpoint returns 200 when server is healthy",
			testFunc: func(t *testing.T, server cryptoutilAppsFrameworkServiceServer.IPublicServer) {
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
							MinVersion: tls.VersionTLS13,
							RootCAs:    cryptoutilAppsFrameworkServiceServerTestutil.PublicRootCAPool(),
						},
						DisableKeepAlives: true,
					},
				}
				url := "https://" + net.JoinHostPort(cryptoutilSharedMagic.IPv4Loopback, strconv.Itoa(port)) + cryptoutilSharedMagic.IME2EHealthEndpoint
				req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
				require.NoError(t, err)
				resp, err := client.Do(req)
				require.NoError(t, err)

				defer func() { _ = resp.Body.Close() }()

				require.Equal(t, http.StatusOK, resp.StatusCode)

				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				require.Contains(t, string(body), cryptoutilSharedMagic.DockerServiceHealthHealthy)

				// Shutdown.
				shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
				defer shutdownCancel()

				err = server.Shutdown(shutdownCtx)
				require.NoError(t, err)

				wg.Wait()
			},
		},
		{
			name:        "BrowserHealth_Healthy",
			description: "Browser health endpoint returns 200 when server is healthy",
			testFunc: func(t *testing.T, server cryptoutilAppsFrameworkServiceServer.IPublicServer) {
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
							MinVersion: tls.VersionTLS13,
							RootCAs:    cryptoutilAppsFrameworkServiceServerTestutil.PublicRootCAPool(),
						},
						DisableKeepAlives: true,
					},
				}
				url := "https://" + net.JoinHostPort(cryptoutilSharedMagic.IPv4Loopback, strconv.Itoa(port)) + "/browser/api/v1/health"
				req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
				require.NoError(t, err)
				resp, err := client.Do(req)
				require.NoError(t, err)

				defer func() { _ = resp.Body.Close() }()

				require.Equal(t, http.StatusOK, resp.StatusCode)

				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				require.Contains(t, string(body), cryptoutilSharedMagic.DockerServiceHealthHealthy)

				// Shutdown.
				shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
				defer shutdownCancel()

				err = server.Shutdown(shutdownCtx)
				require.NoError(t, err)

				wg.Wait()
			},
		},
		{
			name:        "Shutdown_Graceful",
			description: "Shutdown gracefully stops server and waits for connections to drain",
			testFunc: func(t *testing.T, server cryptoutilAppsFrameworkServiceServer.IPublicServer) {
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
				shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
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
			tlsCfg := cryptoutilAppsFrameworkServiceServerTestutil.PublicTLS()
			server, err := cryptoutilAppsFrameworkServiceServerListener.NewPublicHTTPServer(context.Background(), cryptoutilAppsFrameworkServiceServerTestutil.ServiceFrameworkServerSettings(), tlsCfg)
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
		setupFunc     func(t *testing.T) (cryptoutilAppsFrameworkServiceServer.IPublicServer, error)
		expectedError string
	}{
		{
			name:        "NilContext",
			description: "NewPublicHTTPServer rejects nil context",
			setupFunc: func(t *testing.T) (cryptoutilAppsFrameworkServiceServer.IPublicServer, error) {
				t.Helper()

				tlsCfg := cryptoutilAppsFrameworkServiceServerTestutil.PublicTLS()

				return cryptoutilAppsFrameworkServiceServerListener.NewPublicHTTPServer(nil, cryptoutilAppsFrameworkServiceServerTestutil.ServiceFrameworkServerSettings(), tlsCfg) //nolint:staticcheck // Testing nil context.
			},
			expectedError: "context cannot be nil",
		},
		{
			name:        "Start_NilContext",
			description: "Start rejects nil context",
			setupFunc: func(t *testing.T) (cryptoutilAppsFrameworkServiceServer.IPublicServer, error) {
				t.Helper()

				tlsCfg := cryptoutilAppsFrameworkServiceServerTestutil.PublicTLS()

				server, err := cryptoutilAppsFrameworkServiceServerListener.NewPublicHTTPServer(context.Background(), cryptoutilAppsFrameworkServiceServerTestutil.ServiceFrameworkServerSettings(), tlsCfg)
				require.NoError(t, err)

				err = server.Start(nil) //nolint:staticcheck // Testing nil context.

				return server, fmt.Errorf("failed to start server: %w", err)
			},
			expectedError: "context cannot be nil",
		},
		{
			name:        "Shutdown_NilContext",
			description: "Shutdown accepts nil context and uses Background()",
			setupFunc: func(t *testing.T) (cryptoutilAppsFrameworkServiceServer.IPublicServer, error) {
				t.Helper()

				tlsCfg := cryptoutilAppsFrameworkServiceServerTestutil.PublicTLS()

				server, err := cryptoutilAppsFrameworkServiceServerListener.NewPublicHTTPServer(context.Background(), cryptoutilAppsFrameworkServiceServerTestutil.ServiceFrameworkServerSettings(), tlsCfg)
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
			setupFunc: func(t *testing.T) (cryptoutilAppsFrameworkServiceServer.IPublicServer, error) {
				t.Helper()

				tlsCfg := cryptoutilAppsFrameworkServiceServerTestutil.PublicTLS()

				server, err := cryptoutilAppsFrameworkServiceServerListener.NewPublicHTTPServer(context.Background(), cryptoutilAppsFrameworkServiceServerTestutil.ServiceFrameworkServerSettings(), tlsCfg)
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
