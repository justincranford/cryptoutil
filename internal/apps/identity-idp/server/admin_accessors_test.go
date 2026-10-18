// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
)

const adminTestPort = 12345

type adminTestListener struct {
	addr net.Addr
}

func (l *adminTestListener) Accept() (net.Conn, error) {
	return nil, errors.New("accept is not implemented")
}

func (l *adminTestListener) Close() error {
	return nil
}

func (l *adminTestListener) Addr() net.Addr {
	return l.addr
}

func TestAdminServer_Accessors(t *testing.T) {
	t.Parallel()

	cfg := cryptoutilIdentityConfig.RequireNewForTest("idp-admin-accessors")
	admin, err := NewAdminHTTPServer(context.Background(), cfg)
	require.NoError(t, err)
	require.Nil(t, admin.TLSRootCAPool())
	require.Zero(t, admin.ActualPort())
	require.Equal(t, "https://127.0.0.1:0", admin.AdminBaseURL())

	tlsConfig, err := generateSelfSignedTLSConfig()
	require.NoError(t, err)

	admin.tlsCerts = tlsConfig.Certificates
	require.NotNil(t, admin.TLSRootCAPool())

	admin.listener = &adminTestListener{addr: &net.TCPAddr{Port: adminTestPort}}
	require.Equal(t, adminTestPort, admin.ActualPort())
	require.Equal(t, fmt.Sprintf("https://127.0.0.1:%d", adminTestPort), admin.AdminBaseURL())

	admin.listener = &adminTestListener{addr: &net.UnixAddr{Name: "test", Net: "unix"}}
	require.Zero(t, admin.ActualPort())
	require.Equal(t, "https://127.0.0.1:0", admin.AdminBaseURL())
}
