// Copyright (c) 2025-2026 Justin Cranford.
package testutil

import (
	"context"
	"crypto/tls"
	"fmt"
	http "net/http"
	"strings"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps-framework/service/server"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilPoll "cryptoutil/internal/shared/util/poll"
)

const (
	httpTestStartupTimeout  = 30 * time.Second
	httpTestStartupInterval = 100 * time.Millisecond
	httpTestShutdownTimeout = 5 * time.Second
)

// HTTPTestServer bundles the runtime pieces shared by HTTP server tests.
type HTTPTestServer struct {
	Server       cryptoutilAppsFrameworkServiceServer.ServiceServer
	PublicClient *http.Client
	AdminClient  *http.Client
}

// NewUniqueSQLiteMemoryURL returns a unique in-memory SQLite URL for a test instance.
func NewUniqueSQLiteMemoryURL(t testing.TB, serviceName string) string {
	t.Helper()

	instanceID, err := googleUuid.NewV7()
	require.NoError(t, err)

	return fmt.Sprintf("file:%s-%s?mode=memory&cache=shared", serviceName, instanceID.String())
}

// StartHTTPServer starts a service server, waits for both public and admin ports, and
// registers a cleanup that shuts the server down with a bounded timeout.
func StartHTTPServer(t testing.TB, ctx context.Context, srv cryptoutilAppsFrameworkServiceServer.ServiceServer) *HTTPTestServer {
	t.Helper()
	require.NotNil(t, srv)

	errChan := make(chan error, 1)

	go func() {
		errChan <- srv.Start(ctx)
	}()

	require.NoError(t, waitForHTTPServerPorts(ctx, srv, errChan))

	srv.SetReady(true)

	server := &HTTPTestServer{
		Server: srv,
		PublicClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS13,
					RootCAs:    srv.TLSRootCAPool(),
				},
				DisableKeepAlives: true,
			},
			Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
		},
		AdminClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS13,
					RootCAs:    srv.AdminTLSRootCAPool(),
				},
				DisableKeepAlives: true,
			},
			Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
		},
	}

	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), httpTestShutdownTimeout)
		defer cancel()

		shutdownErr := srv.Shutdown(shutdownCtx)
		if shutdownErr != nil && !strings.Contains(shutdownErr.Error(), "already shutdown") {
			require.NoError(t, shutdownErr)
		}
	})

	return server
}

func waitForHTTPServerPorts(ctx context.Context, srv cryptoutilAppsFrameworkServiceServer.ServiceServer, errChan <-chan error) error {
	if err := cryptoutilSharedUtilPoll.Until(ctx, httpTestStartupTimeout, httpTestStartupInterval, func(_ context.Context) (bool, error) {
		select {
		case startErr := <-errChan:
			return false, fmt.Errorf("server failed to start: %w", startErr)
		default:
		}

		return srv.PublicPort() > 0 && srv.AdminPort() > 0, nil
	}); err != nil {
		return fmt.Errorf("waiting for HTTP server ports: %w", err)
	}

	return nil
}
