// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package cli_test

import (
	"bytes"
	"fmt"
	http "net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateCli "cryptoutil/internal/apps/template/service/cli"
)

// newHealthMockServer creates a TLS httptest server that handles health check endpoints.
// It responds 200 OK for registered paths and 404 for unknown paths.
func newHealthMockServer(t *testing.T, path string, statusCode int, body string) *httptest.Server {
	t.Helper()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == path {
			w.WriteHeader(statusCode)
			_, _ = fmt.Fprint(w, body)

			return
		}

		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, "not found")
	})

	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)

	return srv
}

func TestHealthCommand_HelpFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  string
	}{
		{name: "help_word", arg: "help"},
		{name: "help_long", arg: "--help"},
		{name: "help_short", arg: "-h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := cryptoutilAppsTemplateCli.HealthCommand([]string{tt.arg}, &stdout, &stderr, "Usage: health", 8800)
			require.Equal(t, 0, exitCode)
			require.Contains(t, stderr.String(), "Usage: health")
		})
	}
}

func TestHealthCommand_Success(t *testing.T) {
	t.Parallel()

	srv := newHealthMockServer(t, "/health", http.StatusOK, `{"status":"ok"}`)

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsTemplateCli.HealthCommand(
		[]string{"--url", srv.URL},
		&stdout, &stderr,
		"Usage: health",
		8800,
	)
	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "✅")
}

func TestHealthCommand_ServiceUnavailable(t *testing.T) {
	t.Parallel()

	srv := newHealthMockServer(t, "/health", http.StatusServiceUnavailable, "Service Unavailable")

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsTemplateCli.HealthCommand(
		[]string{"--url", srv.URL},
		&stdout, &stderr,
		"Usage: health",
		8800,
	)
	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "❌")
}

func TestHealthCommand_ConnectionRefused(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsTemplateCli.HealthCommand(
		[]string{"--url", "https://127.0.0.1:1"},
		&stdout, &stderr,
		"Usage: health",
		8800,
	)
	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "❌")
}

func TestLivezCommand_HelpFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  string
	}{
		{name: "help_word", arg: "help"},
		{name: "help_long", arg: "--help"},
		{name: "help_short", arg: "-h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := cryptoutilAppsTemplateCli.LivezCommand([]string{tt.arg}, &stdout, &stderr, "Usage: livez")
			require.Equal(t, 0, exitCode)
			require.Contains(t, stderr.String(), "Usage: livez")
		})
	}
}

func TestLivezCommand_Success(t *testing.T) {
	t.Parallel()

	srv := newHealthMockServer(t, "/admin/api/v1/livez", http.StatusOK, `{"status":"alive"}`)

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsTemplateCli.LivezCommand(
		[]string{"--url", srv.URL},
		&stdout, &stderr,
		"Usage: livez",
	)
	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "✅")
}

func TestLivezCommand_Failure(t *testing.T) {
	t.Parallel()

	srv := newHealthMockServer(t, "/admin/api/v1/livez", http.StatusServiceUnavailable, "down")

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsTemplateCli.LivezCommand(
		[]string{"--url", srv.URL},
		&stdout, &stderr,
		"Usage: livez",
	)
	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "❌")
}

func TestLivezCommand_ConnectionRefused(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsTemplateCli.LivezCommand(
		[]string{"--url", "https://127.0.0.1:1"},
		&stdout, &stderr,
		"Usage: livez",
	)
	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "❌")
}

func TestReadyzCommand_HelpFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  string
	}{
		{name: "help_word", arg: "help"},
		{name: "help_long", arg: "--help"},
		{name: "help_short", arg: "-h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := cryptoutilAppsTemplateCli.ReadyzCommand([]string{tt.arg}, &stdout, &stderr, "Usage: readyz")
			require.Equal(t, 0, exitCode)
			require.Contains(t, stderr.String(), "Usage: readyz")
		})
	}
}

func TestReadyzCommand_Success(t *testing.T) {
	t.Parallel()

	srv := newHealthMockServer(t, "/admin/api/v1/readyz", http.StatusOK, `{"ready":true}`)

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsTemplateCli.ReadyzCommand(
		[]string{"--url", srv.URL},
		&stdout, &stderr,
		"Usage: readyz",
	)
	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "✅")
}

