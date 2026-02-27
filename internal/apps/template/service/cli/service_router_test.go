// Copyright (c) 2025 Justin Cranford
//
// SPDX-License-Identifier: MIT

package cli_test

import (
	"bytes"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateCli "cryptoutil/internal/apps/template/service/cli"
)

const testServiceID = "test-svc"
const testProductNameService = "testproduct"
const testServiceNameConst = "svc"

var testDefaultPort = uint16(8042) //nolint:gochecknoglobals // test fixture

var testServiceCfg = cryptoutilAppsTemplateCli.ServiceConfig{ //nolint:gochecknoglobals // test fixture
	ServiceID:         testServiceID,
	ProductName:       testProductNameService,
	ServiceName:       testServiceNameConst,
	DefaultPublicPort: testDefaultPort,
	UsageMain:         "Usage: test-svc <subcommand>",
	UsageServer:       "Usage: test-svc server",
	UsageClient:       "Usage: test-svc client",
	UsageInit:         "Usage: test-svc init",
	UsageHealth:       "Usage: test-svc health",
	UsageLivez:        "Usage: test-svc livez",
	UsageReadyz:       "Usage: test-svc readyz",
	UsageShutdown:     "Usage: test-svc shutdown",
}

func noopSubcmd(_ []string, _, _ io.Writer) int { return 0 }

func TestRouteService_EmptyArgsDefaultsToServer(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	called := false
	serverFn := func(_ []string, _, _ io.Writer) int {
		called = true

		return 0
	}

	exitCode := cryptoutilAppsTemplateCli.RouteService(testServiceCfg, nil, &stdout, &stderr, serverFn, noopSubcmd, noopSubcmd)
	require.Equal(t, 0, exitCode)
	require.True(t, called, "expected server fn to be called as default")
}

func TestRouteService_HelpFlag(t *testing.T) {
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

			exitCode := cryptoutilAppsTemplateCli.RouteService(testServiceCfg, []string{tt.arg}, &stdout, &stderr, noopSubcmd, noopSubcmd, noopSubcmd)
			require.Equal(t, 0, exitCode)
			require.Contains(t, stdout.String(), "Usage: test-svc")
		})
	}
}

func TestRouteService_VersionSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsTemplateCli.RouteService(testServiceCfg, []string{"version"}, &stdout, &stderr, noopSubcmd, noopSubcmd, noopSubcmd)
	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), testServiceID)
	require.Contains(t, stdout.String(), testProductNameService)
}

func TestRouteService_ServerSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	serverCalled := false
	serverFn := func(_ []string, _, _ io.Writer) int {
		serverCalled = true

		return cryptoutilSharedMagic.GitRecentActivityDays
	}

	exitCode := cryptoutilAppsTemplateCli.RouteService(testServiceCfg, []string{"server"}, &stdout, &stderr, serverFn, noopSubcmd, noopSubcmd)
	require.Equal(t, cryptoutilSharedMagic.GitRecentActivityDays, exitCode)
	require.True(t, serverCalled)
}

func TestRouteService_ClientSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	clientCalled := false
	clientFn := func(_ []string, _, _ io.Writer) int {
		clientCalled = true

		return cryptoutilSharedMagic.IMMinPasswordLength
	}

	exitCode := cryptoutilAppsTemplateCli.RouteService(testServiceCfg, []string{"client"}, &stdout, &stderr, noopSubcmd, clientFn, noopSubcmd)
	require.Equal(t, cryptoutilSharedMagic.IMMinPasswordLength, exitCode)
	require.True(t, clientCalled)
}

func TestRouteService_InitSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	initCalled := false
	initFn := func(_ []string, _, _ io.Writer) int {
		initCalled = true

		return 9
	}

	exitCode := cryptoutilAppsTemplateCli.RouteService(testServiceCfg, []string{"init"}, &stdout, &stderr, noopSubcmd, noopSubcmd, initFn)
	require.Equal(t, 9, exitCode)
	require.True(t, initCalled)
}

func TestRouteService_UnknownSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsTemplateCli.RouteService(testServiceCfg, []string{"banana"}, &stdout, &stderr, noopSubcmd, noopSubcmd, noopSubcmd)
	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "Unknown subcommand: banana")
	require.Contains(t, stdout.String(), "Usage: test-svc")
}
