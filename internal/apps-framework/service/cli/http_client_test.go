// Copyright (c) 2025-2026 Justin Cranford.
//
// SPDX-License-Identifier: AGPL-3.0-only
package cli_test

import (
	"bytes"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	http "net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkCli "cryptoutil/internal/apps-framework/service/cli"
)

func TestLoadCACertPool_EmptyPath(t *testing.T) {
	t.Parallel()

	pool, err := cryptoutilAppsFrameworkCli.LoadCACertPool("")
	require.NoError(t, err)
	require.Nil(t, pool, "expected nil pool for empty path")
}

func TestLoadCACertPool_NonExistentPath(t *testing.T) {
	t.Parallel()

	pool, err := cryptoutilAppsFrameworkCli.LoadCACertPool("/nonexistent/path/to/ca.pem")
	require.Error(t, err)
	require.Nil(t, pool)
}

func TestLoadCACertPool_InvalidPEM(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "invalid.pem")

	err := os.WriteFile(certPath, []byte("not a valid PEM certificate"), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	pool, err := cryptoutilAppsFrameworkCli.LoadCACertPool(certPath)
	require.Error(t, err)
	require.Nil(t, pool)
}

func TestLoadCACertPool_InvalidCertDER(t *testing.T) {
	t.Parallel()

	// Create a PEM file with CERTIFICATE type but invalid DER bytes inside.
	invalidCertPEM := pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypeCertificate,
		Bytes: []byte("invalid DER certificate data"),
	})

	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "bad_cert.pem")

	err := os.WriteFile(certPath, invalidCertPEM, cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	pool, err := cryptoutilAppsFrameworkCli.LoadCACertPool(certPath)
	require.Error(t, err)
	require.Nil(t, pool)
	require.Contains(t, err.Error(), "failed to parse CA certificate")
}

func TestHTTPGet_Success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	}))
	t.Cleanup(srv.Close)

	statusCode, body, err := cryptoutilAppsFrameworkCli.HTTPGet(srv.URL+"/test", "", "", "")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.Contains(t, body, cryptoutilSharedMagic.StringStatus)
}

func TestHTTPGet_ServiceUnavailable(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprint(w, "Service Unavailable")
	}))
	t.Cleanup(srv.Close)

	statusCode, body, err := cryptoutilAppsFrameworkCli.HTTPGet(srv.URL+"/test", "", "", "")
	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, statusCode)
	require.Contains(t, body, "Unavailable")
}

func TestHTTPGet_ConnectionRefused(t *testing.T) {
	t.Parallel()

	// Port 1 is always refused
	_, _, err := cryptoutilAppsFrameworkCli.HTTPGet("https://127.0.0.1:1/test", "", "", "")
	require.Error(t, err)
}

func TestHTTPPost_Success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)

			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"result":"done"}`)
	}))
	t.Cleanup(srv.Close)

	statusCode, body, err := cryptoutilAppsFrameworkCli.HTTPPost(srv.URL+cryptoutilSharedMagic.PrivateAdminShutdownRequestPath, "", "", "")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.Contains(t, body, "done")
}

func TestHTTPPost_Accepted(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = fmt.Fprint(w, `{"status":"accepted"}`)
	}))
	t.Cleanup(srv.Close)

	statusCode, body, err := cryptoutilAppsFrameworkCli.HTTPPost(srv.URL+cryptoutilSharedMagic.PrivateAdminShutdownRequestPath, "", "", "")
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, statusCode)
	require.Contains(t, body, "accepted")
}

func TestHTTPPost_ConnectionRefused(t *testing.T) {
	t.Parallel()

	_, _, err := cryptoutilAppsFrameworkCli.HTTPPost("https://127.0.0.1:1/shutdown", "", "", "")
	require.Error(t, err)
}

func writeCACertFromTLSServer(t *testing.T, srv *httptest.Server) string {
	t.Helper()

	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "ca.pem")

	// Export the httptest server's certificate as PEM
	certDER := srv.TLS.Certificates[0].Certificate[0]

	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypeCertificate,
		Bytes: certDER,
	})
	require.NotNil(t, pemData, "failed to encode certificate to PEM")

	err := os.WriteFile(certPath, pemData, cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	return certPath
}

func TestLoadCACertPool_ValidCert(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	certPath := writeCACertFromTLSServer(t, srv)

	pool, err := cryptoutilAppsFrameworkCli.LoadCACertPool(certPath)
	require.NoError(t, err)
	require.NotNil(t, pool, "expected non-nil pool for valid cert")
}

func TestHTTPGet_WithCACert(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	}))
	t.Cleanup(srv.Close)

	certPath := writeCACertFromTLSServer(t, srv)

	statusCode, body, err := cryptoutilAppsFrameworkCli.HTTPGet(srv.URL+"/test", certPath, "", "")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.Contains(t, body, cryptoutilSharedMagic.StringStatus)
}

func TestHTTPPost_WithCACert(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"result":"done"}`)
	}))
	t.Cleanup(srv.Close)

	certPath := writeCACertFromTLSServer(t, srv)

	statusCode, body, err := cryptoutilAppsFrameworkCli.HTTPPost(srv.URL+cryptoutilSharedMagic.PrivateAdminShutdownRequestPath, certPath, "", "")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.Contains(t, body, "done")
}

