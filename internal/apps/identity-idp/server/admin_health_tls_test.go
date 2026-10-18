// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	http "net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	testAdminHealthAliveJSONStatus        = `"status":"alive"`
	testAdminHealthShuttingDownJSONStatus = `"status":"shutting_down"`
)

func TestAdminServerHealthHandlers_StateMatrix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		ready        bool
		shutdown     bool
		livezStatus  int
		livezBody    string
		readyzStatus int
		readyzBody   string
	}{
		{
			name:         "starting",
			livezStatus:  http.StatusOK,
			livezBody:    testAdminHealthAliveJSONStatus,
			readyzStatus: http.StatusServiceUnavailable,
			readyzBody:   `"status":"not ready"`,
		},
		{
			name:         "ready",
			ready:        true,
			livezStatus:  http.StatusOK,
			livezBody:    testAdminHealthAliveJSONStatus,
			readyzStatus: http.StatusOK,
			readyzBody:   `"status":"ready"`,
		},
		{
			name:         cryptoutilSharedMagic.StringStatusShuttingDown,
			ready:        true,
			shutdown:     true,
			livezStatus:  http.StatusServiceUnavailable,
			livezBody:    testAdminHealthShuttingDownJSONStatus,
			readyzStatus: http.StatusServiceUnavailable,
			readyzBody:   testAdminHealthShuttingDownJSONStatus,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server, err := NewAdminHTTPServer(context.Background(), cryptoutilIdentityConfig.RequireNewForTest("idp_admin_coverage"))
			require.NoError(t, err)

			server.ready = tc.ready
			server.shutdown = tc.shutdown

			livezResponse, err := server.app.Test(httptest.NewRequest(http.MethodGet, cryptoutilSharedMagic.DefaultAdminLivezEndpoint, nil), -1)
			require.NoError(t, err)
			require.Equal(t, tc.livezStatus, livezResponse.StatusCode)
			require.NoError(t, livezResponse.Body.Close())

			readyzPath := cryptoutilSharedMagic.DefaultPrivateAdminAPIContextPath + cryptoutilSharedMagic.PrivateAdminReadyzRequestPath
			readyzResponse, err := server.app.Test(httptest.NewRequest(http.MethodGet, readyzPath, nil), -1)
			require.NoError(t, err)
			require.Equal(t, tc.readyzStatus, readyzResponse.StatusCode)
			require.NoError(t, readyzResponse.Body.Close())
		})
	}
}

func TestAdminServer_LoadTLSConfig_StaticCertificate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	generated, err := generateSelfSignedTLSConfig()
	require.NoError(t, err)
	require.Len(t, generated.Certificates, 1)

	privateKey, ok := generated.Certificates[0].PrivateKey.(*ecdsa.PrivateKey)
	require.True(t, ok)

	certPath := filepath.Join(t.TempDir(), "server.crt")
	keyPath := filepath.Join(t.TempDir(), "server.key")
	certificatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypeCertificate,
		Bytes: generated.Certificates[0].Certificate[0],
	})
	privateKeyDER, err := x509.MarshalECPrivateKey(privateKey)
	require.NoError(t, err)

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypeECPrivateKey,
		Bytes: privateKeyDER,
	})

	require.NoError(t, os.WriteFile(certPath, certificatePEM, cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(keyPath, privateKeyPEM, cryptoutilSharedMagic.CacheFilePermissions))

	config := cryptoutilIdentityConfig.RequireNewForTest("idp_admin_static_tls")
	config.IDP.TLSCertFile = certPath
	config.IDP.TLSKeyFile = keyPath
	server, err := NewAdminHTTPServer(ctx, config)
	require.NoError(t, err)

	tlsConfig, err := server.loadTLSConfig(ctx)
	require.NoError(t, err)
	require.Len(t, tlsConfig.Certificates, 1)
	require.Equal(t, uint16(tls.VersionTLS13), tlsConfig.MinVersion)
}

func TestAdminServer_ShutdownNilContext(t *testing.T) {
	t.Parallel()

	server, err := NewAdminHTTPServer(context.Background(), cryptoutilIdentityConfig.RequireNewForTest("idp_admin_shutdown_nil"))
	require.NoError(t, err)

	require.NoError(t, server.Shutdown(context.Background()))
	require.True(t, server.shutdown)
	require.False(t, server.ready)
}
