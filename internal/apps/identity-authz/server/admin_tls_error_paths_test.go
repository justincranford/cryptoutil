// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	"crypto/tls"
	"net"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func nilAuthzContextForTest() context.Context {
	return nil
}

func TestAdminServer_LoadTLSConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ctx       context.Context
		certFile  string
		keyFile   string
		wantErr   string
		wantTLS13 bool
	}{
		{
			name:    "nil context",
			wantErr: "context cannot be nil",
			ctx:     nil,
		},
		{
			name:      "automatic certificate",
			ctx:       context.Background(),
			wantTLS13: true,
		},
		{
			name:      "missing key uses automatic certificate",
			ctx:       context.Background(),
			certFile:  "missing-cert.pem",
			wantTLS13: true,
		},
		{
			name:     "invalid certificate path",
			ctx:      context.Background(),
			certFile: "C:\\cryptoutil-test-missing\\server.crt",
			keyFile:  "C:\\cryptoutil-test-missing\\server.key",
			wantErr:  "failed to load TLS certificate and key",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := cryptoutilIdentityConfig.RequireNewForTest("authz-admin-load-tls")
			config.AuthZ.TLSCertFile = tc.certFile
			config.AuthZ.TLSKeyFile = tc.keyFile
			admin := &AdminServer{config: config}

			tlsConfig, err := admin.loadTLSConfig(tc.ctx)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.Nil(t, tlsConfig)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, tlsConfig)
			require.Equal(t, tc.wantTLS13, tlsConfig.MinVersion == tls.VersionTLS13)
			require.NotEmpty(t, tlsConfig.Certificates)
		})
	}
}

func TestAdminServer_StartAndShutdownErrorPaths(t *testing.T) {
	t.Parallel()

	t.Run("nil start context", func(t *testing.T) {
		t.Parallel()

		admin := &AdminServer{}
		require.ErrorContains(t, admin.Start(nilAuthzContextForTest()), "context cannot be nil")
	})

	t.Run("TLS configuration failure", func(t *testing.T) {
		t.Parallel()

		config := cryptoutilIdentityConfig.RequireNewForTest("authz-admin-start-tls-error")
		config.AuthZ.TLSCertFile = filepath.Join(t.TempDir(), "missing-start-cert.pem")
		config.AuthZ.TLSKeyFile = filepath.Join(t.TempDir(), "missing-start-key.pem")
		admin := &AdminServer{config: config}

		err := admin.Start(context.Background())
		require.ErrorContains(t, err, "failed to load TLS configuration")
		require.ErrorContains(t, err, "failed to load TLS certificate and key")
	})

	t.Run("occupied admin port", func(t *testing.T) {
		t.Parallel()

		listener, err := net.Listen("tcp", net.JoinHostPort(cryptoutilSharedMagic.IPv4Loopback, "0"))

		require.NoError(t, err)
		defer func() { require.NoError(t, listener.Close()) }()

		address, ok := listener.Addr().(*net.TCPAddr)
		require.True(t, ok)

		config := cryptoutilIdentityConfig.RequireNewForTest("authz-admin-occupied-port")
		config.AuthZ.AdminBindAddress = cryptoutilSharedMagic.IPv4Loopback
		config.AuthZ.AdminPort = address.Port
		admin, err := NewAdminHTTPServer(context.Background(), config)
		require.NoError(t, err)

		require.ErrorContains(t, admin.Start(context.Background()), "failed to create TLS listener")
	})

	t.Run("nil shutdown context and application", func(t *testing.T) {
		t.Parallel()

		admin := &AdminServer{ready: true}
		require.NoError(t, admin.Shutdown(nilAuthzContextForTest()))
		require.True(t, admin.shutdown)
		require.False(t, admin.ready)
	})
}

func TestAdminServer_TLSRootCAPoolMalformedCertificate(t *testing.T) {
	t.Parallel()

	admin := &AdminServer{
		tlsCerts: []tls.Certificate{{
			Certificate: [][]byte{{0x01}},
		}},
	}

	require.Nil(t, admin.TLSRootCAPool())
}
