// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	http "net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestIDPAdminServer_StartInvalidTLSFiles(t *testing.T) {
	t.Parallel()

	cfg := cryptoutilIdentityConfig.RequireNewForTest("idp_admin_start_invalid_tls")
	cfg.IDP.TLSCertFile = filepath.Join(t.TempDir(), "missing-cert.pem")
	cfg.IDP.TLSKeyFile = filepath.Join(t.TempDir(), "missing-key.pem")

	admin, err := NewAdminHTTPServer(context.Background(), cfg)
	require.NoError(t, err)

	require.ErrorContains(t, admin.Start(context.Background()), "failed to load TLS configuration")
}

func TestIDPAdminServer_ShutdownNilContext(t *testing.T) {
	t.Parallel()

	admin, err := NewAdminHTTPServer(
		context.Background(),
		cryptoutilIdentityConfig.RequireNewForTest("idp_admin_shutdown_nil_context"),
	)
	require.NoError(t, err)

	var nilContext context.Context
	require.NoError(t, admin.Shutdown(nilContext))
	require.True(t, admin.shutdown)
	require.False(t, admin.ready)
}

func TestIDPAdminServer_ShutdownEndpointCompletesAsyncShutdown(t *testing.T) {
	t.Parallel()

	admin, err := NewAdminHTTPServer(
		context.Background(),
		cryptoutilIdentityConfig.RequireNewForTest("idp_admin_shutdown_async"),
	)
	require.NoError(t, err)

	path := cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath +
		cryptoutilSharedMagic.PrivateAdminShutdownRequestPath
	response, err := admin.app.Test(httptest.NewRequest(http.MethodPost, path, nil), -1)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.NoError(t, response.Body.Close())

	require.Eventually(
		t,
		func() bool {
			admin.mu.RLock()
			defer admin.mu.RUnlock()

			return admin.shutdown
		},
		cryptoutilSharedMagic.DefaultIdentityServerShutdownDelay+time.Second,
		cryptoutilSharedMagic.DefaultIdentityServerShutdownDelay/4,
	)
}