func TestHealthCommand_WithCACert(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	}))
	t.Cleanup(srv.Close)

	certPath := writeCACertFromTLSServer(t, srv)

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsFrameworkCli.HealthCommand(
		[]string{cryptoutilSharedMagic.CLIURLFlag, srv.URL, cryptoutilSharedMagic.CLICACertFlag, certPath},
		&stdout, &stderr,
		"Usage: health",
		cryptoutilSharedMagic.JoseJAServicePort,
	)
	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "\u2705")
}

func TestLivezCommand_WithCACert(t *testing.T) {
	t.Parallel()

	srv := newHealthMockServer(t, "/admin/api/v1/livez", http.StatusOK, `{"status":"alive"}`)

	certPath := writeCACertFromTLSServer(t, srv)

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsFrameworkCli.LivezCommand(
		[]string{cryptoutilSharedMagic.CLIURLFlag, srv.URL, cryptoutilSharedMagic.CLICACertFlag, certPath},
		&stdout, &stderr,
		"Usage: livez",
	)
	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "\u2705")
}

func TestReadyzCommand_WithCACert(t *testing.T) {
	t.Parallel()

	srv := newHealthMockServer(t, "/admin/api/v1/readyz", http.StatusOK, `{"ready":true}`)

	certPath := writeCACertFromTLSServer(t, srv)

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsFrameworkCli.ReadyzCommand(
		[]string{cryptoutilSharedMagic.CLIURLFlag, srv.URL, cryptoutilSharedMagic.CLICACertFlag, certPath},
		&stdout, &stderr,
		"Usage: readyz",
	)
	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "\u2705")
}

func TestShutdownCommand_WithCACert(t *testing.T) {
	t.Parallel()

	srv := newHealthMockServer(t, "/admin/api/v1/shutdown", http.StatusOK, `{"shutdown":"initiated"}`)

	certPath := writeCACertFromTLSServer(t, srv)

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsFrameworkCli.ShutdownCommand(
		[]string{cryptoutilSharedMagic.CLIURLFlag, srv.URL, cryptoutilSharedMagic.CLICACertFlag, certPath},
		&stdout, &stderr,
		"Usage: shutdown",
	)
	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "\u2705")
}

func TestHealthCommand_URLAlreadyHasHealthPath(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	}))
	t.Cleanup(srv.Close)

	var stdout, stderr bytes.Buffer

	// Pass a URL that already ends with /health
	exitCode := cryptoutilAppsFrameworkCli.HealthCommand(
		[]string{cryptoutilSharedMagic.CLIURLFlag, srv.URL + "/health"},
		&stdout, &stderr,
		"Usage: health",
		cryptoutilSharedMagic.JoseJAServicePort,
	)
	// Connection succeeds but response shape may differ
	// The key check is that it doesn't double-append /health
	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "\u2705")
}

