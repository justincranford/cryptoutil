// Copyright (c) 2025-2026 Justin Cranford.
//

package cli_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkCli "cryptoutil/internal/apps-framework/service/cli"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var testIdentity = cryptoutilAppsFrameworkCli.ServiceIdentity{ //nolint:gochecknoglobals // test fixture
	ServiceID:   testServiceID,
	ProductName: testProductNameService,
	ServiceName: testServiceNameConst,
	DisplayName: "Test Service",
	ServicePort: testDefaultPort,
}

func TestBuildServerUsage(t *testing.T) {
	t.Parallel()

	got := cryptoutilAppsFrameworkCli.BuildServerUsage(testIdentity)
	require.Contains(t, got, testProductNameService)
	require.Contains(t, got, testServiceNameConst)
	require.Contains(t, got, "Test Service")
	require.Contains(t, got, "configs/test-svc/test-svc-framework.yml")
}

func TestRouteServiceFromIdentity_HelpFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  string
	}{
		{name: "help_word", arg: cryptoutilSharedMagic.CLIHelpCommand},
		{name: "help_long", arg: cryptoutilSharedMagic.CLIHelpFlag},
		{name: "help_short", arg: "-h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			exitCode := cryptoutilAppsFrameworkCli.RouteServiceFromIdentity(
				testIdentity,
				[]string{tt.arg},
				&stdout, &stderr,
				func(_ []string, _, _ io.Writer) int { return 0 },
			)
			require.Equal(t, 0, exitCode)
			require.Contains(t, stdout.String(), testServiceNameConst)
		})
	}
}

func TestRouteServiceFromIdentity_VersionSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsFrameworkCli.RouteServiceFromIdentity(
		testIdentity,
		[]string{cryptoutilSharedMagic.CLIVersionCommand},
		&stdout, &stderr,
		func(_ []string, _, _ io.Writer) int { return 0 },
	)
	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout.String(), testServiceID)
}

func TestRouteServiceFromIdentity_ServerSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	serverCalled := false
	serverFn := func(_ []string, _, _ io.Writer) int {
		serverCalled = true

		return 0
	}

	exitCode := cryptoutilAppsFrameworkCli.RouteServiceFromIdentity(
		testIdentity,
		[]string{"server"},
		&stdout, &stderr,
		serverFn,
	)
	require.Equal(t, 0, exitCode)
	require.True(t, serverCalled)
}

func TestRouteServiceFromIdentity_ClientHelpSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsFrameworkCli.RouteServiceFromIdentity(
		testIdentity,
		[]string{"client", cryptoutilSharedMagic.CLIHelpFlag},
		&stdout, &stderr,
		func(_ []string, _, _ io.Writer) int { return 0 },
	)
	require.Equal(t, 0, exitCode)
}

func TestRouteServiceFromIdentity_ClientUnknownArg(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsFrameworkCli.RouteServiceFromIdentity(
		testIdentity,
		[]string{"client", "some-arg"},
		&stdout, &stderr,
		func(_ []string, _, _ io.Writer) int { return 0 },
	)
	require.Equal(t, 1, exitCode)
}

func TestRouteServiceFromIdentity_InitHelpSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsFrameworkCli.RouteServiceFromIdentity(
		testIdentity,
		[]string{"init", cryptoutilSharedMagic.CLIHelpFlag},
		&stdout, &stderr,
		func(_ []string, _, _ io.Writer) int { return 0 },
	)
	require.Equal(t, 0, exitCode)
}

func TestRouteServiceFromIdentity_UnknownSubcommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilAppsFrameworkCli.RouteServiceFromIdentity(
		testIdentity,
		[]string{"unknown-subcommand"},
		&stdout, &stderr,
		func(_ []string, _, _ io.Writer) int { return 0 },
	)
	require.Equal(t, 1, exitCode)
	require.Contains(t, stderr.String(), "Unknown subcommand")
}

func TestRouteServiceFromIdentity_EmptyArgsDefaultsToServer(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	serverCalled := false
	serverFn := func(_ []string, _, _ io.Writer) int {
		serverCalled = true

		return 0
	}

	exitCode := cryptoutilAppsFrameworkCli.RouteServiceFromIdentity(
		testIdentity,
		nil,
		&stdout, &stderr,
		serverFn,
	)
	require.Equal(t, 0, exitCode)
	require.True(t, serverCalled)
}

func TestBuildServerUsage_ContainsConfigPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		serviceID string
		wantInfix string
	}{
		{name: "svc_sm_kms", serviceID: cryptoutilSharedMagic.OTLPServiceSMKMS, wantInfix: "configs/sm-kms/sm-kms-framework.yml"},
		{name: "svc_jose_ja", serviceID: cryptoutilSharedMagic.OTLPServiceJoseJA, wantInfix: "configs/jose-ja/jose-ja-framework.yml"},
		{name: "svc_pki_ca", serviceID: cryptoutilSharedMagic.OTLPServicePKICA, wantInfix: "configs/pki-ca/pki-ca-framework.yml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			id := cryptoutilAppsFrameworkCli.ServiceIdentity{
				ServiceID:   tt.serviceID,
				ProductName: "p",
				ServiceName: "s",
				DisplayName: "D",
				ServicePort: cryptoutilSharedMagic.KMSServicePort,
			}
			got := cryptoutilAppsFrameworkCli.BuildServerUsage(id)
			require.True(t, strings.Contains(got, tt.wantInfix),
				"expected usage to contain %q, got: %q", tt.wantInfix, got)
		})
	}
}