func TestReadyzCommand_Failure(t *testing.T) {
	t.Parallel()

	srv := newHealthMockServer(t, "/admin/api/v1/readyz", http.StatusServiceUnavailable, "not ready")

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsTemplateCli.ReadyzCommand(
		[]string{"--url", srv.URL},
		&stdout, &stderr,
		"Usage: readyz",
	)
	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "❌")
}

func TestReadyzCommand_ConnectionRefused(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsTemplateCli.ReadyzCommand(
		[]string{"--url", "https://127.0.0.1:1"},
		&stdout, &stderr,
		"Usage: readyz",
	)
	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "❌")
}

func TestShutdownCommand_HelpFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  string
	}{
		{name: "help_word", arg: "help"},
		{name: "help_long", arg: "--help"},
		{name: "help_short", arg: "-h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := cryptoutilAppsTemplateCli.ShutdownCommand([]string{tt.arg}, &stdout, &stderr, "Usage: shutdown")
			require.Equal(t, 0, exitCode)
			require.Contains(t, stderr.String(), "Usage: shutdown")
		})
	}
}

func TestShutdownCommand_Success(t *testing.T) {
	t.Parallel()

	srv := newHealthMockServer(t, "/admin/api/v1/shutdown", http.StatusOK, `{"shutdown":"initiated"}`)

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsTemplateCli.ShutdownCommand(
		[]string{"--url", srv.URL},
		&stdout, &stderr,
		"Usage: shutdown",
	)
	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "✅")
}

func TestShutdownCommand_SuccessWithAccepted(t *testing.T) {
	t.Parallel()

	srv := newHealthMockServer(t, "/admin/api/v1/shutdown", http.StatusAccepted, `{"shutdown":"accepted"}`)

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsTemplateCli.ShutdownCommand(
		[]string{"--url", srv.URL},
		&stdout, &stderr,
		"Usage: shutdown",
	)
	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "✅")
}

func TestShutdownCommand_Failure(t *testing.T) {
	t.Parallel()

	srv := newHealthMockServer(t, "/admin/api/v1/shutdown", http.StatusInternalServerError, "error")

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsTemplateCli.ShutdownCommand(
		[]string{"--url", srv.URL},
		&stdout, &stderr,
		"Usage: shutdown",
	)
	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "❌")
}

func TestShutdownCommand_ConnectionRefused(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsTemplateCli.ShutdownCommand(
		[]string{"--url", "https://127.0.0.1:1"},
		&stdout, &stderr,
		"Usage: shutdown",
	)
	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "❌")
}

func TestRouteService_HealthSubcommand_Success(t *testing.T) {
	t.Parallel()

	srv := newHealthMockServer(t, "/health", http.StatusOK, `{"status":"ok"}`)

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsTemplateCli.RouteService(testServiceCfg, []string{"health", "--url", srv.URL}, &stdout, &stderr, noopSubcmd, noopSubcmd, noopSubcmd)
	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "✅")
}

func TestRouteService_LivezSubcommand_Success(t *testing.T) {
	t.Parallel()

	srv := newHealthMockServer(t, "/admin/api/v1/livez", http.StatusOK, `{"status":"alive"}`)

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsTemplateCli.RouteService(testServiceCfg, []string{"livez", "--url", srv.URL}, &stdout, &stderr, noopSubcmd, noopSubcmd, noopSubcmd)
	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "✅")
}

func TestRouteService_ReadyzSubcommand_Success(t *testing.T) {
	t.Parallel()

	srv := newHealthMockServer(t, "/admin/api/v1/readyz", http.StatusOK, `{"ready":true}`)

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsTemplateCli.RouteService(testServiceCfg, []string{"readyz", "--url", srv.URL}, &stdout, &stderr, noopSubcmd, noopSubcmd, noopSubcmd)
	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "✅")
}

func TestRouteService_ShutdownSubcommand_Success(t *testing.T) {
	t.Parallel()

	srv := newHealthMockServer(t, "/admin/api/v1/shutdown", http.StatusOK, `{"shutdown":"initiated"}`)

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsTemplateCli.RouteService(testServiceCfg, []string{"shutdown", "--url", srv.URL}, &stdout, &stderr, noopSubcmd, noopSubcmd, noopSubcmd)
	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), "✅")
}