func TestHTTPGet_InvalidCACert(t *testing.T) {
	t.Parallel()

	// An invalid cacert path should cause LoadCACertPool to fail
	_, _, err := cryptoutilAppsFrameworkCli.HTTPGet("https://127.0.0.1:1/test", "/nonexistent/ca.pem", "", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to load CA certificate")
}

func TestHTTPPost_InvalidCACert(t *testing.T) {
	t.Parallel()

	// An invalid cacert path should cause LoadCACertPool to fail
	_, _, err := cryptoutilAppsFrameworkCli.HTTPPost("https://127.0.0.1:1/shutdown", "/nonexistent/ca.pem", "", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to load CA certificate")
}

func TestHTTPGet_InvalidURL(t *testing.T) {
	t.Parallel()

	// A completely invalid URL scheme causes NewRequestWithContext to fail
	_, _, err := cryptoutilAppsFrameworkCli.HTTPGet("://invalid-url", "", "", "")
	require.Error(t, err)
}

func TestHTTPPost_InvalidURL(t *testing.T) {
	t.Parallel()

	// A completely invalid URL scheme causes NewRequestWithContext to fail
	_, _, err := cryptoutilAppsFrameworkCli.HTTPPost("://invalid-url", "", "", "")
	require.Error(t, err)
}

// --- LoadClientCert tests ---

func TestLoadClientCert_BothEmpty(t *testing.T) {
	t.Parallel()

	cert, err := cryptoutilAppsFrameworkCli.LoadClientCert("", "")
	require.NoError(t, err)
	require.Nil(t, cert, "expected nil cert when both paths are empty")
}

func TestLoadClientCert_OnlyCert(t *testing.T) {
	t.Parallel()

	// Only certPath provided without keyPath — must error.
	cert, err := cryptoutilAppsFrameworkCli.LoadClientCert("/some/cert.pem", "")
	require.Error(t, err)
	require.Nil(t, cert)
	require.Contains(t, err.Error(), "--cert and --key must be provided together")
}

func TestLoadClientCert_OnlyKey(t *testing.T) {
	t.Parallel()

	// Only keyPath provided without certPath — must error.
	cert, err := cryptoutilAppsFrameworkCli.LoadClientCert("", "/some/key.pem")
	require.Error(t, err)
	require.Nil(t, cert)
	require.Contains(t, err.Error(), "--cert and --key must be provided together")
}

func TestLoadClientCert_InvalidPaths(t *testing.T) {
	t.Parallel()

	// Both paths provided but files don't exist.
	cert, err := cryptoutilAppsFrameworkCli.LoadClientCert("/nonexistent/cert.pem", "/nonexistent/key.pem")
	require.Error(t, err)
	require.Nil(t, cert)
	require.Contains(t, err.Error(), "failed to load client certificate and key")
}

// --- mTLS flag parsing tests ---

func TestLivezCommand_WithCertAndKey_InvalidPaths(t *testing.T) {
	t.Parallel()

	// Passing --cert and --key with non-existent paths should fail with cert load error.
	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsFrameworkCli.LivezCommand(
		[]string{
			cryptoutilSharedMagic.CLIURLFlag, "https://127.0.0.1:1/admin/api/v1/livez",
			cryptoutilSharedMagic.CLICertFlag, "/nonexistent/client.pem",
			cryptoutilSharedMagic.CLIKeyFlag, "/nonexistent/client.key",
		},
		&stdout, &stderr,
		"Usage: livez",
	)
	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "failed to load client certificate")
}

func TestReadyzCommand_WithCertAndKey_InvalidPaths(t *testing.T) {
	t.Parallel()

	// Passing --cert and --key with non-existent paths should fail with cert load error.
	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsFrameworkCli.ReadyzCommand(
		[]string{
			cryptoutilSharedMagic.CLIURLFlag, "https://127.0.0.1:1/admin/api/v1/readyz",
			cryptoutilSharedMagic.CLICertFlag, "/nonexistent/client.pem",
			cryptoutilSharedMagic.CLIKeyFlag, "/nonexistent/client.key",
		},
		&stdout, &stderr,
		"Usage: readyz",
	)
	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "failed to load client certificate")
}

