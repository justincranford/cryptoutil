// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net"
	http "net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

type adminTestAddress struct{}

func (adminTestAddress) Network() string {
	return testAdminAddressNetwork
}

func (adminTestAddress) String() string {
	return testAdminAddressNetwork
}

type adminTestListener struct{}

func (adminTestListener) Accept() (net.Conn, error) {
	return nil, errors.New("accept is not used")
}

func (adminTestListener) Close() error {
	return nil
}

func (adminTestListener) Addr() net.Addr {
	return adminTestAddress{}
}

func TestAdminServer_HealthHandlersJSONEncodingErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		ready    bool
		shutdown bool
		handler  func(*AdminServer) fiber.Handler
	}{
		{
			name: "livez",
			path: cryptoutilSharedMagic.PrivateAdminLivezRequestPath,
			handler: func(s *AdminServer) fiber.Handler {
				return s.handleLivez
			},
		},
		{
			name:     "livez while shutting down",
			path:     "/livez-shutdown",
			shutdown: true,
			handler: func(s *AdminServer) fiber.Handler {
				return s.handleLivez
			},
		},
		{
			name: "readyz while not ready",
			path: "/readyz-not-ready",
			handler: func(s *AdminServer) fiber.Handler {
				return s.handleReadyz
			},
		},
		{
			name:  "readyz",
			path:  cryptoutilSharedMagic.PrivateAdminReadyzRequestPath,
			ready: true,
			handler: func(s *AdminServer) fiber.Handler {
				return s.handleReadyz
			},
		},
		{
			name:     "readyz while shutting down",
			path:     "/readyz-shutdown",
			ready:    true,
			shutdown: true,
			handler: func(s *AdminServer) fiber.Handler {
				return s.handleReadyz
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
				JSONEncoder: func(any) ([]byte, error) {
					return nil, errors.New("injected JSON encoding failure")
				},
			})
			server := &AdminServer{
				app:      app,
				ready:    tc.ready,
				shutdown: tc.shutdown,
			}
			app.Get(tc.path, tc.handler(server))

			response, err := app.Test(httptest.NewRequest(http.MethodGet, tc.path, nil), -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, response.Body.Close()) }()

			require.Equal(t, http.StatusInternalServerError, response.StatusCode)
		})
	}
}

func TestAdminServer_TLSConfigFileHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		configure        func(*testing.T, *cryptoutilIdentityConfig.Config)
		nilContext       bool
		wantError        string
		wantCertificates bool
	}{
		{
			name: "generated TLS material",
			configure: func(*testing.T, *cryptoutilIdentityConfig.Config) {
			},
			wantCertificates: true,
		},
		{
			name: "partial TLS material generates certificate",
			configure: func(_ *testing.T, cfg *cryptoutilIdentityConfig.Config) {
				cfg.RS.TLSCertFile = "provided-certificate.pem"
			},
			wantCertificates: true,
		},
		{
			name: "missing TLS file",
			configure: func(t *testing.T, cfg *cryptoutilIdentityConfig.Config) {
				cfg.RS.TLSCertFile = filepath.Join(t.TempDir(), "missing-certificate.pem")
				cfg.RS.TLSKeyFile = filepath.Join(t.TempDir(), "missing-key.pem")
			},
			wantError: "failed to load TLS certificate and key",
		},
		{
			name: testAdminNilContextCaseName,
			configure: func(*testing.T, *cryptoutilIdentityConfig.Config) {
			},
			nilContext: true,
			wantError:  testContextCannotBeNilError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := cryptoutilIdentityConfig.RequireNewForTest("test_rs_tls_config")
			tc.configure(t, cfg)
			server, err := NewAdminHTTPServer(context.Background(), cfg)
			require.NoError(t, err)

			ctx := context.Background()
			if tc.nilContext {
				ctx = nil
			}

			tlsConfig, err := server.loadTLSConfig(ctx)

			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)

				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, tlsConfig.Certificates)

			if tc.wantCertificates {
				require.Equal(t, uint16(tls.VersionTLS13), tlsConfig.MinVersion)
			}
		})
	}
}

