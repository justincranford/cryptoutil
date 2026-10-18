// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	"crypto/tls"
	json "encoding/json"
	http "net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const idpAdminReadyState = "ready"

func nilIDPContextForTest() context.Context {
	return nil
}

func TestAdminServer_HealthRouteStates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		path           string
		ready          bool
		shutdown       bool
		wantStatus     int
		wantStateValue string
	}{
		{
			name:           "livez reports alive",
			path:           cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminLivezRequestPath,
			wantStatus:     http.StatusOK,
			wantStateValue: "alive",
		},
		{
			name:           "livez reports shutdown",
			path:           cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminLivezRequestPath,
			shutdown:       true,
			wantStatus:     http.StatusServiceUnavailable,
			wantStateValue: cryptoutilSharedMagic.StringStatusShuttingDown,
		},
		{
			name:           "readyz reports not ready",
			path:           cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath,
			wantStatus:     http.StatusServiceUnavailable,
			wantStateValue: "not ready",
		},
		{
			name:           "readyz reports ready",
			path:           cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath,
			ready:          true,
			wantStatus:     http.StatusOK,
			wantStateValue: idpAdminReadyState,
		},
		{
			name:           "readyz reports shutdown",
			path:           cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath,
			ready:          true,
			shutdown:       true,
			wantStatus:     http.StatusServiceUnavailable,
			wantStateValue: cryptoutilSharedMagic.StringStatusShuttingDown,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			admin, err := NewAdminHTTPServer(context.Background(), cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin_state"))
			require.NoError(t, err)

			admin.mu.Lock()
			admin.ready = tc.ready
			admin.shutdown = tc.shutdown
			admin.mu.Unlock()

			response, err := admin.app.Test(httptest.NewRequest(http.MethodGet, tc.path, nil), -1)
			require.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, response.Body.Close())
			})

			require.Equal(t, tc.wantStatus, response.StatusCode)

			var body map[string]string
			require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
			require.Equal(t, tc.wantStateValue, body[cryptoutilSharedMagic.StringStatus])
		})
	}
}

func TestAdminServer_TLSRootCAPool_MalformedCertificate(t *testing.T) {
	t.Parallel()

	admin, err := NewAdminHTTPServer(context.Background(), cryptoutilIdentityConfig.RequireNewForTest("test_idp_admin_malformed_cert"))
	require.NoError(t, err)

	admin.tlsCerts = []tls.Certificate{{Certificate: [][]byte{[]byte("invalid DER")}}}

	require.Nil(t, admin.TLSRootCAPool())
}

func TestIDPServer_Start_NilContext(t *testing.T) {
	t.Parallel()

	requireTestSetup(t)
	require.ErrorContains(t, testServer.Start(nilIDPContextForTest()), "failed to start application")
}
