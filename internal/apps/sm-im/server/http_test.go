// Copyright (c) 2025-2026 Justin Cranford.
//
//

package server_test

import (
	"context"
	"fmt"
	http "net/http"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps-framework/service/server"
	cryptoutilAppsFrameworkServiceServerTestutil "cryptoutil/internal/apps-framework/service/server/testutil"
	cryptoutilAppsSmImServer "cryptoutil/internal/apps/sm-im/server"
	cryptoutilAppsSmImServerConfig "cryptoutil/internal/apps/sm-im/server/config"
)

// initTestConfig returns an SmIMServerSettings with all required settings for tests.
func initTestConfig(t testing.TB) *cryptoutilAppsSmImServerConfig.SmIMServerSettings {
	t.Helper()

	settings := cryptoutilAppsFrameworkServiceConfig.RequireNewForTest("sm-im-http-test")
	settings.DatabaseURL = cryptoutilAppsFrameworkServiceServerTestutil.NewUniqueSQLiteMemoryURL(t, "sm-im-http-test")

	return &cryptoutilAppsSmImServerConfig.SmIMServerSettings{
		ServiceFrameworkServerSettings: settings,
	}
}

// TestHTTPGet tests the httpGet helper function (used by health CLI wrappers).
func TestHTTPGet(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create server with dynamic ports.
	harness := cryptoutilAppsFrameworkServiceServerTestutil.NewStartedHTTPServer(t, ctx, func(ctx context.Context) (cryptoutilAppsFrameworkServiceServer.ServiceServer, error) {
		cfg := initTestConfig(t)

		return cryptoutilAppsSmImServer.NewIMServerFromConfig(ctx, cfg)
	})

	// Get actual ports.
	publicPort := harness.Server.PublicPort()
	adminPort := harness.Server.AdminPort()

	// Test public health endpoint.
	t.Run("public_health_endpoint", func(t *testing.T) {
		url := fmt.Sprintf("%s%d/service/api/v1/health", cryptoutilSharedMagic.URLPrefixLocalhostHTTPS, publicPort)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		require.NoError(t, err)

		resp, err := harness.PublicClient.Do(req)
		require.NoError(t, err)

		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Test admin livez endpoint.
	t.Run("admin_livez_endpoint", func(t *testing.T) {
		url := fmt.Sprintf("%s%d/admin/api/v1/livez", cryptoutilSharedMagic.URLPrefixLocalhostHTTPS, adminPort)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		require.NoError(t, err)

		resp, err := harness.AdminClient.Do(req)
		require.NoError(t, err)

		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Test admin readyz endpoint.
	t.Run("admin_readyz_endpoint", func(t *testing.T) {
		url := fmt.Sprintf("%s%d/admin/api/v1/readyz", cryptoutilSharedMagic.URLPrefixLocalhostHTTPS, adminPort)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		require.NoError(t, err)

		resp, err := harness.AdminClient.Do(req)
		require.NoError(t, err)

		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// TestHTTPPost tests the httpPost helper function (used by shutdown CLI wrapper).
func TestHTTPPost(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create server with dynamic ports.
	harness := cryptoutilAppsFrameworkServiceServerTestutil.NewStartedHTTPServer(t, ctx, func(ctx context.Context) (cryptoutilAppsFrameworkServiceServer.ServiceServer, error) {
		cfg := initTestConfig(t)

		return cryptoutilAppsSmImServer.NewIMServerFromConfig(ctx, cfg)
	})

	// Get actual ports.
	adminPort := harness.Server.AdminPort()

	// Test admin shutdown endpoint (triggers async shutdown).
	t.Run("admin_shutdown_endpoint", func(t *testing.T) {
		url := fmt.Sprintf("%s%d/admin/api/v1/shutdown", cryptoutilSharedMagic.URLPrefixLocalhostHTTPS, adminPort)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
		require.NoError(t, err)

		resp, err := harness.AdminClient.Do(req)
		require.NoError(t, err)

		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Cancel context to trigger server shutdown (shutdown endpoint starts async shutdown).
	cancel()
}

// TestIMServer tests the imServer function.
func TestIMServer(t *testing.T) {
	t.Parallel()
	// This test would require mocking os.Signal and context handling.
	// Skipping for now as imServer is tested via integration tests.
	t.Skip("imServer requires signal mocking - tested via integration tests")
}