func TestAdminServer_TLSConfigLoadsCertificateFiles(t *testing.T) {
	t.Parallel()

	cfg := cryptoutilIdentityConfig.RequireNewForTest("test_rs_tls_load")
	tlsConfig, err := generateSelfSignedTLSConfig()
	require.NoError(t, err)

	certPath, keyPath := writeAdminTestTLSFiles(t, tlsConfig.Certificates[0])
	cfg.RS.TLSCertFile = certPath
	cfg.RS.TLSKeyFile = keyPath

	server, err := NewAdminHTTPServer(context.Background(), cfg)
	require.NoError(t, err)

	loadedTLSConfig, err := server.loadTLSConfig(context.Background())

	require.NoError(t, err)
	require.Len(t, loadedTLSConfig.Certificates, 1)
	require.Equal(t, uint16(tls.VersionTLS13), loadedTLSConfig.MinVersion)
}

func TestAdminServer_StartRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		nilContext  bool
		badTLSFiles bool
		wantError   string
	}{
		{name: testAdminNilContextCaseName, nilContext: true, wantError: testContextCannotBeNilError},
		{name: "invalid TLS files", badTLSFiles: true, wantError: "failed to load TLS configuration"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := cryptoutilIdentityConfig.RequireNewForTest("test_rs_start_error")
			if tc.badTLSFiles {
				cfg.RS.TLSCertFile = filepath.Join(t.TempDir(), "missing-certificate.pem")
				cfg.RS.TLSKeyFile = filepath.Join(t.TempDir(), "missing-key.pem")
			}

			server, err := NewAdminHTTPServer(context.Background(), cfg)
			require.NoError(t, err)

			ctx := context.Background()
			if tc.nilContext {
				ctx = nil
			}

			require.ErrorContains(t, server.Start(ctx), tc.wantError)
		})
	}
}

func TestAdminServer_StartRejectsOccupiedListenerPort(t *testing.T) {
	t.Parallel()

	occupiedListener, err := net.Listen("tcp", net.JoinHostPort(cryptoutilSharedMagic.IPv4Loopback, "0"))

	require.NoError(t, err)
	defer func() { require.NoError(t, occupiedListener.Close()) }()

	tcpAddr, ok := occupiedListener.Addr().(*net.TCPAddr)
	require.True(t, ok)

	cfg := cryptoutilIdentityConfig.RequireNewForTest("test_rs_start_port_conflict")
	cfg.RS.AdminBindAddress = cryptoutilSharedMagic.IPv4Loopback
	cfg.RS.AdminPort = tcpAddr.Port

	server, err := NewAdminHTTPServer(context.Background(), cfg)
	require.NoError(t, err)
	require.ErrorContains(t, server.Start(context.Background()), "failed to create TLS listener")
}

func writeAdminTestTLSFiles(t *testing.T, cert tls.Certificate) (string, string) {
	t.Helper()

	privateKey, ok := cert.PrivateKey.(*ecdsa.PrivateKey)
	require.True(t, ok)

	certPath := filepath.Join(t.TempDir(), "server.crt")
	keyPath := filepath.Join(t.TempDir(), "server.key")
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypeCertificate,
		Bytes: cert.Certificate[0],
	})
	keyBytes, err := x509.MarshalECPrivateKey(privateKey)
	require.NoError(t, err)

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypeECPrivateKey,
		Bytes: keyBytes,
	})

	require.NoError(t, os.WriteFile(certPath, certPEM, 0o600))
	require.NoError(t, os.WriteFile(keyPath, keyPEM, 0o600))

	return certPath, keyPath
}

func TestAdminServer_AccessorsWithoutTCPListener(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		listener net.Listener
	}{
		{name: "missing listener"},
		{name: "non TCP listener", listener: adminTestListener{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := &AdminServer{listener: tc.listener}

			require.Zero(t, server.ActualPort())
			require.Equal(t, "https://127.0.0.1:0", server.AdminBaseURL())
			require.Nil(t, server.TLSRootCAPool())
		})
	}
}

func TestAdminServer_TLSRootCAPoolRejectsInvalidMaterial(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cert tls.Certificate
	}{
		{name: "certificate bytes absent"},
		{
			name: "malformed certificate",
			cert: tls.Certificate{Certificate: [][]byte{{0x01}}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := &AdminServer{tlsCerts: []tls.Certificate{tc.cert}}
			require.Nil(t, server.TLSRootCAPool())
		})
	}
}

func TestAdminServer_ShutdownWithNilContextAndNoApplication(t *testing.T) {
	t.Parallel()

	var nilContext context.Context
	require.NoError(t, (&AdminServer{}).Shutdown(nilContext))
}
