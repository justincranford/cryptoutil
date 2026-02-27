// Copyright (c) 2025 Justin Cranford
//
//

package listener_test

import (
	"context"
	"crypto/tls"
	json "encoding/json"
	"fmt"
	"io"
	http "net/http"
	"sync"
	"testing"
	"time"

	cryptoutilAppsTemplateServiceServerListener "cryptoutil/internal/apps/template/service/server/listener"
	cryptoutilAppsTemplateServiceServerTestutil "cryptoutil/internal/apps/template/service/server/testutil"
	cryptoutilAppsTemplateServiceTestingHttpservertests "cryptoutil/internal/apps/template/service/testing/httpservertests"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestAdminServer_Shutdown_Endpoint(t *testing.T) {
	// NOT parallel - all admin server tests compete for port 9090.
	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsTemplateServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background and track goroutine.
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		_ = server.Start(ctx)
	}()

	// Wait for server to be ready.
	time.Sleep(200 * time.Millisecond)

	port := server.ActualPort()
	require.Greater(t, port, 0, "Expected dynamic port allocation")

	// Trigger shutdown via HTTP endpoint.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // Self-signed cert in test.
		},
		Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second,
	}

	reqCtx, reqCancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)
	defer reqCancel()

	url := fmt.Sprintf("https://%s:%d/admin/api/v1/shutdown", cryptoutilSharedMagic.IPv4Loopback, port)

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, url, nil)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Verify response.
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]any

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	require.Equal(t, "shutdown initiated", response[cryptoutilSharedMagic.StringStatus])

	// The endpoint triggers shutdown in a goroutine with 100ms delay.
	// Cancel context to let Start() exit cleanly, then wait for goroutine.
	cancel()
	wg.Wait()

	// Wait for OS socket cleanup (TCP TIME_WAIT state).
	// Windows needs longer for socket release.
	time.Sleep(3 * time.Second)
}

// TestAdminServer_Shutdown_NilContext tests Shutdown accepts nil context and uses Background().
func TestAdminServer_Shutdown_NilContext(t *testing.T) {
	// NOT parallel - all admin server tests compete for port 9090.
	createServer := func(t *testing.T) cryptoutilAppsTemplateServiceTestingHttpservertests.HTTPServer {
		t.Helper()

		tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PrivateTLS()
		server, err := cryptoutilAppsTemplateServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
		require.NoError(t, err)

		return server
	}

	cryptoutilAppsTemplateServiceTestingHttpservertests.TestShutdownNilContext(t, createServer)

	// Wait for OS socket cleanup (TCP TIME_WAIT state).
	// Windows needs longer for socket release - ConcurrentRequests runs next.
	time.Sleep(2 * time.Second)
}

// TestAdminServer_ActualPort_BeforeStart tests ActualPort before server starts.
func TestAdminServer_ActualPort_BeforeStart(t *testing.T) {
	t.Parallel()

	tlsCfg := cryptoutilAppsTemplateServiceServerTestutil.PrivateTLS()
	server, err := cryptoutilAppsTemplateServiceServerListener.NewAdminHTTPServer(context.Background(), cryptoutilAppsTemplateServiceServerTestutil.ServiceTemplateServerSettings(), tlsCfg)
	require.NoError(t, err)

	port := server.ActualPort()

	require.Equal(t, 0, port, "Expected port 0 before server starts")
}

// TestAdminServer_ConcurrentRequests tests multiple concurrent requests to admin endpoints.
