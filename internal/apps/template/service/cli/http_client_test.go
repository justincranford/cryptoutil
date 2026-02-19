// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package cli_test

import (
	"bytes"
	"encoding/pem"
	"fmt"
	http "net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateCli "cryptoutil/internal/apps/template/service/cli"
)

func TestLoadCACertPool_EmptyPath(t *testing.T) {
	t.Parallel()

	pool, err := cryptoutilAppsTemplateCli.LoadCACertPool("")
	require.NoError(t, err)
	require.Nil(t, pool, "expected nil pool for empty path")
}

func TestLoadCACertPool_NonExistentPath(t *testing.T) {
	t.Parallel()

	pool, err := cryptoutilAppsTemplateCli.LoadCACertPool("/nonexistent/path/to/ca.pem")
	require.Error(t, err)
	require.Nil(t, pool)
}

func TestLoadCACertPool_InvalidPEM(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "invalid.pem")

	err := os.WriteFile(certPath, []byte("not a valid PEM certificate"), 0o600)
	require.NoError(t, err)

	pool, err := cryptoutilAppsTemplateCli.LoadCACertPool(certPath)
	require.Error(t, err)
	require.Nil(t, pool)
}

func TestHTTPGet_Success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	}))
	t.Cleanup(srv.Close)

	statusCode, body, err := cryptoutilAppsTemplateCli.HTTPGet(srv.URL+"/test", "")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.Contains(t, body, "status")
}

func TestHTTPGet_ServiceUnavailable(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprint(w, "Service Unavailable")
	}))
	t.Cleanup(srv.Close)

	statusCode, body, err := cryptoutilAppsTemplateCli.HTTPGet(srv.URL+"/test", "")
	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, statusCode)
	require.Contains(t, body, "Unavailable")
}

func TestHTTPGet_ConnectionRefused(t *testing.T) {
	t.Parallel()

	// Port 1 is always refused
	_, _, err := cryptoutilAppsTemplateCli.HTTPGet("https://127.0.0.1:1/test", "")
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

	statusCode, body, err := cryptoutilAppsTemplateCli.HTTPPost(srv.URL+"/shutdown", "")
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

	statusCode, body, err := cryptoutilAppsTemplateCli.HTTPPost(srv.URL+"/shutdown", "")
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, statusCode)
	require.Contains(t, body, "accepted")
}

func TestHTTPPost_ConnectionRefused(t *testing.T) {
	t.Parallel()

	_, _, err := cryptoutilAppsTemplateCli.HTTPPost("https://127.0.0.1:1/shutdown", "")
	require.Error(t, err)
}

func writeCACertFromTLSServer(t *testing.T, srv *httptest.Server) string {
	t.Helper()

	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "ca.pem")

	// Export the httptest server's certificate as PEM
	certDER := srv.TLS.Certificates[0].Certificate[0]

	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})
	require.NotNil(t, pemData, "failed to encode certificate to PEM")

	err := os.WriteFile(certPath, pemData, 0o600)
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

	pool, err := cryptoutilAppsTemplateCli.LoadCACertPool(certPath)
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

	statusCode, body, err := cryptoutilAppsTemplateCli.HTTPGet(srv.URL+"/test", certPath)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, statusCode)
	require.Contains(t, body, "status")
}

func TestHTTPPost_WithCACert(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"result":"done"}`)
	}))
	t.Cleanup(srv.Close)

	certPath := writeCACertFromTLSServer(t, srv)

	statusCode, body, err := cryptoutilAppsTemplateCli.HTTPPost(srv.URL+"/shutdown", certPath)
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

	exitCode := cryptoutilAppsTemplateCli.HealthCommand(
		[]string{"--url", srv.URL, "--cacert", certPath},
		&stdout, &stderr,
		"Usage: health",
		8800,
	)
	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "\u2705")
}

func TestLivezCommand_WithCACert(t *testing.T) {
	t.Parallel()

	srv := newHealthMockServer(t, "/admin/api/v1/livez", http.StatusOK, `{"status":"alive"}`)

	certPath := writeCACertFromTLSServer(t, srv)

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsTemplateCli.LivezCommand(
		[]string{"--url", srv.URL, "--cacert", certPath},
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

	exitCode := cryptoutilAppsTemplateCli.ReadyzCommand(
		[]string{"--url", srv.URL, "--cacert", certPath},
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

	exitCode := cryptoutilAppsTemplateCli.ShutdownCommand(
		[]string{"--url", srv.URL, "--cacert", certPath},
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
	exitCode := cryptoutilAppsTemplateCli.HealthCommand(
		[]string{"--url", srv.URL + "/health"},
		&stdout, &stderr,
		"Usage: health",
		8800,
	)
	// Connection succeeds but response shape may differ
	// The key check is that it doesn't double-append /health
	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "\u2705")
}

func TestHTTPGet_InvalidCACert(t *testing.T) {
	t.Parallel()

	// An invalid cacert path should cause LoadCACertPool to fail
	_, _, err := cryptoutilAppsTemplateCli.HTTPGet("https://127.0.0.1:1/test", "/nonexistent/ca.pem")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to load CA certificate")
}

func TestHTTPPost_InvalidCACert(t *testing.T) {
	t.Parallel()

	// An invalid cacert path should cause LoadCACertPool to fail
	_, _, err := cryptoutilAppsTemplateCli.HTTPPost("https://127.0.0.1:1/shutdown", "/nonexistent/ca.pem")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to load CA certificate")
}

func TestHTTPGet_InvalidURL(t *testing.T) {
	t.Parallel()

	// A completely invalid URL scheme causes NewRequestWithContext to fail
	_, _, err := cryptoutilAppsTemplateCli.HTTPGet("://invalid-url", "")
	require.Error(t, err)
}

func TestHTTPPost_InvalidURL(t *testing.T) {
	t.Parallel()

	// A completely invalid URL scheme causes NewRequestWithContext to fail
	_, _, err := cryptoutilAppsTemplateCli.HTTPPost("://invalid-url", "")
	require.Error(t, err)
}