func TestHTTPGet_WithCertOnly_Error(t *testing.T) {
	t.Parallel()

	// Providing cert without key should fail early with a clear error.
	_, _, err := cryptoutilAppsFrameworkCli.HTTPGet("https://127.0.0.1:1/test", "", "/some/cert.pem", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to load client certificate")
}

func TestHTTPPost_WithKeyOnly_Error(t *testing.T) {
	t.Parallel()

	// Providing key without cert should fail early with a clear error.
	_, _, err := cryptoutilAppsFrameworkCli.HTTPPost("https://127.0.0.1:1/shutdown", "", "", "/some/key.pem")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to load client certificate")
}

// --- mTLS round-trip tests (buildTLSConfig clientCert branch) ---

// writeSelfSignedClientCert generates an EC self-signed cert+key, writes them to temp files,
// and returns (certPath, keyPath, caPool). The CA pool trusts the generated cert directly
// so it can also be loaded as a CA cert by the server for client authentication.
func writeSelfSignedClientCert(t *testing.T) (certPath, keyPath string, certDER []byte) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test-client"},
		NotBefore:    time.Now().UTC().Add(-time.Minute),
		NotAfter:     time.Now().UTC().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	der, err := x509.CreateCertificate(crand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)

	tmpDir := t.TempDir()

	// Write cert PEM.
	certFile := filepath.Join(tmpDir, "client.crt")
	certPEM := pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: der})
	require.NoError(t, os.WriteFile(certFile, certPEM, cryptoutilSharedMagic.CacheFilePermissions))

	// Write key PEM.
	keyFile := filepath.Join(tmpDir, "client.key")
	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeECPrivateKey, Bytes: keyDER})
	require.NoError(t, os.WriteFile(keyFile, keyPEM, cryptoutilSharedMagic.CacheFilePermissions))

	return certFile, keyFile, der
}

func TestLoadClientCert_ValidPaths(t *testing.T) {
	t.Parallel()

	certPath, keyPath, _ := writeSelfSignedClientCert(t)

	cert, err := cryptoutilAppsFrameworkCli.LoadClientCert(certPath, keyPath)
	require.NoError(t, err)
	require.NotNil(t, cert, "expected non-nil cert for valid cert+key files")
}

func TestHTTPGet_WithMTLSClientCert(t *testing.T) {
	t.Parallel()

	clientCertPath, clientKeyPath, clientCertDER := writeSelfSignedClientCert(t)

	// Build client CA pool that trusts the self-signed client cert.
	clientCert, err := x509.ParseCertificate(clientCertDER)
	require.NoError(t, err)

	clientCAPool := x509.NewCertPool()
	clientCAPool.AddCert(clientCert)

	// Start mTLS server that requires client cert.
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	}))
	srv.TLS = &tls.Config{
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  clientCAPool,
		MinVersion: tls.VersionTLS13,
	}
	srv.StartTLS()
	t.Cleanup(srv.Close)

	// Write server CA cert for the client to trust.
	serverCACertPath := writeCACertFromTLSServer(t, srv)

	statusCode, body, err := cryptoutilAppsFrameworkCli.HTTPGet(
		srv.URL+"/test", serverCACertPath, clientCertPath, clientKeyPath,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.Contains(t, body, cryptoutilSharedMagic.StringStatus)
}

func TestHTTPPost_WithMTLSClientCert(t *testing.T) {
	t.Parallel()

	clientCertPath, clientKeyPath, clientCertDER := writeSelfSignedClientCert(t)

	clientCert, err := x509.ParseCertificate(clientCertDER)
	require.NoError(t, err)

	clientCAPool := x509.NewCertPool()
	clientCAPool.AddCert(clientCert)

	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"result":"done"}`)
	}))
	srv.TLS = &tls.Config{
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  clientCAPool,
		MinVersion: tls.VersionTLS13,
	}
	srv.StartTLS()
	t.Cleanup(srv.Close)

	serverCACertPath := writeCACertFromTLSServer(t, srv)

	statusCode, body, err := cryptoutilAppsFrameworkCli.HTTPPost(
		srv.URL+cryptoutilSharedMagic.PrivateAdminShutdownRequestPath, serverCACertPath, clientCertPath, clientKeyPath,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.Contains(t, body, "done")
}
